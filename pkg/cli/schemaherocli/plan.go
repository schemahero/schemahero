package schemaherocli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/schemahero/schemahero/pkg/database"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Plan() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "plan",
		Short:        "plan a spec application against a database",
		Long:         `...`,
		SilenceUsage: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			v := viper.GetViper()

			// to support automaticenv, we can't use cobra required flags
			driver := v.GetString("driver")
			specFile := v.GetString("spec-file")
			uri := v.GetString("uri")
			vaultUriRef := v.GetString("vault-uri-ref")

			if driver == "" || specFile == "" || (uri == "" && vaultUriRef == "") {
				missing := []string{}
				if driver == "" {
					missing = append(missing, "driver")
				}
				if specFile == "" {
					missing = append(missing, "spec-file")
				}
				if uri == "" && vaultUriRef == "" {
					missing = append(missing, "uri or vault-uri-ref")
				}

				return fmt.Errorf("missing required params: %v", missing)
			}

			fi, err := os.Stat(v.GetString("spec-file"))
			if err != nil {
				return err
			}

			if _, err = os.Stat(v.GetString("out")); err == nil {
				if !v.GetBool("overwrite") {
					return errors.Errorf("file %s already exists", v.GetString("out"))
				}

				err = os.RemoveAll(v.GetString("out"))
				if err != nil {
					return errors.Wrap(err, "failed remove existing file")
				}
			}

			var f *os.File
			if v.GetString("out") != "" {
				f, err = os.OpenFile(v.GetString("out"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					return err
				}
				defer f.Close()
			}

			db := database.NewDatabase()
			if fi.Mode().IsDir() {
				err := filepath.Walk(v.GetString("spec-file"), func(path string, info os.FileInfo, err error) error {
					if !info.IsDir() {
						statements, err := db.PlanSync(path)
						if err != nil {
							return err
						}

						if f != nil {
							for _, statement := range statements {
								if _, err := f.WriteString(fmt.Sprintf("%s;\n", statement)); err != nil {
									return err
								}
							}
						} else {
							for _, statement := range statements {
								fmt.Printf("%s;\n", statement)
							}
						}
					}

					return nil
				})

				return err
			} else {
				statements, err := db.PlanSync(v.GetString("spec-file"))
				if err != nil {
					return err
				}

				if f != nil {
					for _, statement := range statements {
						if _, err := f.WriteString(fmt.Sprintf("%s;\n", statement)); err != nil {
							return err
						}
					}
				} else {
					for _, statement := range statements {
						fmt.Printf("%s;\n", statement)
					}
				}

				return nil
			}
		},
	}

	cmd.Flags().String("driver", "", "name of the database driver to use")
	cmd.Flags().String("uri", "", "connection string uri to use")
	cmd.Flags().String("vault-uri-ref", "", "URI-reference to Vault-injected connection URI")
	cmd.Flags().String("spec-file", "", "filename or directory name containing the spec(s) to apply")
	cmd.Flags().String("out", "", "filename to write DDL statements to, if not present output file be written to stdout")
	cmd.Flags().Bool("overwrite", false, "when set, will overwrite the out file, if it already exists")

	return cmd
}
