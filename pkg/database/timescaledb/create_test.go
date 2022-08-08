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
	locationHash   = "location_hash"
)

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
			want: `select create_hypertable('table1', 'time', chunk_time_interval => '7 days')`,
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
