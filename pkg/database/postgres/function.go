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
	"strings"

	"github.com/jackc/pgx/v4"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
)

func CreateFunctionStatements(functionName string, functionSchema *schemasv1alpha4.PostgresqlFunctionSchema) []string {
	qualifiedFunctionName := getQualifiedFunctionName(functionName, functionSchema.Schema, functionSchema.Params)

	statement := fmt.Sprintf("create function %s", pgx.Identifier{qualifiedFunctionName}.Sanitize())

	if functionSchema.Return != "" {
		statement = fmt.Sprintf("%s returns", statement)
		if functionSchema.ReturnSet {
			statement = fmt.Sprintf("%s setof", statement)
		}
		statement = fmt.Sprintf("%s %s", statement, functionSchema.Return)
	}

	statements := []string{
		fmt.Sprintf("%s as $$\n%s\n$$ language %s;", statement, functionSchema.As, functionSchema.Lang),
	}

	return statements
}

func DropFunctionStatements(functionName string, functionSchema *schemasv1alpha4.PostgresqlFunctionSchema) []string {
	qualifiedFunctionName := getQualifiedFunctionName(functionName, functionSchema.Schema, functionSchema.Params)

	statements := []string{
		fmt.Sprintf("drop function %s;", pgx.Identifier{qualifiedFunctionName}.Sanitize()),
	}

	return statements
}

func getQualifiedFunctionName(functionName, schema string, params []*schemasv1alpha4.PostgresqlFunctionParameter) string {
	qualifiedFunctionName := functionName
	if schema != "" && schema != "public" {
		qualifiedFunctionName = fmt.Sprintf("%s.%s", schema, functionName)
	}
	return fmt.Sprintf("%s(%s)", qualifiedFunctionName, serializeFunctionParams(params))
}

func serializeFunctionParams(params []*schemasv1alpha4.PostgresqlFunctionParameter) string {
	ps := []string{}
	for _, param := range params {
		p := ""
		if param.Mode != "" {
			p = param.Mode
		}
		if param.Name != "" {
			p = fmt.Sprintf("%s %s", p, param.Name)
		}
		p = fmt.Sprintf("%s %s", p, param.Type)
		ps = append(ps, p)
	}
	return strings.Join(ps, ", ")
}
