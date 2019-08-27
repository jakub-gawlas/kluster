package kluster

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

const DefaultConfigPath = "cluster.yaml"

type Config struct {
	Name      string   `yaml:"name"`
	Charts    []Chart  `yaml:"charts,omitempty"`
	Resources []string `yaml:"resources,omitempty"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := &Config{}

	// TODO: use k8s apimachinery, like in KinD
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
