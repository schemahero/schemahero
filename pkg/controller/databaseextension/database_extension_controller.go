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

package databaseextension

import (
	"context"
	"time"

	databasesv1alpha4 "github.com/schemahero/schemahero/pkg/apis/databases/v1alpha4"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/database"
	"github.com/schemahero/schemahero/pkg/database/postgres"
	"github.com/schemahero/schemahero/pkg/logger"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	kuberneteserrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileDatabaseExtension{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

func add(mgr manager.Manager, r reconcile.Reconciler) error {
	c, err := controller.New("databaseextension-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &schemasv1alpha4.DatabaseExtension{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileDatabaseExtension{}

type ReconcileDatabaseExtension struct {
	client.Client
	scheme *runtime.Scheme
}

func (r *ReconcileDatabaseExtension) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	log := logger.FromContext(ctx).With(
		zap.String("kind", "databaseextension"),
		zap.String("name", request.Name),
		zap.String("namespace", request.Namespace),
	)

	databaseExtension := &schemasv1alpha4.DatabaseExtension{}
	err := r.Get(ctx, request.NamespacedName, databaseExtension)
	if err != nil {
		if kuberneteserrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	database, err := r.getDatabaseFromExtension(ctx, databaseExtension)
	if err != nil {
		if kuberneteserrors.IsNotFound(err) {
			log.Info("database not found, requeuing", zap.String("database", databaseExtension.Spec.Database))
			return reconcile.Result{
				Requeue:      true,
				RequeueAfter: time.Second * 10,
			}, nil
		}
		return reconcile.Result{}, err
	}

	driver, connectionURI, err := database.GetConnection(ctx)
	if err != nil {
		log.Error("failed to get database connection", zap.Error(err))
		return reconcile.Result{}, err
	}

	if driver != "postgres" || databaseExtension.Spec.Postgres == nil {
		log.Info("not a postgres database or no postgres extension specified, skipping")
		return reconcile.Result{}, nil
	}

	statements, err := postgres.CreateExtensionStatements([]*schemasv1alpha4.PostgresDatabaseExtension{databaseExtension.Spec.Postgres})
	if err != nil {
		log.Error("failed to generate extension statements", zap.Error(err))
		return reconcile.Result{}, err
	}

	db := database.Database{
		Driver: driver,
		URI:    connectionURI,
	}

	if err := db.ApplySync(statements); err != nil {
		databaseExtension.Status.Phase = "Failed"
		databaseExtension.Status.Message = err.Error()
		if updateErr := r.Status().Update(ctx, databaseExtension); updateErr != nil {
			log.Error("failed to update status", zap.Error(updateErr))
			return reconcile.Result{}, updateErr
		}
		log.Error("failed to apply extension", zap.Error(err))
		return reconcile.Result{}, err
	}

	databaseExtension.Status.Phase = "Applied"
	databaseExtension.Status.AppliedAt = time.Now().Unix()
	databaseExtension.Status.Message = "Extension successfully applied"

	if err := r.Status().Update(ctx, databaseExtension); err != nil {
		log.Error("failed to update status", zap.Error(err))
		return reconcile.Result{}, err
	}

	log.Info("extension successfully applied")
	return reconcile.Result{}, nil
}

func (r *ReconcileDatabaseExtension) getDatabaseFromExtension(ctx context.Context, databaseExtension *schemasv1alpha4.DatabaseExtension) (*databasesv1alpha4.Database, error) {
	database := &databasesv1alpha4.Database{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      databaseExtension.Spec.Database,
		Namespace: databaseExtension.Namespace,
	}, database)
	if err != nil {
		return nil, err
	}

	return database, nil
}
