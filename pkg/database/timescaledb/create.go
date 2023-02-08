package timescaledb

import (
	"fmt"
	"strings"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/database/postgres"
)

func CreateViewStatements(viewName string, viewSchema *schemasv1alpha4.TimescaleDBViewSchema) ([]string, error) {
	if viewSchema.IsContinuousAggregate != nil && *viewSchema.IsContinuousAggregate {
		// continuous aggregate views are created with "create materialized view"
		withDataStatement := "with data"
		if viewSchema.WithNoData != nil && *viewSchema.WithNoData {
			withDataStatement = "with no data"
		}

		statements := []string{
			fmt.Sprintf(`create materialized view %s with (timescaledb.continuous) as %s %s`,
				pgx.Identifier{viewName}.Sanitize(),
				viewSchema.Query,
				withDataStatement),
		}

		return statements, nil
	}

	// create the views as a postgres view
	// TODO
	return nil, errors.New("not implemented")
}

func CreateTableStatements(tableName string, tableSchema *schemasv1alpha4.TimescaleDBTableSchema) ([]string, error) {
	postgresSchema := toPostgresTableSchema(tableSchema)
	statements, err := postgres.CreateTableStatements(tableName, postgresSchema)
	if err != nil {
		return nil, errors.Wrap(err, "create postgres table statements")
	}

	// add the timecale extensions
	if tableSchema.Hypertable != nil {
		stmt, err := createHypertableStatement(tableName, tableSchema.Hypertable, tableSchema.Columns)
		if err != nil {
			return nil, errors.Wrap(err, "create hypertable statement")
		}
		if stmt != "" {
			statements = append(statements, stmt)
		}
	}

	// check for compression
	if tableSchema.Hypertable != nil && tableSchema.Hypertable.Compression != nil {
		stmts, err := createCompressionStatements(tableName, tableSchema.Hypertable, tableSchema.Columns)
		if err != nil {
			return nil, errors.Wrap(err, "create compression statements")
		}

		statements = append(statements, stmts...)
	}

	// check for retention
	if tableSchema.Hypertable != nil && tableSchema.Hypertable.Retention != nil {
		stmts, err := createRetentionStatements(tableName, tableSchema.Hypertable, tableSchema.Columns)
		if err != nil {
			return nil, errors.Wrap(err, "create retention statements")
		}

		statements = append(statements, stmts...)
	}

	return statements, nil
}

func createRetentionStatements(tableName string, hypertable *schemasv1alpha4.TimescaleDBHypertable, columns []*schemasv1alpha4.PostgresqlTableColumn) ([]string, error) {
	if hypertable.Retention == nil {
		return nil, nil
	}

	stmts := []string{}

	if hypertable.Retention != nil {
		if hypertable.Retention.Interval != "" {
			if !isValidInterval(hypertable.Retention.Interval) {
				return nil, errors.New("invalid interval")
			}

			stmt := fmt.Sprintf(`select add_retention_policy(%s, interval '%s'`, pgx.Identifier{tableName}.Sanitize(), hypertable.Retention.Interval)
			stmts = append(stmts, stmt)
		}
	}

	return stmts, nil
}

func createCompressionStatements(tableName string, hypertable *schemasv1alpha4.TimescaleDBHypertable, columns []*schemasv1alpha4.PostgresqlTableColumn) ([]string, error) {
	if hypertable.Compression == nil {
		return nil, nil
	}

	stmts := []string{}

	if hypertable.Compression.SegmentBy != nil {
		if !columnExists(*hypertable.Compression.SegmentBy, columns) {
			return nil, errors.New("compression column not found")
		}

		stmt := fmt.Sprintf(`alter table %s set (timescaledb.compress, timescaledb.compress_segmentby = %s)`, pgx.Identifier{tableName}.Sanitize(), pgx.Identifier{*hypertable.Compression.SegmentBy}.Sanitize())
		stmts = append(stmts, stmt)
	}

	if hypertable.Compression.Interval != nil {
		if !isValidInterval(*hypertable.Compression.Interval) {
			return nil, errors.New("invalid interval")
		}

		stmt := fmt.Sprintf(`select add_compression_policy(%s, INTERVAL '%s')`, pgx.Identifier{tableName}.Sanitize(), *hypertable.Compression.Interval)
		stmts = append(stmts, stmt)
	}

	return stmts, nil
}

func toPostgresTableSchema(tableSchema *schemasv1alpha4.TimescaleDBTableSchema) *schemasv1alpha4.PostgresqlTableSchema {
	return &schemasv1alpha4.PostgresqlTableSchema{
		PrimaryKey:  tableSchema.PrimaryKey,
		ForeignKeys: tableSchema.ForeignKeys,
		Indexes:     tableSchema.Indexes,
		Columns:     tableSchema.Columns,
		IsDeleted:   tableSchema.IsDeleted,
		Triggers:    tableSchema.Triggers,
	}
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

func SeedDataStatements(tableName string, tableSchema *schemasv1alpha4.TimescaleDBTableSchema, seedData *schemasv1alpha4.SeedData) ([]string, error) {
	postgresTableSchema := toPostgresTableSchema(tableSchema)
	return postgres.SeedDataStatements(tableName, postgresTableSchema, seedData)
}

func columnExists(name string, columns []*schemasv1alpha4.PostgresqlTableColumn) bool {
	for _, column := range columns {
		if column.Name == name {
			return true
		}
	}

	return false
}

func isValidInterval(val string) bool {
	return true // TODO
}

func getHypertableParams(hypertable *schemasv1alpha4.TimescaleDBHypertable, columns []*schemasv1alpha4.PostgresqlTableColumn) ([]string, error) {
	params := []string{}

	if hypertable.PartitioningColumn != nil {
		// numberPartitions is required here
		if hypertable.NumberPartitions == nil || *hypertable.NumberPartitions == 0 {
			return nil, errors.New("must specify number partitions when specifying additional partitioning columns")
		}

		// make sure the partitioning column is a new column
		if hypertable.TimeColumnName != nil && *hypertable.PartitioningColumn == *hypertable.TimeColumnName {
			return nil, errors.New("additional partitioning columns cannot be the same as the time column")
		}

		// make sure the column exists
		if !columnExists(*hypertable.PartitioningColumn, columns) {
			return nil, errors.New("additional partitioning column not defined in schema")
		}

		params = append(params, fmt.Sprintf("partitioning_column => %s", pgx.Identifier{*hypertable.PartitioningColumn}.Sanitize()))
		params = append(params, fmt.Sprintf("number_partitions => %d", *hypertable.NumberPartitions))
	}

	if hypertable.ChunkTimeInterval != nil {
		if !isValidInterval(*hypertable.ChunkTimeInterval) {
			return nil, errors.New("invalid chunk time interval")
		}

		params = append(params, fmt.Sprintf("chunk_time_interval => '%s'", *hypertable.ChunkTimeInterval))
	}

	if hypertable.CreateDefaultIndexes != nil {
		params = append(params, fmt.Sprintf("create_default_indexes => %t", *hypertable.CreateDefaultIndexes))
	}

	if hypertable.IfNotExists != nil {
		params = append(params, fmt.Sprintf("if_not_exists => %t", *hypertable.IfNotExists))
	}

	if hypertable.PartitioningFunc != nil {
		params = append(params, fmt.Sprintf("partitioning_func => '%s'", *hypertable.PartitioningFunc))
	}

	if hypertable.AssociatedSchemaName != nil {
		params = append(params, fmt.Sprintf("associated_schema_name => '%s'", *hypertable.AssociatedSchemaName))
	}

	if hypertable.AssociatedTablePrefix != nil {
		params = append(params, fmt.Sprintf("associated_table_prefix => '%s'", *hypertable.AssociatedTablePrefix))
	}

	if hypertable.MigrateData != nil {
		params = append(params, fmt.Sprintf("migrate_data => %t", *hypertable.MigrateData))
	}

	if hypertable.TimePartitioningFunc != nil {
		params = append(params, fmt.Sprintf("time_partitioning_func => '%s'", *hypertable.TimePartitioningFunc))
	}

	if hypertable.ReplicationFactor != nil {
		params = append(params, fmt.Sprintf("replication_factor => %d", *hypertable.ReplicationFactor))
	}

	// if hypertable.DataNodes != nil {
	// 	// TODO i don't know the format here
	// 	// params = append(params, fmt.Sprintf("data_nodes => '%s'", *hypertable.DataNodes))
	// }
	return params, nil
}
