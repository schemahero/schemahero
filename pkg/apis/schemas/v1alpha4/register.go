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

// NOTE: Boilerplate only.  Ignore this file.

// Package v1alpha4 contains API Schema definitions for the schemas v1alpha4 API group
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen=package,register
// +k8s:conversion-gen=github.com/schemahero/schemahero/pkg/apis/schemas
// +k8s:defaulter-gen=TypeMeta
// +groupName=schemas.schemahero.io
package v1alpha4

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: "schemas.schemahero.io", Version: "v1alpha4"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}

	// AddToScheme is required by pkg/client/...
	AddToScheme = SchemeBuilder.AddToScheme
)

func init() {
	SchemeBuilder.Register(&Table{}, &TableList{})
	SchemeBuilder.Register(&Migration{}, &MigrationList{})
	SchemeBuilder.Register(&View{}, &ViewList{})
	SchemeBuilder.Register(&DataType{}, &DataTypeList{})
	SchemeBuilder.Register(&DatabaseExtension{}, &DatabaseExtensionList{})
	SchemeBuilder.Register(&Function{}, &FunctionList{})
}

// Resource is required by pkg/client/listers/...
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}
