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

package v1alpha3

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type TableSchema struct {
	Postgres *SQLTableSchema `json:"postgres,omitempty" yaml:"postgres,omitempty"`
	Mysql    *SQLTableSchema `json:"mysql,omitempty" yaml:"mysql,omitempty"`
}

// TableSpec defines the desired state of Table
type TableSpec struct {
	Database string   `json:"database" yaml:"database"`
	Name     string   `json:"name" yaml:"name"`
	Requires []string `json:"requires,omitempty" yaml:"requires,omitempty"`

	Schema *TableSchema `json:"schema" yaml:"schema"`
}

type TablePlan struct {
	Name string `json:"name,omitempty"`
	DDL  string `json:"ddl,omitempty"`

	// PlannedAt is the unix nano timestamp when the plan was generated
	PlannedAt int64 `json:"plannedAt,omitempty"`

	// InvalidatedAt is the unix nano timestamp when this plan was determined to be invalid or outdated
	InvalidatedAt int64 `json:"invalidatedAt,omitempty"`

	ApprovedAt int64 `json:"approvedAt,omitempty"`
	RejectedAt int64 `json:"rejectedAt,omitempty"`

	ExecutedAt int64 `json:"executedAt,omitempty"`
}

func (tp TablePlan) GetOutput(format string) ([]byte, error) {
	output := map[string]interface{}{
		"Name":      tp.Name,
		"PlannedAt": time.Unix(tp.PlannedAt, 0).Format(time.RFC3339),
		"Migration": tp.DDL,
	}

	result := []byte("")
	if format == "json" {
		b, err := json.Marshal(output)
		if err != nil {
			return nil, errors.Wrap(err, "failed to marshal output to json")
		}
		result = b
	} else if format == "yaml" {
		b, err := yaml.Marshal(output)
		if err != nil {
			return nil, errors.Wrap(err, "failed to marshal output to yaml")
		}
		result = b
	}

	return result, nil
}

// TableStatus defines the observed state of Table
type TableStatus struct {
	Plans []*TablePlan `json:"plans,omitempty"`
	// Plan string `json:"plans,omitempty"`
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

func (t Table) GetSHA() (string, error) {
	// ignoring the status, json marshal the spec and the metadata
	o := struct {
		Spec TableSpec `json:"spec,omitempty"`
	}{
		Spec: t.Spec,
	}

	b, err := json.Marshal(o)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal")
	}

	sum := sha256.Sum256(b)
	return fmt.Sprintf("%x", sum)[:7], nil
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
