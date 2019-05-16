package runner

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
)

func getApplyableOperator(managerImageName string) ([]byte, error) {
	manager, err := getApplyableManager(managerImageName)
	if err != nil {
		return nil, err
	}
	crds, err := getApplyableCrds()
	if err != nil {
		return nil, err
	}

	return append(manager, crds...), nil
}

func getApplyableManager(managerImageName string) ([]byte, error) {
	crd, err := ioutil.ReadFile("manifests/manager.yaml")
	if err != nil {
		return nil, err
	}

	updatedCrd := strings.Replace(string(crd), "schemahero/schemahero:latest", managerImageName, -1)
	return []byte(updatedCrd), nil
}

func getApplyableCrds() ([]byte, error) {
	crds, err := ioutil.ReadDir("../config/crds")
	if err != nil {
		return nil, err
	}

	manifests := "---"
	for _, crd := range crds {
		manifest, err := ioutil.ReadFile(filepath.Join("../config/crds/", crd.Name()))
		if err != nil {
			return nil, err
		}

		manifests = fmt.Sprintf("%s\n%s\n---", manifests, manifest)
	}

	return []byte(manifests), nil
}
