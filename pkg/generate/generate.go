package generate

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	schemasv1alpha1 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha1"
	"github.com/schemahero/schemahero/pkg/database/postgres"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

type Generator struct {
	Viper *viper.Viper
}

func NewGenerator() *Generator {
	return &Generator{
		Viper: viper.GetViper(),
	}
}

func (g *Generator) RunSync() error {
	fmt.Printf("connecting to %s\n", g.Viper.GetString("uri"))

	db, err := postgres.Connect(g.Viper.GetString("uri"))
	if err != nil {
		return err
	}

	tables, err := db.ListTables()
	if err != nil {
		fmt.Printf("%#v\n", err)
		return err
	}

	filesWritten := make([]string, 0, 0)
	for _, table := range tables {
		primaryKey, err := db.GetTablePrimaryKey(table)
		if err != nil {
			fmt.Printf("%#v\n", err)
			return err
		}

		columns, err := db.GetTableSchema(table)
		if err != nil {
			fmt.Printf("%#v\n", err)
			return err
		}

		postgresTableColumns := make([]*schemasv1alpha1.SQLTableColumn, 0, 0)

		for _, column := range columns {
			postgresTableColumn, err := postgres.PostgresColumnToSchemaColumn(column)
			if err != nil {
				fmt.Printf("%#v\n", err)
				return err
			}

			postgresTableColumns = append(postgresTableColumns, postgresTableColumn)
		}

		postgresTableSchema := schemasv1alpha1.SQLTableSchema{
			PrimaryKey: primaryKey,
			Columns:    postgresTableColumns,
		}

		schemaHeroResource := schemasv1alpha1.TableSpec{
			Database: g.Viper.GetString("dbname"),
			Name:     table,
			Requires: []string{},
			Schema: &schemasv1alpha1.TableSchema{
				Postgres: &postgresTableSchema,
			},
		}

		specDoc := struct {
			Spec schemasv1alpha1.TableSpec `yaml:"spec"`
		}{
			schemaHeroResource,
		}

		b, err := yaml.Marshal(&specDoc)
		if err != nil {
			fmt.Printf("%#v\n", err)
			return err
		}

		// TODO consider marshaling this instead of inline
		tableDoc := fmt.Sprintf(`apiVersion: schemas.schemahero.io/v1alpha1
kind: Table
metadata:
  name: %s
%s`, sanitizeName(table), b)

		// If there was a outputdir set, write it, else print it
		if g.Viper.GetString("output-dir") != "" {
			if err := ioutil.WriteFile(filepath.Join(g.Viper.GetString("output-dir"), fmt.Sprintf("%s.yaml", table)), []byte(tableDoc), 0644); err != nil {
				return err
			}

			filesWritten = append(filesWritten, fmt.Sprintf("./%s.yaml", table))
		} else {
			fmt.Println(tableDoc)
			fmt.Println("---")
		}
	}

	// If there was an output-dir, write a kustomization.yaml too -- this should be optional
	if g.Viper.GetString("output-dir") != "" {
		kustomization := struct {
			Resources []string `yaml:"resources"`
		}{
			filesWritten,
		}

		kustomizeDoc, err := yaml.Marshal(kustomization)
		if err != nil {
			return err
		}

		if err := ioutil.WriteFile(filepath.Join(g.Viper.GetString("output-dir"), "kustomization.yaml"), kustomizeDoc, 0644); err != nil {
			return err
		}
	}
	return nil
}

func sanitizeName(name string) string {
	return strings.Replace(name, "_", "-", -1)
}
