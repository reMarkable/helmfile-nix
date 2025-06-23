// Package utils provides utility functions for helmfile-nix.
package utils

import (
	"fmt"
	"log"
	"maps"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

var l = log.Default()

func UnmarshalOption(val string) (any, error) {
	var v any
	err := yaml.Unmarshal([]byte(val), &v)
	if err != nil {
		return nil, err
	}

	return v, nil
}

// FindFileNameAndBase - find the filename and base directory of the helmfile.nix
func FindFileNameAndBase(input string, wanted []string) (string, string, error) {
	path, err := filepath.Abs(input)
	if err != nil {
		return "", "", err
	}

	fileInfo, err := os.Stat(path)
	if err != nil {
		return "", "", err
	}

	// If it's a directory, we need to find the helmfile
	if fileInfo.IsDir() {
		// Read the contents of the directory
		files, err := os.ReadDir(path)
		if err != nil {
			return "", "", err
		}

		// Check if the desired file exists in the directory
		for _, file := range files {
			for _, w := range wanted {
				if file.Name() == w {
					return w, path, nil
				}
			}
		}

	}
	// Check if the file is one of the wanted files
	for _, w := range wanted {
		if filepath.Base(path) == w {
			return w, filepath.Dir(path), nil
		}
	}

	return "", "", fmt.Errorf("expected %v,  found : %s ", wanted, path)
}

// JSONToYAMLs converts a JSON list to YAML documents.
func JSONToYAMLs(j []byte, preprocess func(any)) ([]byte, error) {
	// Convert the JSON to a list of object.
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

func MergeMaps(a, b map[string]any) map[string]any {
	out := make(map[string]any, len(a))
	maps.Copy(out, a)
	for k, v := range b {
		if v, ok := v.(map[string]any); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(map[string]any); ok {
					out[k] = MergeMaps(bv, v)
					continue
				}
			}
		}
		out[k] = v
	}
	return out
}

func SetNestedMapValue(m map[string]any, dottedKey string, value string) error {
	var err error
	if strings.Contains(dottedKey, ".") {
		// Nested value
		// Split the key into parts
		mref := m
		keys := strings.Split(dottedKey, ".")
		for i, key := range keys {
			if i == len(keys)-1 {
				if mref[key], err = UnmarshalOption(value); err != nil {
					return err
				}
			} else {
				if _, ok := m[key]; !ok {
					mref[key] = make(map[string]any)
				}
				mref = m[key].(map[string]any)
			}
		}
	} else {
		if m[dottedKey], err = UnmarshalOption(value); err != nil {
			return err
		}
	}
	return nil
}

// WriteEvalNix writes the eval.nix to a temporary file. Must be removed after use.
func WriteEvalNix(eval string) (*os.File, error) {
	f, err := os.CreateTemp("", "eval.*.nix")
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := f.Close(); err != nil {
			l.Fatalf("Could not close eval.nix: %s", err)
		}
	}()

	if _, err := f.WriteString(eval); err != nil {
		return nil, err
	}

	return f, nil
}
