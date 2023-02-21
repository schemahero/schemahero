package migration

import (
	"context"
	"testing"

	databasesv1alpha4 "github.com/schemahero/schemahero/pkg/apis/databases/v1alpha4"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	testclient "github.com/schemahero/schemahero/pkg/client/schemaheroclientset/fake"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_getDatabaseFromMigration(t *testing.T) {
	db := &databasesv1alpha4.Database{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testdb",
			Namespace: "namespace1",
		},
	}
	table1 := &schemasv1alpha4.Table{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "table1",
			Namespace: "namespace1",
		},
		Spec: schemasv1alpha4.TableSpec{
			Database: "testdb",
		},
	}
	view1 := &schemasv1alpha4.View{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "view1",
			Namespace: "namespace1",
		},
		Spec: schemasv1alpha4.ViewSpec{
			Database: "testdb",
		},
	}

	schemasClient = testclient.NewSimpleClientset(table1, view1).SchemasV1alpha4()
	databasesClient = testclient.NewSimpleClientset(db).DatabasesV1alpha4()

	tests := []struct {
		name      string
		migration *schemasv1alpha4.Migration
		want      *databasesv1alpha4.Database
	}{
		{
			name: "db from table",
			migration: &schemasv1alpha4.Migration{
				Spec: schemasv1alpha4.MigrationSpec{
					TableNamespace: "namespace1",
					TableName:      "table1",
				},
			},
			want: db,
		},
		{
			name: "db from view",
			migration: &schemasv1alpha4.Migration{
				Spec: schemasv1alpha4.MigrationSpec{
					TableNamespace: "namespace1",
					TableName:      "view1",
				},
			},
			want: db,
		},
		{
			name: "unknown db",
			migration: &schemasv1alpha4.Migration{
				Spec: schemasv1alpha4.MigrationSpec{
					TableNamespace: "namespace1",
					TableName:      "unknown",
				},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			got, err := getDatabaseFromMigration(ctx, tt.migration)
			if tt.want != nil {
				assert.Equal(t, tt.want, got)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
