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

package v1alpha5

import (
	"github.com/schemahero/schemahero/pkg/apis/databases/v1alpha4"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

// ConvertTo converts from this version to v1alpha4
func (src *Database) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1alpha4.Database)

	dst.ObjectMeta = src.ObjectMeta

	panic("Asdasdasd")
}

// ConvertFrom converts from v1alpha4 to this version
func (dst *Database) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1alpha4.Database)

	dst.ObjectMeta = src.ObjectMeta

	panic("weasdasd")
}
