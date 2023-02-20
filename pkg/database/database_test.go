package database

import (
	"testing"

	"github.com/schemahero/schemahero/pkg/database/types"
	"github.com/stretchr/testify/assert"
)

func Test_GetStatementsFromDDL(t *testing.T) {
	tests := []struct {
		name           string
		ddl            string
		wantStatements []string
	}{
		{
			name: "one line without terminator",
			ddl: `alter table "table1" alter column "col1" drop default, alter column "col2" drop not null
			`,
			wantStatements: []string{
				`alter table "table1" alter column "col1" drop default, alter column "col2" drop not null`,
			},
		},
		{
			name: "one line with terminator",
			ddl: `alter table "table1" alter column "col1" drop default, alter column "col2" drop not null;
			`,
			wantStatements: []string{
				`alter table "table1" alter column "col1" drop default, alter column "col2" drop not null;`,
			},
		},
		{
			name: "multiple lines mixed",
			ddl: `alter table "table1" alter column "col1" drop default, alter column "col2" drop not null;
			alter table "table2" alter column "col1" drop default, alter column "col2" drop not null;
			alter table "table3" alter column "col1" drop default, alter column "col2" drop not null;
			create materialized view "some_view" with (timescaledb.continuous) as select time_bucket('1 minute'::interval, created_at) as minute_bucket, id, sum(something) as total
			from some_data
			group by minute_bucket, id with data;
			`,
			wantStatements: []string{
				`alter table "table1" alter column "col1" drop default, alter column "col2" drop not null;`,
				`alter table "table2" alter column "col1" drop default, alter column "col2" drop not null;`,
				`alter table "table3" alter column "col1" drop default, alter column "col2" drop not null;`,
				`create materialized view "some_view" with (timescaledb.continuous) as select time_bucket('1 minute'::interval, created_at) as minute_bucket, id, sum(something) as total from some_data group by minute_bucket, id with data;`,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db := &Database{}
			gotStatements := db.GetStatementsFromDDL(test.ddl)

			assert.Equal(t, test.wantStatements, gotStatements)
		})
	}
}

// Test the SortSpecs function
func TestSortSpecs(t *testing.T) {
	tests := []struct {
		name   string
		driver string
		specs  []types.Spec
		want   []types.Spec
	}{
		{
			name:   "sort views and table",
			driver: "postgres",
			specs: []types.Spec{
				{
					SourceFilename: "table1.yaml",
					Spec: []byte(`
apiVersion: schemas.schemahero.io/v1alpha4
kind: Table
metadata:
  name: table1
spec: {}`,
					),
				},
				{
					SourceFilename: "view1.yaml",
					Spec: []byte(`
apiVersion: schemas.schemahero.io/v1alpha4
kind: View
metadata:
  name: view1
spec:{}`,
					),
				},
				{
					SourceFilename: "table2.yaml",
					Spec: []byte(`
apiVersion: schemas.schemahero.io/v1alpha4
kind: Table
metadata:
  name: table2
spec: {}`,
					),
				},
				{
					SourceFilename: "view2.yaml",
					Spec: []byte(`
apiVersion: schemas.schemahero.io/v1alpha4
kind: View
metadata:
  name: view2
spec:{}`,
					),
				},
			},
			want: []types.Spec{
				{
					SourceFilename: "table1.yaml",
					Spec: []byte(`
apiVersion: schemas.schemahero.io/v1alpha4
kind: Table
metadata:
  name: table1
spec: {}`,
					),
				},
				{
					SourceFilename: "table2.yaml",
					Spec: []byte(`
apiVersion: schemas.schemahero.io/v1alpha4
kind: Table
metadata:
  name: table2
spec: {}`,
					),
				},
				{
					SourceFilename: "view1.yaml",
					Spec: []byte(`
apiVersion: schemas.schemahero.io/v1alpha4
kind: View
metadata:
  name: view1
spec:{}`,
					),
				},
				{
					SourceFilename: "view2.yaml",
					Spec: []byte(`
apiVersion: schemas.schemahero.io/v1alpha4
kind: View
metadata:
  name: view2
spec:{}`,
					),
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db := &Database{
				Driver: test.driver,
			}
			db.SortSpecs(test.specs)

			assert.Equal(t, test.want, test.specs)
		})
	}
}
