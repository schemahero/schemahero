package shell

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/schemahero/schemahero/pkg/config"
	corev1 "k8s.io/api/core/v1"
	kuberneteserrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// StartShellPod will start a new pod in the namespace using the imagename
// The command will be sleep infinity, so the exec will have to start the
// desired command
// the caller is responsible for cleaning up this pod
func StartShellPod(ctx context.Context, namespace string, imageName string) (string, error) {
	pod := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    namespace,
			GenerateName: "schemahero-shell-",
		},
		Spec: corev1.PodSpec{
			// NodeSelector,
			// ServiceAccount
			RestartPolicy: corev1.RestartPolicyNever,
			Containers: []corev1.Container{
				{
					Image:           imageName,
					ImagePullPolicy: corev1.PullIfNotPresent,
					Name:            "shell",
					Command: []string{
						"sleep",
						"infinity",
					},
				},
			},
		},
	}

	cfg, err := config.GetRESTConfig()
	if err != nil {
		return "", errors.Wrap(err, "failed to get config")
	}

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return "", errors.Wrap(err, "failed to get clientset")
	}

	createdPod, err := clientset.CoreV1().Pods(namespace).Create(ctx, pod, metav1.CreateOptions{})
	if err != nil {
		return "", errors.Wrap(err, "failed to create pod")
	}

	// wait for the pod to start
	startedAt := time.Now()
	abortAt := startedAt.Add(time.Minute)
	pollInterval := time.Millisecond * 200

	for {
		if abortAt.Before(time.Now()) {
			return "", errors.New("shell pod did not start")
		}

		currentPod, err := clientset.CoreV1().Pods(namespace).Get(ctx, createdPod.Name, metav1.GetOptions{})
		if kuberneteserrors.IsNotFound(err) {
			time.Sleep(pollInterval)
			continue
		}

		if err != nil {
			return "", errors.Wrap(err, "failed to get pod")
		}

		numNotReady := 0
		numReady := 0
		for _, podCondition := range currentPod.Status.Conditions {
			if podCondition.Type == corev1.ContainersReady {
				if podCondition.Status == corev1.ConditionTrue {
					numReady++
				} else {
					numNotReady++
				}
			}
		}

		if numNotReady == 0 && numReady > 0 {
			return createdPod.Name, nil
		}
	}
}
