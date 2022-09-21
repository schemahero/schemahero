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

type RqliteTableColumnConstraints struct {
	NotNull *bool `json:"notNull,omitempty" yaml:"notNull,omitempty"`
}

type RqliteTableColumnAttributes struct {
	AutoIncrement *bool `json:"autoIncrement,omitempty" yaml:"autoIncrement,omitempty"`
}

type RqliteTableForeignKeyReferences struct {
	Table   string   `json:"table"`
	Columns []string `json:"columns"`
}

type RqliteTableForeignKey struct {
	Columns    []string                        `json:"columns" yaml:"columns"`
	References RqliteTableForeignKeyReferences `json:"references" yaml:"references"`
	OnDelete   string                          `json:"onDelete,omitempty" yaml:"onDelete,omitempty"`
	Name       string                          `json:"name,omitempty" yaml:"name,omitempty"`
}

type RqliteTableIndex struct {
	Columns  []string `json:"columns" yaml:"columns"`
	Name     string   `json:"name,omitempty" yaml:"name,omitempty"`
	IsUnique bool     `json:"isUnique,omitempty" yaml:"isUnique,omitempty"`
	Type     string   `json:"type,omitempty" yaml:"type,omitempty"`
}

type RqliteTableColumn struct {
	Name        string                        `json:"name" yaml:"name"`
	Type        string                        `json:"type" yaml:"type"`
	Constraints *RqliteTableColumnConstraints `json:"constraints,omitempty" yaml:"constraints,omitempty"`
	Attributes  *RqliteTableColumnAttributes  `json:"attributes,omitempty" yaml:"attributes,omitempty"`
	Default     *string                       `json:"default,omitempty" yaml:"default,omitempty"`
}

type RqliteTableSchema struct {
	PrimaryKey  []string                 `json:"primaryKey,omitempty" yaml:"primaryKey,omitempty"`
	ForeignKeys []*RqliteTableForeignKey `json:"foreignKeys,omitempty" yaml:"foreignKeys,omitempty"`
	Indexes     []*RqliteTableIndex      `json:"indexes,omitempty" yaml:"indexes,omitempty"`
	Columns     []*RqliteTableColumn     `json:"columns,omitempty" yaml:"columns,omitempty"`
	IsDeleted   bool                     `json:"isDeleted,omitempty" yaml:"isDeleted,omitempty"`
	Strict      bool                     `json:"strict,omitempty" yaml:"strict,omitempty"`
}
