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

type DatabaseExtensionSpec struct {
	Database string `json:"database" yaml:"database"`

	Postgres *PostgresDatabaseExtension `json:"postgres,omitempty" yaml:"postgres,omitempty"`
}

type PostgresDatabaseExtension struct {
	Name string `json:"name" yaml:"name"`

	Version *string `json:"version,omitempty" yaml:"version,omitempty"`

	Schema *string `json:"schema,omitempty" yaml:"schema,omitempty"`
}

type DatabaseExtensionStatus struct {
	AppliedAt int64 `json:"appliedAt,omitempty" yaml:"appliedAt,omitempty"`

	Phase string `json:"phase,omitempty" yaml:"phase,omitempty"`

	Message string `json:"message,omitempty" yaml:"message,omitempty"`
}

type DatabaseExtension struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DatabaseExtensionSpec   `json:"spec,omitempty"`
	Status DatabaseExtensionStatus `json:"status,omitempty"`
}

type DatabaseExtensionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DatabaseExtension `json:"items"`
}
