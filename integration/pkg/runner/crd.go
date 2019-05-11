package runner

import (
	"fmt"
	"io/ioutil"
	"strings"
)

func getApplyableCRD(managerImageName string) ([]byte, error) {
	fmt.Printf(managerImageName)
	crd, err := ioutil.ReadFile("manifests/crd.yaml")
	if err != nil {
		return nil, err
	}

	updatedCrd := strings.Replace(string(crd), "IMAGE_URL", managerImageName, -1)
	return []byte(updatedCrd), nil
}
