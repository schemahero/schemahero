package cli

import (
	"github.com/schemahero/schemahero/pkg/database"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Apply() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "apply a spec to a database",
		Long:  `...`,
		PreRun: func(cmd *cobra.Command, args []string) {
			// workaround for https://github.com/spf13/viper/issues/233
			viper.BindPFlag("driver", cmd.Flags().Lookup("driver"))
			viper.BindPFlag("uri", cmd.Flags().Lookup("uri"))
			viper.BindPFlag("spec-file", cmd.Flags().Lookup("spec-file"))
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			db := database.NewDatabase()
			return db.ApplySync()
		},
	}

	cmd.Flags().String("driver", "", "name of the database driver to use")
	cmd.Flags().String("uri", "", "connection string uri to use")
	cmd.Flags().String("spec-file", "", "filename containing the spec to apply")

	cmd.MarkFlagRequired("driver")
	cmd.MarkFlagRequired("uri")
	cmd.MarkFlagRequired("spec-file")

	viper.BindPFlags(cmd.Flags())

	return cmd
}
