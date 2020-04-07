package migration

import (
	"context"

	"github.com/pkg/errors"
	databasesv1alpha3 "github.com/schemahero/schemahero/pkg/apis/databases/v1alpha3"
	schemasv1alpha3 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha3"
	databasesclientv1alpha3 "github.com/schemahero/schemahero/pkg/client/schemaheroclientset/typed/databases/v1alpha3"
	schemasclientv1alpha3 "github.com/schemahero/schemahero/pkg/client/schemaheroclientset/typed/schemas/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func TableFromMigration(ctx context.Context, migration *schemasv1alpha3.Migration) (*schemasv1alpha3.Table, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get config")
	}

	schemasClient, err := schemasclientv1alpha3.NewForConfig(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create schemas client")
	}

	table, err := schemasClient.Tables(migration.Spec.TableNamespace).Get(ctx, migration.Spec.TableName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get table")
	}

	return table, nil
}

func DatabaseFromTable(ctx context.Context, table *schemasv1alpha3.Table) (*databasesv1alpha3.Database, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get config")
	}

	databasesClient, err := databasesclientv1alpha3.NewForConfig(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create databases client")
	}

	database, err := databasesClient.Databases(table.Namespace).Get(ctx, table.Spec.Database, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get database")
	}

	return database, nil
}
