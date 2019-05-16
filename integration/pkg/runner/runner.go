package runner

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

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
			if test.Cluster.SkipCleanup {
				fmt.Printf("\n\n WARNING: Skipping Cleanup of %s\n\n\n", test.Cluster.Name)
			} else {
				defer func() {
					fmt.Printf("(%s) -----> Deleting cluster\n", test.Cluster.Name)
					cluster.delete()
				}()
			}

			fmt.Printf("(%s) -----> Applying databases\n", test.Cluster.Name)
			for _, database := range test.Databases {
				fmt.Printf("(%s) -----> ... %s\n", test.Cluster.Name, database)
				databaseManifests, err := ioutil.ReadFile(filepath.Join(root, database))
				if err != nil {
					return err
				}

				if err := cluster.apply(databaseManifests, false); err != nil {
					return err
				}
			}

			// TODO the database has to be started before continuing here...
			// The framework (really the operator itself) should handle this
			time.Sleep(time.Second * 30)

			fmt.Printf("(%s) -----> Applying SchemaHero Operator\n", test.Cluster.Name)
			operator, err := getApplyableOperator(r.Viper.GetString("manager-image-name"))
			if err != nil {
				return nil
			}
			if err := cluster.apply(operator, false); err != nil {
				return err
			}

			// Give the cluster 2 seconds to register the CRDs. This shouldn't be necessary
			// And for production code this would be better handled in almost any other way
			// But this test flow is a very specific pattern and this is working for now.
			time.Sleep(time.Second * 2)

			fmt.Printf("(%s) -----> Applying database connections\n", test.Cluster.Name)
			for _, connection := range test.Connections {
				fmt.Printf("(%s) -----> ... %s\n", test.Cluster.Name, connection)
				connectionManifests, err := ioutil.ReadFile(filepath.Join(root, connection))
				if err != nil {
					return err
				}

				replacedConnetionManifests := strings.Replace(string(connectionManifests), "__SCHEMAHERO_IMAGE_NAME__", r.Viper.GetString("schemahero-image-name"), -1)
				if err := cluster.apply([]byte(replacedConnetionManifests), false); err != nil {
					return err
				}
			}

			fmt.Printf("(%s) -----> Setting up test\n", test.Cluster.Name)
			// TODO

			fmt.Printf("(%s) -----> Running test(s)\n", test.Cluster.Name)
			for _, testStep := range test.Steps {
				fmt.Printf("(%s) -----> ... %s\n", test.Cluster.Name, testStep.Name)

				if testStep.Table != nil {
					sourceManifests, err := ioutil.ReadFile(filepath.Join(root, testStep.Table.Source))
					if err != nil {
						return err
					}

					if err := cluster.apply(sourceManifests, true); err != nil {
						return err
					}

					// Wait up to 5 seconds for verification to pass
					verifyCheckCount := 0
					maxVerifications := 10
					for verifyCheckCount < maxVerifications {
						ok, stdout, stderr, err := verify(cluster, testStep.Verification)
						if err != nil {
							return err
						}

						if ok {
							verifyCheckCount = maxVerifications
						} else {
							verifyCheckCount++
							if verifyCheckCount == maxVerifications {
								fmt.Printf("stderr: %s\n", stderr)
								fmt.Printf("stdout: %s\n", stdout)
								return errors.New("verification failed")
							} else {
								time.Sleep(time.Second)
							}
						}
					}
				}
			}
		}
	}

	return nil
}
