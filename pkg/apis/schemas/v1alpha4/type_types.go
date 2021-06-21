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

type DataTypeSchema struct {
	Cassandra *CassandraDataTypeSchema `json:"cassandra,omitempty" yaml:"cassandra,omitempty"`
}

// DataTypeSpec defines the desired state of Type
type DataTypeSpec struct {
	Database string `json:"database" yaml:"database"`
	Name     string `json:"name" yaml:"name"`

	Schema *DataTypeSchema `json:"schema,omitempty" yaml:"schema,omitempty"`
}

// DataTypeStatus defines the observed state of Type
type DataTypeStatus struct {
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DataType is the Schema for the datatypes API
// +k8s:openapi-gen=true
type DataType struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DataTypeSpec   `json:"spec,omitempty"`
	Status DataTypeStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DataTypeList contains a list of DataType
type DataTypeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DataType `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DataType{}, &DataTypeList{})
}
