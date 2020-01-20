package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"github.com/pkg/errors"
)

type TemplateContext struct {
	ManifestV1Beta1 string
	ManifestV1      string
}

func main() {
	err := templateManifest(
		filepath.Join("config", "crds", "v1beta1", "databases.schemahero.io_databases.yaml"),
		filepath.Join("config", "crds", "v1", "databases.schemahero.io_databases.yaml"),
		filepath.Join("pkg", "installer", "database_objects.tmpl"),
		filepath.Join("pkg", "installer", "database_objects.go"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = templateManifest(
		filepath.Join("config", "crds", "v1beta1", "schemas.schemahero.io_tables.yaml"),
		filepath.Join("config", "crds", "v1", "schemas.schemahero.io_tables.yaml"),
		filepath.Join("pkg", "installer", "table_objects.tmpl"),
		filepath.Join("pkg", "installer", "table_objects.go"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = templateManifest(
		filepath.Join("config", "crds", "v1beta1", "schemas.schemahero.io_migrations.yaml"),
		filepath.Join("config", "crds", "v1", "schemas.schemahero.io_migrations.yaml"),
		filepath.Join("pkg", "installer", "migration_objects.tmpl"),
		filepath.Join("pkg", "installer", "migration_objects.go"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func templateManifest(manifestV1Beta1File string, manifestV1File string, tmplFile string, out string) error {
	if err := os.RemoveAll(out); err != nil {
		return errors.Wrap(err, "error deleting out file")
	}

	manifestv1beta1, err := ioutil.ReadFile(manifestV1Beta1File)
	if err != nil {
		return errors.Wrap(err, "failed to read v1beta1 file")
	}
	manifestv1, err := ioutil.ReadFile(manifestV1File)
	if err != nil {
		return errors.Wrap(err, "failed to read v1 file")
	}

	templateContext := TemplateContext{
		ManifestV1Beta1: string(manifestv1beta1),
		ManifestV1:      string(manifestv1),
	}

	tmpl, err := template.ParseFiles(tmplFile)
	if err != nil {
		return errors.Wrap(err, "failed to parse template")
	}

	f, err := os.Create(out)
	if err != nil {
		return errors.Wrap(err, "failed to create out file")
	}

	if err := tmpl.Execute(f, templateContext); err != nil {
		return errors.Wrap(err, "failed to execute template")
	}

	return nil
}
