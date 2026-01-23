package environment

import (
	"reflect"
	"testing"
)

func TestMergeMaps(t *testing.T) {
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
