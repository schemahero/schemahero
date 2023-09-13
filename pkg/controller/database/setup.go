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
	databasesv1alpha4 "github.com/schemahero/schemahero/pkg/apis/databases/v1alpha4"
	"github.com/schemahero/schemahero/pkg/logger"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var tenSeconds = int64(10)

// Add creates a new Database Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager, managerImage string, managerTag string, debugLogs bool) error {
	return add(mgr, newReconciler(mgr, managerImage, managerTag, debugLogs), "databaseController")
}

// AddForDatabaseSchemasOnly creates a new Database Controller to watch for changes to a specific database object. This is how database-level schema
// changes are populated. Table and migration reconcile loops are separate, this is used for database options (charset, collate, etc)
func AddForDatabaseSchemasOnly(mgr manager.Manager, databaseNames []string) error {
	return add(mgr, newDatabaseSchemaReconciler(mgr, databaseNames), "databaseSchemaController")
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager, managerImage string, managerTag string, debugLogs bool) reconcile.Reconciler {
	return &ReconcileDatabase{
		Client:       mgr.GetClient(),
		scheme:       mgr.GetScheme(),
		managerImage: managerImage,
		managerTag:   managerTag,
		debugLogs:    debugLogs,
	}
}

func newDatabaseSchemaReconciler(mgr manager.Manager, databaseNames []string) reconcile.Reconciler {
	return &ReconcileDatabaseSchema{
		Client:        mgr.GetClient(),
		scheme:        mgr.GetScheme(),
		databaseNames: databaseNames,
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler, name string) error {
	logger.Debugf("adding %s to manager", name)

	// Create a new controller
	c, err := controller.New("database-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to Database kinds
	err = c.Watch(source.Kind(mgr.GetCache(), &databasesv1alpha4.Database{}), &handler.EnqueueRequestForObject{})

	return err
}
