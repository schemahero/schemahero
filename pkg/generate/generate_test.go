package generate

import (
	"testing"

	"github.com/schemahero/schemahero/pkg/database/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	trueValue = true
)

func Test_sanitizeName(t *testing.T) {
	sanitizeNameTests := map[string]string{
		"two_words": "two-words",
	}

	for unsanitized, expectedSanitized := range sanitizeNameTests {
		t.Run(unsanitized, func(t *testing.T) {
			sanitized := sanitizeName(unsanitized)
			assert.Equal(t, expectedSanitized, sanitized)
		})
	}
}

func Test_writeTableFile(t *testing.T) {
	tests := []struct {
		name         string
		driver       string
		dbName       string
		table        types.Table
		primaryKey   []string
		foreignKeys  []*types.ForeignKey
		indexes      []*types.Index
		columns      []*types.Column
		expectedYAML string
	}{
		{
			name:   "postgres -- 1 col",
			driver: "postgres",
			dbName: "db",
			table: types.Table{
				Name: "simple",
			},
			primaryKey:  []string{"one"},
			foreignKeys: []*types.ForeignKey{},
			indexes:     []*types.Index{},
			columns: []*types.Column{
				{
					Name:     "id",
					DataType: "integer",
				},
			},
			expectedYAML: `apiVersion: schemas.schemahero.io/v1alpha4
kind: Table
metadata:
  name: simple
spec:
  database: db
  name: simple
  schema:
    postgres:
      primaryKey:
      - one
      columns:
      - name: id
        type: integer
`,
		},
		{
			name:   "postgres -- foreign key",
			driver: "postgres",
			dbName: "db",
			table: types.Table{
				Name: "withfk",
			},
			primaryKey: []string{"pk"},
			foreignKeys: []*types.ForeignKey{
				{
					ChildColumns:  []string{"cc"},
					ParentTable:   "p",
					ParentColumns: []string{"pc"},
					Name:          "fk_pc_cc",
				},
			},
			indexes: []*types.Index{},
			columns: []*types.Column{
				{
					Name:     "pk",
					DataType: "integer",
				},
				{
					Name:     "cc",
					DataType: "integer",
				},
			},
			expectedYAML: `apiVersion: schemas.schemahero.io/v1alpha4
kind: Table
metadata:
  name: withfk
spec:
  database: db
  name: withfk
  schema:
    postgres:
      primaryKey:
      - pk
      foreignKeys:
      - columns:
        - cc
        references:
          table: p
          columns:
          - pc
        name: fk_pc_cc
      columns:
      - name: pk
        type: integer
      - name: cc
        type: integer
`,
		},
		{
			name:   "postgres -- generating with index",
			driver: "postgres",
			dbName: "db",
			table: types.Table{
				Name: "simple",
			},
			primaryKey:  []string{"id"},
			foreignKeys: []*types.ForeignKey{},
			indexes: []*types.Index{
				{
					Columns:  []string{"other"},
					Name:     "idx_simple_other",
					IsUnique: true,
				},
			},
			columns: []*types.Column{
				{
					Name:     "id",
					DataType: "integer",
				},
				{
					Name:     "other",
					DataType: "varchar (255)",
					Constraints: &types.ColumnConstraints{
						NotNull: &trueValue,
					},
				},
			},
			expectedYAML: `apiVersion: schemas.schemahero.io/v1alpha4
kind: Table
metadata:
  name: simple
spec:
  database: db
  name: simple
  schema:
    postgres:
      primaryKey:
      - id
      indexes:
      - columns:
        - other
        name: idx_simple_other
        isUnique: true
      columns:
      - name: id
        type: integer
      - name: other
        type: varchar (255)
        constraints:
          notNull: true
`,
		},
		{
			name:   "mysql -- generating with auto_increment",
			driver: "mysql",
			dbName: "db",
			table: types.Table{
				Name: "simple",
			},
			primaryKey: []string{"id"},
			columns: []*types.Column{
				{
					Name:     "id",
					DataType: "integer",
					Attributes: &types.ColumnAttributes{
						AutoIncrement: &trueValue,
					},
				},
			},
			expectedYAML: `apiVersion: schemas.schemahero.io/v1alpha4
kind: Table
metadata:
  name: simple
spec:
  database: db
  name: simple
  schema:
    mysql:
      primaryKey:
      - id
      columns:
      - name: id
        type: integer
        attributes:
          autoIncrement: true
`,
		},
		{
			name:   "rqlite -- 1 col",
			driver: "rqlite",
			table: types.Table{
				Name: "simple",
			},
			primaryKey:  []string{"one"},
			foreignKeys: []*types.ForeignKey{},
			indexes:     []*types.Index{},
			columns: []*types.Column{
				{
					Name:     "id",
					DataType: "integer",
				},
			},
			expectedYAML: `apiVersion: schemas.schemahero.io/v1alpha4
kind: Table
metadata:
  name: simple
spec:
  database: ""
  name: simple
  schema:
    rqlite:
      primaryKey:
      - one
      columns:
      - name: id
        type: integer
`,
		},
		{
			name:   "rqlite -- foreign key",
			driver: "rqlite",
			table: types.Table{
				Name: "withfk",
			},
			primaryKey: []string{"pk"},
			foreignKeys: []*types.ForeignKey{
				{
					ChildColumns:  []string{"cc"},
					ParentTable:   "p",
					ParentColumns: []string{"pc"},
					Name:          "fk_pc_cc",
				},
			},
			indexes: []*types.Index{},
			columns: []*types.Column{
				{
					Name:     "pk",
					DataType: "integer",
				},
				{
					Name:     "cc",
					DataType: "integer",
				},
			},
			expectedYAML: `apiVersion: schemas.schemahero.io/v1alpha4
kind: Table
metadata:
  name: withfk
spec:
  database: ""
  name: withfk
  schema:
    rqlite:
      primaryKey:
      - pk
      foreignKeys:
      - columns:
        - cc
        references:
          table: p
          columns:
          - pc
        name: fk_pc_cc
      columns:
      - name: pk
        type: integer
      - name: cc
        type: integer
`,
		},
		{
			name:   "rqlite -- generating with index",
			driver: "rqlite",
			table: types.Table{
				Name: "simple",
			},
			primaryKey:  []string{"id"},
			foreignKeys: []*types.ForeignKey{},
			indexes: []*types.Index{
				{
					Columns:  []string{"other"},
					Name:     "idx_simple_other",
					IsUnique: true,
				},
			},
			columns: []*types.Column{
				{
					Name:     "id",
					DataType: "integer",
				},
				{
					Name:     "other",
					DataType: "text",
					Constraints: &types.ColumnConstraints{
						NotNull: &trueValue,
					},
				},
			},
			expectedYAML: `apiVersion: schemas.schemahero.io/v1alpha4
kind: Table
metadata:
  name: simple
spec:
  database: ""
  name: simple
  schema:
    rqlite:
      primaryKey:
      - id
      indexes:
      - columns:
        - other
        name: idx_simple_other
        isUnique: true
      columns:
      - name: id
        type: integer
      - name: other
        type: text
        constraints:
          notNull: true
`,
		},
		{
			name:   "rqlite -- generating with auto_increment",
			driver: "rqlite",
			table: types.Table{
				Name: "simple",
			},
			primaryKey: []string{"id"},
			columns: []*types.Column{
				{
					Name:     "id",
					DataType: "integer",
					Attributes: &types.ColumnAttributes{
						AutoIncrement: &trueValue,
					},
				},
			},
			expectedYAML: `apiVersion: schemas.schemahero.io/v1alpha4
kind: Table
metadata:
  name: simple
spec:
  database: ""
  name: simple
  schema:
    rqlite:
      primaryKey:
      - id
      columns:
      - name: id
        type: integer
        attributes:
          autoIncrement: true
`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := require.New(t)

			y, err := generateTableYAML(test.driver, test.dbName, &test.table, test.primaryKey, test.foreignKeys, test.indexes, test.columns)
			req.NoError(err)
			assert.Equal(t, test.expectedYAML, y)
		})
	}
}
