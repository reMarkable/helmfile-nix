// Package transform provides data format transformation utilities for helmfile-nix.
package transform

import (
	"gopkg.in/yaml.v3"
)

func UnmarshalOption(val string) (any, error) {
	var v any
	err := yaml.Unmarshal([]byte(val), &v)
	if err != nil {
		return nil, err
	}

	return v, nil
}

// JSONToYAMLs converts a JSON list to YAML documents.
func JSONToYAMLs(j []byte, preprocess func(any)) ([]byte, error) {
	var jsonObj []any
	// We are using yaml.Unmarshal here (instead of json.Unmarshal) because the
	// Go JSON library doesn't try to pick the right number type (int, float,
	// etc.) when unmarshalling to interface{}, it just picks float64
	// universally. go-yaml does go through the effort of picking the right
	// number type, so we can preserve number type throughout this process.
	if err := yaml.Unmarshal(j, &jsonObj); err != nil {
		return nil, err
	}

	var y []byte
	// Marshal this object into YAML.
	for _, v := range jsonObj {
		preprocess(v)
		res, err := yaml.Marshal(v)
		if err != nil {
			return nil, err
		}

		if len(y) > 0 {
			y = append(y, []byte("---\n")...)
		}
		y = append(y, res...)

	}
	return y, nil
}
