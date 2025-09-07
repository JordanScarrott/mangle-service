package file

import (
	"mangle-service/internal/core/domain"
	"os"

	"gopkg.in/yaml.v2"
)

// NewConfigLoader creates a new ConfigLoader.
func NewConfigLoader() *ConfigLoader {
	return &ConfigLoader{}
}

// ConfigLoader is a file-based loader for service relationship configurations.
type ConfigLoader struct{}

// Load reads a YAML file from the given path and returns the RelationshipConfig.
func (l *ConfigLoader) Load(path string) (*domain.RelationshipConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config domain.RelationshipConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
