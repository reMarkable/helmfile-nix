package utils

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

var cwd, _ = os.Getwd()

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

func TestPathAndBase(t *testing.T) {
	inputs := []string{"../../testData/helm", "../../testData/helm/helmfile.nix", "../..//testData/helm/helmfile.nix", cwd + "/../../testData/helm/helmfile.nix", "../../testData/helm/helmfile.nix"}
	for _, input := range inputs {
		hfPath, base, err := FindFileNameAndBase(input, []string{"helmfile.nix"})
		if err != nil {
			t.Error("full path failed:", err)
		}
		absPath, err := filepath.Abs(cwd + "/../../testData/helm")
		if err != nil {
			panic(err)
		}
		if base != absPath {
			t.Error("Base not matched: ", base, " != ", cwd+"/../../testData/helm")
		}
		if hfPath != "helmfile.nix" {
			t.Error("Path not matched: ", hfPath, " != helmfile.nix")
		}
	}
}

func TestFindFileNameAndBase(t *testing.T) {
	tmpDir := t.TempDir()
	// Create wanted files
	wantedFiles := []string{"foo.nix", "bar.nix"}
	for _, name := range wantedFiles {
		f, err := os.Create(filepath.Join(tmpDir, name))
		if err != nil {
			t.Fatalf("failed to create file: %v", err)
		}

		if err := f.Close(); err != nil {
			panic(err)
		}
	}

	// Directory input, finds file
	name, base, err := FindFileNameAndBase(tmpDir, wantedFiles)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if base != tmpDir {
		t.Errorf("expected base %s, got %s", tmpDir, base)
	}
	if name != "foo.nix" && name != "bar.nix" {
		t.Errorf("unexpected file name: %s", name)
	}

	// File input, finds file
	filePath := filepath.Join(tmpDir, "foo.nix")
	name, base, err = FindFileNameAndBase(filePath, wantedFiles)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if base != tmpDir {
		t.Errorf("expected base %s, got %s", tmpDir, base)
	}
	if name != "foo.nix" {
		t.Errorf("expected file name foo.nix, got %s", name)
	}

	// Not found
	_, _, err = FindFileNameAndBase(tmpDir, []string{"notfound.nix"})
	if err == nil {
		t.Errorf("expected error for missing file, got nil")
	}
}
