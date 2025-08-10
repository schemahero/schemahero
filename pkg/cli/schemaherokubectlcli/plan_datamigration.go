package schemaherokubectlcli

import (
	"context"
	"fmt"
	"strings"

	"github.com/schemahero/schemahero/pkg/client/schemaheroclientset"
	"github.com/schemahero/schemahero/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func PlanDataMigrationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "plan-datamigration [name]",
		Short:         "plan a data migration",
		Long:          `Show the SQL that will be executed for a data migration`,
		Args:          cobra.ExactArgs(1),
		SilenceErrors: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			v := viper.GetViper()

			ctx := context.Background()
			cfg, err := config.GetRESTConfig()
			if err != nil {
				return err
			}

			schemaHeroClient, err := schemaheroclientset.NewForConfig(cfg)
			if err != nil {
				return err
			}

			namespace := v.GetString("namespace")
			if namespace == "" {
				namespace = "default"
			}

			// Get the data migration
			dataMigrationName := args[0]
			dataMigration, err := schemaHeroClient.SchemasV1alpha4().DataMigrations(namespace).Get(ctx, dataMigrationName, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("failed to get data migration %s: %v", dataMigrationName, err)
			}

			// Check if there's an associated migration
			if dataMigration.Status.MigrationName == "" {
				fmt.Printf("Data migration '%s' has not generated a migration yet.\n", dataMigrationName)
				fmt.Println("The controller may still be processing. Try again in a few seconds.")
				return nil
			}

			// Get the associated migration
			migration, err := schemaHeroClient.SchemasV1alpha4().Migrations(namespace).Get(ctx, dataMigration.Status.MigrationName, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("failed to get migration %s: %v", dataMigration.Status.MigrationName, err)
			}

			fmt.Printf("Data Migration: %s\n", dataMigrationName)
			fmt.Printf("Database: %s\n", dataMigration.Spec.Database)
			fmt.Printf("Migration: %s\n", dataMigration.Status.MigrationName)
			fmt.Printf("Status: %s\n\n", migration.Status.Phase)

			fmt.Println("Operations to be performed:")
			fmt.Println("============================")
			
			for i, op := range dataMigration.Spec.Migrations {
				fmt.Printf("\n%d. %s operation:\n", i+1, strings.ToUpper(string(op.Type)))
				switch op.Type {
				case "update":
					fmt.Printf("   Table: %s\n", op.Table)
					fmt.Printf("   Set %s = %s\n", op.Column, op.Value)
					if op.Where != "" {
						fmt.Printf("   Where: %s\n", op.Where)
					}
				case "calculate":
					fmt.Printf("   Table: %s\n", op.Table)
					fmt.Printf("   Set %s = %s\n", op.Column, op.Expression)
					if op.Where != "" {
						fmt.Printf("   Where: %s\n", op.Where)
					}
				case "convert":
					fmt.Printf("   Table: %s\n", op.Table)
					fmt.Printf("   Column: %s\n", op.Column)
					fmt.Printf("   From: %s → To: %s\n", op.From, op.To)
				}
			}

			fmt.Println("\nGenerated SQL:")
			fmt.Println("==============")
			fmt.Println(migration.Spec.GeneratedDDL)

			if migration.Status.Phase == "PLANNED" {
				fmt.Println("\nTo apply this migration, run:")
				fmt.Printf("  kubectl schemahero approve migration %s -n %s\n", migration.Name, namespace)
			}

			return nil
		},
	}

	cmd.Flags().String("namespace", "", "namespace of the data migration")

	return cmd
}