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

func PlanCmd() *cobra.Command {
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
			host := v.GetStringSlice("host")

			if driver == "" || specFile == "" || uri == "" || len(host) == 0 {
				missing := []string{}
				if driver == "" {
					missing = append(missing, "driver")
				}
				if specFile == "" {
					missing = append(missing, "spec-file")
				}

				// one of uri or host/port must be specified
				if uri == "" && len(host) == 0 {
					missing = append(missing, "uri or host(s)")
				}

				if len(missing) > 0 {
					return fmt.Errorf("missing required params: %v", missing)
				}
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
				f, err = os.OpenFile(v.GetString("out"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
				if err != nil {
					return err
				}
				defer func() {
					f.Close()
				}()
			}

			db := database.Database{
				InputDir:       v.GetString("input-dir"),
				OutputDir:      v.GetString("output-dir"),
				Driver:         v.GetString("driver"),
				URI:            v.GetString("uri"),
				Hosts:          v.GetStringSlice("host"),
				Username:       v.GetString("username"),
				Password:       v.GetString("password"),
				Keyspace:       v.GetString("keyspace"),
				DeploySeedData: v.GetBool("seed-data"),
			}

			specsFromFiles := []types.Spec{}
			if fi.Mode().IsDir() {
				err := filepath.Walk(v.GetString("spec-file"), func(path string, info os.FileInfo, err error) error {
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
					statements, err := db.PlanSync(spec.Spec, v.GetString("spec-type"))
					if err != nil {
						return fmt.Errorf("plan sync from file %q: %w", spec.SourceFilename, err)
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
			} else {
				statements, err := db.PlanSyncFromFile(v.GetString("spec-file"), v.GetString("spec-type"))
				if err != nil {
					return fmt.Errorf("plan sync from file %q: %w", v.GetString("spec-file"), err)
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

	cmd.Flags().String("username", "", "username to use when connecting")
	cmd.Flags().String("password", "", "password to use when connecting")
	cmd.Flags().StringSlice("host", []string{}, "hostname to use when connecting")
	cmd.Flags().String("keyspace", "", "the keyspace to use for databases that support keyspaces")

	cmd.Flags().String("spec-file", "", "filename or directory name containing the spec(s) to apply")
	cmd.Flags().String("spec-type", "table", "type of spec in spec-file")
	cmd.Flags().String("out", "", "filename to write DDL statements to, if not present output file be written to stdout")
	cmd.Flags().Bool("overwrite", true, "when set, will overwrite the out file, if it already exists")

	cmd.Flags().Bool("seed-data", false, "when set, will deploy seed data")
	return cmd
}
