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
	corev1 "k8s.io/api/core/v1"
	kuberneteserrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func DescribeMigrationCmd() *cobra.Command {
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
				foundMigration, err := schemasClient.Migrations(namespaceName).Get(ctx, migrationName, metav1.GetOptions{})
				if kuberneteserrors.IsNotFound(err) {
					// next namespace
					continue
				}
				if err != nil {
					return err
				}

				baseCommand := "kubectl schemahero"
				if namespaceName != corev1.NamespaceDefault {
					baseCommand = fmt.Sprintf("%s -n %s", baseCommand, namespaceName)
				}

				fmt.Printf("\nMigration Name: %s\n\n", foundMigration.Name)
				fmt.Printf("Generated DDL Statement (generated at %s): \n  %s\n",
					time.Unix(foundMigration.Status.PlannedAt, 0).Format(time.RFC3339),
					foundMigration.Spec.GeneratedDDL)

				fmt.Println("")
				fmt.Println("To apply this migration:")
				fmt.Printf(`  %s approve migration %s`, baseCommand, foundMigration.Name)
				fmt.Println("")

				fmt.Println("")
				fmt.Println("To recalculate this migration against the current schema:")
				fmt.Printf(`  %s recalculate migration %s`, baseCommand, foundMigration.Name)
				fmt.Println("")

				fmt.Println("")
				fmt.Println("To deny and cancel this migration:")
				fmt.Printf(`  %s reject migration %s`, baseCommand, foundMigration.Name)
				fmt.Println("")

				return nil
			}

			err = errors.Errorf("migration %q not found", migrationName)
			return err

		},
	}

	cmd.Flags().Bool("all-namespaces", false, "If present, list the requested object(s) across all namespaces. Namespace in current context is ignored even if specified with --namespace.")
	cmd.Flags().StringP("output", "o", "yaml", "Output format (can be json or yaml")

	return cmd
}
