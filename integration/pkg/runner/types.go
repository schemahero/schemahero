package runner

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Test struct {
	Cluster     TestCluster `yaml:"cluster"`
	Databases   []string    `yaml:"databases"`
	Connections []string    `yaml:"connections"`
}

type TestCluster struct {
	Name string `json:"name"`
}

func unmarshalTestFile(filename string) (*Test, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	test := Test{}
	if err := yaml.Unmarshal(data, &test); err != nil {
		return nil, err
	}

	return &test, nil
}
