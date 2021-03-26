package schemaherokubectlcli

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func RecalculateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "recalculate",
		Short:         "",
		Long:          ``,
		SilenceErrors: true,
		Args:          cobra.MinimumNArgs(1),
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	cmd.AddCommand(RecalculateMigrationCmd())

	return cmd
}
