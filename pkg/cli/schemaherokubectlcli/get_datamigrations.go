package schemaherokubectlcli

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/client/schemaheroclientset"
	"github.com/schemahero/schemahero/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func GetDataMigrationsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "datamigrations",
		Short:         "list datamigrations",
		Long:          `Show all data migrations in a namespace`,
		Args:          cobra.MaximumNArgs(1),
		SilenceErrors: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			v := viper.GetViper()

			namespaceNames, err := getNamespacesOrDefault(v.GetString("namespace"), v.GetStringSlice("all-namespaces"))
			if err != nil {
				return err
			}

			matchingDataMigrations := []schemasv1alpha4.DataMigration{}

			for _, namespace := range namespaceNames {
				// Get data migrations
				ctx := context.Background()

				cfg, err := config.GetRESTConfig()
				if err != nil {
					return err
				}

				schemaHeroClient, err := schemaheroclientset.NewForConfig(cfg)
				if err != nil {
					return err
				}

				dataMigrations, err := schemaHeroClient.SchemasV1alpha4().DataMigrations(namespace).List(ctx, metav1.ListOptions{})
				if err != nil {
					return err
				}

				if len(args) == 1 {
					// Filter by name if provided
					for _, dataMigration := range dataMigrations.Items {
						if dataMigration.Name == args[0] {
							matchingDataMigrations = append(matchingDataMigrations, dataMigration)
						}
					}
				} else {
					matchingDataMigrations = append(matchingDataMigrations, dataMigrations.Items...)
				}
			}

			if len(matchingDataMigrations) == 0 {
				if len(namespaceNames) == 1 {
					fmt.Printf("No resources found in %s namespace.\n", namespaceNames[0])
				} else {
					fmt.Println("No resources found.")
				}
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tDATABASE\tPHASE\tMIGRATION\tAGE")

			for _, dataMigration := range matchingDataMigrations {
				age := time.Since(dataMigration.CreationTimestamp.Time).Round(time.Second)
				
				phase := string(dataMigration.Status.Phase)
				if phase == "" {
					phase = "PENDING"
				}
				
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
					dataMigration.Name,
					dataMigration.Spec.Database,
					phase,
					dataMigration.Status.MigrationName,
					age.String())
			}

			w.Flush()

			return nil
		},
	}

	cmd.Flags().String("namespace", "", "namespace to search for data migrations")
	cmd.Flags().StringSlice("all-namespaces", []string{}, "search all namespaces for data migrations")

	return cmd
}

func getNamespacesOrDefault(namespace string, allNamespaces []string) ([]string, error) {
	if namespace != "" {
		return []string{namespace}, nil
	}

	if len(allNamespaces) > 0 {
		if allNamespaces[0] == "*" {
			cfg, err := config.GetRESTConfig()
			if err != nil {
				return nil, err
			}
			clientset, err := kubernetes.NewForConfig(cfg)
			if err != nil {
				return nil, err
			}

			namespaces, err := clientset.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
			if err != nil {
				return nil, err
			}

			namespaceNames := []string{}
			for _, namespace := range namespaces.Items {
				namespaceNames = append(namespaceNames, namespace.Name)
			}
			return namespaceNames, nil
		}
		return allNamespaces, nil
	}

	// Default to current namespace
	return []string{"default"}, nil
}