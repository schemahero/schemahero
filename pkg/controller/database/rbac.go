/*
Copyright 2019 Replicated, Inc.

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

	"github.com/pkg/errors"
	databasesv1alpha4 "github.com/schemahero/schemahero/pkg/apis/databases/v1alpha4"
	"github.com/schemahero/schemahero/pkg/logger"
	rbacv1 "k8s.io/api/rbac/v1"
	kuberneteserrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *ReconcileDatabase) reconcileRBAC(ctx context.Context, databaseInstance *databasesv1alpha4.Database) error {
	if err := r.reconcileRBACRole(ctx, databaseInstance); err != nil {
		return errors.Wrap(err, "failed to reconcile role")
	}

	if err := r.reconcileRBACRoleBinding(ctx, databaseInstance); err != nil {
		return errors.Wrap(err, "failed to reconcile rolebinding")
	}

	return nil
}

func (r *ReconcileDatabase) reconcileRBACRoleBinding(ctx context.Context, databaseInstance *databasesv1alpha4.Database) error {
	existingRoleBinding := rbacv1.RoleBinding{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      "schemahero-rolebinding",
		Namespace: databaseInstance.Namespace,
	}, &existingRoleBinding)
	if kuberneteserrors.IsNotFound(err) {
		// create
		roleBinding := rbacv1.RoleBinding{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "rbac.authorization.k8s.io/v1",
				Kind:       "RoleBinding",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "schemahero-rolebinding",
				Namespace: databaseInstance.Namespace,
			},
			Subjects: []rbacv1.Subject{
				{
					Kind:      "ServiceAccount",
					Name:      "schemahero",
					Namespace: databaseInstance.Namespace,
				},
			},
			RoleRef: rbacv1.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "Role",
				Name:     "schemahero-role",
			},
		}

		if err := controllerutil.SetControllerReference(databaseInstance, &roleBinding, r.scheme); err != nil {
			return errors.Wrap(err, "failed to set owner ref on rolebinding")
		}

		if err := r.Create(ctx, &roleBinding); err != nil {
			return errors.Wrap(err, "failed to create rolebinding")
		}
	} else if err != nil {
		return errors.Wrap(err, "failed to check rolebinding")
	} else {
		// update
		logger.Error(errors.New("updating rolebinding is not implemented"))
	}

	return nil
}

func (r *ReconcileDatabase) reconcileRBACRole(ctx context.Context, databaseInstance *databasesv1alpha4.Database) error {
	existingRole := rbacv1.Role{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      "schemahero-role",
		Namespace: databaseInstance.Namespace,
	}, &existingRole)
	if kuberneteserrors.IsNotFound(err) {
		// create the role
		role := rbacv1.Role{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "rbac.authorization.k8s.io/v1",
				Kind:       "Role",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "schemahero-role",
				Namespace: databaseInstance.Namespace,
			},
			Rules: []rbacv1.PolicyRule{
				{
					APIGroups: []string{""},
					Resources: []string{"secrets"},
					Verbs:     metav1.Verbs{"get"},
				},
				{
					APIGroups: []string{"databases.schemahero.io"},
					Resources: []string{"databases"},
					Verbs:     metav1.Verbs{"get", "list", "watch", "create", "update", "patch", "delete"},
				},
				{
					APIGroups: []string{"databases.schemahero.io"},
					Resources: []string{"databases/status"},
					Verbs:     metav1.Verbs{"get", "update", "patch"},
				},
				{
					APIGroups: []string{"schemas.schemahero.io"},
					Resources: []string{"migrations"},
					Verbs:     metav1.Verbs{"get", "list", "watch", "create", "update", "patch", "delete"},
				},
				{
					APIGroups: []string{"schemas.schemahero.io"},
					Resources: []string{"migrations/status"},
					Verbs:     metav1.Verbs{"get", "update", "patch"},
				},
				{
					APIGroups: []string{"schemas.schemahero.io"},
					Resources: []string{"tables"},
					Verbs:     metav1.Verbs{"get", "list", "watch", "create", "update", "patch", "delete"},
				},
				{
					APIGroups: []string{"schemas.schemahero.io"},
					Resources: []string{"tables/status"},
					Verbs:     metav1.Verbs{"get", "update", "patch"},
				},
			},
		}

		if err := controllerutil.SetControllerReference(databaseInstance, &role, r.scheme); err != nil {
			return errors.Wrap(err, "failed to set owner ref on rolebinding")
		}

		if err := r.Create(ctx, &role); err != nil {
			return errors.Wrap(err, "failed to create role")
		}
	} else if err != nil {
		return errors.Wrap(err, "failed to get existing role")
	} else {
		// update
		logger.Error(errors.New("updating role is not implemented"))
	}

	return nil
}
