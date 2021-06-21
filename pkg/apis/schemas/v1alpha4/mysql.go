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

type MysqlTableColumnConstraints struct {
	NotNull *bool `json:"notNull,omitempty" yaml:"notNull,omitempty"`
}

type MysqlTableColumnAttributes struct {
	AutoIncrement *bool `json:"autoIncrement,omitempty" yaml:"autoIncrement,omitempty"`
}

type MysqlTableForeignKeyReferences struct {
	Table   string   `json:"table"`
	Columns []string `json:"columns"`
}

type MysqlTableForeignKey struct {
	Columns    []string                       `json:"columns" yaml:"columns"`
	References MysqlTableForeignKeyReferences `json:"references" yaml:"references"`
	OnDelete   string                         `json:"onDelete,omitempty" yaml:"onDelete,omitempty"`
	Name       string                         `json:"name,omitempty" yaml:"name,omitempty"`
}

type MysqlTableIndex struct {
	Columns  []string `json:"columns" yaml:"columns"`
	Name     string   `json:"name,omitempty" yaml:"name,omitempty"`
	IsUnique bool     `json:"isUnique,omitempty" yaml:"isUnique,omitempty"`
	Type     string   `json:"type,omitempty" yaml:"type,omitempty"`
}

type MysqlTableColumn struct {
	Name        string                       `json:"name" yaml:"name"`
	Type        string                       `json:"type" yaml:"type"`
	Constraints *MysqlTableColumnConstraints `json:"constraints,omitempty" yaml:"constraints,omitempty"`
	Attributes  *MysqlTableColumnAttributes  `json:"attributes,omitempty" yaml:"attributes,omitempty"`
	Default     *string                      `json:"default,omitempty" yaml:"default,omitempty"`
	Charset     string                       `json:"charset,omitempty" yaml:"charset,omitempty"`
	Collation   string                       `json:"collation,omitempty" yaml:"collation,omitempty"`
}

type MysqlTableSchema struct {
	PrimaryKey     []string                `json:"primaryKey,omitempty" yaml:"primaryKey,omitempty"`
	ForeignKeys    []*MysqlTableForeignKey `json:"foreignKeys,omitempty" yaml:"foreignKeys,omitempty"`
	Indexes        []*MysqlTableIndex      `json:"indexes,omitempty" yaml:"indexes,omitempty"`
	Columns        []*MysqlTableColumn     `json:"columns,omitempty" yaml:"columns,omitempty"`
	IsDeleted      bool                    `json:"isDeleted,omitempty" yaml:"isDeleted,omitempty"`
	DefaultCharset string                  `json:"defaultCharset,omitempty" yaml:"defaultCharset,omitempty"`
	Collation      string                  `json:"collation,omitempty" yaml:"collation,omitempty"`
}
