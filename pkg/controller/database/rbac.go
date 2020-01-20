package database

import (
	"context"
	"fmt"

	databasesv1alpha3 "github.com/schemahero/schemahero/pkg/apis/databases/v1alpha3"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	kuberneteserrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *ReconcileDatabase) ensureWatchRBAC(instance *databasesv1alpha3.Database) error {
	err := r.ensureRoleInNamespaceForInstance(instance)
	if err != nil {
		fmt.Printf("%#v\n", err)
		return err
	}

	err = r.ensureRoleBindingInNamespaceForInstance(instance)
	if err != nil {
		fmt.Printf("%#v\n", err)
		return err
	}

	err = r.ensureServiceAccountInNamespaceForInstance(instance)
	if err != nil {
		fmt.Printf("%#v\n", err)
		return err
	}

	return nil
}

func (r *ReconcileDatabase) ensureRoleInNamespaceForInstance(instance *databasesv1alpha3.Database) error {
	namespacedName := types.NamespacedName{
		Name:      instance.Name,
		Namespace: instance.Namespace,
	}
	existing := rbacv1.Role{}

	err := r.Get(context.TODO(), namespacedName, &existing)
	if err != nil {
		if kuberneteserrors.IsNotFound(err) {
			role := getRoleSpecForInstance(context.TODO(), namespacedName)
			if err := r.Create(context.Background(), role); err != nil {
				return err
			}
			if err := controllerutil.SetControllerReference(instance, role, r.scheme); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	return nil
}

func (r *ReconcileDatabase) ensureRoleBindingInNamespaceForInstance(instance *databasesv1alpha3.Database) error {
	namespacedName := types.NamespacedName{
		Name:      instance.Name,
		Namespace: instance.Namespace,
	}
	existing := rbacv1.RoleBinding{}

	err := r.Get(context.TODO(), namespacedName, &existing)
	if err != nil {
		if kuberneteserrors.IsNotFound(err) {
			roleBinding := getRoleBindingSpecForInstance(context.TODO(), namespacedName)
			if err := r.Create(context.Background(), roleBinding); err != nil {
				return err
			}
			if err := controllerutil.SetControllerReference(instance, roleBinding, r.scheme); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	return nil
}

func (r *ReconcileDatabase) ensureServiceAccountInNamespaceForInstance(instance *databasesv1alpha3.Database) error {
	namespacedName := types.NamespacedName{
		Name:      instance.Name,
		Namespace: instance.Namespace,
	}
	existing := corev1.ServiceAccount{}

	err := r.Get(context.TODO(), namespacedName, &existing)
	if err != nil {
		if kuberneteserrors.IsNotFound(err) {
			serviceAccount := getServiceAccountSpecForInstance(context.TODO(), namespacedName)
			if err := r.Create(context.Background(), serviceAccount); err != nil {
				return err
			}
			if err := controllerutil.SetControllerReference(instance, serviceAccount, r.scheme); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	return nil
}

func getRoleSpecForInstance(ctx context.Context, namespacedName types.NamespacedName) *rbacv1.Role {
	role := rbacv1.Role{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Role",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      namespacedName.Name,
			Namespace: namespacedName.Namespace,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups:     []string{"databases.schemahero.io"},
				Resources:     []string{"databases"},
				ResourceNames: []string{namespacedName.Name},
				Verbs:         metav1.Verbs{"get", "list", "update"},
			},
		},
	}

	return &role
}

func getRoleBindingSpecForInstance(ctx context.Context, namespacedName types.NamespacedName) *rbacv1.RoleBinding {
	roleBinding := rbacv1.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "RoleBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      namespacedName.Name,
			Namespace: namespacedName.Namespace,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      namespacedName.Name,
				Namespace: namespacedName.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     namespacedName.Name,
		},
	}

	return &roleBinding
}

func getServiceAccountSpecForInstance(ctx context.Context, namespacedName types.NamespacedName) *corev1.ServiceAccount {
	serviceAccount := corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ServiceAccount",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      namespacedName.Name,
			Namespace: namespacedName.Namespace,
		},
	}

	return &serviceAccount
}
