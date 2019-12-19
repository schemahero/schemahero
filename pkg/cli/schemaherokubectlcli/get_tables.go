package schemaherokubectlcli

import (
	"fmt"
	"os"
	"text/tabwriter"

	schemasv1alpha3 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha3"
	schemasclientv1alpha3 "github.com/schemahero/schemahero/pkg/client/schemaheroclientset/typed/schemas/v1alpha3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func GetTablesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tables",
		Short: "",
		Long:  `...`,
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			v := viper.GetViper()
			databaseNameFilter := v.GetString("database")

			cfg, err := config.GetConfig()
			if err != nil {
				return err
			}

			client, err := kubernetes.NewForConfig(cfg)
			if err != nil {
				return err
			}

			schemasClient, err := schemasclientv1alpha3.NewForConfig(cfg)
			if err != nil {
				return err
			}

			namespaces, err := client.CoreV1().Namespaces().List(metav1.ListOptions{})
			if err != nil {
				return err
			}

			matchingTables := []schemasv1alpha3.Table{}

			for _, namespace := range namespaces.Items {
				tables, err := schemasClient.Tables(namespace.Name).List(metav1.ListOptions{})
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
				fmt.Println("No reosurces found.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tDATABASE\tSTATUS")

			for _, table := range matchingTables {
				status := "Current"

				for _, plan := range table.Status.Plans {
					if plan.ExecutedAt == 0 && plan.RejectedAt == 0 && plan.ApprovedAt == 0 {
						status = "Pending"
					} else if plan.PlannedAt == 0 || plan.InvalidatedAt == 0 {
						status = "Planning"
					}
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