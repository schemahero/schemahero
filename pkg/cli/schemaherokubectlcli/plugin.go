package schemaherokubectlcli

import (
	"fmt"

	"github.com/schemahero/schemahero/pkg/database/plugin"
	"github.com/spf13/cobra"
)

func PluginCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plugin",
		Short: "Manage SchemaHero database plugins",
	}

	cmd.AddCommand(PluginDownloadCmd())

	return cmd
}

func PluginDownloadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "download <database>",
		Short:         "Download a SchemaHero database plugin",
		Args:          cobra.ExactArgs(1),
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			pluginManager := plugin.GetGlobalPluginManager()
			if pluginManager == nil {
				pluginManager = plugin.InitializePluginSystem()
			}

			pluginPath, err := pluginManager.DownloadPlugin(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Downloaded SchemaHero %s plugin to %s\n", args[0], pluginPath)
			return nil
		},
	}

	return cmd
}
