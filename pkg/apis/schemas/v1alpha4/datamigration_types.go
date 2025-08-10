/*
Copyright 2025 The SchemaHero Authors

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

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DataMigrationType defines the type of data migration operation
// +kubebuilder:validation:Enum=update;calculate;convert
type DataMigrationType string

const (
	UpdateMigration    DataMigrationType = "update"
	CalculateMigration DataMigrationType = "calculate"
	ConvertMigration   DataMigrationType = "convert"
)

// DataMigrationOperation defines a single data migration operation
type DataMigrationOperation struct {
	// Type of the migration operation
	Type DataMigrationType `json:"type"`

	// Table to operate on
	Table string `json:"table"`

	// Column to modify (for update and convert operations)
	Column string `json:"column,omitempty"`

	// Value to set (for update operations)
	Value string `json:"value,omitempty"`

	// Where clause (for conditional operations)
	Where string `json:"where,omitempty"`

	// Expression for calculated columns
	Expression string `json:"expression,omitempty"`

	// From type (for convert operations)
	From string `json:"from,omitempty"`

	// To type (for convert operations)
	To string `json:"to,omitempty"`
}

// DataMigrationSpec defines the desired state of DataMigration
type DataMigrationSpec struct {
	// Reference to the database
	Database string `json:"database"`

	// List of migration operations to perform
	Migrations []DataMigrationOperation `json:"migrations"`
}

// DataMigrationStatus defines the observed state of DataMigration
type DataMigrationStatus struct {
	// Phase of the data migration
	Phase Phase `json:"phase,omitempty"`

	// Name of the associated Migration resource
	MigrationName string `json:"migrationName,omitempty"`

	// Timestamp when the plan was generated
	PlannedAt int64 `json:"plannedAt,omitempty"`

	// Timestamp when the migration was executed
	ExecutedAt int64 `json:"executedAt,omitempty"`

	// Error message if the migration failed
	Error string `json:"error,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DataMigration is the Schema for the datamigrations API
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Database",type=string,JSONPath=`.spec.database`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Migration",type=string,JSONPath=`.status.migrationName`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +k8s:openapi-gen=true
type DataMigration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DataMigrationSpec   `json:"spec,omitempty"`
	Status DataMigrationStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DataMigrationList contains a list of DataMigration
type DataMigrationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DataMigration `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DataMigration{}, &DataMigrationList{})
}