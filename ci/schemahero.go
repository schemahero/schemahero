package main

import (
	"context"
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"sort"

	"dagger.io/dagger"
)

func buildSchemahero(ctx context.Context) (string, string, error) {
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout), dagger.WithWorkdir("."))
	if err != nil {
		panic(err)
	}
	defer client.Close()

	contextDir := client.Host().Directory(".")

	// this is way overengineered but performance of multiple tests runs make
	// a developer experience better. we need to focus on building a test
	// execution that's quick or people lose context and don't contribute as much
	tagSuffix, err := schemaheroSourceChecksum()
	if err != nil {
		panic(err)
	}

	buildManagerOpts := dagger.DirectoryDockerBuildOpts{
		Dockerfile: filepath.Join(".", "deploy", "Dockerfile.multiarch"),
		Target:     "manager",
		Platform:   "linux/amd64",
	}
	managerImage := fmt.Sprintf("ttl.sh/schemahero-manager-%s", tagSuffix)
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
	schemaheroImage := fmt.Sprintf("ttl.sh/schemahero-schemahero-%s", tagSuffix)
	schemaheroBuilt := contextDir.DockerBuild(buildSchemaheroOpts)
	schemaheroRef, err := schemaheroBuilt.Publish(ctx, schemaheroImage)
	if err != nil {
		panic(err)
	}

	return managerRef, schemaheroRef, nil
}

func schemaheroSourceChecksum() (string, error) {
	// this is deterministically ordered by the order i determined it to be when writing this
	// it's not important that it's ordered, just that it's consistent

	// the checksum of the source code
	shasums := []string{}

	dirs := []string{
		"pkg",
		"cmd",
		"go.mod",
		"go.sum",
		"Makefile",
		"deploy",
	}

	for _, dir := range dirs {
		shasum, err := md5dir(dir)
		if err != nil {
			return "", err
		}
		shasums = append(shasums, shasum)
	}

	return fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%v", shasums)))), nil
}

func md5dir(root string) (string, error) {
	stat, err := os.Stat(root)
	if err != nil {
		return "", err
	}
	if !stat.IsDir() {
		// calculate the shasum of the file
		// where root is the filename
		data, err := ioutil.ReadFile(root)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("%x", md5.Sum(data)), nil
	}

	m := make(map[string][md5.Size]byte)
	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		m[path] = md5.Sum(data)
		return nil
	})
	if err != nil {
		return "", err
	}

	// sort the map by key
	// this is important because we want to ensure that the checksum is the same
	// regardless of the order that the files are walked
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// create a new map with the sorted keys
	sorted := make(map[string][md5.Size]byte)
	for _, k := range keys {
		sorted[k] = m[k]
	}

	// put all of the checksums together
	var joined []byte
	for _, v := range sorted {
		joined = append(joined, v[:]...)
	}

	// calculate a shasum of the joined checksums
	return fmt.Sprintf("%x", md5.Sum(joined)), nil
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
