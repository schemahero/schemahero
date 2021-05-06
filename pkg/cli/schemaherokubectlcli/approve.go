package schemaherokubectlcli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func ApproveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "approve",
		Short:         "",
		Long:          `...`,
		Args:          cobra.MinimumNArgs(1),
		SilenceErrors: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			err := viper.BindPFlags(cmd.Flags())
			if err != nil {
				fmt.Print("error ", err)
			}
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	cmd.AddCommand(ApproveMigrationCmd())

	return cmd
}
