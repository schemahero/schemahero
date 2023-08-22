package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"

	"dagger.io/dagger"
)

func cli(client *dagger.Client) *dagger.Container {
	cli := client.Container().From(Image).
		WithSecretVariable("REPLICATED_API_TOKEN", client.SetSecret("replicatedToken", os.Getenv("REPLICATED_API_TOKEN")))

	if os.Getenv("REPLICATED_API_ORIGIN") != "" {
		cli = cli.WithEnvVariable("REPLICATED_API_ORIGIN", os.Getenv("REPLICATED_API_ORIGIN"))
	}
	if os.Getenv("REPLICATED_ID_ORIGIN") != "" {
		cli = cli.WithEnvVariable("REPLICATED_ID_ORIGIN", os.Getenv("REPLICATED_ID_ORIGIN"))
	}

	return cli
}

func createCluster(ctx context.Context, client *dagger.Client, distribution string, version string) (string, string, error) {
	// generate a name for this cluster
	clusterName := fmt.Sprintf("schemahero-%s-%s-%.0f", distribution, version, math.Floor(rand.Float64()*10000000))

	// create a cluster cluster
	createCluster := cli(client).WithExec([]string{
		"cluster",
		"create",
		"--output", "json",
		"--distribution", distribution,
		"--name", clusterName,
		"--version", version,
		"--ttl", "2h",
		"--wait", "5m",
	})
	output, err := createCluster.Stdout(ctx)
	if err != nil {
		return "", "", err
	}
	cluster := ClusterOutput{}
	if err := json.Unmarshal([]byte(output), &cluster); err != nil {
		return "", "", err
	}

	kubeconfig := cli(client).WithExec([]string{
		"cluster",
		"kubeconfig",
		cluster.ID,
		"--stdout",
	})

	kubeconfigOut, err := kubeconfig.Stdout(ctx)
	if err != nil {
		return "", "", err
	}

	return cluster.ID, kubeconfigOut, nil
}

// deleteCluster will terminate before the ttl expires
// we call this when tests were successful and we don't need to
// keep the cluster around
func deleteCluster(ctx context.Context, client *dagger.Client, id string) error {
	deleteCluster := cli(client).WithExec([]string{
		"cluster",
		"rm",
		id,
	})
	if _, err := deleteCluster.Stdout(ctx); err != nil {
		return err
	}

	return nil
}
