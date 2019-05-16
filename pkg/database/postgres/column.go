package postgres

import (
	goerrors "errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/lib/pq"

	schemasv1alpha1 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha1"
)

var unparameterizedColumnTypes = []string{
	"bigint",
	"bigserial",
	"boolean",
	"box",
	"bytea",
	"cide",
	"circle",
	"date",
	"double precision",
	"inet",
	"integer",
	"json",
	"jsonb",
	"line",
	"lseg",
	"macaddr",
	"money",
	"path",
	"pg_lsn",
	"point",
	"polygon",
	"real",
	"smallint",
	"smallserial",
	"serial",
	"text",
	"tsquery",
	"tsvector",
	"txid_snapshot",
	"uuid",
	"xml",
}

type ColumnConstraints struct {
	NotNull bool
}

type Column struct {
	Name          string
	DataType      string
	CharMaxLength *int64
	ColumnDefault *string
	Constraints   *ColumnConstraints
}

func maybeParseParameterizedColumnType(requestedType string) (string, int64, error) {
	columnType := ""
	maxLength := int64(0)

	// if strings.HasPrefix(requestedType, "bit varying" {
	// 	columnType = "bit varying"

	// 	r := regexp.MustCompile((`bit varying\s*\((?P<max>\d*)\)`))

	// 	matchGroups := r.FindStringSubmatch(requestedType)
	// 	masStr
	// }
	if strings.HasPrefix(requestedType, "character varying") {
		columnType = "character varying"

		r := regexp.MustCompile(`character varying\s*\((?P<max>\d*)\)`)

		matchGroups := r.FindStringSubmatch(requestedType)
		maxStr := matchGroups[1]
		max, err := strconv.Atoi(maxStr)
		if err != nil {
			return "", int64(0), err
		}
		maxLength = int64(max)
	}

	return columnType, maxLength, nil
}

func isParameterizedColumnType(requestedType string) bool {
	for _, unparameterizedColumnType := range unparameterizedColumnTypes {
		if unparameterizedColumnType == requestedType {
			return false
		}
	}

	return true
}

func unaliasUnparameterizedColumnType(requestedType string) string {
	switch requestedType {
	case "int8":
		return "bigint"
	case "serial8":
		return "bigserial"
	case "bool":
		return "boolean"
	case "float8":
		return "double precision"
	case "int":
		return "integer"
	case "int4":
		return "integer"
	case "float4":
		return "real"
	case "int2":
		return "smallint"
	case "serial2":
		return "smallserial"
	case "serial4":
		return "serial"
	}

	for _, unparameterizedColumnType := range unparameterizedColumnTypes {
		if unparameterizedColumnType == requestedType {
			return requestedType
		}
	}

	return ""
}

func unaliasParameterizedColumnType(requestedType string) string {
	if strings.HasPrefix(requestedType, "varbit") {
		r := regexp.MustCompile(`varbit\s*\((?P<max>\d*)\)`)

		matchGroups := r.FindStringSubmatch(requestedType)
		if len(matchGroups) == 0 {
			return "bit varying"
		}

		return fmt.Sprintf("bit varying (%s)", matchGroups[1])
	}
	if strings.HasPrefix(requestedType, "char") {
		r := regexp.MustCompile(`char\s*\((?P<len>\d*)\)`)

		matchGroups := r.FindStringSubmatch(requestedType)
		if len(matchGroups) == 0 {
			return "character"
		}

		return fmt.Sprintf("character (%s)", matchGroups[1])
	}
	if strings.HasPrefix(requestedType, "varchar") {
		r := regexp.MustCompile(`varchar\s*\(\s*(?P<max>\d*)\s*\)`)

		matchGroups := r.FindStringSubmatch(requestedType)
		if len(matchGroups) == 0 {
			return "character varying"
		}

		return fmt.Sprintf("character varying (%s)", matchGroups[1])
	}
	if strings.HasPrefix(requestedType, "decimal") {
		precisionAndScale := regexp.MustCompile(`decimal\s*\(\s*(?P<precision>\d*),\s*(?P<scale>\d)\s*\)`)
		precisionOnly := regexp.MustCompile(`decimal\s*\(\s*(?P<precision>\d*)\s*\)`)

		precisionAndScaleMatchGroups := precisionAndScale.FindStringSubmatch(requestedType)
		precisionOnlyMatchGroups := precisionOnly.FindStringSubmatch(requestedType)

		if len(precisionAndScaleMatchGroups) == 0 && len(precisionOnlyMatchGroups) == 0 {
			return "numeric"
		}

		if len(precisionAndScaleMatchGroups) > 0 {
			return fmt.Sprintf("numeric (%s, %s)", precisionAndScaleMatchGroups[1], precisionAndScaleMatchGroups[2])
		}

		return fmt.Sprintf("numeric (%s)", precisionOnlyMatchGroups[1])
	}
	if strings.HasPrefix(requestedType, "timetz") {
		r := regexp.MustCompile(`timetz\s*\(\s*(?P<precision>.*)\s*\)`)

		matchGroups := r.FindStringSubmatch(requestedType)
		if len(matchGroups) == 0 {
			return "time with time zone"
		}

		return fmt.Sprintf("time (%s) with time zone", matchGroups[1])
	}
	if strings.HasPrefix(requestedType, "timestamptz") {
		r := regexp.MustCompile(`timestamptz\s*\(\s*(?P<precision>.*)\s*\)`)

		matchGroups := r.FindStringSubmatch(requestedType)
		if len(matchGroups) == 0 {
			return "timestamp with time zone"
		}

		return fmt.Sprintf("timestamp (%s) with time zone", matchGroups[1])
	}
	return ""
}

func schemaColumnToPostgresColumn(schemaColumn *schemasv1alpha1.PostgresTableColumn) (*Column, error) {
	column := &Column{}

	if schemaColumn.Constraints != nil {
		column.Constraints = &ColumnConstraints{
			NotNull: schemaColumn.Constraints.NotNull,
		}
	}

	requestedType := schemaColumn.Type
	unaliasedColumnType := unaliasUnparameterizedColumnType(requestedType)
	if unaliasedColumnType != "" {
		requestedType = unaliasedColumnType
	}

	unaliasedColumnType = unaliasParameterizedColumnType(requestedType)
	if unaliasedColumnType != "" {
		requestedType = unaliasedColumnType
	}

	if !isParameterizedColumnType(requestedType) {
		column.DataType = requestedType
		return column, nil
	}

	columnType, maxLength, err := maybeParseParameterizedColumnType(requestedType)
	if err != nil {
		return nil, err
	}

	if columnType != "" {
		column.DataType = columnType
		column.CharMaxLength = &maxLength

		return column, nil
	}

	return nil, goerrors.New("unknown column type")
}

func postgresColumnAsInsert(column *schemasv1alpha1.PostgresTableColumn) (string, error) {
	// Note, we don't always quote the column type becuase of how pg handles these two statement very differently:

	// 1. create table "users" ("id" "bigint","login" "varchar(255)","name" "varchar(255)")
	// 2. create table "users" ("id" bigint,"login" varchar(255),"name" varchar(255))

	// if the column type is a known (safe) type, pass it unquoted, else pass whatever we received as quoted
	postgresColumn, err := schemaColumnToPostgresColumn(column)
	if err != nil {
		return "", err
	}

	formatted := fmt.Sprintf("%s %s", pq.QuoteIdentifier(column.Name), postgresColumn.DataType)

	if postgresColumn.CharMaxLength != nil {
		formatted = fmt.Sprintf("%s(%d)", formatted, *postgresColumn.CharMaxLength)
	}

	fmt.Printf("%#v\n", postgresColumn)

	if postgresColumn.Constraints != nil {
		if postgresColumn.Constraints.NotNull {
			formatted = fmt.Sprintf("%s not null", formatted)
		} else {
			formatted = fmt.Sprintf("%s null", formatted)
		}
	}

	return formatted, nil
}

func InsertColumnStatement(tableName string, desiredColumn *schemasv1alpha1.PostgresTableColumn) (string, error) {
	columnFields, err := postgresColumnAsInsert(desiredColumn)
	if err != nil {
		return "", err
	}

	statement := fmt.Sprintf(`alter table %s add column %s`, pq.QuoteIdentifier(tableName), columnFields)

	return statement, nil
}
