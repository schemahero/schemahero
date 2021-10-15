package types

import (
	"testing"

	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
)

func Test_GenerateMysqlIndexName(t *testing.T) {
	tests := []struct {
		name        string
		tableName   string
		schemaIndex *schemasv1alpha4.MysqlTableIndex
		want        string
	}{
		{
			name:      "short index",
			tableName: "table_name",
			schemaIndex: &schemasv1alpha4.MysqlTableIndex{
				Columns: []string{
					"col1",
					"col2",
				},
			},
			want: "idx_table_name_col1_col2",
		},
		{
			name:      "long index",
			tableName: "very_very_very_long_table_name",
			schemaIndex: &schemasv1alpha4.MysqlTableIndex{
				Columns: []string{
					"collumn_1",
					"collumn_2",
					"collumn_3",
					"collumn_4",
					"collumn_5",
					"collumn_6",
					"collumn_7",
				},
			},
			want: "idx_very_very_very_long_table_name_collumn_1_collumn_2_collumn_3",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GenerateMysqlIndexName(tt.tableName, tt.schemaIndex); got != tt.want {
				t.Errorf("GenerateMysqlIndexName() = %s, want %s", got, tt.want)
			}
		})
	}
}
