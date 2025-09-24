package schemaherokubectlcli

import (
	"context"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"
	"time"

	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	schemasclientv1alpha4 "github.com/schemahero/schemahero/pkg/client/schemaheroclientset/typed/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// matchesMigrationStatus determines if a migration matches the status filter using hierarchy
// (rejected > approved > executed > planned)
func matchesMigrationStatus(migration schemasv1alpha4.Migration, status string) bool {
	switch status {
	case "rejected":
		return migration.Status.RejectedAt > 0
	case "approved":
		return migration.Status.ApprovedAt > 0 && migration.Status.RejectedAt == 0
	case "executed":
		return migration.Status.ExecutedAt > 0 && migration.Status.RejectedAt == 0
	case "planned":
		return migration.Status.PlannedAt > 0 && migration.Status.ExecutedAt == 0 && migration.Status.RejectedAt == 0 && migration.Status.ApprovedAt == 0
	default:
		return false
	}
}

// getMostRecentTimestamp returns the most recent timestamp from a migration
func getMostRecentTimestamp(migration schemasv1alpha4.Migration) int64 {
	timestamps := []int64{
		migration.Status.PlannedAt,
		migration.Status.ExecutedAt,
		migration.Status.ApprovedAt,
		migration.Status.RejectedAt,
	}

	var mostRecent int64
	for _, ts := range timestamps {
		if ts > mostRecent {
			mostRecent = ts
		}
	}
	return mostRecent
}

func GetMigrationsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "migrations",
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
			statusFilter := v.GetString("status")

			// Validate status filter
			if statusFilter != "" {
				validStatuses := []string{"planned", "executed", "approved", "rejected"}
				isValid := false
				for _, validStatus := range validStatuses {
					if statusFilter == validStatus {
						isValid = true
						break
					}
				}
				if !isValid {
					return fmt.Errorf("invalid status filter: %s. Valid options are: planned, executed, approved, rejected", statusFilter)
				}
			}

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

			matchingMigrations := []schemasv1alpha4.Migration{}
			for _, namespaceName := range namespaceNames {
				migrations, err := schemasClient.Migrations(namespaceName).List(ctx, metav1.ListOptions{})
				if err != nil {
					return err
				}

				for _, m := range migrations.Items {
					// Check database filter
					if databaseNameFilter != "" && m.Spec.DatabaseName != databaseNameFilter {
						continue
					}

					// Check status filter
					if statusFilter != "" && !matchesMigrationStatus(m, statusFilter) {
						continue
					}

					matchingMigrations = append(matchingMigrations, m)
				}
			}

			if len(matchingMigrations) == 0 {
				fmt.Println("No resources found.")
				return nil
			}

			// Sort migrations by most recent timestamp (descending)
			sort.Slice(matchingMigrations, func(i, j int) bool {
				return getMostRecentTimestamp(matchingMigrations[i]) > getMostRecentTimestamp(matchingMigrations[j])
			})

			rows := [][]string{}
			for _, m := range matchingMigrations {
				rows = append(rows, []string{
					m.Name,
					m.Spec.DatabaseName,
					m.Spec.TableName,
					timestampToAge(m.Status.PlannedAt),
					timestampToAge(m.Status.ExecutedAt),
					timestampToAge(m.Status.ApprovedAt),
					timestampToAge(m.Status.RejectedAt),
				})
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

	cmd.Flags().String("status", "", "status to filter results to (planned, executed, approved, rejected)")

	return cmd
}

func timestampToAge(t int64) string {
	if t == 0 {
		return ""
	}

	d := time.Since(time.Unix(t, 0))
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
