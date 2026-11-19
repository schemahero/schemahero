package mysql

import (
	"fmt"
	"regexp"
	"strings"

	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/database/types"
)

func schemaColumnToColumn(schemaColumn *schemasv1alpha4.MysqlTableColumn) (*types.Column, error) {
	column := &types.Column{
		Name:          schemaColumn.Name,
		ColumnDefault: schemaColumn.Default,
		Charset:       schemaColumn.Charset,
		Collation:     schemaColumn.Collation,
	}

	if schemaColumn.Constraints != nil {
		column.Constraints = &types.ColumnConstraints{
			NotNull: schemaColumn.Constraints.NotNull,
		}
	}

	if schemaColumn.Attributes != nil {
		column.Attributes = &types.ColumnAttributes{
			AutoIncrement: schemaColumn.Attributes.AutoIncrement,
			OnUpdate:      schemaColumn.Attributes.OnUpdate,
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

	columnType, err := maybeParseParameterizedColumnType(requestedType)
	if err != nil {
		return nil, err
	}

	if columnType != "" {
		column.DataType = columnType
		return column, nil
	}

	return nil, fmt.Errorf("unknown column type. cannot validate column type %q", schemaColumn.Type)
}

var onUpdateExtraRegexp = regexp.MustCompile(`(?i)\bon update\s+(\S+)`)

// onUpdateFromExtra returns the on update expression reported in the EXTRA column
// of information_schema.columns (for example "DEFAULT_GENERATED on update CURRENT_TIMESTAMP"),
// or nil if the column has no on update clause.
func onUpdateFromExtra(extra string) *string {
	matches := onUpdateExtraRegexp.FindStringSubmatch(extra)
	if matches == nil {
		return nil
	}

	onUpdate := normalizeOnUpdate(matches[1])
	return &onUpdate
}

// normalizeOnUpdate makes on update expressions comparable. MySQL reports
// CURRENT_TIMESTAMP while MariaDB reports current_timestamp().
func normalizeOnUpdate(onUpdate string) string {
	onUpdate = strings.ToUpper(strings.TrimSpace(onUpdate))
	return strings.TrimSuffix(onUpdate, "()")
}

// onUpdateOf returns the normalized on update expression of the column, or an empty string if there is none.
func onUpdateOf(column *types.Column) string {
	if column.Attributes == nil || column.Attributes.OnUpdate == nil {
		return ""
	}
	return normalizeOnUpdate(*column.Attributes.OnUpdate)
}

// isCurrentTimestampDefault returns true when the value is the CURRENT_TIMESTAMP default function,
// which must not be quoted for datetime and timestamp columns
func isCurrentTimestampDefault(dataType string, value string) bool {
	return (dataType == "datetime" || dataType == "timestamp") && value == "CURRENT_TIMESTAMP"
}

func mysqlColumnAsInsert(column *schemasv1alpha4.MysqlTableColumn) (string, error) {
	mysqlColumn, err := schemaColumnToColumn(column)
	if err != nil {
		return "", err
	}

	formatted := fmt.Sprintf("`%s` %s", column.Name, mysqlColumn.DataType)

	if mysqlColumn.Charset != "" {
		formatted = fmt.Sprintf("%s character set %s", formatted, mysqlColumn.Charset)
	}
	if mysqlColumn.Collation != "" {
		formatted = fmt.Sprintf("%s collate %s", formatted, mysqlColumn.Collation)
	}

	if mysqlColumn.Constraints != nil && mysqlColumn.Constraints.NotNull != nil {
		if *mysqlColumn.Constraints.NotNull {
			formatted = fmt.Sprintf("%s not null", formatted)
		} else {
			formatted = fmt.Sprintf("%s null", formatted)
		}
	}

	if mysqlColumn.Attributes != nil && mysqlColumn.Attributes.AutoIncrement != nil && *mysqlColumn.Attributes.AutoIncrement {
		formatted = fmt.Sprintf("%s auto_increment", formatted)
	}

	if mysqlColumn.ColumnDefault != nil {
		if isCurrentTimestampDefault(mysqlColumn.DataType, *mysqlColumn.ColumnDefault) {
			formatted = fmt.Sprintf("%s default %s", formatted, *mysqlColumn.ColumnDefault)
		} else {
			formatted = fmt.Sprintf("%s default '%s'", formatted, *mysqlColumn.ColumnDefault)
		}
	}

	if onUpdate := onUpdateOf(mysqlColumn); onUpdate != "" {
		formatted = fmt.Sprintf("%s on update %s", formatted, onUpdate)
	}

	return formatted, nil
}

func InsertColumnStatement(tableName string, desiredColumn *schemasv1alpha4.MysqlTableColumn) (string, error) {
	columnFields, err := mysqlColumnAsInsert(desiredColumn)
	if err != nil {
		return "", err
	}

	statement := fmt.Sprintf("alter table `%s` add column %s", tableName, columnFields)

	return statement, nil
}
