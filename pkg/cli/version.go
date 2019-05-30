package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/schemahero/schemahero/pkg/version"
)

func Version() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "schemahero version information",
		Long:  `...`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("%+v\n", version.GetBuild())
			return nil
		},
	}

	return cmd
}
