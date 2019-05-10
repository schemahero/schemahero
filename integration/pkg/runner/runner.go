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

	for _, test := range tests {
		if test.IsDir() {
			fmt.Printf("-----> Beginning test %q\n", test.Name())

			cluster, err := createCluster(test.Name())
			if err != nil {
				return err
			}
			defer func() {
				cluster.delete()
			}()

			fmt.Printf("(%s) -----> Applying database.yaml\n", test.Name())
			databaseManifests, err := ioutil.ReadFile(filepath.Join(currentDir, "tests", test.Name(), "database.yaml"))
			if err != nil {
				return err
			}

			if err := cluster.apply(databaseManifests); err != nil {
				return err
			}

			fmt.Printf("(%s) -----> Applying SchemaHero Operator\n", test.Name())

			fmt.Printf("(%s) -----> Applying database connection\n", test.Name())

		}
	}

	return nil
}
