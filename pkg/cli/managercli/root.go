package managercli

import (
	"fmt"
	"os"
	"strings"

	"github.com/schemahero/schemahero/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func RootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "manager",
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

	config.AddFlags(cmd.PersistentFlags())

	cmd.PersistentFlags().String("log-level", "info", "set the log level")

	cmd.AddCommand(Version())
	cmd.AddCommand(RunCmd())

	viper.BindPFlags(cmd.Flags())

	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

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
