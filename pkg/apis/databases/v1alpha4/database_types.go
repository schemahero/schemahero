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

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DatabaseConnection defines connection parameters for the database driver
type DatabaseConnection struct {
	Postgres    *PostgresConnection    `json:"postgres,omitempty"`
	Mysql       *MysqlConnection       `json:"mysql,omitempty"`
	CockroachDB *CockroachDBConnection `json:"cockroachdb,omitempty"`
	Cassandra   *CassandraConnection   `json:"cassandra,omitempty"`
	SQLite      *SqliteConnection      `json:"sqlite,omitempty"`
}

type SchemaHero struct {
	Image        string            `json:"image,omitempty"`
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
}

type DatabaseSpec struct {
	Connection         DatabaseConnection `json:"connection,omitempty"`
	EnableShellCommand bool               `json:"enableShellCommand,omitempty"`
	ImmediateDeploy    bool               `json:"immediateDeploy,omitempty"`
	SchemaHero         *SchemaHero        `json:"schemahero,omitempty"`
}

// DatabaseStatus defines the observed state of Database
type DatabaseStatus struct {
	IsConnected bool   `json:"isConnected"`
	LastPing    string `json:"lastPing"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Database is the Schema for the databases API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type Database struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DatabaseSpec   `json:"spec"`
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
