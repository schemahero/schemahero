package managercli

import (
	"os"

	"github.com/schemahero/schemahero/pkg/apis"
	"github.com/schemahero/schemahero/pkg/controller"
	"github.com/schemahero/schemahero/pkg/logger"
	"github.com/schemahero/schemahero/pkg/version"
	"github.com/schemahero/schemahero/pkg/webhook"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

func RunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "run",
		Short:         "runs the schemahero manager",
		Long:          `...`,
		SilenceErrors: true,
		SilenceUsage:  true,
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			logger.Infof("Starting schemahero version %+v", version.GetBuild())

			v := viper.GetViper()

			// Get a config to talk to the apiserver
			cfg, err := config.GetConfig()
			if err != nil {
				logger.Error(err)
				os.Exit(1)
			}

			// Create a new Cmd to provide shared dependencies and start components
			mgr, err := manager.New(cfg, manager.Options{MetricsBindAddress: v.GetString("metrics-addr")})
			if err != nil {
				logger.Error(err)
				os.Exit(1)
			}

			// Setup Scheme for all resources
			if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
				logger.Error(err)
				os.Exit(1)
			}

			// Setup all Controllers
			if err := controller.AddToManager(mgr); err != nil {
				logger.Error(err)
				os.Exit(1)
			}

			if err := webhook.AddToManager(mgr); err != nil {
				logger.Error(err)
				os.Exit(1)
			}

			// Start the Cmd
			if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
				logger.Error(err)
				os.Exit(1)
			}

			return nil
		},
	}

	cmd.Flags().String("metrics-addr", ":8088", "The address the metric endpoint binds to.")

	return cmd
}
