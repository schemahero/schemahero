package cli

import (
	"github.com/schemahero/schemahero/integration/pkg/runner"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Run() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "run the integration tests",
		Long:  `...`,
		RunE: func(cmd *cobra.Command, args []string) error {
			r := runner.NewRunner()
			return r.RunSync()
		},
	}

	cmd.Flags().String("manager-image-name", "", "docker image for the manager pod")
	cmd.Flags().String("schemahero-image-name", "", "docker image for the schemahero pod")

	viper.BindPFlags(cmd.Flags())

	return cmd
}
