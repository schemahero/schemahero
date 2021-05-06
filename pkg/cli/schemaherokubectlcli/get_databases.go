package schemaherokubectlcli

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	databasesv1alpha4 "github.com/schemahero/schemahero/pkg/apis/databases/v1alpha4"
	databasesclientv1alpha4 "github.com/schemahero/schemahero/pkg/client/schemaheroclientset/typed/databases/v1alpha4"
	"github.com/schemahero/schemahero/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func GetDatabasesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "databases",
		Short:         "",
		Long:          `...`,
		SilenceErrors: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			cfg, err := config.GetRESTConfig()
			if err != nil {
				return err
			}

			client, err := kubernetes.NewForConfig(cfg)
			if err != nil {
				return err
			}

			databasesClient, err := databasesclientv1alpha4.NewForConfig(cfg)
			if err != nil {
				return err
			}

			namespaces, err := client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
			if err != nil {
				return err
			}

			matchingDatabases := []databasesv1alpha4.Database{}

			for _, namespace := range namespaces.Items {
				databases, err := databasesClient.Databases(namespace.Name).List(ctx, metav1.ListOptions{})
				if err != nil {
					return err
				}

				matchingDatabases = append(matchingDatabases, databases.Items...)
			}

			if len(matchingDatabases) == 0 {
				fmt.Println("No reosurces found.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tNAMESPACE\tPENDING")

			for _, database := range matchingDatabases {
				fmt.Fprintln(w, fmt.Sprintf("%s\t%s\t%s", database.Name, database.Namespace, "0"))
			}

			w.Flush()

			return nil
		},
	}

	return cmd
}
