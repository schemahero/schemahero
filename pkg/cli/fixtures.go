package cli

import (
	"github.com/schemahero/schemahero/pkg/fixtures"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Fixtures() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fixtures",
		Short: "fixtures creates sql statements from a schemahero definition",
		Long:  `...`,
		PreRun: func(cmd *cobra.Command, args []string) {
			// workaround for https://github.com/spf13/viper/issues/233
			viper.BindPFlag("driver", cmd.Flags().Lookup("driver"))
			viper.BindPFlag("dbname", cmd.Flags().Lookup("dbname"))
			viper.BindPFlag("input-dir", cmd.Flags().Lookup("input-dir"))
			viper.BindPFlag("output-dir", cmd.Flags().Lookup("output-dir"))
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			f := fixtures.NewFixturator()
			return f.RunSync()
		},
	}

	cmd.Flags().String("driver", "", "name of the database driver to use")
	cmd.Flags().String("dbname", "", "schemahero database name to write in the yaml")
	cmd.Flags().String("input-dir", "", "directory to read schema files from")
	cmd.Flags().String("output-dir", "", "directory to write fixture files to")

	cmd.MarkFlagRequired("driver")
	cmd.MarkFlagRequired("dbname")
	cmd.MarkFlagRequired("input-dir")
	cmd.MarkFlagRequired("output-dir")

	viper.BindPFlags(cmd.Flags())

	return cmd
}
