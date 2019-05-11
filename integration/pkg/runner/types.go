package runner

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Test struct {
	Cluster     TestCluster `yaml:"cluster"`
	Databases   []string    `yaml:"databases"`
	Connections []string    `yaml:"connections"`
	Steps       []*TestStep `yaml:"steps"`
}

type TestCluster struct {
	Name string `json:"name"`
}

type TestStep struct {
	Name  string         `yaml:"name"`
	Table *TestStepTable `yaml:"table"`
}

type TestStepTable struct {
	Source string `yaml:"source"`
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
