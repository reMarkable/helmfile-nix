package environment

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadYamlFile_Success(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.yaml")

	content := `
foo: bar
nested:
  key: keyValue
  number: 42
`
	if err := os.WriteFile(testFile, []byte(content), 0o600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	result, err := LoadYamlFile(testFile)
	if err != nil {
		t.Fatalf("LoadYamlFile() error: %v", err)
	}

	if result["foo"] != "bar" {
		t.Errorf("LoadYamlFile() expected foo=bar, got: %v", result["foo"])
	}

	if nested, ok := result["nested"].(map[string]any); ok {
		if nested["key"] != "keyValue" {
			t.Errorf("LoadYamlFile() expected nested.key=value, got: %v", nested["key"])
		}
		if nested["number"] != 42 {
			t.Errorf("LoadYamlFile() expected nested.number=42, got: %v", nested["number"])
		}
	} else {
		t.Error("LoadYamlFile() nested key is not a map")
	}
}

func TestLoadYamlFile_MissingFile(t *testing.T) {
	t.Parallel()
	result, err := LoadYamlFile("/nonexistent/file.yaml")
	if err != nil {
		t.Fatalf("LoadYamlFile() with missing file should not error: %v", err)
	}

	// Should return empty map for missing files
	if len(result) != 0 {
		t.Errorf("LoadYamlFile() with missing file should return empty map, got: %v", result)
	}
}

func TestLoadYamlFile_EmptyFile(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "empty.yaml")

	if err := os.WriteFile(testFile, []byte(""), 0o600); err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}

	result, err := LoadYamlFile(testFile)
	if err != nil {
		t.Fatalf("LoadYamlFile() with empty file error: %v", err)
	}

	// Empty YAML should parse to empty map
	if len(result) != 0 {
		t.Errorf("LoadYamlFile() with empty file should return empty map, got: %v", result)
	}
}

func TestLoadYamlFile_InvalidYAML(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "invalid.yaml")

	invalidYAML := `
foo: bar
  bad indentation
	mixed tabs and spaces
- invalid list
  with bad structure
`
	if err := os.WriteFile(testFile, []byte(invalidYAML), 0o600); err != nil {
		t.Fatalf("Failed to create invalid YAML file: %v", err)
	}

	_, err := LoadYamlFile(testFile)

	if err == nil {
		t.Error("LoadYamlFile() expected error for invalid YAML, got nil")
	}
}

func TestLoadYamlFile_ComplexStructure(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "complex.yaml")

	complexYAML := `
string: rootValue
number: 123
boolean: true
list:
  - item1
  - item2
  - item3
nested:
  deep:
    deeper:
      key: deepValue
mixed:
  string: text
  number: 456
  list:
    - a
    - b
`
	if err := os.WriteFile(testFile, []byte(complexYAML), 0o600); err != nil {
		t.Fatalf("Failed to create complex YAML file: %v", err)
	}

	result, err := LoadYamlFile(testFile)
	if err != nil {
		t.Fatalf("LoadYamlFile() error: %v", err)
	}

	// Verify various types were parsed correctly
	if result["string"] != "rootValue" {
		t.Errorf("LoadYamlFile() string value incorrect: %v", result["string"])
	}

	if result["number"] != 123 {
		t.Errorf("LoadYamlFile() number value incorrect: %v", result["number"])
	}

	if result["boolean"] != true {
		t.Errorf("LoadYamlFile() boolean value incorrect: %v", result["boolean"])
	}

	if list, ok := result["list"].([]any); ok {
		if len(list) != 3 {
			t.Errorf("LoadYamlFile() list length incorrect: %d", len(list))
		}
	} else {
		t.Error("LoadYamlFile() list is not a slice")
	}
}

func TestLoadYamlFile_PermissionDenied(t *testing.T) {
	t.Parallel()
	// Skip this test on systems where we can't easily test permission denied
	if os.Getuid() == 0 {
		t.Skip("Skipping permission test when running as root")
	}

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "noperm.yaml")

	if err := os.WriteFile(testFile, []byte("test: value"), 0o600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Remove read permissions
	if err := os.Chmod(testFile, 0o000); err != nil {
		t.Fatalf("Failed to chmod file: %v", err)
	}
	defer func() {
		_ = os.Chmod(testFile, 0o644) // Restore for cleanup
	}()

	_, err := LoadYamlFile(testFile)

	if err == nil {
		t.Error("LoadYamlFile() expected error for permission denied, got nil")
	}
}
