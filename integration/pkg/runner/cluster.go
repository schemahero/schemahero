package runner

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"sigs.k8s.io/kind/pkg/cluster"
	"sigs.k8s.io/kind/pkg/cluster/config/encoding"
	"sigs.k8s.io/kind/pkg/cluster/create"
)

type Cluster struct {
	Name                     string
	KubeConfigPath           string
	KubeConfigFromDockerPath string
}

func createCluster(name string) (*Cluster, error) {
	cfg, err := encoding.Load("")
	if err != nil {
		return nil, err
	}

	cfg.Networking.APIServerAddress = "0.0.0.0"

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	ctx := cluster.NewContext(name)
	err = ctx.Create(cfg,
		create.Retain(true),
		create.WaitForReady(time.Second*90),
		create.SetupKubernetes(true))
	if err != nil {
		return nil, err
	}

	kubeConfig, err := ioutil.ReadFile(ctx.KubeConfigPath())
	if err != nil {
		return nil, err
	}

	rewrittenKubeConfig := strings.Replace(string(kubeConfig), "localhost", "kubernetes", -1)
	tmpFile, err := ioutil.TempFile(os.TempDir(), "kubeconfig-")
	if err != nil {
		return nil, err
	}
	if _, err = tmpFile.Write([]byte(rewrittenKubeConfig)); err != nil {
		return nil, err
	}

	if err = os.Chmod(tmpFile.Name(), 0644); err != nil {
		return nil, err
	}

	cluster := Cluster{
		Name:                     name,
		KubeConfigPath:           ctx.KubeConfigPath(),
		KubeConfigFromDockerPath: tmpFile.Name(),
	}

	return &cluster, nil
}

func (c Cluster) delete() error {
	os.Remove(c.KubeConfigFromDockerPath)
	ctx := cluster.NewContext(c.Name)
	return ctx.Delete()
}

func (c Cluster) kubectl(ctx context.Context, cmd []string) (*container.Config, *container.HostConfig, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, nil, err
	}

	pullReader, err := cli.ImagePull(ctx, "docker.io/bitnami/kubectl:1.14", types.ImagePullOptions{})
	if err != nil {
		return nil, nil, err
	}
	io.Copy(ioutil.Discard, pullReader)

	containerConfig := &container.Config{
		Image: "bitnami/kubectl:1.14",
		Env: []string{
			"KUBECONFIG=/kubeconfig",
		},
		Cmd: cmd,
	}
	hostConfig := &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   "bind",
				Source: c.KubeConfigFromDockerPath,
				Target: "/kubeconfig",
			},
		},
		ExtraHosts: []string{
			"kubernetes:172.17.0.1",
		},
	}

	return containerConfig, hostConfig, nil
}

func (c Cluster) apply(manifests []byte, showStdOut bool) error {
	ctx := context.Background()

	tmpFile, err := ioutil.TempFile(os.TempDir(), "manifests-")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())
	if _, err = tmpFile.Write(manifests); err != nil {
		return err
	}
	if err = os.Chmod(tmpFile.Name(), 0644); err != nil {
		return err
	}

	cmd := []string{
		"apply",
		"-f",
		"/manifests.yaml",
	}
	containerConfig, hostConfig, err := c.kubectl(ctx, cmd)
	if err != nil {
		return err
	}

	hostConfig.Mounts = append(hostConfig.Mounts,
		mount.Mount{
			Type:   "bind",
			Source: tmpFile.Name(),
			Target: "/manifests.yaml",
		})

	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	resp, err := cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, "")
	if err != nil {
		return err
	}
	defer cli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{})

	startOptions := types.ContainerStartOptions{}
	err = cli.ContainerStart(ctx, resp.ID, startOptions)
	if err != nil {
		return err
	}

	exitCode, err := cli.ContainerWait(ctx, resp.ID)
	if err != nil {
		return err
	}

	data, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		return err
	}

	stdOut := new(bytes.Buffer)
	stdErr := new(bytes.Buffer)

	stdcopy.StdCopy(stdOut, stdErr, data)

	if exitCode != 0 {
		return fmt.Errorf("unexpected exit code running kubectl: %d\nstderr:%s\b\bstdout:%s", exitCode, stdErr, stdOut)
	}

	if showStdOut {
		fmt.Printf("%s\n", stdOut)
	}

	return nil
}

func (c Cluster) exec(podName string, command string, args []string) (int64, []byte, []byte, error) {
	ctx := context.Background()

	cmd := []string{
		"exec",
		podName,
		command,
		"--",
	}
	cmd = append(cmd, args...)

	containerConfig, hostConfig, err := c.kubectl(ctx, cmd)
	if err != nil {
		return -1, nil, nil, err
	}

	cli, err := client.NewEnvClient()
	if err != nil {
		return -1, nil, nil, err
	}

	resp, err := cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, "")
	if err != nil {
		return -1, nil, nil, err
	}
	defer cli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{})

	startOptions := types.ContainerStartOptions{}
	err = cli.ContainerStart(ctx, resp.ID, startOptions)
	if err != nil {
		return -1, nil, nil, err
	}

	exitCode, err := cli.ContainerWait(ctx, resp.ID)
	if err != nil {
		return -1, nil, nil, err
	}

	data, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		return -1, nil, nil, err
	}

	stdOut := new(bytes.Buffer)
	stdErr := new(bytes.Buffer)

	stdcopy.StdCopy(stdOut, stdErr, data)

	return exitCode, stdOut.Bytes(), stdErr.Bytes(), nil
}
