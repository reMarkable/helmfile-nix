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
