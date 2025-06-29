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

	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
)

// NOTE: the following keyword is chosen since we assume users will never need it in a function
const FunctionLogicTag = "$_SCHEMAHERO_$"

func CreateFunctionStatements(functionName string, functionSchema *schemasv1alpha4.PostgresqlFunctionSchema) []string {
	qualifiedFunctionName := getQualifiedExecuteName(functionName, functionSchema.Schema, functionSchema.Params)

	statement := fmt.Sprintf("create function %s", qualifiedFunctionName)

	if functionSchema.Return != "" {
		statement = fmt.Sprintf("%s returns", statement)
		if functionSchema.ReturnSet {
			statement = fmt.Sprintf("%s setof", statement)
		}
		statement = fmt.Sprintf("%s %s", statement, functionSchema.Return)
	}

	statements := []string{
		// it is important to keep the function logic tags on their own respective lines
		fmt.Sprintf("%s as\n%s\n%s\n%s\nlanguage %s", statement, FunctionLogicTag, functionSchema.As, FunctionLogicTag, functionSchema.Lang),
	}

	return statements
}

func DropFunctionStatements(functionName string, functionSchema *schemasv1alpha4.PostgresqlFunctionSchema) []string {
	qualifiedFunctionName := getQualifiedExecuteName(functionName, functionSchema.Schema, functionSchema.Params)

	statements := []string{
		fmt.Sprintf("drop function %s", qualifiedFunctionName),
	}

	return statements
}
