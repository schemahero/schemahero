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

func CreateExtensionStatements(extensions []*schemasv1alpha4.PostgresDatabaseExtension) ([]string, error) {
	statements := []string{}

	for _, extension := range extensions {
		statement := fmt.Sprintf("CREATE EXTENSION IF NOT EXISTS %s", pgx.Identifier{extension.Name}.Sanitize())

		if extension.Version != nil {
			statement += fmt.Sprintf(" VERSION %s", *extension.Version)
		}

		if extension.Schema != nil {
			statement += fmt.Sprintf(" SCHEMA %s", pgx.Identifier{*extension.Schema}.Sanitize())
		}

		statement += ";"
		statements = append(statements, statement)
	}

	return statements, nil
}
