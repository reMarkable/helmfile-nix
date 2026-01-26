package environment

import (
	"reflect"
	"testing"
)

func TestMergeMaps(t *testing.T) {
	t.Parallel()
	a := map[string]any{"a": 1, "b": map[string]any{"x": 1}}
	b := map[string]any{"b": map[string]any{"y": 2}, "c": 3}
	expected := map[string]any{
		"a": 1,
		"b": map[string]any{"x": 1, "y": 2},
		"c": 3,
	}
	got := MergeMaps(a, b)
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("MergeMaps(a, b) = %#v, want %#v", got, expected)
	}
}

func TestMergeMaps_EmptyMaps(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		a        map[string]any
		b        map[string]any
		expected map[string]any
	}{
		{
			name:     "both empty",
			a:        map[string]any{},
			b:        map[string]any{},
			expected: map[string]any{},
		},
		{
			name:     "first empty",
			a:        map[string]any{},
			b:        map[string]any{"key": "value"},
			expected: map[string]any{"key": "value"},
		},
		{
			name:     "second empty",
			a:        map[string]any{"key": "value"},
			b:        map[string]any{},
			expected: map[string]any{"key": "value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := MergeMaps(tt.a, tt.b)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("MergeMaps() = %#v, want %#v", got, tt.expected)
			}
		})
	}
}

func TestMergeMaps_NestedConflict(t *testing.T) {
	t.Parallel()
	// Test when both maps have nested structures at same key
	a := map[string]any{
		"nested": map[string]any{
			"a": 1,
			"shared": map[string]any{
				"x": "original",
			},
		},
	}
	b := map[string]any{
		"nested": map[string]any{
			"b": 2,
			"shared": map[string]any{
				"y": "new",
			},
		},
	}
	expected := map[string]any{
		"nested": map[string]any{
			"a": 1,
			"b": 2,
			"shared": map[string]any{
				"x": "original",
				"y": "new",
			},
		},
	}

	got := MergeMaps(a, b)
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("MergeMaps() = %#v, want %#v", got, expected)
	}
}

func TestMergeMaps_ValueOverwrite(t *testing.T) {
	t.Parallel()
	// Test that b's values overwrite a's values
	a := map[string]any{"key": "original"}
	b := map[string]any{"key": "overwritten"}
	expected := map[string]any{"key": "overwritten"}

	got := MergeMaps(a, b)
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("MergeMaps() = %#v, want %#v", got, expected)
	}
}

func TestMergeMaps_DeepNesting(t *testing.T) {
	t.Parallel()
	a := map[string]any{
		"level1": map[string]any{
			"level2": map[string]any{
				"level3": map[string]any{
					"level4": "deep-a",
				},
			},
		},
	}
	b := map[string]any{
		"level1": map[string]any{
			"level2": map[string]any{
				"level3": map[string]any{
					"level4-new": "deep-b",
				},
			},
		},
	}
	expected := map[string]any{
		"level1": map[string]any{
			"level2": map[string]any{
				"level3": map[string]any{
					"level4":     "deep-a",
					"level4-new": "deep-b",
				},
			},
		},
	}

	got := MergeMaps(a, b)
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("MergeMaps() deep nesting = %#v, want %#v", got, expected)
	}
}

func TestSetNestedMapValue_SimpleKey(t *testing.T) {
	t.Parallel()
	m := map[string]any{}

	err := SetNestedMapValue(m, "simple", "nestedValue")
	if err != nil {
		t.Errorf("SetNestedMapValue() error: %v", err)
	}

	if m["simple"] != "nestedValue" {
		t.Errorf("SetNestedMapValue() = %#v, want 'value'", m["simple"])
	}
}

func TestSetNestedMapValue_DottedKey(t *testing.T) {
	t.Parallel()
	m := map[string]any{}

	err := SetNestedMapValue(m, "foo.bar.baz", "value")
	if err != nil {
		t.Errorf("SetNestedMapValue() error: %v", err)
	}

	// Navigate to nested value
	foo, ok := m["foo"].(map[string]any)
	if !ok {
		t.Fatal("SetNestedMapValue() foo is not a map")
	}

	bar, ok := foo["bar"].(map[string]any)
	if !ok {
		t.Fatal("SetNestedMapValue() bar is not a map")
	}

	if bar["baz"] != "value" {
		t.Errorf("SetNestedMapValue() baz = %#v, want 'value'", bar["baz"])
	}
}

func TestSetNestedMapValue_DeepNesting(t *testing.T) {
	t.Parallel()
	m := map[string]any{}

	// Test 5 levels deep
	err := SetNestedMapValue(m, "a.b.c.d.e", "deep")
	if err != nil {
		t.Errorf("SetNestedMapValue() error: %v", err)
	}

	// Verify the deep structure was created
	a, _ := m["a"].(map[string]any)
	b, _ := a["b"].(map[string]any)
	c, _ := b["c"].(map[string]any)
	d, _ := c["d"].(map[string]any)

	if d["e"] != "deep" {
		t.Errorf("SetNestedMapValue() deep value = %#v, want 'deep'", d["e"])
	}
}

func TestSetNestedMapValue_NumericValue(t *testing.T) {
	t.Parallel()
	m := map[string]any{}

	err := SetNestedMapValue(m, "number", "123")
	if err != nil {
		t.Errorf("SetNestedMapValue() error: %v", err)
	}

	if m["number"] != 123 {
		t.Errorf("SetNestedMapValue() number = %#v, want 123", m["number"])
	}
}

func TestSetNestedMapValue_BooleanValue(t *testing.T) {
	t.Parallel()
	m := map[string]any{}

	err := SetNestedMapValue(m, "flag", "true")
	if err != nil {
		t.Errorf("SetNestedMapValue() error: %v", err)
	}

	if m["flag"] != true {
		t.Errorf("SetNestedMapValue() flag = %#v, want true", m["flag"])
	}
}

func TestSetNestedMapValue_ArrayValue(t *testing.T) {
	t.Parallel()
	m := map[string]any{}

	err := SetNestedMapValue(m, "list", "[1, 2, 3]")
	if err != nil {
		t.Errorf("SetNestedMapValue() error: %v", err)
	}

	list, ok := m["list"].([]any)
	if !ok {
		t.Fatal("SetNestedMapValue() list is not a slice")
	}

	if len(list) != 3 {
		t.Errorf("SetNestedMapValue() list length = %d, want 3", len(list))
	}
}

func TestSetNestedMapValue_InvalidJSON(t *testing.T) {
	t.Parallel()
	m := map[string]any{}

	// Use something that's invalid in both JSON and YAML
	err := SetNestedMapValue(m, "bad", "[unclosed")
	if err == nil {
		t.Error("SetNestedMapValue() expected error for invalid YAML, got nil")
	}
}

func TestSetNestedMapValue_InvalidYAML_SimpleKey(t *testing.T) {
	t.Parallel()
	m := map[string]any{}

	// Test error path for simple (non-dotted) key with invalid value
	err := SetNestedMapValue(m, "simplekey", "[invalid")
	if err == nil {
		t.Error("SetNestedMapValue() expected error for invalid YAML on simple key, got nil")
	}
}

func TestSetNestedMapValue_ExistingPath(t *testing.T) {
	t.Parallel()
	// Test setting a value when part of the path already exists
	m := map[string]any{
		"existing": map[string]any{
			"nested": "original",
		},
	}

	err := SetNestedMapValue(m, "existing.new", "value")
	if err != nil {
		t.Errorf("SetNestedMapValue() error: %v", err)
	}

	existing, _ := m["existing"].(map[string]any)
	if existing["new"] != "value" {
		t.Errorf("SetNestedMapValue() new = %#v, want 'value'", existing["new"])
	}

	// Original value should still exist
	if existing["nested"] != "original" {
		t.Errorf("SetNestedMapValue() overwrote existing nested value")
	}
}
