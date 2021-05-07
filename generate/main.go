package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

type TemplateContext struct {
	ManifestV1 string
}

func main() {
	err := templateManifest(
		filepath.Join("config", "crds", "v1", "databases.schemahero.io_databases.yaml"),
		filepath.Join("pkg", "installer", "database_objects.tmpl"),
		filepath.Join("pkg", "installer", "database_objects.go"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = templateManifest(
		filepath.Join("config", "crds", "v1", "schemas.schemahero.io_tables.yaml"),
		filepath.Join("pkg", "installer", "table_objects.tmpl"),
		filepath.Join("pkg", "installer", "table_objects.go"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = templateManifest(
		filepath.Join("config", "crds", "v1", "schemas.schemahero.io_migrations.yaml"),
		filepath.Join("pkg", "installer", "migration_objects.tmpl"),
		filepath.Join("pkg", "installer", "migration_objects.go"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func templateManifest(manifestV1File string, tmplFile string, out string) error {
	if err := os.RemoveAll(out); err != nil {
		return errors.Wrap(err, "error deleting out file")
	}

	manifestv1, err := ioutil.ReadFile(filepath.Clean(manifestV1File))
	if err != nil {
		return errors.Wrap(err, "failed to read v1 file")
	}

	templateContext := TemplateContext{
		ManifestV1: string(manifestv1),
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
