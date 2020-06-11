package schemaherokubectlcli

import (
	"context"
	"fmt"
	"os"

	databasesclientv1alpha4 "github.com/schemahero/schemahero/pkg/client/schemaheroclientset/typed/databases/v1alpha4"
	"github.com/schemahero/schemahero/pkg/generate"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func GenerateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "generate",
		Short:         "",
		Long:          `...`,
		SilenceErrors: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			v := viper.GetViper()

			cfg, err := config.GetConfig()
			if err != nil {
				return err
			}

			databasesClient, err := databasesclientv1alpha4.NewForConfig(cfg)
			if err != nil {
				return err
			}

			namespace := v.GetString("namespace")
			if namespace == "" {
				namespace = corev1.NamespaceDefault
			}

			ctx := context.Background()
			database, err := databasesClient.Databases(namespace).Get(ctx, v.GetString("database"), metav1.GetOptions{})
			if err != nil {
				return err
			}

			driver := ""
			connectionURI := ""

			if database.Spec.Connection.Postgres != nil {
				driver = "postgres"
				connectionURI = database.Spec.Connection.Postgres.URI.Value
			} else if database.Spec.Connection.Mysql != nil {
				driver = "mysql"
				connectionURI = database.Spec.Connection.Mysql.URI.Value
			} else if database.Spec.Connection.CockroachDB != nil {
				driver = "cockroachdb"
				connectionURI = database.Spec.Connection.CockroachDB.URI.Value
			}

			// TODO support Vault
			// TODO support non URI connection params

			g := generate.Generator{
				Driver:    driver,
				URI:       connectionURI,
				DBName:    database.Name,
				OutputDir: v.GetString("output-dir"),
			}

			return g.RunSync()
		},
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("unable to get workdir: %s\n", err.Error())
		cwd = "."
	}

	cmd.Flags().StringP("database", "d", "", "database name to generate table yaml for")
	cmd.Flags().String("output-dir", cwd, "directory to write schema files to")

	return cmd
}
