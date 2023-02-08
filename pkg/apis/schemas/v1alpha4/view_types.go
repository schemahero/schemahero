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

type NotImplementedViewSchema struct {
}

type ViewSchema struct {
	Postgres    *NotImplementedViewSchema `json:"postgres,omitempty" yaml:"postgres,omitempty"`
	Mysql       *NotImplementedViewSchema `json:"mysql,omitempty" yaml:"mysql,omitempty"`
	CockroachDB *NotImplementedViewSchema `json:"cockroachdb,omitempty" yaml:"cockroachdb,omitempty"`
	RQLite      *NotImplementedViewSchema `json:"rqlite,omitempty" yaml:"rqlite,omitempty"`
	SQLite      *NotImplementedViewSchema `json:"sqlite,omitempty" yaml:"sqlite,omitempty"`
	TimescaleDB *TimescaleDBViewSchema    `json:"timescaledb,omitempty" yaml:"timescaledb,omitempty"`
	Cassandra   *NotImplementedViewSchema `json:"cassandra,omitempty" yaml:"cassandra,omitempty"`
}

// ViewSpec defines the desired state of View
type ViewSpec struct {
	Database string   `json:"database" yaml:"database"`
	Name     string   `json:"name" yaml:"name"`
	Requires []string `json:"requires,omitempty" yaml:"requires,omitempty"`

	Schema *ViewSchema `json:"schema,omitempty" yaml:"schema,omitempty"`
}

// ViewStatus defines the observed state of View
type ViewStatus struct {
	// We store the SHA of the view spec from the last time we executed a plan to
	// make startup less noisy by skipping re-planning objects that have been planned
	// we cannot use the resourceVersion or generation fields because updating them
	// would cause the object to be modified again
	LastPlannedViewSpecSHA string `json:"lastPlannedViewSpecSHA,omitempty" yaml:"lastPlannedViewSpecSHA,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// View is the Schema for the view API
// +kubebuilder:printcolumn:name="Namespace",type=string,JSONPath=`.metadata.namespace`,priority=1
// +kubebuilder:printcolumn:name="View",type=string,JSONPath=`.spec.name`
// +kubebuilder:printcolumn:name="Database",type=string,JSONPath=`.spec.database`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +k8s:openapi-gen=true
type View struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ViewSpec   `json:"spec,omitempty"`
	Status ViewStatus `json:"status,omitempty"`
}

func (v View) GetSHA() (string, error) {
	// ignoring the status, json marshal the spec and the metadata
	o := struct {
		Spec ViewSpec `json:"spec,omitempty"`
	}{
		Spec: v.Spec,
	}

	b, err := json.Marshal(o)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal")
	}

	sum := sha256.Sum256(b)
	return fmt.Sprintf("%x", sum), nil
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ViewList contains a list of View
type ViewList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []View `json:"items"`
}

func init() {
	SchemeBuilder.Register(&View{}, &ViewList{})
}
