package database

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	databasesv1alpha2 "github.com/schemahero/schemahero/pkg/apis/databases/v1alpha2"
	schemasv1alpha2 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha2"
	schemaheroscheme "github.com/schemahero/schemahero/pkg/client/schemaheroclientset/scheme"
	schemasclientv1alpha2 "github.com/schemahero/schemahero/pkg/client/schemaheroclientset/typed/schemas/v1alpha2"
	"gopkg.in/src-d/go-git.v4"
	gitconfig "gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	kuberneteserrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var (
	gitopsPlanLoop *GitOpsPlanLoop
)

type GitOpsPlanLoop struct {
	workingDir      string
	isAccessAllowed bool
	authMethod      transport.AuthMethod
	pulls           map[string]string
}

func (r *ReconcileDatabase) ensureGitOpsPlan(instance *databasesv1alpha2.Database) error {
	if instance.Status.GitopsPlanStatus == "" {
		if gitopsPlanLoop != nil {
			return nil
		}

		fmt.Println("initializing gitops planning loop")

		gitopsPlanLoop = &GitOpsPlanLoop{}

		// attempt to connect to the repo to see if we have access
		hasPrintedPublicKey := false

		for gitopsPlanLoop.isAccessAllowed == false {
			authMethod, publicKeyBytes, err := r.mustGetPrivateKey(instance)
			if err != nil {
				return errors.Wrap(err, "failed to get private key")
			}

			gitopsPlanLoop.authMethod = authMethod

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
				gitopsPlanLoop.isAccessAllowed = true
			}
		}

		workingDir, err := ioutil.TempDir("", "shplan")
		if err != nil {
			return errors.Wrap(err, "failed to crreate temp dir")
		}
		defer os.RemoveAll(workingDir)

		gitopsPlanLoop.workingDir = workingDir

		_, err = git.PlainClone(gitopsPlanLoop.workingDir, false, &git.CloneOptions{
			URL:               instance.GitOps.URL,
			RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
			Auth:              gitopsPlanLoop.authMethod,
		})
		if err != nil {
			return errors.Wrap(err, "failed to perform initial clone")
		}

		for {
			repo, err := git.PlainOpen(gitopsPlanLoop.workingDir)
			if err != nil {
				return errors.Wrap(err, "failed to open git repo")
			}

			fetchOptions := git.FetchOptions{
				RemoteName: "origin",
				Auth:       gitopsPlanLoop.authMethod,
				RefSpecs: []gitconfig.RefSpec{
					"+refs/pull/*/head:refs/remotes/origin/pr/*",
				},
			}
			if err := repo.Fetch(&fetchOptions); err != nil {
				if err == git.NoErrAlreadyUpToDate {
					time.Sleep(time.Second * 10)
					continue
				}

				return errors.Wrap(err, "failed to fetch")
			}

			referencesIter, err := repo.References()
			if err != nil {
				return errors.Wrap(err, "failed to get references iterator")
			}

			pulls := map[string]string{}

			err = referencesIter.ForEach(func(ref *plumbing.Reference) error {
				if ref.Type() == plumbing.SymbolicReference {
					return nil
				}

				if !strings.HasPrefix(ref.Name().String(), "refs/remotes/origin/pr/") {
					return nil
				}

				s := strings.Split(ref.Name().String(), "/")
				prNumber := s[len(s)-1]

				pulls[prNumber] = ref.Hash().String()

				return nil
			})
			if err != nil {
				return errors.Wrap(err, "failed to walk refs")
			}

			// look for new or updated pulls
			for prNumber, currentHash := range pulls {
				knownHash, ok := gitopsPlanLoop.pulls[prNumber]
				if !ok || currentHash != knownHash {
					// check out this branch
					w, err := repo.Worktree()
					if err != nil {
						return errors.Wrap(err, "failed to get working tree")
					}

					checkoutOptions := git.CheckoutOptions{
						Hash: plumbing.NewHash(currentHash),
					}
					if err := w.Checkout(&checkoutOptions); err != nil {
						return errors.Wrap(err, "failed to check out hash")
					}

					plan, err := r.executePlan(workingDir, instance, currentHash)
					if err != nil {
						return errors.Wrap(err, "failed to execute plan")
					}

					fmt.Printf("plan = %s\n", plan)
				}
			}

			gitopsPlanLoop.pulls = pulls

			// Write the current state of the pulls ot the status object to
			// allow quicker startup in the future and less thrashing

			time.Sleep(time.Second * 10)
		}
	}

	return nil
}

func (r *ReconcileDatabase) executePlan(workingDir string, instance *databasesv1alpha2.Database, hash string) (string, error) {
	plan := ""

	schemaheroscheme.AddToScheme(scheme.Scheme)
	decode := scheme.Codecs.UniversalDeserializer().Decode

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

		// We could (maybe should) deploy this and let the table reconciler run the plan...
		// but that async nature will create a lot more code, and these are the same codebase/project.
		// so let's just call over to the table controller directly for now

		table.Spec.IsPlan = true
		table.Name = fmt.Sprintf("%s-%s", table.Name, hash[0:7])
		if table.Namespace == "" {
			table.Namespace = "default"
		}

		tablePlan, err := r.getTablePlan(table, instance)
		if err != nil {
			return err
		}

		if len(plan) > 0 {
			plan = plan + "\n"
		}

		plan = plan + tablePlan
		return nil
	})

	if err != nil {
		return "", errors.Wrap(err, "failed to walk file path")
	}

	return plan, nil
}

func (r *ReconcileDatabase) getTablePlan(table *schemasv1alpha2.Table, instance *databasesv1alpha2.Database) (string, error) {

	// we need to deploy the table plan as a CR here because we need all of the logic that
	// the reconcile loop in the schema package manages...

	// this definitely makes this process a lot more complex than it really should be

	existingObj := &schemasv1alpha2.Table{}
	err := r.Get(context.Background(), types.NamespacedName{Name: table.Name, Namespace: table.Namespace}, existingObj)
	if err != nil && !kuberneteserrors.IsNotFound(err) {
		return "", errors.Wrap(err, "failed to look for existing object")
	}

	if kuberneteserrors.IsNotFound(err) {
		// create
		err = r.Create(context.Background(), table)
		if err != nil {
			return "", errors.Wrap(err, "failed to create table object")
		}
	} else {
		// TODO how can we handle this?
		// Delete and recreate?
		// This won't happen in the "happy path" gitops flow, but certainly can happen in
		// other use cases
		return "", errors.New("cannot update existing table plan")
	}

	// Watch the object that we just created using an informer to get the plan
	plan := ""

	cfg, err := config.GetConfig()
	if err != nil {
		return "", errors.Wrap(err, "failed to get config")
	}
	schemasClient, err := schemasclientv1alpha2.NewForConfig(cfg)
	if err != nil {
		return "", errors.Wrap(err, "failed to create schema client")
	}

	watchlist := cache.NewListWatchFromClient(schemasClient.RESTClient(), "tables", table.Namespace, fields.Everything())
	resyncPeriod := 10 * time.Second
	_, controller := cache.NewInformer(watchlist, &schemasv1alpha2.Table{}, resyncPeriod,
		cache.ResourceEventHandlerFuncs{
			UpdateFunc: func(oldObj interface{}, newObj interface{}) {
				updatedTable := newObj.(*schemasv1alpha2.Table)
				if updatedTable.Status.Plan != "" {
					plan = updatedTable.Status.Plan
				}
			},
		},
	)

	ctx := context.Background()
	controller.Run(ctx.Done())

	start := time.Now()
	abort := start.Add(time.Minute * 2)
	for plan == "" {
		if time.Now().After(abort) {
			return "", errors.New("timeout waiting for plan")
		}

		if plan != "" {
			break
		}
	}

	return plan, nil
}
