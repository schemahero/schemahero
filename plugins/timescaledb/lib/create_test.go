package timescaledb

import (
	"errors"
	"testing"

	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	timeColumnName = "time"
	idColumnName   = "id"
	one            = 1
	interval       = "7 days"
	trueVar        = true
	falseVar       = false
	locationHash   = "location_hash"
)

func TestCreateViewStatements(t *testing.T) {
	tests := []struct {
		name      string
		viewName  string
		viewSchema *schemasv1alpha4.TimescaleDBViewSchema
		want      []string
		wantErr   bool
	}{
		{
			name:     "continuous aggregate without refresh policy",
			viewName: "my_view",
			viewSchema: &schemasv1alpha4.TimescaleDBViewSchema{
				IsContinuousAggregate: &trueVar,
				Query:                 "select * from foo",
			},
			want: []string{
				`create materialized view "my_view" with (timescaledb.continuous) as select * from foo with data`,
			},
		},
		{
			name:     "continuous aggregate with refresh policy required fields only",
			viewName: "my_view",
			viewSchema: &schemasv1alpha4.TimescaleDBViewSchema{
				IsContinuousAggregate: &trueVar,
				Query:                 "select * from foo",
				RefreshPolicy: &schemasv1alpha4.TimescaleDBViewRefreshPolicy{
					StartOffset:      "1 day",
					EndOffset:        "1 hour",
					ScheduleInterval: "1 hour",
				},
			},
			want: []string{
				`create materialized view "my_view" with (timescaledb.continuous) as select * from foo with data`,
				`select add_continuous_aggregate_policy('my_view', start_offset => INTERVAL '1 day', end_offset => INTERVAL '1 hour', schedule_interval => INTERVAL '1 hour')`,
			},
		},
		{
			name:     "continuous aggregate with refresh policy all optional fields",
			viewName: "my_view",
			viewSchema: &schemasv1alpha4.TimescaleDBViewSchema{
				IsContinuousAggregate: &trueVar,
				Query:                 "select * from foo",
				RefreshPolicy: &schemasv1alpha4.TimescaleDBViewRefreshPolicy{
					StartOffset:            "1 day",
					EndOffset:              "1 hour",
					ScheduleInterval:       "1 hour",
					IfNotExists:            &trueVar,
					InitialStart:           "2024-01-01 00:00:00+00",
					Timezone:               "UTC",
					IncludeTieredData:      &trueVar,
					BucketsPerBatch:        &one,
					MaxBatchesPerExecution: &one,
					RefreshNewestFirst:     &falseVar,
				},
			},
			want: []string{
				`create materialized view "my_view" with (timescaledb.continuous) as select * from foo with data`,
				`select add_continuous_aggregate_policy('my_view', start_offset => INTERVAL '1 day', end_offset => INTERVAL '1 hour', schedule_interval => INTERVAL '1 hour', if_not_exists => true, initial_start => '2024-01-01 00:00:00+00', timezone => 'UTC', include_tiered_data => true, buckets_per_batch => 1, max_batches_per_execution => 1, refresh_newest_first => false)`,
			},
		},
		{
			name:     "continuous aggregate missing startOffset",
			viewName: "my_view",
			viewSchema: &schemasv1alpha4.TimescaleDBViewSchema{
				IsContinuousAggregate: &trueVar,
				Query:                 "select * from foo",
				RefreshPolicy: &schemasv1alpha4.TimescaleDBViewRefreshPolicy{
					EndOffset:        "1 hour",
					ScheduleInterval: "1 hour",
				},
			},
			wantErr: true,
		},
		{
			name:     "continuous aggregate missing endOffset",
			viewName: "my_view",
			viewSchema: &schemasv1alpha4.TimescaleDBViewSchema{
				IsContinuousAggregate: &trueVar,
				Query:                 "select * from foo",
				RefreshPolicy: &schemasv1alpha4.TimescaleDBViewRefreshPolicy{
					StartOffset:      "1 day",
					ScheduleInterval: "1 hour",
				},
			},
			wantErr: true,
		},
		{
			name:     "continuous aggregate missing scheduleInterval",
			viewName: "my_view",
			viewSchema: &schemasv1alpha4.TimescaleDBViewSchema{
				IsContinuousAggregate: &trueVar,
				Query:                 "select * from foo",
				RefreshPolicy: &schemasv1alpha4.TimescaleDBViewRefreshPolicy{
					StartOffset: "1 day",
					EndOffset:   "1 hour",
				},
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := require.New(t)

			got, err := CreateViewStatements(test.viewName, test.viewSchema)
			if test.wantErr {
				req.Error(err)
				return
			}
			req.NoError(err)
			assert.Equal(t, test.want, got)
		})
	}
}

func Test_createHypertableStatement(t *testing.T) {
	tests := []struct {
		name       string
		tableName  string
		hypertable *schemasv1alpha4.TimescaleDBHypertable
		columns    []*schemasv1alpha4.PostgresqlTableColumn
		want       string
		wantErr    error
	}{
		{
			name:      "simple hypertable",
			tableName: "table1",
			hypertable: &schemasv1alpha4.TimescaleDBHypertable{
				TimeColumnName: &timeColumnName,
			},
			columns: []*schemasv1alpha4.PostgresqlTableColumn{
				{
					Name: timeColumnName,
					Type: "timestamptz",
				},
			},
			want: `select create_hypertable('table1', 'time')`,
		},
		{
			name:      "hypertable on non existant column",
			tableName: "table1",
			hypertable: &schemasv1alpha4.TimescaleDBHypertable{
				TimeColumnName: &timeColumnName,
			},
			columns: []*schemasv1alpha4.PostgresqlTableColumn{
				{
					Name: idColumnName,
					Type: "text",
				},
			},
			wantErr: errors.New("cannot create hypertable on column time because column not included in schema"),
		},
		{
			name:      "with partitioning column and number partitions",
			tableName: "table1",
			hypertable: &schemasv1alpha4.TimescaleDBHypertable{
				TimeColumnName:     &timeColumnName,
				PartitioningColumn: &idColumnName,
				NumberPartitions:   &one,
			},
			columns: []*schemasv1alpha4.PostgresqlTableColumn{
				{
					Name: idColumnName,
					Type: "text",
				},
				{
					Name: timeColumnName,
					Type: "timestamptz",
				},
			},
			want: `select create_hypertable('table1', 'time', partitioning_column => "id", number_partitions => 1)`,
		},
		{
			name:      "with chunk time interval",
			tableName: "table1",
			hypertable: &schemasv1alpha4.TimescaleDBHypertable{
				TimeColumnName:    &timeColumnName,
				ChunkTimeInterval: &interval,
			},
			columns: []*schemasv1alpha4.PostgresqlTableColumn{
				{
					Name: timeColumnName,
					Type: "timestamptz",
				},
			},
			want: `select create_hypertable('table1', 'time', chunk_time_interval => INTERVAL '7 days')`,
		},
		{
			name:      "create default indexes",
			tableName: "table1",
			hypertable: &schemasv1alpha4.TimescaleDBHypertable{
				TimeColumnName:       &timeColumnName,
				CreateDefaultIndexes: &trueVar,
			},
			columns: []*schemasv1alpha4.PostgresqlTableColumn{
				{
					Name: timeColumnName,
					Type: "timestamptz",
				},
			},
			want: `select create_hypertable('table1', 'time', create_default_indexes => true)`,
		},
		{
			name:      "if not exists",
			tableName: "table1",
			hypertable: &schemasv1alpha4.TimescaleDBHypertable{
				TimeColumnName: &timeColumnName,
				IfNotExists:    &trueVar,
			},
			columns: []*schemasv1alpha4.PostgresqlTableColumn{
				{
					Name: timeColumnName,
					Type: "timestamptz",
				},
			},
			want: `select create_hypertable('table1', 'time', if_not_exists => true)`,
		},
		{
			name:      "partitioning func",
			tableName: "table1",
			hypertable: &schemasv1alpha4.TimescaleDBHypertable{
				TimeColumnName:   &timeColumnName,
				PartitioningFunc: &locationHash,
			},
			columns: []*schemasv1alpha4.PostgresqlTableColumn{
				{
					Name: timeColumnName,
					Type: "timestamptz",
				},
			},
			want: `select create_hypertable('table1', 'time', partitioning_func => 'location_hash')`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := require.New(t)

			got, err := createHypertableStatement(test.tableName, test.hypertable, test.columns)
			if test.wantErr != nil {
				assert.Equal(t, test.wantErr, err)
			} else {
				req.NoError(err)
				assert.Equal(t, test.want, got)
			}
		})
	}
}
