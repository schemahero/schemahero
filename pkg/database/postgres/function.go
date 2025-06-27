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
	"fmt"

	"github.com/jackc/pgx/v4"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
)

func CreateFunctionStatements(functionName string, functionSchema *schemasv1alpha4.PostgresqlFunctionSchema) []string {
	qualifiedFunctionName := getQualifiedFunctionName(functionName, functionSchema.Schema)

	statement := fmt.Sprintf("create function %s", pgx.Identifier{qualifiedFunctionName}.Sanitize())

	if functionSchema.Return != nil {
		statement = fmt.Sprintf("%s returns %s", *functionSchema.Return)
	}

	statements := []string{
		fmt.Sprintf(`%s AS $$\n%s\n$$ language %s;`, statement, functionSchema.As, functionSchema.Lang),
	}

	return statements
}

func DropFunctionStatements(functionName string, functionSchema *schemasv1alpha4.PostgresqlFunctionSchema) []string {
	qualifiedFunctionName := getQualifiedFunctionName(functionName, functionSchema.Schema)

	statements := []string{
		fmt.Sprintf("drop function if exists %s;", pgx.Identifier{qualifiedFunctionName}.Sanitize()),
	}

	return statements
}

func getQualifiedFunctionName(functionName, schema string) string {
	qualifiedFunctionName := functionName
	if schema != "" && schema != "public" {
		qualifiedFunctionName = fmt.Sprintf("%s.%s", schema, functionName)
	}
	return qualifiedFunctionName
}
