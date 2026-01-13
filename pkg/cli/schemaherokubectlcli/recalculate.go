package schemaherokubectlcli

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func RecalculateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "recalculate",
		Short:         "Recalculate planned migration",
		Long:          "Recalculate annotates table from migration to trigger reconcile loop",
		Args:          cobra.ExactArgs(0),
		SilenceErrors: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.Flags())
		},
	}

	cmd.AddCommand(RecalculateMigrationCmd())

	return cmd
}
