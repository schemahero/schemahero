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
	"fmt"

	"github.com/pkg/errors"
	databasesv1alpha4 "github.com/schemahero/schemahero/pkg/apis/databases/v1alpha4"
	"github.com/schemahero/schemahero/pkg/logger"
	corev1 "k8s.io/api/core/v1"
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

	if err := r.reconcileServiceAccount(ctx, databaseInstance); err != nil {
		return errors.Wrap(err, "failed to reconcile serviceaccount")
	}

	return nil
}

func (r *ReconcileDatabase) reconcileRBACRoleBinding(ctx context.Context, databaseInstance *databasesv1alpha4.Database) error {
	roleBindingName := fmt.Sprintf("schemahero-%s", databaseInstance.Name)
	roleName := fmt.Sprintf("schemahero-%s", databaseInstance.Name)
	serviceAccountName := fmt.Sprintf("schemahero-%s", databaseInstance.Name)

	existingRoleBinding := rbacv1.RoleBinding{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      roleBindingName,
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
				Name:      roleBindingName,
				Namespace: databaseInstance.Namespace,
			},
			Subjects: []rbacv1.Subject{
				{
					Kind:      "ServiceAccount",
					Name:      serviceAccountName,
					Namespace: databaseInstance.Namespace,
				},
			},
			RoleRef: rbacv1.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "Role",
				Name:     roleName,
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
	roleName := fmt.Sprintf("schemahero-%s", databaseInstance.Name)

	existingRole := rbacv1.Role{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      roleName,
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
				Name:      roleName,
				Namespace: databaseInstance.Namespace,
			},
			Rules: []rbacv1.PolicyRule{
				{
					APIGroups: []string{""},
					Resources: []string{"secrets", "serviceaccounts"},
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

func (r *ReconcileDatabase) reconcileServiceAccount(ctx context.Context, databaseInstance *databasesv1alpha4.Database) error {
	serviceAccountName := fmt.Sprintf("schemahero-%s", databaseInstance.Name)

	existingServiceAccount := corev1.ServiceAccount{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      serviceAccountName,
		Namespace: databaseInstance.Namespace,
	}, &existingServiceAccount)
	if kuberneteserrors.IsNotFound(err) {
		// create the service account
		serviceAccount := corev1.ServiceAccount{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "ServiceAccount",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      serviceAccountName,
				Namespace: databaseInstance.Namespace,
			},
		}

		if err := controllerutil.SetControllerReference(databaseInstance, &serviceAccount, r.scheme); err != nil {
			return errors.Wrap(err, "failed to set owner ref on service account")
		}

		if err := r.Create(ctx, &serviceAccount); err != nil {
			return errors.Wrap(err, "failed to create service account")
		}
	} else if err != nil {
		return errors.Wrap(err, "failed to get existing service account")
	} else {
		// update
		logger.Error(errors.New("updating serviceaccount is not implemented"))
	}

	return nil
}
