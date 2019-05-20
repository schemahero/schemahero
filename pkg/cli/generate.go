package cli

import (
	"github.com/schemahero/schemahero/pkg/generate"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Generate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "generate schemahero custom resources from a running database instance",
		Long:  `...`,
		RunE: func(cmd *cobra.Command, args []string) error {
			g := generate.NewGenerator()
			return g.RunSync()
		},
	}

	cmd.Flags().StringP("driver", "d", "", "name of the database driver to use")
	cmd.Flags().StringP("uri", "u", "", "connection string uri")
	cmd.Flags().StringP("namespace", "n", "default", "namespace to put the custom resources into")
	cmd.Flags().StringP("dbname", "", "", "schemahero database name to write in the yaml")
	cmd.Flags().StringP("output-dir", "o", "", "directory to write schema files to")

	cmd.MarkFlagRequired("driver")
	cmd.MarkFlagRequired("uri")
	cmd.MarkFlagRequired("dbname")

	viper.BindPFlags(cmd.Flags())

	return cmd
}
