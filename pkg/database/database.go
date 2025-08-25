package database

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pkg/errors"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/client/schemaheroclientset/scheme"
	"github.com/schemahero/schemahero/pkg/database/interfaces"
	"github.com/schemahero/schemahero/pkg/database/plugin"
	"github.com/schemahero/schemahero/pkg/database/types"
	"github.com/schemahero/schemahero/pkg/logger"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

type Database struct {
	InputDir       string
	OutputDir      string
	Driver         string
	URI            string
	Hosts          []string
	Username       string
	Password       string
	Keyspace       string
	DeploySeedData bool
	pluginManager  *plugin.PluginManager
}

// SetPluginManager sets the plugin manager for this database instance.
// The plugin manager is used to resolve database connections through plugins
// before falling back to in-tree drivers.
func (d *Database) SetPluginManager(manager *plugin.PluginManager) {
	d.pluginManager = manager
}

// getConnection attempts to establish a database connection through plugins.
// For postgres and mysql, plugins are required. Other drivers still use in-tree implementations.
func (d *Database) GetConnection(ctx context.Context) (interfaces.SchemaHeroDatabaseConnection, error) {
	// Try plugin-based connection first if plugin manager is set
	if d.pluginManager != nil {
		// Prepare connection options based on driver
		options := map[string]interface{}{}
		if d.Driver == "cassandra" {
			// Cassandra requires special connection parameters
			if len(d.Hosts) > 0 {
				options["hosts"] = d.Hosts
			}
			if d.Username != "" {
				options["username"] = d.Username
			}
			if d.Password != "" {
				options["password"] = d.Password
			}
			if d.Keyspace != "" {
				options["keyspace"] = d.Keyspace
			}
		}

		conn, err := d.pluginManager.GetConnection(ctx, d.Driver, d.URI, options)
		if err == nil {
			return conn, nil
		}
		// For postgres, mysql, sqlite, cassandra, and rqlite, plugins are required - no fallback
		if d.Driver == "postgres" || d.Driver == "postgresql" || d.Driver == "cockroachdb" || d.Driver == "timescaledb" {
			return nil, errors.Wrapf(err, "postgres driver requires plugin")
		}
		if d.Driver == "mysql" || d.Driver == "mariadb" {
			return nil, errors.Wrapf(err, "mysql driver requires plugin")
		}
		if d.Driver == "sqlite" || d.Driver == "sqlite3" {
			return nil, errors.Wrapf(err, "sqlite driver requires plugin")
		}
		if d.Driver == "rqlite" {
			return nil, errors.Wrapf(err, "rqlite driver requires plugin")
		}
		if d.Driver == "cassandra" {
			return nil, errors.Wrapf(err, "cassandra driver requires plugin")
		}
		// Log the plugin failure for other drivers
		logger.Info("Plugin connection failed for driver",
			zap.String("driver", d.Driver),
			zap.Error(err))
	}

	// For drivers that require plugins, return error if no plugin manager
	switch d.Driver {
	case "postgres", "postgresql", "cockroachdb", "timescaledb":
		return nil, errors.New("postgres driver requires plugin - install schemahero-postgres plugin")

	case "mysql", "mariadb":
		return nil, errors.New("mysql driver requires plugin - install schemahero-mysql plugin")

	case "rqlite":
		return nil, errors.New("rqlite driver requires plugin - install schemahero-rqlite plugin")

	case "sqlite", "sqlite3":
		return nil, errors.New("sqlite driver requires plugin - install schemahero-sqlite plugin")

	case "cassandra":
		return nil, errors.New("cassandra driver requires plugin - install schemahero-cassandra plugin")

	default:
		return nil, errors.Errorf("unknown database driver: %q", d.Driver)
	}
}

func (d *Database) CreateFixturesSync() error {
	logger.Info("generating fixtures",
		zap.String("input-dir", d.InputDir))

	statements := []string{}
	handleFile := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		fileData, err := ioutil.ReadFile(filepath.Join(d.InputDir, info.Name()))
		if err != nil {
			return err
		}

		var spec *schemasv1alpha4.TableSpec

		parsedK8sObject := schemasv1alpha4.Table{}
		if err := yaml.Unmarshal(fileData, &parsedK8sObject); err == nil {
			if parsedK8sObject.Spec.Schema != nil {
				spec = &parsedK8sObject.Spec
			}
		}

		if spec == nil {
			plainSpec := schemasv1alpha4.TableSpec{}
			if err := yaml.Unmarshal(fileData, &plainSpec); err != nil {
				return err
			}

			spec = &plainSpec
		}

		if spec.Schema == nil {
			return nil
		}

		// For fixtures, we don't need a database connection - just generate DDL
		// We'll use the plugin manager with a special fixture-only mode
		if d.pluginManager != nil {
			// Create a dummy URI for fixture generation - we won't actually connect
			dummyURI := "fixture-only://localhost/schemahero"

			// Pass a special option to indicate this is fixture-only mode
			options := map[string]interface{}{
				"fixture-only": true,
			}

			conn, err := d.pluginManager.GetConnection(context.Background(), d.Driver, dummyURI, options)
			if err == nil {
				defer conn.Close()

				// Generate fixtures statements
				stmts, err := conn.GenerateFixtures(spec)
				if err != nil {
					return errors.Wrap(err, "failed to generate fixtures")
				}

				statements = append(statements, stmts...)
				return nil
			}
			// If plugin connection fails, fall through to error
		}

		// Fallback to error for drivers that require plugins
		if d.Driver == "postgres" || d.Driver == "postgresql" {
			return errors.New("postgres driver requires plugin - install schemahero-postgres plugin")
		} else if d.Driver == "mysql" || d.Driver == "mariadb" {
			return errors.New("mysql driver requires plugin - install schemahero-mysql plugin")
		} else if d.Driver == "cockroachdb" {
			return errors.New("cockroachdb driver requires plugin - install schemahero-postgres plugin")
		} else if d.Driver == "sqlite" || d.Driver == "sqlite3" {
			return errors.New("sqlite driver requires plugin - install schemahero-sqlite plugin")
		} else if d.Driver == "rqlite" {
			return errors.New("rqlite driver requires plugin - install schemahero-rqlite plugin")
		} else if d.Driver == "timescaledb" {
			return errors.New("postgres driver requires plugin - install schemahero-postgres plugin")
		} else if d.Driver == "cassandra" {
			return errors.New("cassandra driver requires plugin - install schemahero-cassandra plugin")
		}

		return nil
	}

	err := filepath.Walk(d.InputDir, handleFile)
	if err != nil {
		return err
	}

	output := strings.Join(statements, ";\n")
	if output != "" {
		output = output + ";\n"
	}
	output = fmt.Sprintf("/* Auto generated file. Do not edit by hand. This file was generated by SchemaHero. */\n\n%s\n", output)

	if _, err := os.Stat(d.OutputDir); os.IsNotExist(err) {
		os.MkdirAll(d.OutputDir, 0o750)
	}

	err = ioutil.WriteFile(filepath.Join(d.OutputDir, "fixtures.sql"), []byte(output), 0o600)
	if err != nil {
		return err
	}

	return nil
}

func (d *Database) PlanSyncFromFile(filename string, specType string) ([]string, error) {
	specContents, err := ioutil.ReadFile(filepath.Clean(filename))
	if err != nil {
		return nil, errors.Wrap(err, "failed to read file")
	}

	plan, err := d.PlanSync(specContents, specType)
	if err != nil {
		return nil, errors.Wrap(err, "failed to plan sync")
	}

	return plan, nil
}

func (d *Database) PlanSync(specContents []byte, specType string) ([]string, error) {
	// Try GVK first and fall back to plain spec for backwards compatibility
	plan, err := d.planGVKSync(specContents)
	if err == nil {
		return plan, nil
	}

	logger.Debugf("failed to plan using GVK, falling back on spec type parameter: %s", err)

	if specType == "table" {
		plan, err := d.planTableSync(specContents)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to plan table sync")
		}
		return plan, nil
	} else if specType == "type" {
		plan, err := d.planTypeSync(specContents)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to plan type sync")
		}
		return plan, nil
	} else if specType == "view" {
		plan, err := d.planViewSync(specContents)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to plan view sync")
		}
		return plan, nil
	} else if specType == "extension" {
		plan, err := d.planExtensionSync(specContents)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to plan extension sync")
		}
		return plan, nil
	} else if specType == "function" {
		plan, err := d.planFunctionSync(specContents)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to plan function sync")
		}
		return plan, nil
	}

	return nil, errors.New("unknown spec type")
}

func (d *Database) planGVKSync(specContents []byte) ([]string, error) {
	decode := scheme.Codecs.UniversalDeserializer().Decode

	obj, gvk, err := decode(specContents, nil, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode spec")
	}

	if gvk.Group == "schemas.schemahero.io" && gvk.Version == "v1alpha4" && gvk.Kind == "Table" {
		table := obj.(*schemasv1alpha4.Table)
		plan, err := d.PlanSyncTableSpec(&table.Spec)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to plan table %s", table.Name)
		}
		return plan, nil
	} else if gvk.Group == "schemas.schemahero.io" && gvk.Version == "v1alpha4" && gvk.Kind == "View" {
		view := obj.(*schemasv1alpha4.View)
		plan, err := d.PlanSyncViewSpec(&view.Spec)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to plan view %s", view.Name)
		}
		return plan, nil
	} else if gvk.Group == "schemas.schemahero.io" && gvk.Version == "v1alpha4" && gvk.Kind == "DatabaseExtension" {
		extension := obj.(*schemasv1alpha4.DatabaseExtension)
		plan, err := d.PlanSyncDatabaseExtensionSpec(&extension.Spec)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to plan extension %s", extension.Name)
		}
		return plan, nil
	} else {
		return nil, errors.Errorf("unknown gvk %s", gvk)
	}
}

func (d *Database) planTableSync(specContents []byte) ([]string, error) {
	parsedK8sObject := schemasv1alpha4.Table{}
	var spec *schemasv1alpha4.TableSpec
	if err := yaml.Unmarshal(specContents, &parsedK8sObject); err == nil {
		if parsedK8sObject.Spec.Schema != nil {
			spec = &parsedK8sObject.Spec
		}
	}

	if spec == nil {
		plainSpec := schemasv1alpha4.TableSpec{}
		if err := yaml.Unmarshal(specContents, &plainSpec); err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal table spec")
		}

		spec = &plainSpec
	}

	plan, err := d.PlanSyncTableSpec(spec)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to plan table sync for %s", spec.Name)
	}

	return plan, nil
}

func (d *Database) planViewSync(specContents []byte) ([]string, error) {
	parsedK8sObject := schemasv1alpha4.View{}
	var spec *schemasv1alpha4.ViewSpec
	if err := yaml.Unmarshal(specContents, &parsedK8sObject); err == nil {
		if parsedK8sObject.Spec.Schema != nil {
			spec = &parsedK8sObject.Spec
		}
	}

	if spec == nil {
		plainSpec := schemasv1alpha4.ViewSpec{}
		if err := yaml.Unmarshal(specContents, &plainSpec); err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal view spec")
		}

		spec = &plainSpec
	}

	plan, err := d.PlanSyncViewSpec(spec)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to plan view sync for %s", spec.Name)
	}

	return plan, nil
}

func (d *Database) SortSpecs(specs []types.Spec) {
	switch d.Driver {
	case "postgres", "timescaledb":
		sort.Sort(types.Specs(specs))
	}
}

func (d *Database) PlanSyncViewSpec(spec *schemasv1alpha4.ViewSpec) ([]string, error) {
	if spec.Schema == nil {
		return []string{}, nil
	}

	// Views are supported through plugins for databases that support them
	if d.Driver == "postgres" || d.Driver == "cockroachdb" || d.Driver == "timescaledb" {
		conn, err := d.GetConnection(context.Background())
		if err != nil {
			return nil, errors.Wrap(err, "failed to get database connection")
		}
		defer conn.Close()

		var schema interface{}
		if d.Driver == "postgres" || d.Driver == "cockroachdb" {
			schema = spec.Schema.Postgres
		} else if d.Driver == "timescaledb" {
			schema = spec.Schema.TimescaleDB
		}

		if schema == nil {
			return []string{}, nil
		}

		return conn.PlanViewSchema(spec.Name, schema)
	} else if d.Driver == "mysql" {
		conn, err := d.GetConnection(context.Background())
		if err != nil {
			return nil, errors.Wrap(err, "failed to get database connection")
		}
		defer conn.Close()

		if spec.Schema.Mysql == nil {
			return []string{}, nil
		}

		return conn.PlanViewSchema(spec.Name, spec.Schema.Mysql)
	}

	// Other drivers don't support views yet
	return nil, errors.Errorf("driver %s does not support views", d.Driver)
}

func (d *Database) PlanSyncTableSpec(spec *schemasv1alpha4.TableSpec) ([]string, error) {
	if spec.Schema == nil {
		// If there's no schema but there is seed data, process only the seed data
		if d.DeploySeedData && spec.SeedData != nil {
			return d.PlanSyncSeedData(spec)
		}
		return []string{}, nil
	}

	var seedData *schemasv1alpha4.SeedData
	if d.DeploySeedData {
		seedData = spec.SeedData
	}

	// Use connection-based planning for postgres, mysql, sqlite, rqlite, timescaledb, and cassandra
	if d.Driver == "postgres" || d.Driver == "cockroachdb" || d.Driver == "mysql" || d.Driver == "timescaledb" || d.Driver == "sqlite" || d.Driver == "sqlite3" || d.Driver == "rqlite" || d.Driver == "cassandra" {
		conn, err := d.GetConnection(context.Background())
		if err != nil {
			return nil, errors.Wrap(err, "failed to get database connection")
		}
		defer conn.Close()

		// Get the appropriate schema based on driver
		var schema interface{}
		switch d.Driver {
		case "postgres":
			schema = spec.Schema.Postgres
		case "cockroachdb":
			schema = spec.Schema.CockroachDB
		case "mysql":
			schema = spec.Schema.Mysql
		case "timescaledb":
			schema = spec.Schema.TimescaleDB
		case "sqlite", "sqlite3":
			schema = spec.Schema.SQLite
		case "rqlite":
			schema = spec.Schema.RQLite
		case "cassandra":
			schema = spec.Schema.Cassandra
		}

		return conn.PlanTableSchema(spec.Name, schema, seedData)
	}

	return nil, errors.Errorf("unknown database driver: %q", d.Driver)
}

func (d *Database) PlanSyncSeedData(spec *schemasv1alpha4.TableSpec) ([]string, error) {
	if spec.SeedData == nil {
		return []string{}, nil
	}

	// For seed data without schema, we need to connect to the database to get the existing schema
	// This is supported through the plugin system for all drivers
	if d.Driver == "postgres" || d.Driver == "cockroachdb" || d.Driver == "mysql" || d.Driver == "timescaledb" || 
		d.Driver == "sqlite" || d.Driver == "sqlite3" || d.Driver == "rqlite" || d.Driver == "cassandra" {
		
		conn, err := d.GetConnection(context.Background())
		if err != nil {
			return nil, errors.Wrap(err, "failed to get database connection")
		}
		defer conn.Close()

		// When there's no schema, pass nil for the schema and let the plugin handle it
		// The plugin should retrieve the existing schema from the database
		return conn.PlanTableSchema(spec.Name, nil, spec.SeedData)
	}

	return nil, errors.Errorf("unknown database driver: %q", d.Driver)
}

func (d *Database) planTypeSync(specContents []byte) ([]string, error) {
	var spec *schemasv1alpha4.DataTypeSpec
	parsedK8sObject := schemasv1alpha4.DataType{}
	if err := yaml.Unmarshal(specContents, &parsedK8sObject); err == nil {
		if parsedK8sObject.Spec.Schema != nil {
			spec = &parsedK8sObject.Spec
		}
	}

	if spec == nil {
		plainSpec := schemasv1alpha4.DataTypeSpec{}
		if err := yaml.Unmarshal(specContents, &plainSpec); err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal type sync spec")
		}

		spec = &plainSpec
	}

	return d.PlanSyncTypeSpec(spec)
}

func (d *Database) PlanSyncTypeSpec(spec *schemasv1alpha4.DataTypeSpec) ([]string, error) {
	if spec.Schema == nil {
		return []string{}, nil
	}

	if d.Driver == "cassandra" {
		return nil, errors.New("cassandra driver requires plugin - install schemahero-cassandra plugin")
	}

	return nil, errors.Errorf("planning types is not supported for driver %q", d.Driver)
}

func (d *Database) ApplySync(statements []string) error {
	// Print each statement before executing
	for _, statement := range statements {
		if statement != "" {
			fmt.Printf("Executing query %q\n", statement)
		}
	}

	// Use connection-based deployment for postgres, mysql, sqlite, rqlite, timescaledb, and cassandra
	if d.Driver == "postgres" || d.Driver == "postgresql" || d.Driver == "cockroachdb" || d.Driver == "mysql" || d.Driver == "timescaledb" || d.Driver == "sqlite" || d.Driver == "sqlite3" || d.Driver == "rqlite" || d.Driver == "cassandra" {
		conn, err := d.GetConnection(context.Background())
		if err != nil {
			return errors.Wrap(err, "failed to get database connection")
		}
		defer conn.Close()

		// Use the DeployStatements method on the connection
		return conn.DeployStatements(statements)
	}

	return errors.Errorf("unknown database driver: %q", d.Driver)
}

// Combine lines that don't terminate with a semicolon.
// Semicolon on the last line is optional.
func (d *Database) GetStatementsFromDDL(ddl string) []string {
	lines := strings.Split(ddl, "\n")

	statements := []string{}

	functionTagOpen := false
	dollarQuoteTag := ""
	statement := ""
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check for dollar-quoted string delimiters (e.g., $$ or $_SCHEMAHERO_$)
		// These are used in PostgreSQL functions to avoid escaping issues
		if dollarQuoteTag == "" {
			// Look for opening dollar quote
			if idx := strings.Index(line, "$"); idx >= 0 {
				// Find the closing $ for the tag
				endIdx := strings.Index(line[idx+1:], "$")
				if endIdx >= 0 {
					tag := line[idx : idx+endIdx+2]
					// Check if this same tag appears again on the line (opening and closing on same line)
					if !strings.Contains(line[idx+endIdx+2:], tag) {
						dollarQuoteTag = tag
					}
				}
			}
		} else {
			// Look for closing dollar quote
			if strings.Contains(line, dollarQuoteTag) {
				dollarQuoteTag = ""
			}
		}

		// we assume the function tag to always have its own line and never be the last line for simplicity
		if line == "-- Function body follows" {
			functionTagOpen = !functionTagOpen
		}

		// Don't split on semicolons if we're inside a dollar-quoted string or function tag
		if !functionTagOpen && dollarQuoteTag == "" && (i == len(lines)-1 || strings.HasSuffix(line, ";")) {
			statement = statement + " " + line
			statements = append(statements, strings.TrimSpace(statement))
			statement = ""
		} else {
			statement = statement + " " + line
		}
	}

	if statement != "" {
		statements = append(statements, strings.TrimSpace(statement))
	}

	return statements
}

func (d *Database) planExtensionSync(specContents []byte) ([]string, error) {
	var spec *schemasv1alpha4.DatabaseExtensionSpec
	parsedK8sObject := schemasv1alpha4.DatabaseExtension{}
	if err := yaml.Unmarshal(specContents, &parsedK8sObject); err == nil {
		if parsedK8sObject.Spec.Database != "" {
			spec = &parsedK8sObject.Spec
		}
	}

	if spec == nil {
		plainSpec := schemasv1alpha4.DatabaseExtensionSpec{}
		if err := yaml.Unmarshal(specContents, &plainSpec); err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal extension spec")
		}

		spec = &plainSpec
	}

	return d.PlanSyncDatabaseExtensionSpec(spec)
}

func (d *Database) PlanSyncDatabaseExtensionSpec(spec *schemasv1alpha4.DatabaseExtensionSpec) ([]string, error) {
	// Use connection-based planning for postgres
	if d.Driver == "postgres" || d.Driver == "cockroachdb" || d.Driver == "timescaledb" {
		conn, err := d.GetConnection(context.Background())
		if err != nil {
			return nil, errors.Wrap(err, "failed to get database connection")
		}
		defer conn.Close()

		// Get the appropriate schema based on driver
		var schema interface{}
		var extensionName string
		if d.Driver == "postgres" && spec.Postgres != nil {
			schema = spec.Postgres
			extensionName = spec.Postgres.Name
		}
		// CockroachDB and TimescaleDB would also use postgres schema if they support extensions

		return conn.PlanExtensionSchema(extensionName, schema)
	}

	return nil, errors.Errorf("planning extensions is not supported for driver %q or extension type not specified", d.Driver)
}

func (d *Database) planFunctionSync(specContents []byte) ([]string, error) {
	var spec *schemasv1alpha4.FunctionSpec
	parsedK8sObject := schemasv1alpha4.Function{}
	if err := yaml.Unmarshal(specContents, &parsedK8sObject); err == nil {
		if parsedK8sObject.Spec.Database != "" {
			spec = &parsedK8sObject.Spec
		}
	}

	if spec == nil {
		plainSpec := schemasv1alpha4.FunctionSpec{}
		if err := yaml.Unmarshal(specContents, &plainSpec); err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal function spec")
		}

		spec = &plainSpec
	}

	return d.PlanSyncFunctionSpec(spec)
}

func (d *Database) PlanSyncFunctionSpec(spec *schemasv1alpha4.FunctionSpec) ([]string, error) {
	// Use connection-based planning for postgres
	if d.Driver == "postgres" || d.Driver == "cockroachdb" || d.Driver == "timescaledb" {
		conn, err := d.GetConnection(context.Background())
		if err != nil {
			return nil, errors.Wrap(err, "failed to get database connection")
		}
		defer conn.Close()

		// Get the appropriate schema based on driver
		var schema interface{}
		if d.Driver == "postgres" && spec.Schema != nil && spec.Schema.Postgres != nil {
			schema = spec.Schema.Postgres
		}
		// CockroachDB and TimescaleDB would also use postgres schema if they support functions

		return conn.PlanFunctionSchema(spec.Name, schema)
	}

	return nil, errors.Errorf("planning functions is not supported for driver %q or function database schema not specified", d.Driver)
}

// generateFixturesFromLib generates fixtures using the plugin lib packages directly
// without requiring a database connection. This is used for fixture generation.
