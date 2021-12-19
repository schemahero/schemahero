package generate

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/database/postgres"
	"github.com/schemahero/schemahero/pkg/database/types"
	"gopkg.in/yaml.v2"
)

func (g *Generator) generatePostgres(db *postgres.PostgresConnection) error {
	tables, err := db.ListTables()
	if err != nil {
		return errors.Wrap(err, "failed to list tables")
	}

	filesWritten := make([]string, 0)
	for _, table := range tables {
		primaryKey, err := db.GetTablePrimaryKey(table.Name)
		if err != nil {
			return errors.Wrap(err, "failed to get table primary key")
		}

		foreignKeys, err := db.ListTableForeignKeys(g.DBName, table.Name)
		if err != nil {
			return errors.Wrap(err, "failed to list table foreign keys")
		}

		indexes, err := db.ListTableIndexes(g.DBName, table.Name)
		if err != nil {
			return errors.Wrap(err, "failed to list table indexes")
		}

		columns, err := db.GetTableSchema(table.Name)
		if err != nil {
			return errors.Wrap(err, "failed to get table schema")
		}

		var primaryKeyColumns []string
		if primaryKey != nil {
			primaryKeyColumns = primaryKey.Columns
		}
		tableYAML, err := generatePostgresqlTableYAML(g.DBName, table, primaryKeyColumns, foreignKeys, indexes, columns)
		if err != nil {
			return errors.Wrap(err, "failed to generate table yaml")
		}

		// ensure that outputdir exists
		fi, err := os.Stat(g.OutputDir)
		if os.IsNotExist(err) {
			if err := os.MkdirAll(g.OutputDir, 0750); err != nil {
				return errors.Wrap(err, "failed to create output dir")
			}
		} else if err != nil {
			return errors.Wrap(err, "failed to check if output dir exists")
		} else if !fi.IsDir() {
			return errors.New("output dir already exists and is not a directory")
		}

		// If there was a outputdir set, write it, else print it
		if g.OutputDir != "" {
			if err := ioutil.WriteFile(filepath.Join(g.OutputDir, fmt.Sprintf("%s.yaml", sanitizeName(table.Name))), []byte(tableYAML), 0600); err != nil {
				return err
			}

			filesWritten = append(filesWritten, fmt.Sprintf("./%s.yaml", sanitizeName(table.Name)))
		} else {

			fmt.Println(tableYAML)
			fmt.Println("---")
		}
	}

	// If there was an output-dir, write a kustomization.yaml too -- this should be optional
	if g.OutputDir != "" {
		kustomization := struct {
			Resources []string `yaml:"resources"`
		}{
			filesWritten,
		}

		kustomizeDoc, err := yaml.Marshal(kustomization)
		if err != nil {
			return err
		}

		if err := ioutil.WriteFile(filepath.Join(g.OutputDir, "kustomization.yaml"), kustomizeDoc, 0644); err != nil {
			return err
		}
	}
	return nil
}

func generatePostgresqlTableYAML(dbName string, table *types.Table, primaryKey []string, foreignKeys []*types.ForeignKey, indexes []*types.Index, columns []*types.Column) (string, error) {
	schemaForeignKeys := make([]*schemasv1alpha4.PostgresqlTableForeignKey, 0)
	for _, foreignKey := range foreignKeys {
		schemaForeignKey := types.ForeignKeyToPostgresqlSchemaForeignKey(foreignKey)
		schemaForeignKeys = append(schemaForeignKeys, schemaForeignKey)
	}

	schemaIndexes := make([]*schemasv1alpha4.PostgresqlTableIndex, 0)
	for _, index := range indexes {
		schemaIndex := types.IndexToPostgresqlSchemaIndex(index)
		schemaIndexes = append(schemaIndexes, schemaIndex)
	}

	schemaTableColumns := make([]*schemasv1alpha4.PostgresqlTableColumn, 0)
	for _, column := range columns {
		schemaTableColumn, err := types.ColumnToPostgresqlSchemaColumn(column)
		if err != nil {
			fmt.Printf("%#v\n", err)
			return "", err
		}

		schemaTableColumns = append(schemaTableColumns, schemaTableColumn)
	}

	tableSchema := &schemasv1alpha4.PostgresqlTableSchema{
		PrimaryKey:  primaryKey,
		Columns:     schemaTableColumns,
		ForeignKeys: schemaForeignKeys,
		Indexes:     schemaIndexes,
	}

	schema := &schemasv1alpha4.TableSchema{
		Postgres: tableSchema,
	}

	schemaHeroResource := schemasv1alpha4.TableSpec{
		Database: dbName,
		Name:     table.Name,
		Requires: []string{},
		Schema:   schema,
	}

	specDoc := struct {
		Spec schemasv1alpha4.TableSpec `yaml:"spec"`
	}{
		schemaHeroResource,
	}

	b, err := yaml.Marshal(&specDoc)
	if err != nil {
		fmt.Printf("%#v\n", err)
		return "", err
	}

	// TODO consider marshaling this instead of inline
	tableDoc := fmt.Sprintf(`apiVersion: schemas.schemahero.io/v1alpha4
kind: Table
metadata:
  name: %s
%s`, sanitizeName(table.Name), b)

	return tableDoc, nil

}
