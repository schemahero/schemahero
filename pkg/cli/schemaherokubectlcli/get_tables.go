package schemaherokubectlcli

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	schemasclientv1alpha4 "github.com/schemahero/schemahero/pkg/client/schemaheroclientset/typed/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func GetTablesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "tables",
		Short:         "",
		Long:          `...`,
		SilenceErrors: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			v := viper.GetViper()
			ctx := context.Background()

			databaseNameFilter := v.GetString("database")

			cfg, err := config.GetRESTConfig()
			if err != nil {
				return err
			}

			client, err := kubernetes.NewForConfig(cfg)
			if err != nil {
				return err
			}

			schemasClient, err := schemasclientv1alpha4.NewForConfig(cfg)
			if err != nil {
				return err
			}

			namespaces, err := client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
			if err != nil {
				return err
			}

			matchingTables := []schemasv1alpha4.Table{}

			for _, namespace := range namespaces.Items {
				tables, err := schemasClient.Tables(namespace.Name).List(ctx, metav1.ListOptions{})
				if err != nil {
					return err
				}

				for _, table := range tables.Items {
					if databaseNameFilter != "" {
						if table.Spec.Database != databaseNameFilter {
							continue
						}
					}

					matchingTables = append(matchingTables, table)
				}
			}

			if len(matchingTables) == 0 {
				fmt.Println("No resources found.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tDATABASE\tPENDING")

			namespaceNames := map[string]struct{}{}
			for _, table := range matchingTables {
				namespaceNames[table.Namespace] = struct{}{}
			}

			migrations := []schemasv1alpha4.Migration{}

			for namespaceName := range namespaceNames {
				namespaceMigrations, err := schemasClient.Migrations(namespaceName).List(context.Background(), metav1.ListOptions{})
				if err != nil {
					return err
				}

				migrations = append(migrations, namespaceMigrations.Items...)
			}

			for _, table := range matchingTables {
				pendingMigrations := 0

				for _, migration := range migrations {
					if migration.Namespace == table.Namespace {
						if migration.Spec.TableName == table.Name {
							if migration.Status.ExecutedAt == int64(0) && migration.Status.RejectedAt == int64(0) {
								pendingMigrations++
							}
						}
					}
				}

				status := "0"
				if pendingMigrations == 1 {
					status = "1"
				} else if pendingMigrations > 1 {
					status = fmt.Sprintf("%d", pendingMigrations)
				}

				fmt.Fprintln(w, fmt.Sprintf("%s\t%s\t%s", table.Name, table.Spec.Database, status))
			}
			w.Flush()

			return nil
		},
	}

	cmd.Flags().StringP("database", "d", "", "database name to filter to results to")

	return cmd
}
