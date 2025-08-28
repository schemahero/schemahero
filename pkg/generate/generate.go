package generate

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/database/interfaces"
	"github.com/schemahero/schemahero/pkg/database/plugin"
	"github.com/schemahero/schemahero/pkg/database/types"
	"gopkg.in/yaml.v2"
)

type Generator struct {
	Driver        string
	URI           string
	DBName        string
	OutputDir     string
	Schemas       []string
	pluginManager *plugin.PluginManager
}

// SetPluginManager sets the plugin manager for this generator instance.
// The plugin manager is used to resolve database connections through plugins.
func (g *Generator) SetPluginManager(manager *plugin.PluginManager) {
	g.pluginManager = manager
}

// getConnection attempts to establish a database connection through plugins.
func (g *Generator) getConnection(ctx context.Context) (interfaces.SchemaHeroDatabaseConnection, error) {
	if g.pluginManager == nil {
		return nil, errors.New("plugin manager not set - use SetPluginManager() before calling RunSync()")
	}

	// Handle PostgreSQL schema parameters
	uri := g.URI
	if g.Driver == "postgres" || g.Driver == "postgresql" || g.Driver == "cockroachdb" || g.Driver == "timescaledb" {
		if !strings.Contains(uri, "schema=") && !strings.Contains(uri, "schemas=") && len(g.Schemas) > 0 {
			schemasStr := strings.Join(g.Schemas, ",")

			// If there's only one schema and it's not "public", use schema parameter
			if len(g.Schemas) == 1 && g.Schemas[0] != "public" {
				if strings.Contains(uri, "?") {
					uri = fmt.Sprintf("%s&schema=%s", uri, g.Schemas[0])
				} else {
					uri = fmt.Sprintf("%s?schema=%s", uri, g.Schemas[0])
				}
			} else if len(g.Schemas) > 1 || (len(g.Schemas) == 1 && g.Schemas[0] != "public") {
				if strings.Contains(uri, "?") {
					uri = fmt.Sprintf("%s&schemas=%s", uri, schemasStr)
				} else {
					uri = fmt.Sprintf("%s?schemas=%s", uri, schemasStr)
				}
			}
		}
	}

	// Prepare connection options (currently no special options needed for generate)
	options := map[string]interface{}{}

	// Get connection through plugin manager
	conn, err := g.pluginManager.GetConnection(ctx, g.Driver, uri, options)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to connect to %s database", g.Driver)
	}

	return conn, nil
}

func (g *Generator) RunSync() error {
	fmt.Printf("connecting to %s\n", g.URI)

	ctx := context.Background()
	
	// Get database connection through plugin manager
	db, err := g.getConnection(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to connect to database")
	}
	defer db.Close()

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
		tableYAML, err := generateTableYAML(g.Driver, g.DBName, table, primaryKeyColumns, foreignKeys, indexes, columns)
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

func generateTableYAML(driver string, dbName string, table *types.Table, primaryKey []string, foreignKeys []*types.ForeignKey, indexes []*types.Index, columns []*types.Column) (string, error) {
	switch driver {
	case "mysql", "mariadb":
		return generateMysqlTableYAML(dbName, table, primaryKey, foreignKeys, indexes, columns)
	case "rqlite":
		return generateRqliteTableYAML(dbName, table, primaryKey, foreignKeys, indexes, columns)
	case "sqlite", "sqlite3":
		return generateSqliteTableYAML(dbName, table, primaryKey, foreignKeys, indexes, columns)
	case "cassandra":
		return generateCassandraTableYAML(dbName, table, primaryKey, foreignKeys, indexes, columns)
	case "postgres", "postgresql", "cockroachdb", "timescaledb":
		return generatePostgresqlTableYAML(driver, dbName, table, primaryKey, foreignKeys, indexes, columns)
	default:
		return "", errors.Errorf("unsupported database driver for generate: %s", driver)
	}
}

func generateMysqlTableYAML(dbName string, table *types.Table, primaryKey []string, foreignKeys []*types.ForeignKey, indexes []*types.Index, columns []*types.Column) (string, error) {
	schemaForeignKeys := make([]*schemasv1alpha4.MysqlTableForeignKey, 0)
	for _, foreignKey := range foreignKeys {
		schemaForeignKey := types.ForeignKeyToMysqlSchemaForeignKey(foreignKey)
		schemaForeignKeys = append(schemaForeignKeys, schemaForeignKey)
	}

	schemaIndexes := make([]*schemasv1alpha4.MysqlTableIndex, 0)
	for _, index := range indexes {
		schemaIndex := types.IndexToMysqlSchemaIndex(index)
		schemaIndexes = append(schemaIndexes, schemaIndex)
	}

	schemaTableColumns := make([]*schemasv1alpha4.MysqlTableColumn, 0)
	for _, column := range columns {
		schemaTableColumn, err := types.ColumnToMysqlSchemaColumn(column)
		if err != nil {
			return "", err
		}

		schemaTableColumns = append(schemaTableColumns, schemaTableColumn)
	}

	tableSchema := &schemasv1alpha4.MysqlTableSchema{
		PrimaryKey:     primaryKey,
		Columns:        schemaTableColumns,
		ForeignKeys:    schemaForeignKeys,
		Indexes:        schemaIndexes,
		DefaultCharset: table.Charset,
		Collation:      table.Collation,
	}

	schema := &schemasv1alpha4.TableSchema{}

	schema.Mysql = tableSchema

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

func generatePostgresqlTableYAML(driver string, dbName string, table *types.Table, primaryKey []string, foreignKeys []*types.ForeignKey, indexes []*types.Index, columns []*types.Column) (string, error) {
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

	schema := &schemasv1alpha4.TableSchema{}

	if driver == "postgres" {
		schema.Postgres = tableSchema
	} else if driver == "cockroachdb" {
		schema.CockroachDB = tableSchema
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

func generateRqliteTableYAML(dbName string, table *types.Table, primaryKey []string, foreignKeys []*types.ForeignKey, indexes []*types.Index, columns []*types.Column) (string, error) {
	schemaForeignKeys := make([]*schemasv1alpha4.RqliteTableForeignKey, 0)
	for _, foreignKey := range foreignKeys {
		schemaForeignKey := types.ForeignKeyToRqliteSchemaForeignKey(foreignKey)
		schemaForeignKeys = append(schemaForeignKeys, schemaForeignKey)
	}

	schemaIndexes := make([]*schemasv1alpha4.RqliteTableIndex, 0)
	for _, index := range indexes {
		schemaIndex := types.IndexToRqliteSchemaIndex(index)
		schemaIndexes = append(schemaIndexes, schemaIndex)
	}

	schemaTableColumns := make([]*schemasv1alpha4.RqliteTableColumn, 0)
	for _, column := range columns {
		schemaTableColumn, err := types.ColumnToRqliteSchemaColumn(column)
		if err != nil {
			return "", err
		}

		schemaTableColumns = append(schemaTableColumns, schemaTableColumn)
	}

	tableSchema := &schemasv1alpha4.RqliteTableSchema{
		PrimaryKey:  primaryKey,
		Columns:     schemaTableColumns,
		ForeignKeys: schemaForeignKeys,
		Indexes:     schemaIndexes,
	}

	schema := &schemasv1alpha4.TableSchema{}
	schema.RQLite = tableSchema

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

func generateSqliteTableYAML(dbName string, table *types.Table, primaryKey []string, foreignKeys []*types.ForeignKey, indexes []*types.Index, columns []*types.Column) (string, error) {
	schemaForeignKeys := make([]*schemasv1alpha4.SqliteTableForeignKey, 0)
	for _, foreignKey := range foreignKeys {
		schemaForeignKey := types.ForeignKeyToSqliteSchemaForeignKey(foreignKey)
		schemaForeignKeys = append(schemaForeignKeys, schemaForeignKey)
	}

	schemaIndexes := make([]*schemasv1alpha4.SqliteTableIndex, 0)
	for _, index := range indexes {
		schemaIndex := types.IndexToSqliteSchemaIndex(index)
		schemaIndexes = append(schemaIndexes, schemaIndex)
	}

	schemaTableColumns := make([]*schemasv1alpha4.SqliteTableColumn, 0)
	for _, column := range columns {
		schemaTableColumn, err := types.ColumnToSqliteSchemaColumn(column)
		if err != nil {
			return "", err
		}

		schemaTableColumns = append(schemaTableColumns, schemaTableColumn)
	}

	tableSchema := &schemasv1alpha4.SqliteTableSchema{
		PrimaryKey:  primaryKey,
		Columns:     schemaTableColumns,
		ForeignKeys: schemaForeignKeys,
		Indexes:     schemaIndexes,
	}

	schema := &schemasv1alpha4.TableSchema{}
	schema.SQLite = tableSchema

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

func generateCassandraTableYAML(dbName string, table *types.Table, primaryKey []string, foreignKeys []*types.ForeignKey, indexes []*types.Index, columns []*types.Column) (string, error) {
	// Cassandra doesn't support foreign keys, so we ignore them
	_ = foreignKeys

	// Cassandra doesn't have index support in SchemaHero schemas (it uses secondary indexes differently)
	_ = indexes

	// Convert columns to Cassandra columns
	schemaTableColumns := make([]*schemasv1alpha4.CassandraColumn, 0)
	for _, column := range columns {
		// Cassandra columns have a different structure
		cassandraColumn := &schemasv1alpha4.CassandraColumn{
			Name: column.Name,
			Type: column.DataType,
		}

		schemaTableColumns = append(schemaTableColumns, cassandraColumn)
	}

	// Convert primary key to Cassandra format ([][]string instead of []string)
	// For simple primary key, we wrap it in another array
	cassandraPrimaryKey := [][]string{}
	if len(primaryKey) > 0 {
		cassandraPrimaryKey = [][]string{primaryKey}
	}

	tableSchema := &schemasv1alpha4.CassandraTableSchema{
		PrimaryKey: cassandraPrimaryKey,
		Columns:    schemaTableColumns,
	}

	schema := &schemasv1alpha4.TableSchema{}
	schema.Cassandra = tableSchema

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

func sanitizeName(name string) string {
	return strings.Replace(name, "_", "-", -1)
}
