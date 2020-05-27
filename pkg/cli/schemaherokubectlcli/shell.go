package schemaherokubectlcli

import (
	"context"
	"errors"
	"fmt"
	"os"

	databasesclientv1alpha4 "github.com/schemahero/schemahero/pkg/client/schemaheroclientset/typed/databases/v1alpha4"
	"github.com/schemahero/schemahero/pkg/database/mysql"
	"github.com/schemahero/schemahero/pkg/shell"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
	kuberneteserrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func ShellCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "shell",
		Short:         "create a new shell to the database",
		Long:          `...`,
		SilenceErrors: true,
		SilenceUsage:  true,
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			v := viper.GetViper()

			if len(args) == 0 {
				return errors.New("database name required")
			}

			databaseName := args[0]
			namespace := v.GetString("namespace")

			if namespace == "" {
				namespace = corev1.NamespaceDefault
			}

			cfg, err := config.GetConfig()
			if err != nil {
				return err
			}

			clientset, err := kubernetes.NewForConfig(cfg)
			if err != nil {
				return err
			}

			databasesClient, err := databasesclientv1alpha4.NewForConfig(cfg)
			if err != nil {
				return err
			}

			database, err := databasesClient.Databases(namespace).Get(ctx, databaseName, metav1.GetOptions{})
			if kuberneteserrors.IsNotFound(err) {
				return fmt.Errorf("database %q not found in %q namespace", databaseName, namespace)
			}
			if err != nil {
				return err
			}

			if !database.Spec.EnableShellCommand {
				return fmt.Errorf("shell command is not allowed for database %q", databaseName)
			}

			podImage := v.GetString("image")
			podCommand := []string{}
			// podArgs := []string{}

			if database.Spec.Connection.Postgres != nil {
				// TODO versions
				podImage = "postgres:11"

				connectionURI, err := database.Spec.Connection.Postgres.URI.Read(clientset, namespace)
				if err != nil {
					return err
				}
				podCommand = []string{
					"psql",
					connectionURI,
				}
			} else if database.Spec.Connection.Mysql != nil {
				// TODO versions
				podImage = "mysql:latest"

				connectionURI, err := database.Spec.Connection.Mysql.URI.Read(clientset, namespace)
				if err != nil {
					return err
				}

				username, err := mysql.UsernameFromURI(connectionURI)
				if err != nil {
					return err
				}

				password, err := mysql.PasswordFromURI(connectionURI)
				if err != nil {
					return err
				}

				hostname, err := mysql.HostnameFromURI(connectionURI)
				if err != nil {
					return err
				}

				database, err := mysql.DatabaseNameFromURI(connectionURI)
				if err != nil {
					return err
				}

				// TODO add port
				podCommand = []string{
					"mysql",
					"-u",
					username,
					fmt.Sprintf("-p%s", password),
					"-h",
					hostname,
					"-D",
					database,
				}
			}

			if podImage == "" {
				return errors.New("unable to determine image for shell -- consider specifying it with --image tag")
			}

			podName, err := shell.StartShellPod(ctx, namespace, podImage)
			if err != nil {
				return err
			}

			defer func() error {
				// a lot of time could pass, so we need to get a
				// new clientset here to re-authc
				cfg, err := config.GetConfig()
				if err != nil {
					return err
				}

				clientset, err := kubernetes.NewForConfig(cfg)
				if err != nil {
					return err
				}

				fmt.Println("deleting pod...")
				err = clientset.CoreV1().Pods(namespace).Delete(ctx, podName, metav1.DeleteOptions{})
				if err != nil {
					return err
				}

				fmt.Println("pod deleted.")
				return nil
			}()

			// exec and pipe the stdin/out/err
			req := clientset.CoreV1().RESTClient().Post().Resource("pods").Name(podName).Namespace(namespace).SubResource("exec")

			req.VersionedParams(&corev1.PodExecOptions{
				Container: "shell",
				Stdin:     true,
				Stdout:    true,
				Stderr:    true,
				TTY:       true,
				Command:   podCommand,
			}, scheme.ParameterCodec)

			exec, err := remotecommand.NewSPDYExecutor(cfg, "POST", req.URL())
			if err != nil {
				return err
			}
			err = exec.Stream(remotecommand.StreamOptions{
				Stdin:  os.Stdin,
				Stdout: os.Stdout,
				Stderr: os.Stderr,
				Tty:    true,
			})
			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().String("image", "", "the image to use when executing the shell. if not provided, one will be selected based on the database provider")

	return cmd
}
