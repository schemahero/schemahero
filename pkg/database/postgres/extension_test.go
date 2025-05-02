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

package postgres

import (
	"testing"

	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateExtensionStatements(t *testing.T) {
	tests := []struct {
		name       string
		extensions []*schemasv1alpha4.PostgresDatabaseExtension
		expected   []string
		wantErr    bool
	}{
		{
			name: "basic extension",
			extensions: []*schemasv1alpha4.PostgresDatabaseExtension{
				{
					Name: "vector",
				},
			},
			expected: []string{
				"CREATE EXTENSION IF NOT EXISTS \"vector\";",
			},
			wantErr: false,
		},
		{
			name: "extension with version",
			extensions: []*schemasv1alpha4.PostgresDatabaseExtension{
				{
					Name:    "vector",
					Version: strPtr("1.0"),
				},
			},
			expected: []string{
				"CREATE EXTENSION IF NOT EXISTS \"vector\" VERSION 1.0;",
			},
			wantErr: false,
		},
		{
			name: "extension with schema",
			extensions: []*schemasv1alpha4.PostgresDatabaseExtension{
				{
					Name:   "vector",
					Schema: strPtr("public"),
				},
			},
			expected: []string{
				"CREATE EXTENSION IF NOT EXISTS \"vector\" SCHEMA \"public\";",
			},
			wantErr: false,
		},
		{
			name: "extension with version and schema",
			extensions: []*schemasv1alpha4.PostgresDatabaseExtension{
				{
					Name:    "vector",
					Version: strPtr("1.0"),
					Schema:  strPtr("public"),
				},
			},
			expected: []string{
				"CREATE EXTENSION IF NOT EXISTS \"vector\" VERSION 1.0 SCHEMA \"public\";",
			},
			wantErr: false,
		},
		{
			name: "multiple extensions",
			extensions: []*schemasv1alpha4.PostgresDatabaseExtension{
				{
					Name: "vector",
				},
				{
					Name: "pg_stat_statements",
				},
			},
			expected: []string{
				"CREATE EXTENSION IF NOT EXISTS \"vector\";",
				"CREATE EXTENSION IF NOT EXISTS \"pg_stat_statements\";",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statements, err := CreateExtensionStatements(tt.extensions)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, statements)
			}
		})
	}
}

func strPtr(s string) *string {
	return &s
}
