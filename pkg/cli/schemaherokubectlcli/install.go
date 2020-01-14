package schemaherokubectlcli

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/schemahero/schemahero/pkg/installer"
)

func InstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "install",
		Short:         "install the schemahero operator to the cluster",
		Long:          `...`,
		SilenceErrors: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := installer.InstallOperator(); err != nil {
				return err
			}

			return nil
		},
	}

	return cmd
}
