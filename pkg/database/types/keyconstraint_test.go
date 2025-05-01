package types

import (
	"testing"
)

func TestKeyConstraintEquals(t *testing.T) {
	tests := []struct {
		name     string
		k1       *KeyConstraint
		k2       *KeyConstraint
		expected bool
	}{
		{
			name: "both nil",
			k1:   nil,
			k2:   nil,
			expected: true,
		},
		{
			name: "first nil",
			k1:   nil,
			k2:   &KeyConstraint{},
			expected: false,
		},
		{
			name: "second nil",
			k1:   &KeyConstraint{},
			k2:   nil,
			expected: false,
		},
		{
			name: "different primary status",
			k1:   &KeyConstraint{IsPrimary: true},
			k2:   &KeyConstraint{IsPrimary: false},
			expected: false,
		},
		{
			name: "different column count",
			k1:   &KeyConstraint{IsPrimary: true, Columns: []string{"col1"}},
			k2:   &KeyConstraint{IsPrimary: true, Columns: []string{"col1", "col2"}},
			expected: false,
		},
		{
			name: "same columns, same order, primary",
			k1:   &KeyConstraint{IsPrimary: true, Columns: []string{"col1", "col2"}},
			k2:   &KeyConstraint{IsPrimary: true, Columns: []string{"col1", "col2"}},
			expected: true,
		},
		{
			name: "same columns, different order, primary",
			k1:   &KeyConstraint{IsPrimary: true, Columns: []string{"col1", "col2"}},
			k2:   &KeyConstraint{IsPrimary: true, Columns: []string{"col2", "col1"}},
			expected: true,
		},
		{
			name: "different columns, primary",
			k1:   &KeyConstraint{IsPrimary: true, Columns: []string{"col1", "col2"}},
			k2:   &KeyConstraint{IsPrimary: true, Columns: []string{"col1", "col3"}},
			expected: false,
		},
		{
			name: "same columns, same order, not primary",
			k1:   &KeyConstraint{IsPrimary: false, Columns: []string{"col1", "col2"}},
			k2:   &KeyConstraint{IsPrimary: false, Columns: []string{"col1", "col2"}},
			expected: true,
		},
		{
			name: "same columns, different order, not primary",
			k1:   &KeyConstraint{IsPrimary: false, Columns: []string{"col1", "col2"}},
			k2:   &KeyConstraint{IsPrimary: false, Columns: []string{"col2", "col1"}},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := test.k1.Equals(test.k2)
			if actual != test.expected {
				t.Errorf("Expected %v, got %v", test.expected, actual)
			}
		})
	}
}
