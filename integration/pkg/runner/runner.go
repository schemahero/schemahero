package runner

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Runner struct {
	Viper *viper.Viper
}

func NewRunner() *Runner {
	return &Runner{
		Viper: viper.GetViper(),
	}
}

func (r *Runner) RunSync() error {
	fmt.Println("running integration tests")

	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}
	tests, err := ioutil.ReadDir(filepath.Join(currentDir, "tests"))
	if err != nil {
		return err
	}

	for _, testFile := range tests {
		if testFile.IsDir() {
			fmt.Printf("-----> Beginning test %q\n", testFile.Name())

			root := filepath.Join(currentDir, "tests", testFile.Name())

			test, err := unmarshalTestFile(filepath.Join(root, "test.yaml"))
			if err != nil {
				return err
			}

			cluster, err := createCluster(test.Cluster.Name)
			if err != nil {
				return err
			}
			defer func() {
				cluster.delete()
			}()

			fmt.Printf("(%s) -----> Applying databases\n", test.Cluster.Name)
			for _, database := range test.Databases {
				fmt.Printf("(%s) -----> ... %s\n", test.Cluster.Name, database)
				databaseManifests, err := ioutil.ReadFile(filepath.Join(root, database))
				if err != nil {
					return err
				}

				if err := cluster.apply(databaseManifests); err != nil {
					return err
				}
			}

			fmt.Printf("(%s) -----> Applying SchemaHero Operator\n", test.Cluster.Name)

			fmt.Printf("(%s) -----> Applying database connection\n", test.Cluster.Name)

		}
	}

	return nil
}
