package cli

import (
	"github.com/schemahero/schemahero/pkg/watcher"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Watch() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "watch",
		Short: "watch a database for any changes",
		Long:  `...`,
		PreRun: func(cmd *cobra.Command, args []string) {
			// workaround for https://github.com/spf13/viper/issues/233
			viper.BindPFlag("driver", cmd.Flags().Lookup("driver"))
			viper.BindPFlag("uri", cmd.Flags().Lookup("uri"))
			viper.BindPFlag("namespace", cmd.Flags().Lookup("namespace"))
			viper.BindPFlag("instance", cmd.Flags().Lookup("instance"))
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			w := watcher.NewWatcher()
			return w.RunSync()
		},
	}

	cmd.Flags().String("driver", "", "name of the database driver to use")
	cmd.Flags().String("uri", "", "connection string uri")
	cmd.Flags().String("namespace", "default", "namespace of the spwaning object")
	cmd.Flags().String("instance", "", "instance name of the spawning object")

	cmd.MarkFlagRequired("driver")
	cmd.MarkFlagRequired("uri")
	cmd.MarkFlagRequired("namespace")
	cmd.MarkFlagRequired("instance")

	viper.BindPFlags(cmd.Flags())

	return cmd
}
