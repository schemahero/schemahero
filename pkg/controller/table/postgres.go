package table

import (
	"context"
	"database/sql"

	databasesv1alpha1 "github.com/schemahero/schemahero/pkg/apis/databases/v1alpha1"
	schemasv1alpha1 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha1"

	"github.com/lib/pq"
)

func (r *ReconcileTable) deployPostgres(connection *databasesv1alpha1.PostgresConnection, tableName string, postgresTableSchema *schemasv1alpha1.PostgresTableSchema) error {
	query := `create table if not exists ` + pq.QuoteIdentifier(tableName) + ` (version bigint not null primary key, dirty boolean not null)`

	db, err := sql.Open("postgres", connection.URI.Value)
	if err != nil {
		return err
	}
	defer db.Close()

	if _, err = db.ExecContext(context.Background(), query); err != nil {
		return err
	}

	return nil
}
