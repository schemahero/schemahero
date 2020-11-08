/*
Copyright 2019 Replicated, Inc.

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

type SqliteTableColumnConstraints struct {
	NotNull *bool `json:"notNull,omitempty" yaml:"notNull,omitempty"`
}

type SqliteTableColumnAttributes struct {
	AutoIncrement *bool `json:"autoIncrement,omitempty" yaml:"autoIncrement,omitempty"`
}

type SqliteTableForeignKeyReferences struct {
	Table   string   `json:"table"`
	Columns []string `json:"columns"`
}

type SqliteTableForeignKey struct {
	Columns    []string                        `json:"columns" yaml:"columns"`
	References SqliteTableForeignKeyReferences `json:"references" yaml:"references"`
	OnDelete   string                          `json:"onDelete,omitempty" yaml:"onDelete,omitempty"`
	Name       string                          `json:"name,omitempty" yaml:"name,omitempty"`
}

type SqliteTableIndex struct {
	Columns  []string `json:"columns" yaml:"columns"`
	Name     string   `json:"name,omitempty" yaml:"name,omitempty"`
	IsUnique bool     `json:"isUnique,omitempty" yaml:"isUnique,omitempty"`
	Type     string   `json:"type,omitempty" yaml:"type,omitempty"`
}

type SqliteTableColumn struct {
	Name        string                        `json:"name" yaml:"name"`
	Type        string                        `json:"type" yaml:"type"`
	Constraints *SqliteTableColumnConstraints `json:"constraints,omitempty" yaml:"constraints,omitempty"`
	Attributes  *SqliteTableColumnAttributes  `json:"attributes,omitempty" yaml:"attributes,omitempty"`
	Default     *string                       `json:"default,omitempty" yaml:"default,omitempty"`
}

type SqliteTableSchema struct {
	PrimaryKey  []string                 `json:"primaryKey,omitempty" yaml:"primaryKey,omitempty"`
	ForeignKeys []*SqliteTableForeignKey `json:"foreignKeys,omitempty" yaml:"foreignKeys,omitempty"`
	Indexes     []*SqliteTableIndex      `json:"indexes,omitempty" yaml:"indexes,omitempty"`
	Columns     []*SqliteTableColumn     `json:"columns,omitempty" yaml:"columns,omitempty"`
	IsDeleted   bool                     `json:"isDeleted,omitempty" yaml:"isDeleted,omitempty"`
}
