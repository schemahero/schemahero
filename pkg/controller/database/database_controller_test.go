package database

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"

	"github.com/schemahero/schemahero/pkg/apis"
	databasesv1alpha4 "github.com/schemahero/schemahero/pkg/apis/databases/v1alpha4"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	kuberneteserrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

// TestDbControllerTest uses envtest (see https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/envtest#Environment
// and https://github.com/kubernetes-sigs/controller-runtime/blob/master/FAQ.md#q-wheres-the-fake-client--how-do-i-use-it)
// This starts a real kube-api and etcd and runs the controller against it. The tests then
// become about submitting requests to the kube API and asserting on the state at the end
//
// You will need to run 'make envtest' to install the kube api/etcd binaries
func TestDbControllerTest(t *testing.T) {
	// create envtest instance, point to binaries
	// and load CRDs
	env := &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("..", "..", "..", "config", "crds", "v1"),
		},
		BinaryAssetsDirectory: filepath.Join(os.TempDir(), "kubebuilder", "bin"),
	}

	// start the kube api server and etcd
	cfg, err := env.Start()
	if err != nil {
		t.Fatalf("Failed to start environment: %v", err)
	}

	if cfg == nil {
		t.Fatal("Config was nil")
	}

	mgr, err := manager.New(cfg, manager.Options{})
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// add types to Scheme
	err = apis.AddToScheme(mgr.GetScheme())
	if err != nil {
		t.Fatalf("Failed to setup Scheme: %v", err)
	}

	// add the Controller to the Manager
	err = Add(mgr, "", "", true)
	if err != nil {
		t.Fatalf("Failed to add controller to manager: %v", err)
	}

	// start the controller manager in a goroutine so it doesn't
	// block envtest cleanup
	go func() {
		err = mgr.Start(signals.SetupSignalHandler())
		if err != nil {
			panic(err)
		}
	}()

	client := mgr.GetClient()

	tests := []struct {
		name                  string
		namespace             corev1.Namespace
		customLabels          map[string]string
		expectedLabels        map[string]string
		customNodeSelectors   map[string]string
		expectedNodeSelectors map[string]string
	}{
		{
			name: "creates_statefulset_with_default_labels",
			expectedLabels: map[string]string{
				"control-plane": "schemahero",
				"database":      "test",
			},
		},
		{
			name: "creates_statefulset_with_default_and_custom_labels",
			customLabels: map[string]string{
				"custom": "label",
			},
			expectedLabels: map[string]string{
				"control-plane": "schemahero",
				"database":      "test",
				"custom":        "label",
			},
			customNodeSelectors: map[string]string{
				"custom": "node-selector",
			},
			expectedNodeSelectors: map[string]string{
				"custom": "node-selector",
			},
		},
	}

	// wrapping the subtests in a t.Run so we wait till they're all
	// done before bringing down envtest
	t.Run("group", func(t *testing.T) {
		for _, test := range tests {
			test := test

			t.Run(test.name, func(t *testing.T) {
				t.Parallel()

				// create a new Namespace for each run to
				// create Objects in to avoid collisions
				ns, err := CreateNamespace(client)
				assert.NoError(t, err)

				db := databasesv1alpha4.Database{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: ns.Name,
					},
					Spec: databasesv1alpha4.DatabaseSpec{
						Template: &databasesv1alpha4.DatabaseTemplate{
							ObjectMeta: metav1.ObjectMeta{
								Labels: test.customLabels,
							},
						},
					},
				}

				// Use nil SchemaHero value in some tests to cover that case as well.
				if len(test.customNodeSelectors) > 0 {
					db.Spec.SchemaHero = &databasesv1alpha4.SchemaHero{
						NodeSelector: test.customNodeSelectors,
					}
				}

				err = client.Create(context.Background(), &db)
				assert.NoError(t, err)

				ss, err := GetStatefulSet(client, ns.Name, fmt.Sprintf("%s-controller", db.Name), 5, 1)
				assert.NoError(t, err)

				actualLabels := ss.Spec.Template.ObjectMeta.Labels
				assert.True(t, MapHasValues(test.expectedLabels, actualLabels), "Wanted: %v\ngot: %v", test.expectedLabels, actualLabels)

				actualNodeSelectors := ss.Spec.Template.Spec.NodeSelector
				assert.True(t, MapHasValues(test.expectedNodeSelectors, actualNodeSelectors), "Wanted: %v\ngot: %v", test.expectedNodeSelectors, actualNodeSelectors)
			})
		}
	})

	err = env.Stop()
	if err != nil {
		t.Fatalf("Failed to clean up environment: %v", err)
	}
}

func CreateNamespace(client runtimeclient.Client) (*corev1.Namespace, error) {
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "test-",
		},
	}
	err := client.Create(context.Background(), namespace)

	return namespace, err
}

// GetStatefulSet will retry 'retries' times, sleeping for 'sleepSeconds'
// to get the StatefulSet created by the Controller. This caters for the
// time between submitting a Database CRD and the Controller Reconcile loop
// finishing
func GetStatefulSet(client runtimeclient.Client, namespace string, name string, retries int, sleepSeconds time.Duration) (appsv1.StatefulSet, error) {
	ss := appsv1.StatefulSet{}
	lk := runtimeclient.ObjectKey{
		Namespace: namespace,
		Name:      name,
	}

	tries := 0
	var err error
	for tries < retries {
		err = client.Get(context.Background(), lk, &ss)
		if kuberneteserrors.IsNotFound(err) {
			time.Sleep(sleepSeconds * time.Second)
			tries++
		} else {
			break
		}
	}

	return ss, err
}

func MapHasValues(expected map[string]string, actual map[string]string) bool {
	for ek, ev := range expected {
		av, ok := actual[ek]
		if !ok {
			return false
		} else if ev != av {
			return false
		}
	}
	return true
}
