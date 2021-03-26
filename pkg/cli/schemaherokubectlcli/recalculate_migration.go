package schemaherokubectlcli

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	schemasclientv1alpha4 "github.com/schemahero/schemahero/pkg/client/schemaheroclientset/typed/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func RecalculateMigrationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migration",
		Short: "recalculate a generated migration",
		Long: `Executing the recalculate command will discard the previously generated migration and create a new one.

This is useful if the target database schema has changed and the generate migration is no longer valid or ideal.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.MinimumNArgs(1),
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			v := viper.GetViper()
			ctx := context.Background()
			migrationName := args[0]

			// namespace comes from the kubectl flags that we merged
			namespace := v.GetString("namespace")

			// but namespace is optional
			if namespace == "" {
				namespace = metav1.NamespaceDefault
			}

			cfg, err := config.GetRESTConfig()
			if err != nil {
				return err
			}

			schemasClient, err := schemasclientv1alpha4.NewForConfig(cfg)
			if err != nil {
				return err
			}

			// get the migration
			migration, err := schemasClient.Migrations(namespace).Get(ctx, migrationName, metav1.GetOptions{})
			if err != nil {
				return err
			}

			// delete the migration
			if err := schemasClient.Migrations(migration.Namespace).Delete(ctx, migration.Name, metav1.DeleteOptions{}); err != nil {
				return err
			}

			// touch the table so that it's regenerate in the normal reconcile loop
			fmt.Printf("--> %s/%s \n\n <--", migration.Spec.TableNamespace, migration.Spec.TableName)
			table, err := schemasClient.Tables(migration.Spec.TableNamespace).Get(ctx, migration.Spec.TableName, metav1.GetOptions{})
			if err != nil {
				return errors.Wrap(err, "get table")
			}

			table.Status.LastPlannedTableSpecSHA = ""
			fmt.Printf("!!! %s !!! \n", table.Namespace)
			_, err = schemasClient.Tables(table.Namespace).Update(ctx, table, metav1.UpdateOptions{})
			if err != nil {
				return errors.Wrap(err, "update table")
			}

			return nil
		},
	}

	return cmd
}
