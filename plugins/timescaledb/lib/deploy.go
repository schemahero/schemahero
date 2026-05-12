package timescaledb

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	postgres "github.com/schemahero/schemahero/plugins/postgres/lib"
	"github.com/schemahero/schemahero/pkg/database/types"
)

// currentRefreshPolicy represents the live state of a continuous aggregate refresh policy in the database.
type currentRefreshPolicy struct {
	scheduleInterval         string
	startOffset              string
	endOffset                string
	timezone                 string
	bucketsPerBatch          string
	maxBatchesPerExecution   string
	refreshNewestFirst       string
	includeTieredData        string
}

func getContinuousAggregatePolicy(p *postgres.PostgresConnection, viewName string) (*currentRefreshPolicy, error) {
	query := `select
		j.schedule_interval::text,
		j.config->>'start_offset',
		j.config->>'end_offset',
		j.timezone,
		j.config->>'buckets_per_batch',
		j.config->>'max_batches_per_execution',
		j.config->>'refresh_newest_first',
		j.config->>'include_tiered_data'
	from _timescaledb_catalog.continuous_agg c
	join _timescaledb_catalog.bgw_job j on j.hypertable_id = c.mat_hypertable_id
	where c.user_view_name = $1`

	row := p.GetConnection().QueryRow(context.Background(), query, viewName)

	var policy currentRefreshPolicy
	var scheduleInterval, startOffset, endOffset, timezone, bucketsPerBatch, maxBatchesPerExecution, refreshNewestFirst, includeTieredData *string

	if err := row.Scan(&scheduleInterval, &startOffset, &endOffset, &timezone, &bucketsPerBatch, &maxBatchesPerExecution, &refreshNewestFirst, &includeTieredData); err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil
		}
		return nil, errors.Wrap(err, "failed to scan continuous aggregate policy")
	}

	if scheduleInterval != nil {
		policy.scheduleInterval = *scheduleInterval
	}
	if startOffset != nil {
		policy.startOffset = *startOffset
	}
	if endOffset != nil {
		policy.endOffset = *endOffset
	}
	if timezone != nil {
		policy.timezone = *timezone
	}
	if bucketsPerBatch != nil {
		policy.bucketsPerBatch = *bucketsPerBatch
	}
	if maxBatchesPerExecution != nil {
		policy.maxBatchesPerExecution = *maxBatchesPerExecution
	}
	if refreshNewestFirst != nil {
		policy.refreshNewestFirst = *refreshNewestFirst
	}
	if includeTieredData != nil {
		policy.includeTieredData = *includeTieredData
	}

	return &policy, nil
}

func policyDiffers(desired *schemasv1alpha4.TimescaleDBViewRefreshPolicy, current *currentRefreshPolicy) bool {
	if desired == nil && current == nil {
		return false
	}
	if desired == nil || current == nil {
		return true
	}

	if desired.StartOffset != current.startOffset {
		return true
	}
	if desired.EndOffset != current.endOffset {
		return true
	}
	// NOTE: schedule_interval may be canonicalized by PostgreSQL (e.g. "1 hour" -> "01:00:00").
	// For simplicity we do exact string comparison. Users should match the canonical form.
	if desired.ScheduleInterval != current.scheduleInterval {
		return true
	}
	if desired.Timezone != current.timezone {
		return true
	}
	if boolPtrString(desired.IncludeTieredData) != current.includeTieredData {
		return true
	}
	if intPtrString(desired.BucketsPerBatch) != current.bucketsPerBatch {
		return true
	}
	if intPtrString(desired.MaxBatchesPerExecution) != current.maxBatchesPerExecution {
		return true
	}
	if boolPtrString(desired.RefreshNewestFirst) != current.refreshNewestFirst {
		return true
	}

	return false
}

func boolPtrString(v *bool) string {
	if v == nil {
		return ""
	}
	if *v {
		return "true"
	}
	return "false"
}

func intPtrString(v *int) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%d", *v)
}

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
	} else if viewExists > 0 && !viewSchema.IsDeleted {
		// TODO: Alter view.  Some properties can be altered, but to alter the SQL, a new view should be created.
		// For now, handle refresh policy drift for continuous aggregates.
		if isContinuousAggregate {
			return planPolicyChanges(p, viewName, viewSchema)
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

func planPolicyChanges(p *postgres.PostgresConnection, viewName string, viewSchema *schemasv1alpha4.TimescaleDBViewSchema) ([]string, error) {
	statements := []string{}

	currentPolicy, err := getContinuousAggregatePolicy(p, viewName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get current refresh policy")
	}

	if viewSchema.RefreshPolicy == nil {
		if currentPolicy != nil {
			// policy exists but is no longer desired
			statements = append(statements,
				fmt.Sprintf("select remove_continuous_aggregate_policy(%s)", strings.ReplaceAll(pgx.Identifier{viewName}.Sanitize(), "\"", "'")),
			)
		}
		return statements, nil
	}

	// desired policy exists
	if policyDiffers(viewSchema.RefreshPolicy, currentPolicy) {
		if currentPolicy != nil {
			statements = append(statements,
				fmt.Sprintf("select remove_continuous_aggregate_policy(%s)", strings.ReplaceAll(pgx.Identifier{viewName}.Sanitize(), "\"", "'")),
			)
		}

		createStmts, err := CreateViewStatements(viewName, viewSchema)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create policy statement")
		}
		// CreateViewStatements returns both CREATE MATERIALIZED VIEW and the policy.
		// Filter out the CREATE statement when the view already exists.
		for _, stmt := range createStmts {
			if !strings.Contains(strings.ToLower(stmt), "create materialized view") {
				statements = append(statements, stmt)
			}
		}
	}

	return statements, nil
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
	indexStatements, err := BuildIndexStatements(p, tableName, tableSchema)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build index statements")
	}
	statements = append(statements, indexStatements...)

	// hypertable changes
	hypertableStatements, err := BuildHypertableStatements(p, tableName, tableSchema)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build hypertable statements")
	}
	statements = append(statements, hypertableStatements...)

	// seed data
	statements = append(statements, seedDataStatements...)

	return statements, nil
}

// This is slightly different than the postgres version because we need to handle indices created for hypertables
func BuildIndexStatements(p *postgres.PostgresConnection, tableName string, tableSchema *schemasv1alpha4.TimescaleDBTableSchema) ([]string, error) {
	postgresTableSchema := toPostgresTableSchema(tableSchema)

	indexStatements := []string{}
	droppedIndexes := []string{}
	currentIndexes, err := p.ListTableIndexes(p.DatabaseName(), tableName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list table indexes")
	}
	currentConstraints, err := p.ListTableConstraints(p.DatabaseName(), tableName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list table constraints")
	}

DesiredIndexLoop:
	for _, index := range postgresTableSchema.Indexes {
		if index.Name == "" {
			index.Name = types.GeneratePostgresqlIndexName(tableName, index)
		}

		var statement string
		var matchedIndex *types.Index
		for _, currentIndex := range currentIndexes {
			if currentIndex.Equals(types.PostgresqlSchemaIndexToIndex(index)) {
				continue DesiredIndexLoop
			}

			if currentIndex.Name == index.Name {
				matchedIndex = currentIndex
			}
		}

		// drop and readd? pg supports a little bit of alter index we should support (rename)
		if matchedIndex != nil {
			isConstraint := false
			for _, currentConstraint := range currentConstraints {
				if matchedIndex.Name == currentConstraint {
					isConstraint = true
				}
			}

			if isConstraint {
				statement = postgres.RemoveConstraintStatement(tableName, matchedIndex)
			} else {
				statement = postgres.RemoveIndexStatement(tableName, matchedIndex)
			}
			droppedIndexes = append(droppedIndexes, matchedIndex.Name)
			indexStatements = append(indexStatements, statement)
		}

		statement = postgres.AddIndexStatement(tableName, index)
		indexStatements = append(indexStatements, statement)
	}

ExistingIndexLoop:
	for _, currentIndex := range currentIndexes {
		var statement string
		isConstraint := false

		for _, index := range postgresTableSchema.Indexes {
			if currentIndex.Equals(types.PostgresqlSchemaIndexToIndex(index)) {
				continue ExistingIndexLoop
			}
		}

		for _, droppedIndex := range droppedIndexes {
			if droppedIndex == currentIndex.Name {
				continue ExistingIndexLoop
			}
		}

		// This check does not exist in the postgres version
		if tableSchema.Hypertable != nil && tableSchema.Hypertable.TimeColumnName != nil {
			if len(currentIndex.Columns) == 1 && currentIndex.Columns[0] == *tableSchema.Hypertable.TimeColumnName {
				continue ExistingIndexLoop
			}
		}

		for _, currentConstraint := range currentConstraints {
			if currentIndex.Name == currentConstraint {
				isConstraint = true
			}
		}

		if isConstraint {
			statement = postgres.RemoveConstraintStatement(tableName, currentIndex)
		} else {
			statement = postgres.RemoveIndexStatement(tableName, currentIndex)
		}

		indexStatements = append(indexStatements, statement)
	}

	return indexStatements, nil
}
