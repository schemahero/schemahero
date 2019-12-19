package schemaherocli

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
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			g := generate.NewGenerator()
			return g.RunSync()
		},
	}

	cmd.Flags().String("driver", "", "name of the database driver to use")
	cmd.Flags().String("uri", "", "connection string uri")
	cmd.Flags().String("namespace", "default", "namespace to put the custom resources into")
	cmd.Flags().String("dbname", "", "schemahero database name to write in the yaml")
	cmd.Flags().String("output-dir", "", "directory to write schema files to")

	cmd.MarkFlagRequired("driver")
	cmd.MarkFlagRequired("uri")
	cmd.MarkFlagRequired("dbname")

	return cmd
}
