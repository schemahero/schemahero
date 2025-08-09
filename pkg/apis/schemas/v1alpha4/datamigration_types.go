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
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DataMigrationOperation defines different types of data migration operations
type DataMigrationOperation struct {
	// StaticUpdate performs a simple UPDATE with static values
	StaticUpdate *StaticUpdateOperation `json:"staticUpdate,omitempty" yaml:"staticUpdate,omitempty"`
	
	// CalculatedUpdate performs an UPDATE with calculated values from other columns
	CalculatedUpdate *CalculatedUpdateOperation `json:"calculatedUpdate,omitempty" yaml:"calculatedUpdate,omitempty"`
	
	// TransformUpdate performs data transformation operations
	TransformUpdate *TransformUpdateOperation `json:"transformUpdate,omitempty" yaml:"transformUpdate,omitempty"`
	
	// CustomSQL allows arbitrary SQL execution for complex migrations
	CustomSQL *CustomSQLOperation `json:"customSQL,omitempty" yaml:"customSQL,omitempty"`
}

// StaticUpdateOperation represents updating columns with static values
type StaticUpdateOperation struct {
	Table     string            `json:"table" yaml:"table"`
	Set       map[string]string `json:"set" yaml:"set"`
	Where     string            `json:"where,omitempty" yaml:"where,omitempty"`
	Condition string            `json:"condition,omitempty" yaml:"condition,omitempty"`
}

// CalculatedUpdateOperation represents updating columns with calculated values
type CalculatedUpdateOperation struct {
	Table       string                    `json:"table" yaml:"table"`
	Calculations []ColumnCalculation      `json:"calculations" yaml:"calculations"`
	Where       string                    `json:"where,omitempty" yaml:"where,omitempty"`
	Condition   string                    `json:"condition,omitempty" yaml:"condition,omitempty"`
}

// ColumnCalculation defines how to calculate a new column value
type ColumnCalculation struct {
	Column     string `json:"column" yaml:"column"`
	Expression string `json:"expression" yaml:"expression"`
}

// TransformUpdateOperation represents complex data transformations
type TransformUpdateOperation struct {
	Table         string             `json:"table" yaml:"table"`
	Transformations []DataTransformation `json:"transformations" yaml:"transformations"`
	Where         string             `json:"where,omitempty" yaml:"where,omitempty"`
	Condition     string             `json:"condition,omitempty" yaml:"condition,omitempty"`
}

// DataTransformation defines a specific data transformation
type DataTransformation struct {
	Column        string `json:"column" yaml:"column"`
	TransformType string `json:"transformType" yaml:"transformType"` // e.g., "timezone_convert", "type_cast", "format_change"
	FromValue     string `json:"fromValue,omitempty" yaml:"fromValue,omitempty"`
	ToValue       string `json:"toValue,omitempty" yaml:"toValue,omitempty"`
	Parameters    map[string]string `json:"parameters,omitempty" yaml:"parameters,omitempty"`
}

// CustomSQLOperation allows executing arbitrary SQL
type CustomSQLOperation struct {
	SQL       string `json:"sql" yaml:"sql"`
	Validate  bool   `json:"validate,omitempty" yaml:"validate,omitempty"` // Whether to validate SQL syntax
}

// DataMigrationSchema contains database-specific migration configurations
type DataMigrationSchema struct {
	Postgres    []DataMigrationOperation `json:"postgres,omitempty" yaml:"postgres,omitempty"`
	Mysql       []DataMigrationOperation `json:"mysql,omitempty" yaml:"mysql,omitempty"`
	CockroachDB []DataMigrationOperation `json:"cockroachdb,omitempty" yaml:"cockroachdb,omitempty"`
	Cassandra   []DataMigrationOperation `json:"cassandra,omitempty" yaml:"cassandra,omitempty"`
	TimescaleDB []DataMigrationOperation `json:"timescaledb,omitempty" yaml:"timescaledb,omitempty"`
	SQLite      []DataMigrationOperation `json:"sqlite,omitempty" yaml:"sqlite,omitempty"`
	RQLite      []DataMigrationOperation `json:"rqlite,omitempty" yaml:"rqlite,omitempty"`
}

// DataMigrationSpec defines the desired state of DataMigration
type DataMigrationSpec struct {
	Database string               `json:"database" yaml:"database"`
	Name     string               `json:"name" yaml:"name"`
	Requires []string             `json:"requires,omitempty" yaml:"requires,omitempty"`
	
	// ExecutionOrder controls when this migration runs relative to schema changes
	// Values: "before_schema", "after_schema" (default)
	ExecutionOrder string `json:"executionOrder,omitempty" yaml:"executionOrder,omitempty"`
	
	// Idempotent indicates if this migration can be run multiple times safely
	Idempotent bool `json:"idempotent,omitempty" yaml:"idempotent,omitempty"`
	
	Schema *DataMigrationSchema `json:"schema,omitempty" yaml:"schema,omitempty"`
}

// DataMigrationStatus defines the observed state of DataMigration
type DataMigrationStatus struct {
	// SHA of the migration spec from the last time we executed
	LastExecutedMigrationSpecSHA string `json:"lastExecutedMigrationSpecSHA,omitempty" yaml:"lastExecutedMigrationSpecSHA,omitempty"`
	
	// ExecutionTimestamp records when this migration was last executed
	ExecutionTimestamp *metav1.Time `json:"executionTimestamp,omitempty" yaml:"executionTimestamp,omitempty"`
	
	// ExecutionStatus indicates the result of the last execution
	ExecutionStatus string `json:"executionStatus,omitempty" yaml:"executionStatus,omitempty"` // "success", "failed", "pending"
	
	// ErrorMessage contains details if execution failed
	ErrorMessage string `json:"errorMessage,omitempty" yaml:"errorMessage,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DataMigration is the Schema for the data migrations API
// +kubebuilder:printcolumn:name="Namespace",type=string,JSONPath=`.metadata.namespace`,priority=1
// +kubebuilder:printcolumn:name="Migration",type=string,JSONPath=`.spec.name`
// +kubebuilder:printcolumn:name="Database",type=string,JSONPath=`.spec.database`
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.executionStatus`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +k8s:openapi-gen=true
type DataMigration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DataMigrationSpec   `json:"spec,omitempty"`
	Status DataMigrationStatus `json:"status,omitempty"`
}

func (dm DataMigration) GetSHA() (string, error) {
	// ignoring the status, json marshal the spec and the metadata
	o := struct {
		Spec DataMigrationSpec `json:"spec,omitempty"`
	}{
		Spec: dm.Spec,
	}

	b, err := json.Marshal(o)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal")
	}

	sum := sha256.Sum256(b)
	return fmt.Sprintf("%x", sum), nil
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
