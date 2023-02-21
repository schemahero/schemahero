package installer

import (
	"bytes"
	"context"

	"github.com/pkg/errors"
	"github.com/schemahero/schemahero/pkg/client/schemaheroclientset/scheme"
	rbacv1 "k8s.io/api/rbac/v1"
	kuberneteserrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/kubernetes"
)

func clusterRoleYAML() ([]byte, error) {
	s := json.NewYAMLSerializer(json.DefaultMetaFactory, scheme.Scheme, scheme.Scheme)
	var result bytes.Buffer
	if err := s.Encode(clusterRole(), &result); err != nil {
		return nil, errors.Wrap(err, "failed to marshal cluster role")
	}

	return result.Bytes(), nil
}

func ensureClusterRole(ctx context.Context, clientset *kubernetes.Clientset) error {
	_, err := clientset.RbacV1().ClusterRoles().Get(ctx, "schemahero-role", metav1.GetOptions{})
	if err != nil {
		if !kuberneteserrors.IsNotFound(err) {
			return errors.Wrap(err, "failed to get clusterrole")
		}

		_, err := clientset.RbacV1().ClusterRoles().Create(ctx, clusterRole(), metav1.CreateOptions{})
		if err != nil {
			return errors.Wrap(err, "failed to create cluster role")
		}
	}

	return nil
}

func clusterRole() *rbacv1.ClusterRole {
	clusterRole := &rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "ClusterRole",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "schemahero-role",
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"apps"},
				Resources: []string{"deployments", "statefulsets"},
				Verbs:     metav1.Verbs{"get", "list", "watch", "create", "update", "patch", "delete"},
			},
			{
				APIGroups: []string{"apps"},
				Resources: []string{"deployments/status", "statefulset/status"},
				Verbs:     metav1.Verbs{"get", "update", "patch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"pods"},
				Verbs:     metav1.Verbs{"get", "list", "watch", "create", "update", "patch", "delete"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"pods/log"},
				Verbs:     metav1.Verbs{"get"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"secrets"},
				Verbs:     metav1.Verbs{"get", "list", "watch", "create", "update", "patch", "delete"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"serviceaccounts"},
				Verbs:     metav1.Verbs{"get", "list", "create", "update", "delete", "watch"},
			},
			{
				APIGroups: []string{"rbac.authorization.k8s.io"},
				Resources: []string{"roles", "rolebindings"},
				Verbs:     metav1.Verbs{"get", "list", "create", "update", "delete", "watch"},
			},
			{
				APIGroups: []string{"admissionregistration.k8s.io"},
				Resources: []string{"mutatingwebhookconfigurations", "validatingwebhookconfigurations"},
				Verbs:     metav1.Verbs{"get", "list", "watch", "create", "update", "patch", "delete"},
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
			{
				APIGroups: []string{"schemas.schemahero.io"},
				Resources: []string{"views"},
				Verbs:     metav1.Verbs{"get", "list", "watch", "create", "update", "patch", "delete"},
			},
			{
				APIGroups: []string{"schemas.schemahero.io"},
				Resources: []string{"views/status"},
				Verbs:     metav1.Verbs{"get", "update", "patch"},
			},
		},
	}

	return clusterRole
}

func clusterRoleBindingYAML(namespace string) ([]byte, error) {
	s := json.NewYAMLSerializer(json.DefaultMetaFactory, scheme.Scheme, scheme.Scheme)
	var result bytes.Buffer
	if err := s.Encode(clusterRoleBinding(namespace), &result); err != nil {
		return nil, errors.Wrap(err, "failed to marshal cluster role binding")
	}

	return result.Bytes(), nil
}

func ensureClusterRoleBinding(ctx context.Context, clientset *kubernetes.Clientset, namespace string) error {
	_, err := clientset.RbacV1().ClusterRoleBindings().Get(ctx, "schemahero-rolebinding", metav1.GetOptions{})
	if err != nil {
		if !kuberneteserrors.IsNotFound(err) {
			return errors.Wrap(err, "failed to get clusterrolebinding")
		}

		_, err := clientset.RbacV1().ClusterRoleBindings().Create(ctx, clusterRoleBinding(namespace), metav1.CreateOptions{})
		if err != nil {
			return errors.Wrap(err, "failed to create cluster rolebinding")
		}
	}

	return nil
}

func clusterRoleBinding(namespace string) *rbacv1.ClusterRoleBinding {
	clusterRoleBinding := &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "ClusterRoleBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "schemahero-rolebinding",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "default",
				Namespace: namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "schemahero-role",
		},
	}

	return clusterRoleBinding
}
