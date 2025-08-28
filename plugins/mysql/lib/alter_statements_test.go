package mysql

import (
	"testing"

	"github.com/schemahero/schemahero/pkg/database/types"
	"github.com/stretchr/testify/assert"
)

func TestAlterDDL(t *testing.T) {
	default11 := "11"
	tests := []struct {
		name           string
		tableName      string
		existingColumn types.Column
		column         types.Column
		expect         []string
	}{
		{
			name:      "change charset and collation only",
			tableName: "t",
			existingColumn: types.Column{
				Name:      "col",
				DataType:  "datatype",
				Charset:   "charset",
				Collation: "collation",
				Constraints: &types.ColumnConstraints{
					NotNull: &trueValue,
				},
				ColumnDefault: &default11,
			},
			column: types.Column{
				Name:      "col",
				DataType:  "datatype",
				Charset:   "charset_new",
				Collation: "collation_new",
			},
			expect: []string{
				"alter table `t` modify column `col` datatype character set charset_new collate collation_new",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := AlterModifyColumnStatement{
				TableName:      tt.tableName,
				Column:         tt.column,
				ExistingColumn: tt.existingColumn,
			}

			actual := s.DDL()
			assert.Equal(t, tt.expect, actual)
		})
	}
}

func TestAlterAddConstrantStatement_String(t *testing.T) {
	tests := []struct {
		name       string
		tableName  string
		constraint types.KeyConstraint
		want       string
	}{
		{
			name:      "basic",
			tableName: "my_table",
			constraint: types.KeyConstraint{
				Columns: []string{"id", "sequence"},
			},
			want: "alter table `my_table` add constraint `my_table_id_sequence_key` (`id`, `sequence`)",
		},
		{
			name:      "named",
			tableName: "my_table",
			constraint: types.KeyConstraint{
				Name:    "my_table_custome_key_name",
				Columns: []string{"id", "sequence"},
			},
			want: "alter table `my_table` add constraint `my_table_custome_key_name` (`id`, `sequence`)",
		},
		{
			name:      "primary",
			tableName: "my_table",
			constraint: types.KeyConstraint{
				Columns:   []string{"id", "sequence"},
				IsPrimary: true,
			},
			want: "alter table `my_table` add constraint `my_table_pkey` primary key (`id`, `sequence`)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AlterAddConstrantStatement{TableName: tt.tableName, Constraint: tt.constraint}.String()
			if got != tt.want {
				t.Errorf("AddConstrantStatement() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAlterRemoveConstrantStatement_String(t *testing.T) {
	tests := []struct {
		name       string
		tableName  string
		constraint types.KeyConstraint
		want       string
	}{
		{
			name:      "basic",
			tableName: "my_table",
			constraint: types.KeyConstraint{
				Name:    "my_table_id_sequence_key",
				Columns: []string{"id", "sequence"},
			},
			want: "alter table `my_table` drop index `my_table_id_sequence_key`",
		},
		{
			name:      "primary",
			tableName: "my_table",
			constraint: types.KeyConstraint{
				Name:      "my_table_id_sequence_key",
				Columns:   []string{"id", "sequence"},
				IsPrimary: true,
			},
			want: "alter table `my_table` drop primary key",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AlterRemoveConstrantStatement{TableName: tt.tableName, Constraint: tt.constraint}.String()
			if got != tt.want {
				t.Errorf("RemoveConstrantStatement() = %v, want %v", got, tt.want)
			}
		})
	}
}
