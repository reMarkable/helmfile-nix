package transform

import (
	"reflect"
	"testing"
)

func TestUnmarshalOption(t *testing.T) {
	tests := []struct {
		input    string
		expected any
	}{
		{`"foo"`, "foo"},
		{`123`, 123},
		{`true`, true},
		{`[1, 2, 3]`, []any{1, 2, 3}},
		{`{"a": 1}`, map[string]any{"a": 1}},
	}

	for _, tt := range tests {
		got, err := UnmarshalOption(tt.input)
		if err != nil {
			t.Errorf("UnmarshalOption(%q) error: %v", tt.input, err)
		}
		if !reflect.DeepEqual(got, tt.expected) {
			t.Errorf("UnmarshalOption(%q) = %#v, want %#v", tt.input, got, tt.expected)
		}
	}
}

func TestJSONToYAMLs_Success(t *testing.T) {
	json := []byte(`[{"key": "value"}, {"foo": "bar"}]`)

	yaml, err := JSONToYAMLs(json, func(v any) {})
	if err != nil {
		t.Fatalf("JSONToYAMLs() error: %v", err)
	}

	expected := "key: value\n---\nfoo: bar\n"
	if string(yaml) != expected {
		t.Errorf("JSONToYAMLs() = %q, want %q", string(yaml), expected)
	}
}

func TestJSONToYAMLs_EmptyArray(t *testing.T) {
	json := []byte(`[]`)

	yaml, err := JSONToYAMLs(json, func(v any) {})
	if err != nil {
		t.Fatalf("JSONToYAMLs() error: %v", err)
	}

	if len(yaml) != 0 {
		t.Errorf("JSONToYAMLs() with empty array should return empty, got: %q", string(yaml))
	}
}

func TestJSONToYAMLs_SingleItem(t *testing.T) {
	json := []byte(`[{"test": "single"}]`)

	yaml, err := JSONToYAMLs(json, func(v any) {})
	if err != nil {
		t.Fatalf("JSONToYAMLs() error: %v", err)
	}

	expected := "test: single\n"
	if string(yaml) != expected {
		t.Errorf("JSONToYAMLs() = %q, want %q", string(yaml), expected)
	}
}

func TestJSONToYAMLs_WithPreprocessing(t *testing.T) {
	json := []byte(`[{"key": "value"}]`)

	preprocessed := false
	yaml, err := JSONToYAMLs(json, func(v any) {
		preprocessed = true
		// Add a field during preprocessing
		if m, ok := v.(map[string]any); ok {
			m["added"] = "field"
		}
	})
	if err != nil {
		t.Fatalf("JSONToYAMLs() error: %v", err)
	}

	if !preprocessed {
		t.Error("JSONToYAMLs() preprocess function not called")
	}

	// Check that preprocessing modified the output
	yamlStr := string(yaml)
	if !contains(yamlStr, "added: field") {
		t.Errorf("JSONToYAMLs() preprocessing didn't modify output: %q", yamlStr)
	}
}

func TestJSONToYAMLs_InvalidJSON(t *testing.T) {
	json := []byte(`{invalid json}`)

	_, err := JSONToYAMLs(json, func(v any) {})
	if err == nil {
		t.Error("JSONToYAMLs() expected error for invalid JSON, got nil")
	}
}

func TestJSONToYAMLs_ComplexStructure(t *testing.T) {
	json := []byte(`[
		{"name": "release1", "values": {"key": "val1"}},
		{"name": "release2", "values": {"key": "val2"}}
	]`)

	yaml, err := JSONToYAMLs(json, func(v any) {})
	if err != nil {
		t.Fatalf("JSONToYAMLs() error: %v", err)
	}

	yamlStr := string(yaml)
	// Should have separator between documents
	if !contains(yamlStr, "---") {
		t.Error("JSONToYAMLs() missing document separator")
	}

	// Should contain both releases
	if !contains(yamlStr, "release1") || !contains(yamlStr, "release2") {
		t.Errorf("JSONToYAMLs() missing expected content: %q", yamlStr)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && (s[0:len(substr)] == substr || contains(s[1:], substr))))
}
