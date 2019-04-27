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
		RunE: func(cmd *cobra.Command, args []string) error {
			w := watcher.NewWatcher()
			return w.RunSync()
		},
	}

	cmd.Flags().StringP("driver", "d", "", "name of the database to connect to")
	cmd.Flags().StringP("uri", "u", "", "connection string uri")

	cmd.MarkFlagRequired("driver")
	cmd.MarkFlagRequired("uri")

	viper.BindPFlags(cmd.Flags())

	return cmd
}
