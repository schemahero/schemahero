package postgres

import (
	"testing"

	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/database/types"
)

func Test_formatDefault(t *testing.T) {
	tests := []struct {
		name     string
		dataType string
		value    string
		expected string
	}{
		{"number is quoted", "integer", "11", "'11'"},
		{"empty is quoted", "text", "", "''"},
		{"text is quoted", "text", "pending", "'pending'"},
		{"now", "timestamp with time zone", "now()", "now()"},
		{"gen_random_uuid", "uuid", "gen_random_uuid()", "gen_random_uuid()"},
		{"schema qualified function", "uuid", "public.uuid_generate_v4()", "public.uuid_generate_v4()"},
		{"function with arguments", "timestamp", "date_trunc('day', now())", "date_trunc('day', now())"},
		{"current_timestamp", "timestamp", "CURRENT_TIMESTAMP", "CURRENT_TIMESTAMP"},
		{"lowercase current_date", "date", "current_date", "current_date"},
		{"surrounding whitespace is trimmed", "timestamp", " now() ", "now()"},
		{"text containing parentheses is quoted", "text", "hello (world)", "'hello (world)'"},
		{"call-like literal on text is quoted", "text", "random(x)", "'random(x)'"},
		{"call-like literal on varchar is quoted", "character varying (255)", "random(x)", "'random(x)'"},
		{"schema qualified unknown function on text is quoted", "text", "public.fn()", "'public.fn()'"},
		{"known function on text", "text", "gen_random_uuid()", "gen_random_uuid()"},
		{"keyword on varchar", "character varying", "current_user", "current_user"},
		{"nextval on text", "text", "nextval('seq')", "nextval('seq')"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := formatDefault(test.dataType, test.value)
			if actual != test.expected {
				t.Errorf("formatDefault(%q, %q) = %s, expected %s", test.dataType, test.value, actual, test.expected)
			}
		})
	}
}

func Test_callLikeTextDefaultIsQuoted(t *testing.T) {
	value := "random(x)"
	column := &schemasv1alpha4.PostgresqlTableColumn{
		Name:    "label",
		Type:    "varchar(64)",
		Default: &value,
	}

	create, err := columnAsInsert(column)
	if err != nil {
		t.Fatal(err)
	}
	expectedCreate := `"label" character varying (64) default 'random(x)'`
	if create != expectedCreate {
		t.Errorf("create: got %s, expected %s", create, expectedCreate)
	}

	existing := &types.Column{Name: "label", DataType: "character varying (64)"}
	statements, err := AlterColumnStatements("t", []string{}, []*schemasv1alpha4.PostgresqlTableColumn{column}, existing)
	if err != nil {
		t.Fatal(err)
	}
	expectedAlter := `alter table "t" alter column "label" set default 'random(x)'`
	if len(statements) != 1 || statements[0] != expectedAlter {
		t.Errorf("alter: got %v, expected [%s]", statements, expectedAlter)
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
