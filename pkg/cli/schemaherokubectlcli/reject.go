package schemaherokubectlcli

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func RejectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "reject",
		Short:         "",
		Long:          `...`,
		Args:          cobra.ExactArgs(0),
		SilenceErrors: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	cmd.AddCommand(RejectMigrationCmd())

	return cmd
}
