package schemaherokubectlcli

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
	schemasclientv1alpha4 "github.com/schemahero/schemahero/pkg/client/schemaheroclientset/typed/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	kuberneteserrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

func RecalculateMigrationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "migration",
		Short:         "",
		Long:          `...`,
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

				recalculatedAt := time.Now().Format(time.RFC3339)

				patch := map[string]interface{}{
					"metadata": map[string]interface{}{
						"annotations": map[string]string{
							"recalculatedAt": recalculatedAt,
						},
					},
				}
				patchData, err := json.Marshal(patch)
				if err != nil {
					fmt.Printf("Failed to marshal patch data: %v\n", err)
					return err
				}

				if _, err := schemasClient.Tables(namespaceName).Patch(ctx, migration.Spec.TableName, types.MergePatchType, patchData, metav1.PatchOptions{}); err != nil {
					return err
				}

				fmt.Printf("Migration %s updated\n", migrationName)
				return nil
			}

			err = errors.Errorf("migration %q not found", migrationName)
			return err
		},
	}

	cmd.Flags().Bool("all-namespaces", false, "If present, list the requested object(s) across all namespaces. Namespace in current context is ignored even if specified with --namespace.")

	return cmd
}
