package schemaherokubectlcli

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/schemahero/schemahero/pkg/database"
	"github.com/schemahero/schemahero/pkg/database/types"
	"github.com/schemahero/schemahero/pkg/files"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func LintCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "lint",
		Short:        "lint database schema specifications",
		Long:         `Validates schema specifications for correctness and best practices`,
		SilenceUsage: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			v := viper.GetViper()

			driver := v.GetString("driver")
			specFile := v.GetString("spec-file")

			if driver == "" || specFile == "" {
				missing := []string{}
				if driver == "" {
					missing = append(missing, "driver")
				}
				if specFile == "" {
					missing = append(missing, "spec-file")
				}

				if len(missing) > 0 {
					return fmt.Errorf("missing required params: %v", missing)
				}
			}

			fi, err := os.Stat(specFile)
			if err != nil {
				return err
			}

			db := database.Database{
				Driver: driver,
			}

			hasErrors := false
			specsFromFiles := []types.Spec{}

			if fi.Mode().IsDir() {
				err := filepath.Walk(specFile, func(path string, info os.FileInfo, err error) error {
					isHidden, err := files.IsHidden(path)
					if err != nil {
						return err
					}

					if info.IsDir() {
						if isHidden {
							return filepath.SkipDir
						}
						return nil
					}

					if isHidden {
						return nil
					}

					specContents, err := ioutil.ReadFile(filepath.Clean(path))
					if err != nil {
						return errors.Wrap(err, "failed to read file")
					}
					specsFromFiles = append(specsFromFiles, types.Spec{
						SourceFilename: path,
						Spec:           specContents,
					})

					return nil
				})

				if err != nil {
					return errors.Wrap(err, "failed to walk directory")
				}

				db.SortSpecs(specsFromFiles)

				for _, spec := range specsFromFiles {
					if err := db.ValidateSpec(spec.Spec, v.GetString("spec-type")); err != nil {
						fmt.Printf("Error in %s: %v\n", spec.SourceFilename, err)
						hasErrors = true
					} else {
						fmt.Printf("✓ %s is valid\n", spec.SourceFilename)
					}
				}
			} else {
				specContents, err := ioutil.ReadFile(filepath.Clean(specFile))
				if err != nil {
					return errors.Wrap(err, "failed to read file")
				}

				if err := db.ValidateSpec(specContents, v.GetString("spec-type")); err != nil {
					fmt.Printf("Error in %s: %v\n", specFile, err)
					hasErrors = true
				} else {
					fmt.Printf("✓ %s is valid\n", specFile)
				}
			}

			if hasErrors {
				return errors.New("one or more specs failed validation")
			}

			return nil
		},
	}

	cmd.Flags().String("driver", "", "name of the database driver to use")
	cmd.Flags().String("spec-file", "", "filename or directory name containing the spec(s) to lint")
	cmd.Flags().String("spec-type", "table", "type of spec in spec-file")

	return cmd
}
