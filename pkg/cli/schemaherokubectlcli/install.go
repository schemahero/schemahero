package schemaherokubectlcli

import (
	"fmt"

	"github.com/schemahero/schemahero/pkg/installer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
			v := viper.GetViper()

			if v.GetBool("yaml") {
				manifests, err := installer.GenerateOperatorYAML()
				if err != nil {
					fmt.Printf("Error: %s\n", err.Error())
					return err
				}

				fmt.Printf("%s\n", manifests)
				return nil
			}
			if err := installer.InstallOperator(); err != nil {
				fmt.Printf("Error: %s\n", err.Error())
				return err
			}

			fmt.Println("The SchemaHero operator has been installed to the cluster")
			return nil
		},
	}

	cmd.Flags().Bool("yaml", false, "Is present, don't install the operator, just generate the yaml")

	return cmd
}
