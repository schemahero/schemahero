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

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DatabaseConnection defines connection parameters for the database driver
type DatabaseConnection struct {
	Postgres *PostgresConnection `json:"postgres,omitempty"`
	Mysql    *MysqlConnection    `json:"mysql,omitempty"`
}

type SchemaHero struct {
	Image        string            `json:"image,omitempty"`
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
}

type GitOps struct {
	URL           string   `json:"url"`
	PollInterval  string   `json:"pollInterval,omitempty"`
	Branch        string   `json:"branch,omitempty"`
	Paths         []string `json:"paths,omitempty"`
	IsPlanEnabled bool     `json:"isPlanEnabled,omitEmpty"`
}

// DatabaseStatus defines the observed state of Database
type DatabaseStatus struct {
	IsConnected      bool              `json:"isConnected"`
	LastPing         string            `json:"lastPing"`
	GitRepoStatus    string            `json:"gitRepoStatus,omitempty"`
	GitopsPlanStatus string            `json:"gitopsPlanStatus,omitempty"`
	GitopsPlanPulls  map[string]string `json:"gitopsPlanPulls,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Database is the Schema for the databases API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
type Database struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	SchemaHero *SchemaHero        `json:"schemahero,omitempty"`
	Connection DatabaseConnection `json:"connection,omitempty"`
	GitOps     *GitOps            `json:"gitops,omitempty"`

	Status DatabaseStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DatabaseList contains a list of Database
type DatabaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Database `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Database{}, &DatabaseList{})
}
