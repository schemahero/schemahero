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

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:validation:Enum=PLANNED;APPROVED;EXECUTED;INVALID;REJECTED
type Phase string

const (
	Planned  Phase = "PLANNED"
	Approved Phase = "APPROVED"
	Executed Phase = "EXECUTED"
	Invalid  Phase = "INVALID"
	Rejected Phase = "REJECTED"
)

// MigrationSpec defines the desired state of Migration
type MigrationSpec struct {
	DatabaseName   string `json:"databaseName,omitempty"`
	TableName      string `json:"tableName"`
	TableNamespace string `json:"tableNamespace"`
	GeneratedDDL   string `json:"generatedDDL,omitempty"`
	EditedDDL      string `json:"editedDDL,omitempty"`
}

// MigrationStatus defines the observed state of Migration
type MigrationStatus struct {
	Phase Phase `json:"phase,omitempty"`

	// PlannedAt is the unix nano timestamp when the plan was generated
	PlannedAt int64 `json:"plannedAt,omitempty"`

	// InvalidatedAt is the unix nano timestamp when this plan was determined to be invalid or outdated
	InvalidatedAt int64 `json:"invalidatedAt,omitempty"`

	ApprovedAt int64 `json:"approvedAt,omitempty"`
	RejectedAt int64 `json:"rejectedAt,omitempty"`
	ExecutedAt int64 `json:"executedAt,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Migration is the Schema for the migrations API
// +kubebuilder:printcolumn:name="Database",type=string,JSONPath=`.spec.databaseName`
// +kubebuilder:printcolumn:name="Table",type=string,JSONPath=`.spec.tableName`
// +kubebuilder:printcolumn:name="Namespace",type=string,JSONPath=`.metadata.namespace`,priority=1
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +k8s:openapi-gen=true
type Migration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MigrationSpec   `json:"spec,omitempty"`
	Status MigrationStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MigrationList contains a list of Migration
type MigrationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Migration `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Migration{}, &MigrationList{})
}
