package table

import (
	"testing"

	schemasv1alpha3 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
)

func Test_planConfigMap(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		tableName string
		tableSpec schemasv1alpha3.TableSpec
		expect    corev1.ConfigMap
	}{
		{
			name:      "basic test",
			namespace: "foo",
			tableName: "name",
			tableSpec: schemasv1alpha3.TableSpec{
				Database: "db",
				Name:     "name",
				Schema: &schemasv1alpha3.TableSchema{
					Postgres: &schemasv1alpha3.SQLTableSchema{},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := require.New(t)

			actual, err := planConfigMap(test.namespace, test.tableName, test.tableSpec)
			req.NoError(err)

			// check some of the fields on the config map
			assert.Equal(t, test.tableName, actual.Name)
			assert.Len(t, actual.Data, 1)
			assert.NotNil(t, actual.Data["table.yaml"], actual.Data["table.yaml"])
		})
	}
}
