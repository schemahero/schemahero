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
			viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			db := database.NewDatabase()
			return db.ApplySync()
		},
	}

	cmd.Flags().String("driver", "", "name of the database driver to use")
	cmd.Flags().String("uri", "", "connection string uri to use")
	cmd.Flags().StringArray("spec-file", []string{}, "filename(s) containing the spec to apply")

	cmd.MarkFlagRequired("driver")
	cmd.MarkFlagRequired("uri")
	cmd.MarkFlagRequired("spec-file")

	return cmd
}
