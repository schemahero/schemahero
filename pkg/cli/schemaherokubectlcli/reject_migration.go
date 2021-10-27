package schemaherokubectlcli

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	schemasclientv1alpha4 "github.com/schemahero/schemahero/pkg/client/schemaheroclientset/typed/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	kuberneteserrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func RejectMigrationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "migration",
		Short:         "reject a migration, which will prevent it from being applied to the database without recalculation",
		Long:          `reject a migration, which will prevent it from being applied to the database without recalculation`,
		Args:          cobra.MinimumNArgs(1),
		SilenceErrors: true,
		SilenceUsage:  true,
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			v := viper.GetViper()
			ctx := context.Background()
			migrationName := args[0]

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

			namespaceNames := []string{}

			if viper.GetBool("all-namespaces") {
				namespaces, err := client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
				if err != nil {
					return err
				}

				for _, namespace := range namespaces.Items {
					namespaceNames = append(namespaceNames, namespace.Name)
				}
			} else {
				if v.GetString("namespace") != "" {
					namespaceNames = []string{v.GetString("namespace")}
				} else {
					namespaceNames = []string{"default"}
				}
			}

			for _, namespaceName := range namespaceNames {
				migration, err := schemasClient.Migrations(namespaceName).Get(ctx, migrationName, metav1.GetOptions{})
				if kuberneteserrors.IsNotFound(err) {
					// continue to the next namespace
					continue
				}
				if err != nil {
					return err
				}

				// approved migrations cannot be rejected
				if migration.Status.ApprovedAt > 0 || migration.Status.ExecutedAt > 0 {
					return errors.Errorf("Already approved migration %s cannot be rejected\n", migrationName)
				}

				migration.Status.RejectedAt = time.Now().Unix()
				if _, err := schemasClient.Migrations(namespaceName).Update(ctx, migration, metav1.UpdateOptions{}); err != nil {
					return err
				}

				fmt.Printf("Migration %s rejected\n", migrationName)
				return nil
			}

			err = errors.Errorf("migration %q not found", migrationName)
			return err
		},
	}

	cmd.Flags().Bool("all-namespaces", false, "If present, look to reject the specified migration in any namespace. Namespace in current context is ignored even if specified with --namespace.")

	return cmd
}
