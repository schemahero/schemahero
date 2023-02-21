package migration

import (
	"context"

	"github.com/pkg/errors"
	databasesv1alpha4 "github.com/schemahero/schemahero/pkg/apis/databases/v1alpha4"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	databasesclientv1alpha4 "github.com/schemahero/schemahero/pkg/client/schemaheroclientset/typed/databases/v1alpha4"
	schemasclientv1alpha4 "github.com/schemahero/schemahero/pkg/client/schemaheroclientset/typed/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	schemasClient   schemasclientv1alpha4.SchemasV1alpha4Interface
	databasesClient databasesclientv1alpha4.DatabasesV1alpha4Interface
)

func getSchemasClient() (schemasclientv1alpha4.SchemasV1alpha4Interface, error) {
	if schemasClient != nil {
		return schemasClient, nil
	}

	cfg, err := config.GetRESTConfig()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get config")
	}

	schemasClient, err = schemasclientv1alpha4.NewForConfig(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create schemas client")
	}

	return schemasClient, nil
}

func getDatabasesClient() (databasesclientv1alpha4.DatabasesV1alpha4Interface, error) {
	if databasesClient != nil {
		return databasesClient, nil
	}

	cfg, err := config.GetRESTConfig()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get config")
	}

	databasesClient, err = databasesclientv1alpha4.NewForConfig(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create schemas client")
	}

	return databasesClient, nil
}

func TableFromMigration(ctx context.Context, migration *schemasv1alpha4.Migration) (*schemasv1alpha4.Table, error) {
	schemasClient, err := getSchemasClient()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get schemas client")
	}

	table, err := schemasClient.Tables(migration.Spec.TableNamespace).Get(ctx, migration.Spec.TableName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get table")
	}

	return table, nil
}

func DatabaseFromTable(ctx context.Context, table *schemasv1alpha4.Table) (*databasesv1alpha4.Database, error) {
	databasesClient, err := getDatabasesClient()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get databases client")
	}

	database, err := databasesClient.Databases(table.Namespace).Get(ctx, table.Spec.Database, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get database")
	}

	return database, nil
}

func ViewFromMigration(ctx context.Context, migration *schemasv1alpha4.Migration) (*schemasv1alpha4.View, error) {
	schemasClient, err := getSchemasClient()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get schemas client")
	}

	table, err := schemasClient.Views(migration.Spec.TableNamespace).Get(ctx, migration.Spec.TableName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get view")
	}

	return table, nil
}

func DatabaseFromView(ctx context.Context, view *schemasv1alpha4.View) (*databasesv1alpha4.Database, error) {
	databasesClient, err := getDatabasesClient()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get databases client")
	}

	database, err := databasesClient.Databases(view.Namespace).Get(ctx, view.Spec.Database, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get database")
	}

	return database, nil
}
