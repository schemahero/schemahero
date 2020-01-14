package schemaherokubectlcli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var (
	KubernetesConfigFlags *genericclioptions.ConfigFlags
)

func RootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "schemahero",
		Short: "SchemaHero is a cloud-native database schema management tool",
		Long:  `...`,
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.Flags())
		},
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
			os.Exit(1)
		},
	}

	cobra.OnInitialize(initConfig)

	cmd.AddCommand(Version())
	cmd.AddCommand(GetCmd())
	cmd.AddCommand(DescribeCmd())
	cmd.AddCommand(UpdateCmd())
	cmd.AddCommand(ApproveCmd())

	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	KubernetesConfigFlags = genericclioptions.NewConfigFlags(true)
	KubernetesConfigFlags.AddFlags(cmd.PersistentFlags())

	return cmd
}

func InitAndExecute() {
	if err := RootCmd().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initConfig() {
	viper.SetEnvPrefix("SCHEMAHERO")
	viper.AutomaticEnv()
}
