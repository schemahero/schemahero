package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"path/filepath"

	"dagger.io/dagger"
)

func buildSchemahero(ctx context.Context) (string, string, error) {
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout), dagger.WithWorkdir("."))
	if err != nil {
		panic(err)
	}
	defer client.Close()

	contextDir := client.Host().Directory(".")

	buildManagerOpts := dagger.DirectoryDockerBuildOpts{
		Dockerfile: filepath.Join(".", "deploy", "Dockerfile.multiarch"),
		Target:     "manager",
		Platform:   "linux/amd64",
	}
	managerImage := fmt.Sprintf("ttl.sh/schemahero-manager-%.0f", math.Floor(rand.Float64()*10000000))
	managerBuilt := contextDir.DockerBuild(buildManagerOpts)
	managerRef, err := managerBuilt.Publish(ctx, managerImage)
	if err != nil {
		panic(err)
	}

	buildSchemaheroOpts := dagger.DirectoryDockerBuildOpts{
		Dockerfile: filepath.Join(".", "deploy", "Dockerfile.multiarch"),
		Target:     "schemahero",
		Platform:   "linux/amd64",
	}
	schemaheroImage := fmt.Sprintf("ttl.sh/schemahero-schemahero-%.0f", math.Floor(rand.Float64()*10000000))
	schemaheroBuilt := contextDir.DockerBuild(buildSchemaheroOpts)
	schemaheroRef, err := schemaheroBuilt.Publish(ctx, schemaheroImage)
	if err != nil {
		panic(err)
	}

	return managerRef, schemaheroRef, nil
}

func runTests(ctx context.Context, kubeconfig string, managerImage string, schemaheroImage string) error {
	// install the schemahero operator, using the manager image provided
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout), dagger.WithWorkdir("."))
	if err != nil {
		panic(err)
	}
	defer client.Close()

	// write the kubeconfig into a temp dir so that we can pass it into the container
	tmpDir, err := ioutil.TempDir("", "schemahero")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpDir)

	// we need a random filename, but it must be in the context that dagger is in
	kubeconfigDir := filepath.Join(".", "e2e", fmt.Sprintf("kubeconfig-%.0f", math.Floor(rand.Float64()*10000000)))
	if err := os.MkdirAll(kubeconfigDir, 0755); err != nil {
		panic(err)
	}
	defer os.RemoveAll(kubeconfigDir)

	kubeconfigFile := filepath.Join(kubeconfigDir, "kubeconfig")
	if err := ioutil.WriteFile(kubeconfigFile, []byte(kubeconfig), 0644); err != nil {
		panic(err)
	}

	// use the schemahero image to install
	schemahero := client.Container().From(schemaheroImage).
		WithDirectory("/kubeconfig", client.Host().Directory(kubeconfigDir)).
		WithExec([]string{
			"install",
			"--kubeconfig", "/kubeconfig/kubeconfig",
		})
	output, err := schemahero.Stdout(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Printf("schemahero install output: %s\n", output)

	return nil
}
