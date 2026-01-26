package tempfiles

import (
	"os"
	"strings"
	"testing"
)

func TestWriteEvalNix_Success(t *testing.T) {
	t.Parallel()
	evalContent := `
{
  test = "content";
}
`

	f, err := WriteEvalNix(evalContent)
	if err != nil {
		t.Fatalf("WriteEvalNix() error: %v", err)
	}

	// File should exist
	if _, err := os.Stat(f.Name()); os.IsNotExist(err) {
		t.Errorf("WriteEvalNix() file does not exist: %s", f.Name())
	}

	// Verify filename pattern
	if !strings.HasPrefix(f.Name(), os.TempDir()) {
		t.Errorf("WriteEvalNix() file not in temp dir: %s", f.Name())
	}

	if !strings.Contains(f.Name(), "eval.") || !strings.HasSuffix(f.Name(), ".nix") {
		t.Errorf("WriteEvalNix() filename pattern incorrect: %s", f.Name())
	}

	// Read and verify content
	content, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(content) != evalContent {
		t.Errorf("WriteEvalNix() content mismatch.\nExpected: %s\nGot: %s", evalContent, content)
	}

	// Cleanup
	if err := os.Remove(f.Name()); err != nil {
		t.Logf("Failed to cleanup temp file: %v", err)
	}
}

func TestWriteEvalNix_EmptyContent(t *testing.T) {
	t.Parallel()
	f, err := WriteEvalNix("")
	if err != nil {
		t.Fatalf("WriteEvalNix() error with empty content: %v", err)
	}

	// Verify empty file was created
	info, err := os.Stat(f.Name())
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	if info.Size() != 0 {
		t.Errorf("WriteEvalNix() with empty content should create 0-byte file, got size: %d", info.Size())
	}

	// Cleanup
	if err := os.Remove(f.Name()); err != nil {
		t.Logf("Failed to cleanup temp file: %v", err)
	}
}

func TestWriteEvalNix_LargeContent(t *testing.T) {
	t.Parallel()
	// Create large Nix expression (10KB)
	largeContent := strings.Repeat("# comment\n", 500)

	f, err := WriteEvalNix(largeContent)
	if err != nil {
		t.Fatalf("WriteEvalNix() error with large content: %v", err)
	}

	// Verify size
	info, err := os.Stat(f.Name())
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	if info.Size() != int64(len(largeContent)) {
		t.Errorf("WriteEvalNix() size mismatch. Expected: %d, got: %d", len(largeContent), info.Size())
	}

	// Cleanup
	if err := os.Remove(f.Name()); err != nil {
		t.Logf("Failed to cleanup temp file: %v", err)
	}
}

func TestWriteEvalNix_SpecialCharacters(t *testing.T) {
	t.Parallel()
	// Test with special characters that might cause issues
	evalContent := `
{
  # Comment with "quotes" and 'apostrophes'
  test = "value with\nnewlines\tand\ttabs";
  unicode = "Unicode: 你好世界 🎉";
  escaped = "path\\with\\backslashes";
}
`

	f, err := WriteEvalNix(evalContent)
	if err != nil {
		t.Fatalf("WriteEvalNix() error: %v", err)
	}

	// Read and verify content preserved special characters
	content, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(content) != evalContent {
		t.Errorf("WriteEvalNix() special characters not preserved")
	}

	// Cleanup
	if err := os.Remove(f.Name()); err != nil {
		t.Logf("Failed to cleanup temp file: %v", err)
	}
}
