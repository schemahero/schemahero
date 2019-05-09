package runner

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
	"sigs.k8s.io/kind/pkg/cluster"
	"sigs.k8s.io/kind/pkg/cluster/config/encoding"
	"sigs.k8s.io/kind/pkg/cluster/create"
)

type Runner struct {
	Viper *viper.Viper
}

func NewRunner() *Runner {
	return &Runner{
		Viper: viper.GetViper(),
	}
}

func (r *Runner) RunSync() error {
	fmt.Println("running integration tests")

	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}
	tests, err := ioutil.ReadDir(filepath.Join(currentDir, "tests"))
	if err != nil {
		return err
	}

	for _, test := range tests {
		if test.IsDir() {
			fmt.Printf("-----> Beginning test %q\n", test.Name())

			if err := r.createCluster("test"); err != nil {
				return err
			}
			defer func() {
				r.deleteCluster("test")
			}()
		}
	}

	return nil
}

func (r *Runner) createCluster(name string) error {
	cfg, err := encoding.Load("")
	if err != nil {
		return err
	}

	if err := cfg.Validate(); err != nil {
		return err
	}

	ctx := cluster.NewContext(name)
	return ctx.Create(cfg,
		create.Retain(true),
		create.WaitForReady(time.Second*90),
		create.SetupKubernetes(true))
}

func (r *Runner) deleteCluster(name string) error {
	ctx := cluster.NewContext(name)
	return ctx.Delete()
}
