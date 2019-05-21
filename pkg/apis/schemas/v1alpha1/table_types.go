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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PostgresTableColumnConstraints struct {
	NotNull   bool   `json:"notNull" yaml:"notNull"`
	MaxLength *int64 `json:"maxLength" yaml:"maxLength"`
}

type PostgresTableColumn struct {
	Name        string                          `json:"name"`
	Type        string                          `json:"type"`
	Constraints *PostgresTableColumnConstraints `json:"constraints,omitempty"`
	Default     string                          `json:"default,omitempty"`
}

type PostgresTableSchema struct {
	PrimaryKey []string               `json:"primaryKey" yaml:"primaryKey"`
	Columns    []*PostgresTableColumn `json:"columns"`
}

type TableSchema struct {
	Postgres *PostgresTableSchema `json:"postgres"`
}

// TableSpec defines the desired state of Table
type TableSpec struct {
	Database string       `json:"database"`
	Name     string       `json:"name"`
	Requires []string     `json:"requires"`
	Schema   *TableSchema `json:"schema"`
}

// TableStatus defines the observed state of Table
type TableStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Table is the Schema for the tables API
// +k8s:openapi-gen=true
type Table struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TableSpec   `json:"spec,omitempty"`
	Status TableStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TableList contains a list of Table
type TableList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Table `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Table{}, &TableList{})
}
