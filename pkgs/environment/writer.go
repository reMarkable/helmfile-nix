package environment

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// ErrInvalidStateValue is returned when a state value has invalid format.
var ErrInvalidStateValue = errors.New("invalid state value")

// ValuesWriter handles writing environment values to JSON files.
type ValuesWriter struct {
	logger *log.Logger
}

// NewValuesWriter creates a new values writer.
func NewValuesWriter(logger *log.Logger) *ValuesWriter {
	return &ValuesWriter{
		logger: logger,
	}
}

// WriteJSON merges environment YAML values and writes them to a temporary JSON file.
// The caller is responsible for removing the file after use.
func (w *ValuesWriter) WriteJSON(state string, env string, overrides []string) (*os.File, error) {
	f, err := os.CreateTemp("", "val.*.json")
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := f.Close(); err != nil {
			log.Fatalf("Could not close values.json: %s", err)
		}
	}()

	// Get defaults
	defaultsPath := filepath.Join(state, "env", "defaults.yaml")
	envPath := filepath.Join(state, "env", env+".yaml")

	m, err := LoadYamlFile(defaultsPath)
	if err != nil {
		return nil, err
	}

	n, err := LoadYamlFile(envPath)
	if err != nil {
		return nil, err
	}

	m = MergeMaps(m, n)
	// Handle state overrides
	for _, v := range overrides {
		vals := strings.SplitSeq(v, ",")
		for val := range vals {
			kv := strings.Split(val, "=")
			if len(kv) != 2 {
				return nil, fmt.Errorf("%w: %s", ErrInvalidStateValue, val)
			}

			if err := SetNestedMapValue(m, kv[0], kv[1]); err != nil {
				return nil, fmt.Errorf("could not set nested map value %s: %w", kv[0], err)
			}
		}
	}

	// Serialize the values
	envStr, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	// Write the values
	if _, err := f.Write(envStr); err != nil {
		return nil, err
	}

	return f, nil
}
