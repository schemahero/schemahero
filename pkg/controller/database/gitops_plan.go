package database

import (
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
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"k8s.io/client-go/kubernetes/scheme"
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
				RefSpecs: []config.RefSpec{
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

					plan, err := executePlan(workingDir, instance)
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

func executePlan(workingDir string, instance *databasesv1alpha2.Database) (string, error) {
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

		tablePlan, err := getTablePlan(table, instance)
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

func getTablePlan(table *schemasv1alpha2.Table, instance *databasesv1alpha2.Database) (string, error) {
	return "", nil
}
