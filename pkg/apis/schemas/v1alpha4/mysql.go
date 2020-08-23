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

type MysqlSQLTableColumn struct {
	Name        string                     `json:"name" yaml:"name"`
	Type        string                     `json:"type" yaml:"type"`
	Constraints *SQLTableColumnConstraints `json:"constraints,omitempty" yaml:"constraints,omitempty"`
	Attributes  *SQLTableColumnAttributes  `json:"attributes,omitempty" yaml:"attributes,omitempty"`
	Default     *string                    `json:"default,omitempty" yaml:"default,omitempty"`
	Charset     string                     `json:"charset,omitempty" yaml:"charset,omitempty"`
	Collation   string                     `json:"collation,omitempty" yaml:"collation,omitempty"`
}

type MysqlSQLTableSchema struct {
	PrimaryKey     []string               `json:"primaryKey,omitempty" yaml:"primaryKey,omitempty"`
	ForeignKeys    []*SQLTableForeignKey  `json:"foreignKeys,omitempty" yaml:"foreignKeys,omitempty"`
	Indexes        []*SQLTableIndex       `json:"indexes,omitempty" yaml:"indexes,omitempty"`
	Columns        []*MysqlSQLTableColumn `json:"columns,omitempty" yaml:"columns,omitempty"`
	IsDeleted      bool                   `json:"isDeleted,omitempty" yaml:"isDeleted,omitempty"`
	DefaultCharset string                 `json:"defaultCharset,omitempty" yaml:"defaultCharset,omitempty"`
	Collation      string                 `json:"collation,omitempty" yaml:"collation,omitempty"`
}
