package schemaherokubectlcli

import (
	"errors"
	"fmt"
	"strings"

	schemasv1alpha3 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha3"
	schemasclientv1alpha3 "github.com/schemahero/schemahero/pkg/client/schemaheroclientset/typed/schemas/v1alpha3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func DescribeMigrationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "migration",
		Short:         "",
		Long:          `...`,
		Args:          cobra.MinimumNArgs(1),
		SilenceErrors: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			v := viper.GetViper()
			migrationName := args[0]

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

			matchingTables := []*schemasv1alpha3.Table{}
			matchingPlans := []*schemasv1alpha3.TablePlan{}

			for _, namespaceName := range namespaceNames {
				// TODO this could be rewritten to use a fieldselector and find the table quicker
				tables, err := schemasClient.Tables(namespaceName).List(metav1.ListOptions{})
				if err != nil {
					return err
				}

				for _, table := range tables.Items {
					for _, plan := range table.Status.Plans {
						if strings.HasPrefix(plan.Name, migrationName) {
							matchingTables = append(matchingTables, &table)
							matchingPlans = append(matchingPlans, plan)
						}
					}
				}
			}

			if len(matchingTables) == 0 {
				fmt.Println("No resources found.")
				return nil
			}

			if len(matchingTables) > 1 {
				fmt.Println("Ambiguious migration name. Multiple migrations found with prefix.")
				return nil
			}

			if len(matchingTables) != len(matchingPlans) {
				fmt.Println("Unexpected error, tables and plan counts do not match")
				return errors.New("error")
			}

			b, err := matchingPlans[0].GetOutput(v.GetString("output"))
			if err != nil {
				return err
			}

			fmt.Printf("%s\n", b)
			return nil
		},
	}

	cmd.Flags().Bool("all-namespaces", false, "If present, list the requested object(s) across all namespaces. Namespace in current context is ignored even if specified with --namespace.")
	cmd.Flags().StringP("output", "o", "yaml", "Output format (can be json or yaml")

	return cmd
}
