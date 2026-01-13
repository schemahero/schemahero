package schemaherokubectlcli

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func GetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "get",
		Short:         "",
		Long:          `...`,
		Args:          cobra.ExactArgs(0),
		SilenceErrors: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.Flags())
		},
	}

	cmd.AddCommand(GetDatabasesCmd())
	cmd.AddCommand(GetTablesCmd())
	cmd.AddCommand(GetMigrationsCmd())

	return cmd
}
