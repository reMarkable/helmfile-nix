package filesystem

import (
	"os"
	"path/filepath"
	"testing"
)

var cwd, _ = os.Getwd()

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
