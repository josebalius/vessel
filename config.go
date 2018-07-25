package vessel

import (
	"io/ioutil"
	"path/filepath"

	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v1"
)

//
type Config struct {
	Profile string `yaml:"profile"`
	Region  string `yaml:"region"`
	ECR     struct {
		RegistryID string
	}
}

//
func NewConfig(projectPath string) (*Config, error) {
	configPath := filepath.Join(projectPath, "config.yaml")
	file, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, errors.Wrap(err, "read config file")
	}

	config := &Config{}
	if err := yaml.Unmarshal(file, config); err != nil {
		return nil, errors.Wrap(err, "parse config file")
	}

	if config.Profile == "" {
		return nil, errors.New("A config AWS profile is required")
	}

	if config.Region == "" {
		config.Region = "us-east-1"
	}

	return config, nil
}
