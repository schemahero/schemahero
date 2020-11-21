package schemaherokubectlcli

import (
	"fmt"
	"os"

	"github.com/schemahero/schemahero/pkg/generate"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func GenerateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "generate",
		Short:         "",
		Long:          `...`,
		SilenceErrors: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			v := viper.GetViper()

			uri := v.GetString("uri")
			driver := v.GetString("driver")
			dbName := v.GetString("dbname")

			if uri == "" || driver == "" || dbName == "" {
				cmd.PrintErr("missing required parameters")
				return cmd.Help()
			}

			g := generate.Generator{
				Driver:    driver,
				URI:       uri,
				DBName:    dbName,
				OutputDir: v.GetString("output-dir"),
			}
			return g.RunSync()

		},
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("unable to get workdir: %s\n", err.Error())
		cwd = "."
	}

	cmd.Flags().String("uri", "", "connection string uri (required)")
	cmd.Flags().String("driver", "", "name of the database driver to run (required)")
	cmd.Flags().String("dbname", "", "schemahero database name to write in the yaml (required)")

	cmd.MarkFlagRequired("uri")
	cmd.MarkFlagRequired("driver")
	cmd.MarkFlagRequired("dbname")

	cmd.Flags().String("output-dir", cwd, "directory to write schema files to")

	return cmd
}
