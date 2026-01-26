package environment

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andreyvit/diff"
)

var testVals = `{"bad":123,"bar":"true","foo":{"bad":"hello","bar":false,"baz":true,"foo":true}}`

func TestValuesWriter_WriteJSON_Success(t *testing.T) {
	t.Parallel()
	logger := log.Default()
	writer := NewValuesWriter(logger)

	// Use the existing test data
	cwd, _ := os.Getwd()
	testDataPath := filepath.Join(cwd, "../../testData/helm")

	f, err := writer.WriteJSON(testDataPath, "test", []string{"foo.bar=false", "bad=123", "foo.bad=hello"})
	if err != nil {
		t.Fatalf("WriteJSON() error: %v", err)
	}

	defer func() {
		if err := os.Remove(f.Name()); err != nil {
			t.Errorf("Failed to remove temp file: %v", err)
		}
	}()

	// Read and verify content
	content, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	if string(content) != testVals {
		t.Errorf("WriteJSON() content mismatch:\n%v", diff.LineDiff(string(content), testVals))
	}
}

func TestValuesWriter_WriteJSON_MissingEnvironmentFiles(t *testing.T) {
	t.Parallel()
	logger := log.Default()
	writer := NewValuesWriter(logger)

	tmpDir := t.TempDir()

	// No env directory exists, should return empty values
	f, err := writer.WriteJSON(tmpDir, "dev", []string{})
	if err != nil {
		t.Fatalf("WriteJSON() with missing env files should not error: %v", err)
	}

	defer func() {
		if err = os.Remove(f.Name()); err != nil {
			t.Errorf("Failed to remove temp file: %v", err)
		}
	}()

	// Verify empty JSON object was written
	content, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	if string(content) != "{}" {
		t.Errorf("WriteJSON() with missing files should return empty object, got: %s", content)
	}
}

func TestValuesWriter_WriteJSON_InvalidOverrideFormat(t *testing.T) {
	t.Parallel()
	logger := log.Default()
	writer := NewValuesWriter(logger)

	tmpDir := t.TempDir()
	envDir := filepath.Join(tmpDir, "env")
	if err := os.MkdirAll(envDir, 0o755); err != nil {
		t.Fatalf("Failed to create env dir: %v", err)
	}

	// Create empty defaults file
	if err := os.WriteFile(filepath.Join(envDir, "defaults.yaml"), []byte("{}"), 0o600); err != nil {
		t.Fatalf("Failed to create defaults: %v", err)
	}

	// Test various invalid override formats that should fail
	invalidOverrides := []struct {
		name     string
		override string
	}{
		{"no equals sign", "no_equals_sign"},
		{"multiple equals", "multiple=equals=signs"},
	}

	for _, tc := range invalidOverrides {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			f, err := writer.WriteJSON(tmpDir, "dev", []string{tc.override})
			if err == nil {
				if f != nil {
					if err := os.Remove(f.Name()); err != nil {
						t.Logf("Failed to cleanup temp file: %v", err)
					}
				}
				t.Errorf("WriteJSON() expected error for invalid override %s, got nil", tc.override)
			}

			if err != nil && !strings.Contains(err.Error(), "invalid state value") {
				t.Errorf("WriteJSON() error should mention 'invalid state value', got: %v", err)
			}
		})
	}
}

func TestValuesWriter_WriteJSON_InvalidYAMLSyntax(t *testing.T) {
	t.Parallel()
	logger := log.Default()
	writer := NewValuesWriter(logger)

	tmpDir := t.TempDir()
	envDir := filepath.Join(tmpDir, "env")
	if err := os.MkdirAll(envDir, 0o755); err != nil {
		t.Fatalf("Failed to create env dir: %v", err)
	}

	// Create invalid YAML file
	invalidYAML := `
foo: bar
  invalid indentation
	mixed tabs
`
	if err := os.WriteFile(filepath.Join(envDir, "defaults.yaml"), []byte(invalidYAML), 0o600); err != nil {
		t.Fatalf("Failed to create invalid YAML: %v", err)
	}

	_, err := writer.WriteJSON(tmpDir, "dev", []string{})

	if err == nil {
		t.Error("WriteJSON() expected error for invalid YAML, got nil")
	}
}

func TestValuesWriter_WriteJSON_NestedOverrides(t *testing.T) {
	t.Parallel()
	logger := log.Default()
	writer := NewValuesWriter(logger)

	tmpDir := t.TempDir()
	envDir := filepath.Join(tmpDir, "env")
	if err := os.MkdirAll(envDir, 0o755); err != nil {
		t.Fatalf("Failed to create env dir: %v", err)
	}

	// Create defaults
	defaultsYAML := `
foo:
  bar:
    baz: original
`
	if err := os.WriteFile(filepath.Join(envDir, "defaults.yaml"), []byte(defaultsYAML), 0o600); err != nil {
		t.Fatalf("Failed to create defaults: %v", err)
	}

	f, err := writer.WriteJSON(tmpDir, "dev", []string{"foo.bar.baz=updated"})
	if err != nil {
		t.Fatalf("WriteJSON() error: %v", err)
	}

	defer func() {
		if err := os.Remove(f.Name()); err != nil {
			t.Errorf("Failed to remove temp file: %v", err)
		}
	}()

	// Verify the nested value was updated
	content, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}

	if !strings.Contains(string(content), "updated") {
		t.Errorf("WriteJSON() nested override not applied: %s", content)
	}
}

func TestValuesWriter_WriteJSON_MultipleOverrides(t *testing.T) {
	t.Parallel()
	logger := log.Default()
	writer := NewValuesWriter(logger)

	tmpDir := t.TempDir()
	envDir := filepath.Join(tmpDir, "env")
	if err := os.MkdirAll(envDir, 0o755); err != nil {
		t.Fatalf("Failed to create env dir: %v", err)
	}

	// Create defaults
	if err := os.WriteFile(filepath.Join(envDir, "defaults.yaml"), []byte("{}"), 0o600); err != nil {
		t.Fatalf("Failed to create defaults: %v", err)
	}

	// Test comma-separated overrides
	f, err := writer.WriteJSON(tmpDir, "dev", []string{"foo=1,bar=2,baz=3"})
	if err != nil {
		t.Fatalf("WriteJSON() error: %v", err)
	}

	defer func() {
		if err := os.Remove(f.Name()); err != nil {
			t.Errorf("Failed to remove temp file: %v", err)
		}
	}()

	// Verify all values were set
	content, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "foo") || !strings.Contains(contentStr, "bar") || !strings.Contains(contentStr, "baz") {
		t.Errorf("WriteJSON() missing overrides in output: %s", contentStr)
	}
}

func TestValuesWriter_NewValuesWriter(t *testing.T) {
	t.Parallel()
	logger := log.Default()
	writer := NewValuesWriter(logger)

	if writer == nil {
		t.Fatal("NewValuesWriter() returned nil")
	}

	if writer.logger != logger {
		t.Error("NewValuesWriter() logger not set correctly")
	}
}
