package schemaherokubectlcli

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/schemahero/schemahero/pkg/database"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func ApplyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "apply",
		Short:        "apply a spec to a database",
		Long:         `...`,
		SilenceUsage: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			v := viper.GetViper()

			// to support automaticenv, we can't use cobra required flags
			driver := v.GetString("driver")
			ddl := v.GetString("ddl")
			uri := v.GetString("uri")
			host := v.GetStringSlice("host")

			if driver == "" || ddl == "" || uri == "" || len(host) == 0 {
				missing := []string{}
				if driver == "" {
					missing = append(missing, "driver")
				}
				if ddl == "" {
					missing = append(missing, "ddl")
				}

				// one of uri or host/port must be specified
				if uri == "" && len(host) == 0 {
					missing = append(missing, "uri or host(s)")
				}

				if len(missing) > 0 {
					return fmt.Errorf("missing required params: %v", missing)
				}
			}

			fi, err := os.Stat(v.GetString("ddl"))
			if err != nil {
				return errors.Wrap(err, "failed to stat ddl file")
			}

			db := database.Database{
				InputDir:  v.GetString("input-dir"),
				OutputDir: v.GetString("output-dir"),
				Driver:    v.GetString("driver"),
				URI:       v.GetString("uri"),
				Hosts:     v.GetStringSlice("host"),
				Username:  v.GetString("username"),
				Password:  v.GetString("password"),
				Keyspace:  v.GetString("keyspace"),
			}

			commands := []string{}
			if fi.Mode().IsDir() {
				err := filepath.Walk(v.GetString("ddl"), func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if info.IsDir() {
						return nil
					}

					ddl, err := ioutil.ReadFile(filepath.Clean(path))
					if err != nil {
						return errors.Wrap(err, "failed to read file in directory")
					}

					statements := db.GetStatementsFromDDL(string(ddl))
					commands = append(commands, statements...)

					return nil
				})

				if err != nil {
					return errors.Wrap(err, "failed to walk ddl directory")
				}

				return nil
			} else {
				ddl, err := ioutil.ReadFile(v.GetString("ddl"))
				if err != nil {
					return errors.Wrap(err, "failed to read file")
				}

				statements := db.GetStatementsFromDDL(string(ddl))
				commands = append(commands, statements...)
			}

			if err := db.ApplySync(commands); err != nil {
				return errors.Wrap(err, "failed to apply commands")
			}

			return nil
		},
	}

	cmd.Flags().String("driver", "", "name of the database driver to use")

	cmd.Flags().String("uri", "", "connection string uri to use")

	cmd.Flags().String("username", "", "username to use when connecting")
	cmd.Flags().String("password", "", "password to use when connecting")
	cmd.Flags().StringSlice("host", []string{}, "hostname to use when connecting")
	cmd.Flags().String("keyspace", "", "the keyspace to use for databases that support keyspaces")

	cmd.Flags().String("ddl", "", "filename or directory name containing the rendered DDL commands to execute")

	return cmd
}
