/*
Copyright 2019 The SchemaHero Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha4

type PostgresqlTableTrigger struct {
	Name              string   `json:"name,omitempty" yaml:"name,omitempty"`
	ConstraintTrigger *bool    `json:"constraintTrigger,omitempty" yaml:"constraintTrigger,omitempty"`
	Events            []string `json:"events" yaml:"events"`
	ForEachStatement  *bool    `json:"forEachStatement,omitempty" yaml:"forEachStatement,omitempty"`
	ForEachRow        *bool    `json:"forEachRun,omitempty" yaml:"forEachRow,omitempty"`
	Condition         *string  `json:"condition,omitempty" yaml:"condition,omitempty"`
	//+kubebuilder:validation:Enum=Procedure;Function
	//+kubebuilder:default:=Procedure
	ExecuteType string `json:"executeType" yaml:"executeType"`
	Execute     string `json:"execute" yaml:"execute"`
	// Deprecated: we support multiple execute types from now on.
	// You are encouraged to use ExecuteType and Execute instead.
	ExecuteProcedure string   `json:"executeProcedure,omitempty" yaml:"executeProcedure,omitempty"`
	Arguments        []string `json:"arguments,omitempty" yaml:"arguments,omitempty"`
}

type PostgresqlTableForeignKeyReferences struct {
	Table   string   `json:"table"`
	Columns []string `json:"columns"`
}

type PostgresqlTableForeignKey struct {
	Columns    []string                            `json:"columns" yaml:"columns"`
	References PostgresqlTableForeignKeyReferences `json:"references" yaml:"references"`
	OnDelete   string                              `json:"onDelete,omitempty" yaml:"onDelete,omitempty"`
	Name       string                              `json:"name,omitempty" yaml:"name,omitempty"`
}

type PostgresqlTableIndex struct {
	Columns  []string `json:"columns" yaml:"columns"`
	Name     string   `json:"name,omitempty" yaml:"name,omitempty"`
	IsUnique bool     `json:"isUnique,omitempty" yaml:"isUnique,omitempty"`
	Type     string   `json:"type,omitempty" yaml:"type,omitempty"`
}

type PostgresqlTableColumnConstraints struct {
	NotNull *bool `json:"notNull,omitempty" yaml:"notNull,omitempty"`
}

type PostgresqlTableColumnAttributes struct {
	AutoIncrement *bool `json:"autoIncrement,omitempty" yaml:"autoIncrement,omitempty"`
}

type PostgresqlTableColumn struct {
	Name        string                            `json:"name" yaml:"name"`
	Type        string                            `json:"type" yaml:"type"`
	Constraints *PostgresqlTableColumnConstraints `json:"constraints,omitempty" yaml:"constraints,omitempty"`
	Attributes  *PostgresqlTableColumnAttributes  `json:"attributes,omitempty" yaml:"attributes,omitempty"`
	Default     *string                           `json:"default,omitempty" yaml:"default,omitempty"`
}

type PostgresqlTableSchema struct {
	Schema      string                       `json:"schema,omitempty" yaml:"schema,omitempty"`
	PrimaryKey  []string                     `json:"primaryKey,omitempty" yaml:"primaryKey,omitempty"`
	ForeignKeys []*PostgresqlTableForeignKey `json:"foreignKeys,omitempty" yaml:"foreignKeys,omitempty"`
	Indexes     []*PostgresqlTableIndex      `json:"indexes,omitempty" yaml:"indexes,omitempty"`
	Columns     []*PostgresqlTableColumn     `json:"columns,omitempty" yaml:"columns,omitempty"`
	IsDeleted   bool                         `json:"isDeleted,omitempty" yaml:"isDeleted,omitempty"`
	Triggers    []*PostgresqlTableTrigger    `json:"json:triggers,omitempty" yaml:"triggers,omitempty"`
}

type PostgresqlFunctionSchema struct {
	Schema string `json:"schema,omitempty" yaml:"schema,omitempty"`
	//+kubebuilder:default:=PLpgSQL
	Lang   string  `json:"lang" yaml:"lang"`
	Return *string `json:"return,omitempty" yaml:"return,omitempty"`
	// As represents the function logic. An example looks as follows:
	// ```
	// DECLARE
	//     user_count bigint;
	// BEGIN
	//     SELECT COUNT(*) INTO user_count FROM users;
	//     RETURN user_count;
	// END;
	// ```
	As string `json:"as" yaml:"as"`
}
