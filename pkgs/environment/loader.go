// Package environment provides environment and configuration management for helmfile-nix.
package environment

import (
	"os"

	"gopkg.in/yaml.v3"
)

// LoadYamlFile reads a YAML file and unmarshals it into a map[string]any.
// If the file does not exist, it returns an empty map.
func LoadYamlFile(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]any{}, nil
		}

		return nil, err
	}

	var m map[string]any
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, err
	}

	return m, nil
}
