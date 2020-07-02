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

			g := generate.Generator{
				Driver:    v.GetString("driver"),
				URI:       v.GetString("uri"),
				DBName:    v.GetString("dbname"),
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

	cmd.Flags().String("uri", "", "connection string uri")
	cmd.Flags().String("driver", "", "name of the database driver to run")
	cmd.Flags().String("dbname", "", "schemahero database name to write in the yaml")

	cmd.Flags().String("output-dir", cwd, "directory to write schema files to")

	return cmd
}
