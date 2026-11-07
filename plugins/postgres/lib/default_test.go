package postgres

import (
	"testing"

	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/database/types"
)

func Test_formatDefault(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{"number is quoted", "11", "'11'"},
		{"empty is quoted", "", "''"},
		{"text is quoted", "pending", "'pending'"},
		{"now", "now()", "now()"},
		{"gen_random_uuid", "gen_random_uuid()", "gen_random_uuid()"},
		{"schema qualified function", "public.uuid_generate_v4()", "public.uuid_generate_v4()"},
		{"function with arguments", "date_trunc('day', now())", "date_trunc('day', now())"},
		{"current_timestamp", "CURRENT_TIMESTAMP", "CURRENT_TIMESTAMP"},
		{"lowercase current_date", "current_date", "current_date"},
		{"surrounding whitespace is trimmed", " now() ", "now()"},
		{"text containing parentheses is quoted", "hello (world)", "'hello (world)'"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := formatDefault(test.value)
			if actual != test.expected {
				t.Errorf("formatDefault(%q) = %s, expected %s", test.value, actual, test.expected)
			}
		})
	}
}

func Test_functionDefaultIsNotQuoted(t *testing.T) {
	now := "now()"
	column := &schemasv1alpha4.PostgresqlTableColumn{
		Name:    "created_at",
		Type:    "timestamptz",
		Default: &now,
	}

	// creating a column
	create, err := columnAsInsert(column)
	if err != nil {
		t.Fatal(err)
	}
	expectedCreate := `"created_at" timestamp with time zone default now()`
	if create != expectedCreate {
		t.Errorf("create: got %s, expected %s", create, expectedCreate)
	}

	// altering an existing column
	existing := &types.Column{Name: "created_at", DataType: "timestamp with time zone"}
	statements, err := AlterColumnStatements("t", []string{}, []*schemasv1alpha4.PostgresqlTableColumn{column}, existing)
	if err != nil {
		t.Fatal(err)
	}
	expectedAlter := `alter table "t" alter column "created_at" set default now()`
	if len(statements) != 1 || statements[0] != expectedAlter {
		t.Errorf("alter: got %v, expected [%s]", statements, expectedAlter)
	}
}
