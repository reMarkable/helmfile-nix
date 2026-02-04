package helmfile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriter_WriteYAML_Success(t *testing.T) {
	t.Parallel()
	writer := NewWriter()
	tmpDir := t.TempDir()

	content := []byte("test: yaml\nkey: value\n")

	f, err := writer.WriteYAML("helmfile.nix", tmpDir, content)
	if err != nil {
		t.Fatalf("WriteYAML() error: %v", err)
	}

	// Verify file was created
	if f == nil {
		t.Fatal("WriteYAML() returned nil file")
	}

	// Verify file exists
	if _, err := os.Stat(f.Name()); os.IsNotExist(err) {
		t.Errorf("WriteYAML() file does not exist: %s", f.Name())
	}

	// Verify filename pattern
	if !strings.HasPrefix(filepath.Base(f.Name()), "helmfile.") {
		t.Errorf("WriteYAML() filename should start with 'helmfile.', got: %s", f.Name())
	}

	if !strings.HasSuffix(f.Name(), ".yaml") {
		t.Errorf("WriteYAML() filename should end with '.yaml', got: %s", f.Name())
	}

	// Verify content was written
	written, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatalf("Failed to read written file: %v", err)
	}

	if string(written) != string(content) {
		t.Errorf("WriteYAML() content mismatch.\nExpected: %s\nGot: %s", content, written)
	}

	// Cleanup
	if err := os.Remove(f.Name()); err != nil {
		t.Logf("Failed to cleanup temp file: %v", err)
	}
}

func TestWriter_WriteYAML_GotmplExtension(t *testing.T) {
	t.Parallel()
	writer := NewWriter()
	tmpDir := t.TempDir()

	content := []byte("test: {{ .Value }}\n")

	f, err := writer.WriteYAML("helmfile.gotmpl.nix", tmpDir, content)
	if err != nil {
		t.Fatalf("WriteYAML() error: %v", err)
	}

	// Verify file has .yaml.gotmpl extension
	if !strings.HasSuffix(f.Name(), ".yaml.gotmpl") {
		t.Errorf("WriteYAML() with gotmpl.nix should create .yaml.gotmpl file, got: %s", f.Name())
	}

	// Cleanup
	if err := os.Remove(f.Name()); err != nil {
		t.Logf("Failed to cleanup temp file: %v", err)
	}
}

func TestWriter_WriteYAML_InvalidDirectory(t *testing.T) {
	t.Parallel()
	writer := NewWriter()

	content := []byte("test: yaml\n")

	// Try to write to a non-existent directory
	_, err := writer.WriteYAML("helmfile.nix", "/nonexistent/directory/path", content)

	if err == nil {
		t.Error("WriteYAML() expected error for non-existent directory, got nil")
	}
}

func TestWriter_WriteYAML_EmptyContent(t *testing.T) {
	t.Parallel()
	writer := NewWriter()
	tmpDir := t.TempDir()

	content := []byte("")

	f, err := writer.WriteYAML("helmfile.nix", tmpDir, content)
	if err != nil {
		t.Fatalf("WriteYAML() error with empty content: %v", err)
	}

	// Verify empty file was created
	info, err := os.Stat(f.Name())
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	if info.Size() != 0 {
		t.Errorf("WriteYAML() with empty content should create 0-byte file, got size: %d", info.Size())
	}

	// Cleanup
	if err := os.Remove(f.Name()); err != nil {
		t.Logf("Failed to cleanup temp file: %v", err)
	}
}

func TestWriter_WriteYAML_LargeContent(t *testing.T) {
	t.Parallel()
	writer := NewWriter()
	tmpDir := t.TempDir()

	// Create a large content (1MB)
	largeContent := make([]byte, 1024*1024)
	for i := range largeContent {
		largeContent[i] = byte('a' + (i % 26))
	}

	f, err := writer.WriteYAML("helmfile.nix", tmpDir, largeContent)
	if err != nil {
		t.Fatalf("WriteYAML() error with large content: %v", err)
	}

	// Verify size
	info, err := os.Stat(f.Name())
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	if info.Size() != int64(len(largeContent)) {
		t.Errorf("WriteYAML() size mismatch. Expected: %d, got: %d", len(largeContent), info.Size())
	}

	// Cleanup
	if err := os.Remove(f.Name()); err != nil {
		t.Logf("Failed to cleanup temp file: %v", err)
	}
}
