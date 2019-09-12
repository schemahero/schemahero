package database

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	databasesv1alpha2 "github.com/schemahero/schemahero/pkg/apis/databases/v1alpha2"
	schemasv1alpha2 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha2"
	schemaheroscheme "github.com/schemahero/schemahero/pkg/client/schemaheroclientset/scheme"
	// schemasclientv1alpha2 "github.com/schemahero/schemahero/pkg/client/schemaheroclientset/typed/schemas/v1alpha2"
	"golang.org/x/crypto/ssh"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	srcdssh "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
	corev1 "k8s.io/api/core/v1"
	kuberneteserrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	// "sigs.k8s.io/controller-runtime/pkg/client/config"
)

var (
	gitopsLoop *GitOpsLoop
)

type GitOpsLoop struct {
	workingDir      string
	isAccessAllowed bool
	authMethod      transport.AuthMethod
	lastCommit      string
}

func (r *ReconcileDatabase) ensureGitOps(instance *databasesv1alpha2.Database) error {
	if instance.Status.GitRepoStatus == "" {
		if gitopsLoop != nil {
			return nil
		}

		fmt.Println("initializing gitops loop")

		gitopsLoop := &GitOpsLoop{}

		// attempt to connect to the repo to see if we have access
		hasPrintedPublicKey := false

		for gitopsLoop.isAccessAllowed == false {
			authMethod, publicKeyBytes, err := r.mustGetPrivateKey(instance)
			if err != nil {
				return errors.Wrap(err, "failed to get private key")
			}

			gitopsLoop.authMethod = authMethod

			err = testRepoAccess(instance.GitOps.URL, authMethod)
			if err != nil {
				if !hasPrintedPublicKey {
					fmt.Printf("Cannot access %s. Please add the followinrg public key to the repo as a deploy key\n\n%s\n\n", instance.GitOps.URL, publicKeyBytes)
				}
				hasPrintedPublicKey = true
				time.Sleep(time.Second * 10)
			} else {
				if hasPrintedPublicKey {
					fmt.Printf("Access to repo is functional. Continuing to set up gitops loop\n")
				}
				gitopsLoop.isAccessAllowed = true
			}
		}

		branch := instance.GitOps.Branch
		if branch == "" {
			branch = "master"
		}

		pollInterval := time.Second * 10

		if instance.GitOps.PollInterval != "" {
			destiredPollInterval, err := time.ParseDuration(instance.GitOps.PollInterval)
			if err != nil {
				return errors.Wrap(err, "failed to parse duration")
			}

			pollInterval = destiredPollInterval
		}

		workingDir, err := ioutil.TempDir("", "sh")
		if err != nil {
			return errors.Wrap(err, "failed to create temp dir")
		}
		defer os.RemoveAll(workingDir)

		gitopsLoop.workingDir = workingDir

		repo, err := git.PlainClone(gitopsLoop.workingDir, false, &git.CloneOptions{
			URL:               instance.GitOps.URL,
			RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
			Auth:              gitopsLoop.authMethod,
		})
		if err != nil {
			return errors.Wrap(err, "failed to perform initial clone")
		}
		ref, err := repo.Head()
		if err != nil {
			return errors.Wrap(err, "failed to get head commit initial")
		}
		gitopsLoop.lastCommit = ref.String()

		if err := r.handleGitOpsCommit(gitopsLoop.workingDir); err != nil {
			return errors.Wrap(err, "failed to handle initial commit")
		}

		fmt.Printf("Gitops loop started	\n")

		for {
			repo, err := git.PlainOpen(gitopsLoop.workingDir)
			if err != nil {
				return errors.Wrap(err, "failed to open git repo")
			}

			w, err := repo.Worktree()
			if err != nil {
				return errors.Wrap(err, "failed to get working tree")
			}

			pullOptions := git.PullOptions{
				RemoteName: "origin",
				Auth:       gitopsLoop.authMethod,
			}
			if err = w.Pull(&pullOptions); err != nil {
				if err == git.NoErrAlreadyUpToDate {
					time.Sleep(pollInterval)
					continue
				}

				return errors.Wrap(err, "failed to pull")
			}

			ref, err := repo.Head()
			if err != nil {
				return errors.Wrap(err, "failed to get head commit")
			}

			gitopsLoop.lastCommit = ref.String()

			if err := r.handleGitOpsCommit(gitopsLoop.workingDir); err != nil {
				return errors.Wrap(err, "failed to handle initial commit")
			}

			time.Sleep(pollInterval)
		}
	}

	return nil
}

func (r *ReconcileDatabase) mustGetPrivateKey(instance *databasesv1alpha2.Database) (transport.AuthMethod, []byte, error) {
	secret := corev1.Secret{}
	secretNamespacedName := types.NamespacedName{
		Name:      fmt.Sprintf("%s-key", instance.Name),
		Namespace: instance.Namespace,
	}
	if err := r.Get(context.Background(), secretNamespacedName, &secret); err != nil {
		if !kuberneteserrors.IsNotFound(err) {
			return nil, nil, errors.Wrap(err, "failed to get gitops secret")
		}

		privateKey, publicKey, err := generateKeyPair()
		if err != nil {
			return nil, nil, errors.Wrap(err, "failed to generate keypair")
		}

		secret = corev1.Secret{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "Secret",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretNamespacedName.Name,
				Namespace: instance.Namespace,
			},
			Data: map[string][]byte{
				"private": privateKey,
				"public":  publicKey,
			},
		}

		if err := r.Create(context.Background(), &secret); err != nil {
			return nil, nil, errors.Wrap(err, "failed to create key secret")
		}
	}

	signer, err := ssh.ParsePrivateKey([]byte(secret.Data["private"]))
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to parse private key")
	}

	auth := srcdssh.PublicKeys{User: "git", Signer: signer}
	return &auth, secret.Data["public"], nil
}

func testRepoAccess(url string, authMethod transport.AuthMethod) error {
	tmpDir, err := ioutil.TempDir("", "sh")
	if err != nil {
		return errors.Wrap(err, "failed to create temp dir")
	}
	defer os.RemoveAll(tmpDir)

	_, err = git.PlainClone(tmpDir, true, &git.CloneOptions{
		URL:               url,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		Auth:              authMethod,
	})
	if err != nil {
		return errors.Wrap(err, "failed to test repo access")
	}

	return nil
}

func generateKeyPair() ([]byte, []byte, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to generate private key")
	}

	publicRsaKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to generate public key")
	}

	publicKeyBytes := ssh.MarshalAuthorizedKey(publicRsaKey)

	privDER := x509.MarshalPKCS1PrivateKey(privateKey)
	privBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privDER,
	}
	privateKeyBytes := pem.EncodeToMemory(&privBlock)

	return privateKeyBytes, publicKeyBytes, nil
}

func (r *ReconcileDatabase) handleGitOpsCommit(workingDir string) error {
	// walk, looking for table schemas...we need to reconcile each
	// relaying on the reconcile loop to diff

	schemaheroscheme.AddToScheme(scheme.Scheme)
	decode := scheme.Codecs.UniversalDeserializer().Decode

	// cfg, err := config.GetConfig()
	// if err != nil {
	// 	return errors.Wrap(err, "failed to get client config")
	// }
	// schemasClient, err := schemasclientv1alpha2.NewForConfig(cfg)
	// if err != nil {
	// 	return errors.Wrap(err, "failewd to create clientset")
	// }

	err := filepath.Walk(workingDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		content, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		obj, gvk, err := decode(content, nil, nil)
		if err != nil {
			return nil // ignore, this isn't a schemahero file
		}

		if gvk == nil {
			return nil // ignore, this isn't a schemahero file
		}

		if gvk.Group != "schemas.schemahero.io" || gvk.Version != "v1alpha2" || gvk.Kind != "Table" {
			return nil
		}

		table := obj.(*schemasv1alpha2.Table)

		if table.Namespace == "" {
			table.Namespace = "default"
		}

		existingObj := &schemasv1alpha2.Table{}
		err = r.Get(context.Background(), types.NamespacedName{Name: table.Name, Namespace: table.Namespace}, existingObj)
		if err != nil && !kuberneteserrors.IsNotFound(err) {
			return errors.Wrap(err, "failed to look for existing object")
		}

		if kuberneteserrors.IsNotFound(err) {
			// create
			err = r.Create(context.Background(), table)
			if err != nil {
				return errors.Wrap(err, "failed to create table object")
			}
		} else {
			// update
			err = r.Update(context.Background(), table)
			if err != nil {
				return errors.Wrap(err, "failed to update table object")
			}
		}

		return nil
	})

	if err != nil {
		return errors.Wrap(err, "failed to walk file path")
	}

	return nil
}
