package runner

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
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

func (c Cluster) apply(manifests []byte) error {
	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}

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

	containerConfig := &container.Config{
		Image: "bitnami/kubectl:1.14",
		Env: []string{
			"KUBECONFIG=/kubeconfig",
		},
		Cmd: []string{
			"apply",
			"-f",
			"/manifests.yaml",
		},
	}
	hostConfig := &container.HostConfig{
		AutoRemove: true,
		Mounts: []mount.Mount{
			{
				Type:   "bind",
				Source: c.KubeConfigFromDockerPath,
				Target: "/kubeconfig",
			},
			{
				Type:   "bind",
				Source: tmpFile.Name(),
				Target: "/manifests.yaml",
			},
		},
		ExtraHosts: []string{
			"kubernetes:172.17.0.1",
		},
	}

	resp, err := cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, "")
	if err != nil {
		return err
	}

	startOptions := types.ContainerStartOptions{}
	err = cli.ContainerStart(ctx, resp.ID, startOptions)
	if err != nil {
		return err
	}

	exitCode, err := cli.ContainerWait(ctx, resp.ID)
	if err != nil {
		return err
	}

	if exitCode != 0 {
		return fmt.Errorf("unexpected exit code running kubectl: %d", exitCode)
	}

	return nil
}
