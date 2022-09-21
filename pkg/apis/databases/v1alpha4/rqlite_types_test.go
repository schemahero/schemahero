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
	"testing"

	"github.com/onsi/gomega"
	"gopkg.in/yaml.v2"
)

func TestUnmarshalRqliteSpec(t *testing.T) {
	manifest := `
uri:
    value: foo`

	g := gomega.NewGomegaWithT(t)

	r := RqliteConnection{}
	err := yaml.Unmarshal([]byte(manifest), &r)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(r.URI.Value).To(gomega.Equal("foo"))
}
