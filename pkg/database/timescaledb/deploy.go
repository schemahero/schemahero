package timescaledb

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/database/postgres"
)

func PlanTimescaleDBView(uri string, viewName string, viewSchema *schemasv1alpha4.TimescaleDBViewSchema) ([]string, error) {
	p, err := postgres.Connect(uri)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to timescaledb")
	}
	defer p.Close()

	// determine if the view exists
	query := `select count(1) from information_schema.tables where table_name = $1 and table_type = 'VIEW'`
	row := p.GetConnection().QueryRow(context.Background(), query, viewName)
	viewExists := 0
	if err := row.Scan(&viewExists); err != nil {
		return nil, errors.Wrap(err, "failed to scan")
	}

	// if the view exists, we need to check if it's a continuous aggregate view
	isContinuousAggregate := false
	if viewExists > 0 {
		query = `select count(1) from _timescaledb_catalog.continuous_agg where user_view_name = $1`
		row = p.GetConnection().QueryRow(context.Background(), query, viewName)
		continuousAggregateCount := 0
		if err := row.Scan(&continuousAggregateCount); err != nil {
			return nil, errors.Wrap(err, "failed to scan")
		}

		if continuousAggregateCount > 0 {
			isContinuousAggregate = true
		}
	}

	if viewExists == 0 && viewSchema.IsDeleted {
		return []string{}, nil
	} else if viewExists > 0 && viewSchema.IsDeleted {
		if !isContinuousAggregate {
			// regular views are dropped with "drop view"
			return []string{
				fmt.Sprintf(`drop view %s`, pgx.Identifier{viewName}.Sanitize()),
			}, nil
		} else {
			// continuous aggregate views are dropped with "drop materialized view"
			return []string{
				fmt.Sprintf(`drop materialized view %s`, pgx.Identifier{viewName}.Sanitize()),
			}, nil
		}
	}

	// if the view doesn't exist, shortcut to create
	if viewExists == 0 {
		queries, err := CreateViewStatements(viewName, viewSchema)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create view statement")
		}

		return queries, nil
	}

	return []string{}, nil
}

func PlanTimescaleDBTable(uri string, tableName string, tableSchema *schemasv1alpha4.TimescaleDBTableSchema, seedData *schemasv1alpha4.SeedData) ([]string, error) {
	p, err := postgres.Connect(uri)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to timescaledb")
	}
	defer p.Close()

	// determine if the table exists
	query := `select count(1) from information_schema.tables where table_name = $1 and table_type = 'BASE TABLE'`
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
