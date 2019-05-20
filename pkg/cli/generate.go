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
		PreRun: func(cmd *cobra.Command, args []string) {
			// workaround for https://github.com/spf13/viper/issues/233
			viper.BindPFlag("driver", cmd.Flags().Lookup("driver"))
			viper.BindPFlag("uri", cmd.Flags().Lookup("uri"))
			viper.BindPFlag("namespace", cmd.Flags().Lookup("namespace"))
			viper.BindPFlag("dbname", cmd.Flags().Lookup("dbname"))
			viper.BindPFlag("output-dir", cmd.Flags().Lookup("output-dir"))
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

	viper.BindPFlags(cmd.Flags())

	return cmd
}
