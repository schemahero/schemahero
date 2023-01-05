package timescaledb

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/database/postgres"
	"github.com/schemahero/schemahero/pkg/trace"
	"go.opentelemetry.io/otel"
	oteltrace "go.opentelemetry.io/otel/trace"
)

func PlanTimescaleDBTable(ctx context.Context, uri string, tableName string, tableSchema *schemasv1alpha4.TimescaleDBTableSchema, seedData *schemasv1alpha4.SeedData) ([]string, error) {
	var span oteltrace.Span
	ctx, span = otel.Tracer(trace.TraceName).Start(ctx, "PlanTimescaleDBTable")
	defer span.End()

	p, err := postgres.Connect(uri)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to timescaledb")
	}
	defer p.Close()

	// determine if the table exists
	query := `select count(1) from information_schema.tables where table_name = $1`
	row := p.GetConnection().QueryRow(context.Background(), query, tableName)
	tableExists := 0
	if err := row.Scan(&tableExists); err != nil {
		return nil, errors.Wrap(err, "failed to scan")
	}

	if tableExists == 0 && tableSchema.IsDeleted {
		return []string{}, nil
	} else if tableExists > 0 && tableSchema.IsDeleted {
		return []string{
			fmt.Sprintf(`drop table %s`, pgx.Identifier{tableName}.Sanitize()),
		}, nil
	}

	postgresTableSchema := toPostgresTableSchema(tableSchema)

	seedDataStatements := []string{}
	if seedData != nil {
		seedDataStatements, err = postgres.SeedDataStatements(tableName, postgresTableSchema, seedData)
		if err != nil {
			return nil, errors.Wrap(err, "create seed data statements")
		}
	}

	if tableExists == 0 {
		// shortcut to just create it
		queries, err := CreateTableStatements(tableName, tableSchema)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create table statement")
		}

		return append(queries, seedDataStatements...), nil
	}

	statements := []string{}

	// table needs to be altered?
	columnStatements, err := postgres.BuildColumnStatements(p, tableName, postgresTableSchema)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build column statement")
	}
	statements = append(statements, columnStatements...)

	// primary key changes
	primaryKeyStatements, err := postgres.BuildPrimaryKeyStatements(p, tableName, postgresTableSchema)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build primary key statements")
	}
	statements = append(statements, primaryKeyStatements...)

	// foreign key changes
	foreignKeyStatements, err := postgres.BuildForeignKeyStatements(p, tableName, postgresTableSchema)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build foreign key statements")
	}
	statements = append(statements, foreignKeyStatements...)

	// index changes
	indexStatements, err := postgres.BuildIndexStatements(p, tableName, postgresTableSchema)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build index statements")
	}
	statements = append(statements, indexStatements...)

	statements = append(statements, seedDataStatements...)

	return statements, nil
}
