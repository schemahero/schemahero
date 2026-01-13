package schemaherokubectlcli

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func DescribeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "describe",
		Short:         "",
		Long:          `...`,
		Args:          cobra.ExactArgs(0),
		SilenceErrors: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.Flags())
		},
	}

	cmd.AddCommand(DescribeMigrationCmd())

	return cmd
}
