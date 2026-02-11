package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestForeignKey_Equals(t *testing.T) {
	tests := []struct {
		name     string
		fk       *ForeignKey
		other    *ForeignKey
		expected bool
	}{
		{
			name: "identical foreign keys",
			fk: &ForeignKey{
				Name:          "assignment_employee_id_fkey",
				ChildColumns:  []string{"employee_id"},
				ParentTable:   "employee",
				ParentColumns: []string{"id"},
				OnDelete:      "CASCADE",
			},
			other: &ForeignKey{
				Name:          "assignment_employee_id_fkey",
				ChildColumns:  []string{"employee_id"},
				ParentTable:   "employee",
				ParentColumns: []string{"id"},
				OnDelete:      "CASCADE",
			},
			expected: true,
		},
		{
			name: "different on delete action",
			fk: &ForeignKey{
				Name:          "fk1",
				ChildColumns:  []string{"col1"},
				ParentTable:   "parent",
				ParentColumns: []string{"id"},
				OnDelete:      "CASCADE",
			},
			other: &ForeignKey{
				Name:          "fk1",
				ChildColumns:  []string{"col1"},
				ParentTable:   "parent",
				ParentColumns: []string{"id"},
				OnDelete:      "SET NULL",
			},
			expected: false,
		},
		{
			// This is the key bug scenario: the DB returns the auto-generated
			// constraint name and uppercase OnDelete, but the spec has no
			// explicit name and lowercase onDelete. These should be considered
			// equal because they represent the same foreign key.
			name: "db has name and uppercase CASCADE, spec has no name and lowercase cascade",
			fk: &ForeignKey{
				Name:          "assignment_employee_id_fkey",
				ChildColumns:  []string{"employee_id"},
				ParentTable:   "employee",
				ParentColumns: []string{"id"},
				OnDelete:      "CASCADE",
			},
			other: &ForeignKey{
				Name:          "",
				ChildColumns:  []string{"employee_id"},
				ParentTable:   "employee",
				ParentColumns: []string{"id"},
				OnDelete:      "cascade",
			},
			expected: true,
		},
		{
			// DB returns "NO ACTION" as the default when no ON DELETE is specified
			name: "db has NO ACTION, spec has empty onDelete",
			fk: &ForeignKey{
				Name:          "assignment_employee_id_fkey",
				ChildColumns:  []string{"employee_id"},
				ParentTable:   "employee",
				ParentColumns: []string{"id"},
				OnDelete:      "NO ACTION",
			},
			other: &ForeignKey{
				Name:          "",
				ChildColumns:  []string{"employee_id"},
				ParentTable:   "employee",
				ParentColumns: []string{"id"},
				OnDelete:      "",
			},
			expected: true,
		},
		{
			name: "db has name, spec has matching auto-generated name pattern",
			fk: &ForeignKey{
				Name:          "assignment_department_id_fkey",
				ChildColumns:  []string{"department_id"},
				ParentTable:   "department",
				ParentColumns: []string{"id"},
				OnDelete:      "CASCADE",
			},
			other: &ForeignKey{
				Name:          "",
				ChildColumns:  []string{"department_id"},
				ParentTable:   "department",
				ParentColumns: []string{"id"},
				OnDelete:      "cascade",
			},
			expected: true,
		},
		{
			name: "different parent table",
			fk: &ForeignKey{
				Name:          "fk1",
				ChildColumns:  []string{"col1"},
				ParentTable:   "table_a",
				ParentColumns: []string{"id"},
				OnDelete:      "CASCADE",
			},
			other: &ForeignKey{
				Name:          "fk1",
				ChildColumns:  []string{"col1"},
				ParentTable:   "table_b",
				ParentColumns: []string{"id"},
				OnDelete:      "CASCADE",
			},
			expected: false,
		},
		{
			name: "different child columns",
			fk: &ForeignKey{
				Name:          "fk1",
				ChildColumns:  []string{"col_a"},
				ParentTable:   "parent",
				ParentColumns: []string{"id"},
			},
			other: &ForeignKey{
				Name:          "fk1",
				ChildColumns:  []string{"col_b"},
				ParentTable:   "parent",
				ParentColumns: []string{"id"},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fk.Equals(tt.other)
			assert.Equal(t, tt.expected, result)
		})
	}
}
