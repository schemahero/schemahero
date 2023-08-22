package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"dagger.io/dagger"
)

const (
	Image = "replicated/vendor-cli:0.48.0"
)

type ClusterOutput struct {
	ID                     string    `json:"id"`
	Name                   string    `json:"name"`
	KubernetesDistribution string    `json:"kubernetes_distribution"`
	KubernetesVersion      string    `json:"kubernetes_version"`
	NodeCount              int       `json:"node_count"`
	VCPUs                  int       `json:"vcpus"`
	MemoryGB               int       `json:"memory_gib"`
	Status                 string    `json:"status"`
	CreatedAt              time.Time `json:"created_at"`
	ExpiresAt              time.Time `json:"expires_at"`
}

func main() {
	ctx := context.Background()

	// initialize Dagger client
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		panic(err)
	}
	defer client.Close()

	managerImage, schemaheroImage, err := buildSchemahero(ctx)
	if err != nil {
		panic(err)
	}

	distributions := map[string][]string{
		"kind": []string{"v1.25.8", "v1.26.3"},
		"k3s":  []string{"v1.25", "v1.26"},
		"eks":  []string{"1.26", "1.27"},
	}

	// for each distribution/version, create a cluster, run tests, and delete the cluster
	// all in parallel
	var wg sync.WaitGroup
	for distribution, versions := range distributions {
		for _, version := range versions {
			wg.Add(1)
			go func(wg *sync.WaitGroup, d string, v string, managerImage string, schemaheroImage string) {
				defer wg.Done()
				if err := testDistributionVersion(ctx, client, d, v, managerImage, schemaheroImage); err != nil {
					panic(err)
				}
				fmt.Printf("Finished tests for distribution %s/%s\n", d, v)

			}(&wg, distribution, version, managerImage, schemaheroImage)
		}
	}

	fmt.Println("waiting for tests to finish")
	wg.Wait()
}

func testDistributionVersion(ctx context.Context, client *dagger.Client, distribution string, version string, managerImage string, schemaheroImage string) error {
	fmt.Printf("Creating cluster for distribution %s/%s\n", distribution, version)
	clusterID, kubeconfig, err := createCluster(ctx, client, distribution, version)
	if err != nil {
		return err
	}

	fmt.Printf("Created cluster %s for distribution %s/%s\n", clusterID, distribution, version)

	// install schemahero into the cluster
	if err := runTests(ctx, kubeconfig, managerImage, schemaheroImage); err != nil {
		return err
	}

	// delete the cluster
	fmt.Printf("Deleting cluster %s for distribution %s/%s\n", clusterID, distribution, version)
	if err := deleteCluster(ctx, client, clusterID); err != nil {
		return err
	}

	return nil
}
