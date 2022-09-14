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

package database

import (
	"context"
	"database/sql"
	"fmt"

	mysqldriver "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	databasesv1alpha4 "github.com/schemahero/schemahero/pkg/apis/databases/v1alpha4"
	"github.com/schemahero/schemahero/pkg/logger"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ reconcile.Reconciler = &ReconcileDatabaseSchema{}

// ReconcileDatabaseSchema reconciles a Database object
type ReconcileDatabaseSchema struct {
	client.Client
	scheme        *runtime.Scheme
	databaseNames []string
}

// Reconcile reads that state of the cluster for a Database object and makes changes based on the state read
// and what is in the Database.Spec for schemas
// +kubebuilder:rbac:groups=databases.schemahero.io,resources=databases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=databases.schemahero.io,resources=databases/status,verbs=get;update;patch
func (r *ReconcileDatabaseSchema) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	databaseInstance, err := r.getInstance(request)
	if err != nil {
		return reconcile.Result{}, err
	}

	// SchemaHero does not current support any database-wide schema properties in rqlite
	if databaseInstance.Spec.Connection.RQLite != nil {
		logger.Debug("ignoring rqlite database schema reconcile request")
		return reconcile.Result{}, nil
	}

	// SchemaHero does not current support any database-wide schema properties in postgresql
	if databaseInstance.Spec.Connection.Postgres != nil {
		logger.Debug("ignoring postgres database schema reconcile request")
		return reconcile.Result{}, nil
	}

	// SchemaHero does not current support any database-wide schema properties in cockroachdb
	if databaseInstance.Spec.Connection.CockroachDB != nil {
		logger.Debug("ignoring cockroachdb database schema reconcile request")
		return reconcile.Result{}, nil
	}

	if databaseInstance.Spec.Connection.Mysql != nil {
		result, err := r.reconcileMysqlDatabaseSchema(databaseInstance)
		if err != nil {
			logger.Error(err)
		}
		return result, err
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileDatabaseSchema) getInstance(request reconcile.Request) (*databasesv1alpha4.Database, error) {
	instance := &databasesv1alpha4.Database{}
	err := r.Get(context.Background(), request.NamespacedName, instance)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get databasesv1alpha4 instance")
	}

	return instance, nil
}

func (r *ReconcileDatabaseSchema) reconcileMysqlDatabaseSchema(databaseInstance *databasesv1alpha4.Database) (reconcile.Result, error) {
	logger.Debug("reconciling mysql database schema")

	ctx := context.Background()

	_, connectionURI, err := databaseInstance.GetConnection(ctx)
	if err != nil {
		return reconcile.Result{}, errors.Wrap(err, "failed to get database connection")
	}

	cfg, err := mysqldriver.ParseDSN(connectionURI)
	if err != nil {
		return reconcile.Result{}, errors.Wrap(err, "failed to parse mysql connection uri")
	}

	// compare the databaseInstance defaultCharset and collation to the running database
	db, err := sql.Open("mysql", connectionURI)
	if err != nil {
		return reconcile.Result{}, errors.Wrap(err, "failed to open connection to mysql")
	}
	defer db.Close()

	db.SetMaxIdleConns(0)
	db.SetMaxOpenConns(1)

	query := `SELECT default_character_set_name, default_collation_name FROM information_schema.SCHEMATA 
	WHERE schema_name = ?`
	row := db.QueryRow(query, cfg.DBName)

	var defaultCharset, defaultCollation string
	if err := row.Scan(&defaultCharset, &defaultCollation); err != nil {
		return reconcile.Result{}, errors.Wrap(err, "failed to scan")
	}

	if databaseInstance.Spec.Connection.Mysql.DefaultCharset == defaultCharset && databaseInstance.Spec.Connection.Mysql.Collation == defaultCollation {
		logger.Debug("no action needed as the charset and collation already match")
		return reconcile.Result{}, nil
	}

	// if they are different, it might be the server default, but that's hard to tell from a connection
	// if the user passed "", then we want the database set to the default
	// https://dev.mysql.com/doc/refman/8.0/en/charset-database.html  <-- see this for application
	query = `show variables like 'collation_server'`
	row = db.QueryRow(query)

	var variableName string
	var defaultServerCollation, defaultServerCharset string
	if err := row.Scan(&variableName, &defaultServerCollation); err != nil {
		return reconcile.Result{}, errors.Wrap(err, "failed to scan server default collation")
	}

	query = `show variables like 'character_set_server'`
	row = db.QueryRow(query)
	if err := row.Scan(&variableName, &defaultServerCharset); err != nil {
		return reconcile.Result{}, errors.Wrap(err, "failed to scan server default charset")
	}

	/*
		from https://dev.mysql.com/doc/refman/8.0/en/charset-database.html

		MySQL chooses the database character set and database collation in the following manner:

			If both CHARACTER SET charset_name and COLLATE collation_name are specified, character set charset_name and collation collation_name are used.

			If CHARACTER SET charset_name is specified without COLLATE, character set charset_name and its default collation are used. To see the default collation for each character set, use the SHOW CHARACTER SET statement or query the INFORMATION_SCHEMA CHARACTER_SETS table.

			If COLLATE collation_name is specified without CHARACTER SET, the character set associated with collation_name and collation collation_name are used.

			Otherwise (neither CHARACTER SET nor COLLATE is specified), the server character set and server collation are used.

	*/

	if databaseInstance.Spec.Connection.Mysql.DefaultCharset != "" && databaseInstance.Spec.Connection.Mysql.Collation != "" {
		// If both CHARACTER SET charset_name and COLLATE collation_name are specified, character set charset_name and collation collation_name are used.
		query = fmt.Sprintf("alter database `%s` character set '%s' collate '%s'", cfg.DBName, databaseInstance.Spec.Connection.Mysql.DefaultCharset, databaseInstance.Spec.Connection.Mysql.Collation)
		if _, err := db.Exec(query); err != nil {
			return reconcile.Result{}, errors.Wrap(err, "failed to update database character set and collation")
		}

		return reconcile.Result{}, nil
	}

	if databaseInstance.Spec.Connection.Mysql.DefaultCharset != "" && databaseInstance.Spec.Connection.Mysql.Collation == "" {
		// If CHARACTER SET charset_name is specified without COLLATE, character set charset_name and its default collation are used. To see the default collation for each character set, use the SHOW CHARACTER SET statement or query the INFORMATION_SCHEMA CHARACTER_SETS table.
		query = `select DEFAULT_COLLATE_NAME from information_schema.character_sets where CHARACTER_SET_NAME = ?`
		row := db.QueryRow(query, databaseInstance.Spec.Connection.Mysql.DefaultCharset)
		var charsetDefaultCollation string
		if err := row.Scan(&charsetDefaultCollation); err != nil {
			return reconcile.Result{}, errors.Wrap(err, "failed to select default collation for character set")
		}

		query = fmt.Sprintf("alter database `%s` character set '%s' collate '%s'", cfg.DBName, databaseInstance.Spec.Connection.Mysql.DefaultCharset, charsetDefaultCollation)
		if _, err := db.Exec(query); err != nil {
			return reconcile.Result{}, errors.Wrap(err, "failed to update database character set and collation")
		}

		return reconcile.Result{}, nil
	}

	if databaseInstance.Spec.Connection.Mysql.DefaultCharset == "" && databaseInstance.Spec.Connection.Mysql.Collation != "" {
		// If COLLATE collation_name is specified without CHARACTER SET, the character set associated with collation_name and collation collation_name are used.
		query = `select CHARACTER_SET_NAME from information_schema.collations where COLLATION_NAME = ?`
		row := db.QueryRow(query, databaseInstance.Spec.Connection.Mysql.Collation)
		var collationCharset string
		if err := row.Scan(&collationCharset); err != nil {
			return reconcile.Result{}, errors.Wrap(err, "failed to select charset for collaction")
		}

		query = fmt.Sprintf("alter database `%s` character set '%s' collate '%s'", cfg.DBName, collationCharset, databaseInstance.Spec.Connection.Mysql.Collation)
		if _, err := db.Exec(query); err != nil {
			return reconcile.Result{}, errors.Wrap(err, "failed to update database character set and collation")
		}

		return reconcile.Result{}, nil
	}

	query = fmt.Sprintf("alter database `%s` character set '%s' collate '%s'", cfg.DBName, defaultServerCharset, defaultServerCollation)
	if _, err := db.Exec(query); err != nil {
		return reconcile.Result{}, errors.Wrap(err, "failed to update database character set and collation")
	}

	return reconcile.Result{}, nil
}
