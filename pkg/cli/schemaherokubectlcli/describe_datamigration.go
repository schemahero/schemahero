package schemaherokubectlcli

import (
	"context"
	"fmt"
	"time"

	"github.com/schemahero/schemahero/pkg/client/schemaheroclientset"
	"github.com/schemahero/schemahero/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func DescribeDataMigrationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "datamigration [name]",
		Short:         "describe a data migration",
		Long:          `Show detailed information about a data migration`,
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

			dataMigrationName := args[0]
			dataMigration, err := schemaHeroClient.SchemasV1alpha4().DataMigrations(namespace).Get(ctx, dataMigrationName, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("failed to get data migration %s: %v", dataMigrationName, err)
			}

			fmt.Printf("Name:         %s\n", dataMigration.Name)
			fmt.Printf("Namespace:    %s\n", dataMigration.Namespace)
			fmt.Printf("Created:      %s\n", dataMigration.CreationTimestamp.Format(time.RFC3339))
			fmt.Printf("Database:     %s\n", dataMigration.Spec.Database)
			
			fmt.Println("\nStatus:")
			fmt.Printf("  Phase:           %s\n", dataMigration.Status.Phase)
			fmt.Printf("  Migration Name:  %s\n", dataMigration.Status.MigrationName)
			if dataMigration.Status.PlannedAt > 0 {
				fmt.Printf("  Planned At:      %s\n", time.Unix(dataMigration.Status.PlannedAt, 0).Format(time.RFC3339))
			}
			if dataMigration.Status.ExecutedAt > 0 {
				fmt.Printf("  Executed At:     %s\n", time.Unix(dataMigration.Status.ExecutedAt, 0).Format(time.RFC3339))
			}
			if dataMigration.Status.Error != "" {
				fmt.Printf("  Error:           %s\n", dataMigration.Status.Error)
			}

			fmt.Println("\nMigrations:")
			for i, migration := range dataMigration.Spec.Migrations {
				fmt.Printf("  %d. Type: %s\n", i+1, migration.Type)
				fmt.Printf("     Table: %s\n", migration.Table)
				
				switch migration.Type {
				case "update":
					fmt.Printf("     Column: %s\n", migration.Column)
					fmt.Printf("     Value: %s\n", migration.Value)
					if migration.Where != "" {
						fmt.Printf("     Where: %s\n", migration.Where)
					}
				case "calculate":
					fmt.Printf("     Column: %s\n", migration.Column)
					fmt.Printf("     Expression: %s\n", migration.Expression)
					if migration.Where != "" {
						fmt.Printf("     Where: %s\n", migration.Where)
					}
				case "convert":
					fmt.Printf("     Column: %s\n", migration.Column)
					fmt.Printf("     From: %s\n", migration.From)
					fmt.Printf("     To: %s\n", migration.To)
				}
				fmt.Println()
			}

			// If there's an associated migration, show its status too
			if dataMigration.Status.MigrationName != "" {
				migration, err := schemaHeroClient.SchemasV1alpha4().Migrations(namespace).Get(ctx, dataMigration.Status.MigrationName, metav1.GetOptions{})
				if err == nil {
					fmt.Println("Associated Migration:")
					fmt.Printf("  Name: %s\n", migration.Name)
					fmt.Printf("  Phase: %s\n", migration.Status.Phase)
					
					if migration.Status.Phase == "PLANNED" {
						fmt.Println("\nTo apply this data migration, run:")
						fmt.Printf("  kubectl schemahero approve migration %s -n %s\n", migration.Name, namespace)
					} else if migration.Status.Phase == "EXECUTED" {
						fmt.Println("\nThis data migration has been executed successfully.")
					}
				}
			}

			return nil
		},
	}

	cmd.Flags().String("namespace", "", "namespace of the data migration")

	return cmd
}