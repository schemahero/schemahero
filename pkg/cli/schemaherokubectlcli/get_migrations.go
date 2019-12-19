package schemaherokubectlcli

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	schemasv1alpha3 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha3"
	schemasclientv1alpha3 "github.com/schemahero/schemahero/pkg/client/schemaheroclientset/typed/schemas/v1alpha3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func GetMigrationsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrations",
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

			namespaceNames := []string{}

			if viper.GetBool("all-namespaces") {
				namespaces, err := client.CoreV1().Namespaces().List(metav1.ListOptions{})
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

			matchingTables := []schemasv1alpha3.Table{}

			for _, namespaceName := range namespaceNames {
				tables, err := schemasClient.Tables(namespaceName).List(metav1.ListOptions{})
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

			rows := [][]string{}
			for _, table := range matchingTables {
				for _, plan := range table.Status.Plans {
					// TODO should we show these?
					if plan.ExecutedAt > 0 {
						continue
					}
					if plan.RejectedAt > 0 {
						continue
					}
					if plan.ApprovedAt > 0 {
						continue
					}

					rows = append(rows, []string{
						plan.Name,
						table.Spec.Database,
						table.Name,
						timestampToAge(plan.PlannedAt),
						timestampToAge(plan.ExecutedAt),
						timestampToAge(plan.ApprovedAt),
						timestampToAge(plan.RejectedAt),
					})
				}
			}

			if len(rows) == 0 {
				fmt.Println("No resources found.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tDATABASE\tTABLE\tPLANNED\tEXECUTED\tAPPROVED\tREJECTED")

			for _, row := range rows {
				fmt.Fprintln(w, fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s", row[0], row[1], row[2], row[3], row[4], row[5], row[6]))
			}
			w.Flush()

			return nil
		},
	}

	cmd.Flags().StringP("database", "d", "", "database name to filter to results to")
	cmd.Flags().Bool("all-namespaces", false, "If present, list the requested object(s) across all namespaces. Namespace in current context is ignored even if specified with --namespace.")

	// cmd.Flags().StringP("status", "s", "", "status to filter to results to")

	return cmd
}

func timestampToAge(t int64) string {
	if t == 0 {
		return ""
	}

	d := time.Now().Sub(time.Unix(t, 0))
	if d < time.Duration(time.Minute) {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	} else if d < time.Duration(time.Minute*10) {
		return fmt.Sprintf("%dm%ds", int(d.Minutes()), int(d.Seconds())%60)
	} else if d < time.Duration(time.Hour) {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	} else if d < time.Duration(time.Hour*24) {
		return fmt.Sprintf("%dh", int(d.Hours()))
	} else {
		return fmt.Sprintf("%dd%dh", int(d.Hours()/24), int(d.Hours())%24)
	}
}
