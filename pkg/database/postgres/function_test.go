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
)

func TestCreateFunctionStatements(t *testing.T) {
	tests := []struct {
		name     string
		function schemasv1alpha4.PostgresqlFunctionSchema
		expected []string
	}{
		{
			name: "get_user_count",
			function: schemasv1alpha4.PostgresqlFunctionSchema{
				Schema:    "public",
				Lang:      "PLpgSQL",
				Params:    []*schemasv1alpha4.PostgresqlExecuteParameter{},
				ReturnSet: false,
				Return:    "bigint",
				As: `DECLARE
  user_count bigint;
BEGIN
  SELECT COUNT(*) INTO user_count FROM users;
  RETURN user_count;
END;`,
			},
			expected: []string{
				`create function "get_user_count()" returns bigint as $$
DECLARE
  user_count bigint;
BEGIN
  SELECT COUNT(*) INTO user_count FROM users;
  RETURN user_count;
END;
$$ language PLpgSQL;`,
			},
		},
		{
			name: "get_user_count_with_param",
			function: schemasv1alpha4.PostgresqlFunctionSchema{
				Schema: "public",
				Lang:   "PLpgSQL",
				Params: []*schemasv1alpha4.PostgresqlExecuteParameter{
					{
						Name: "users_table",
						Type: "text",
					},
				},
				ReturnSet: false,
				Return:    "bigint",
				As: `DECLARE
  user_count bigint;
BEGIN
  SELECT COUNT(*) INTO user_count FROM $1;
  RETURN user_count;
END;`,
			},
			expected: []string{
				`create function "get_user_count_with_param(users_table text)" returns bigint as $$
DECLARE
  user_count bigint;
BEGIN
  SELECT COUNT(*) INTO user_count FROM $1;
  RETURN user_count;
END;
$$ language PLpgSQL;`,
			},
		},
		{
			name: "find_film_by_id",
			function: schemasv1alpha4.PostgresqlFunctionSchema{
				Schema: "television",
				Lang:   "PLpgSQL",
				Params: []*schemasv1alpha4.PostgresqlExecuteParameter{
					{
						Name: "p_id",
						Type: "int",
					},
				},
				ReturnSet: true,
				Return:    "film",
				As: `BEGIN
   RETURN query SELECT * FROM film WHERE film_id = p_id;
END;`,
			},
			expected: []string{
				`create function "television.find_film_by_id(p_id int)" returns setof film as $$
BEGIN
   RETURN query SELECT * FROM film WHERE film_id = p_id;
END;
$$ language PLpgSQL;`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statements := CreateFunctionStatements(tt.name, &tt.function)
			assert.Equal(t, tt.expected, statements)
		})
	}
}
