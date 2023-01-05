package schemaherokubectlcli

import (
	"context"

	"github.com/schemahero/schemahero/pkg/database"
	"github.com/schemahero/schemahero/pkg/trace"
	"go.opentelemetry.io/otel"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func FixturesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fixtures",
		Short: "fixtures creates sql statements from a schemahero definition",
		Long:  `...`,
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, span := otel.Tracer(trace.TraceName).Start(context.Background(), "FixturesCmd")
			defer span.End()

			v := viper.GetViper()

			db := database.Database{
				InputDir:  v.GetString("input-dir"),
				OutputDir: v.GetString("output-dir"),
				Driver:    v.GetString("driver"),
				URI:       v.GetString("uri")}

			return db.CreateFixturesSync(ctx)
		},
	}

	cmd.Flags().String("driver", "", "name of the database driver to use")
	cmd.Flags().String("dbname", "", "schemahero database name to write in the yaml")
	cmd.Flags().String("input-dir", "", "directory to read schema files from")
	cmd.Flags().String("output-dir", "", "directory to write fixture files to")

	cmd.MarkFlagRequired("driver")
	cmd.MarkFlagRequired("dbname")
	cmd.MarkFlagRequired("input-dir")
	cmd.MarkFlagRequired("output-dir")

	return cmd
}
