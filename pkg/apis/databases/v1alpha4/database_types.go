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

// DatabaseConnection defines connection parameters for the database driver
type DatabaseConnection struct {
	Postgres    *PostgresConnection    `json:"postgres,omitempty"`
	Mysql       *MysqlConnection       `json:"mysql,omitempty"`
	CockroachDB *CockroachDBConnection `json:"cockroachdb,omitempty"`
	Cassandra   *CassandraConnection   `json:"cassandra,omitempty"`
	SQLite      *SqliteConnection      `json:"sqlite,omitempty"`
	RQLite      *RqliteConnection      `json:"rqlite,omitempty"`
	TimescaleDB *PostgresConnection    `json:"timescaledb,omitempty"`
}

type Toleration struct {
	Key string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
	Operator string `json:"operator,omitempty"`
	Effect string `json:"effect,omitempty"`
}

type SchemaHero struct {
	Image        string            `json:"image,omitempty"`
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	Tolerations []Toleration `json:"tolerations,omitempty"`
}

type DatabaseSpec struct {
	Connection         DatabaseConnection `json:"connection,omitempty"`
	EnableShellCommand bool               `json:"enableShellCommand,omitempty"`

	//+kubebuilder:default:=false
	ImmediateDeploy bool              `json:"immediateDeploy,omitempty"`
	DeploySeedData  bool              `json:"deploySeedData,omitempty"` // TODO remove this for envs in 0.13.0
	SchemaHero      *SchemaHero       `json:"schemahero,omitempty"`
	Template        *DatabaseTemplate `json:"template,omitempty"`
}

type DatabaseTemplate struct {
	metav1.ObjectMeta `json:"metadata,omitempty"`
}

// DatabaseStatus defines the observed state of Database
type DatabaseStatus struct {
	IsConnected bool   `json:"isConnected"`
	LastPing    string `json:"lastPing"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Database is the Schema for the databases API
// +kubebuilder:printcolumn:name="Namespace",type=string,JSONPath=`.metadata.namespace`,priority=1
// +kubebuilder:printcolumn:name="Deploy Immediately",type=boolean,JSONPath=`.spec.immediateDeploy`,priority=1
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
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
