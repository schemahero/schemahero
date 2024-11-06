package timescaledb

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/database/postgres"
)

func BuildHypertableStatements(p *postgres.PostgresConnection, tableName string, tableSchema *schemasv1alpha4.TimescaleDBTableSchema) ([]string, error) {
	hypertableStatements := []string{}

	currentHypertable, err := getHypertableForTable(p, tableName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get hypertable for table")
	}

	if currentHypertable == nil && tableSchema.Hypertable == nil {
		return []string{}, nil
	}

	if currentHypertable == nil && tableSchema.Hypertable != nil {
		// create hypertable
		createStmt, err := createHypertableStatement(tableName, tableSchema.Hypertable, tableSchema.Columns)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create hypertable statement")
		}

		hypertableStatements = append(hypertableStatements, createStmt)

		return hypertableStatements, nil
	}

	// else we need to drop the hypertable
	// but that's not possible, we need to migrate the table to a regular table
	// This is not currently supported

	// else we need to modify the hypertable
	// This is not currently supported

	return hypertableStatements, nil
}

func getHypertableForTable(p *postgres.PostgresConnection, tableName string) (*schemasv1alpha4.TimescaleDBHypertable, error) {
	conn := p.GetConnection()

	hypertableQuery := `
		SELECT hypertable_name
		FROM timescaledb_information.hypertables
		WHERE hypertable_name = $1;
	`
	var hypertableName string
	err := conn.QueryRow(context.Background(), hypertableQuery, tableName).Scan(&hypertableName)
	if err == pgx.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrap(err, "failed to query hypertable information")
	}

	timeColumnQuery := `SELECT column_name
FROM timescaledb_information.dimensions
WHERE hypertable_name = $1
AND dimension_type = 'Time'`
	var timeColumnName string
	err = conn.QueryRow(context.Background(), timeColumnQuery, tableName).Scan(&timeColumnName)
	if err == pgx.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrap(err, "failed to query time column name")
	}

	hypertable := &schemasv1alpha4.TimescaleDBHypertable{
		TimeColumnName: &timeColumnName,
	}

	return hypertable, nil
}

func createHypertableStatement(tableName string, hypertable *schemasv1alpha4.TimescaleDBHypertable, columns []*schemasv1alpha4.PostgresqlTableColumn) (string, error) {
	// if there isn't a time column name, abort
	if hypertable.TimeColumnName == nil {
		return "", nil
	}

	// if the time column name is not a column name, abort
	if !columnExists(*hypertable.TimeColumnName, columns) {
		return "", fmt.Errorf("cannot create hypertable on column %s because column not included in schema", *hypertable.TimeColumnName)
	}

	params, err := getHypertableParams(hypertable, columns)
	if err != nil {
		return "", errors.Wrap(err, "get hypertable params")
	}

	serializedParams := strings.Join(params, ", ")

	stmt := fmt.Sprintf(`select create_hypertable(%s, %s`,
		strings.ReplaceAll(pgx.Identifier{tableName}.Sanitize(), "\"", "'"),
		strings.ReplaceAll(pgx.Identifier{*hypertable.TimeColumnName}.Sanitize(), "\"", "'"))

	if len(serializedParams) > 0 {
		stmt = fmt.Sprintf("%s, %s)", stmt, serializedParams)
	} else {
		stmt = fmt.Sprintf("%s)", stmt)
	}

	return stmt, nil
}
