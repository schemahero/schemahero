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

func Test_SQLTypes(t *testing.T) {
	const fk = `
isDeleted: false
columns:
  - name: id
    type: integer
  - name: order_id
    type: integer
primaryKey:
  - id
foreignKeys:
  - columns:
      - order_id
    references:
      table: order
      columns:
        - id
`

	g := gomega.NewGomegaWithT(t)

	table := PostgresqlTableSchema{}
	err := yaml.Unmarshal([]byte(fk), &table)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(table.Columns).To(gomega.HaveLen(2))
	g.Expect(table.ForeignKeys).To(gomega.HaveLen(1))

	g.Expect(table.ForeignKeys[0].Columns).To(gomega.HaveLen(1))
	g.Expect(table.ForeignKeys[0].Columns[0]).To(gomega.Equal("order_id"))

	g.Expect(table.ForeignKeys[0].References.Table).To(gomega.Equal("order"))
	g.Expect(table.ForeignKeys[0].References.Columns).To(gomega.HaveLen(1))
	g.Expect(table.ForeignKeys[0].References.Columns[0]).To(gomega.Equal("id"))

}
