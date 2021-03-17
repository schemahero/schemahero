package schemaherokubectlcli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/schemahero/schemahero/pkg/version"
)

func Version() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "version",
		Short:         "schemahero version information",
		Long:          `Prints the current version of the SchemaHero CLI. This may or may not match the version in the cluster.`,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("SchemaHero %s\n", version.Version())
			return nil
		},
	}

	return cmd
}
