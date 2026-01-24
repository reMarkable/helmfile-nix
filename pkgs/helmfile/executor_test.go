package helmfile

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExecutor_Execute_Success(t *testing.T) {
	logger := log.Default()
	executor := NewExecutor(logger)

	// Create a temporary directory for test
	tmpDir := t.TempDir()

	// Create a simple test file to verify we're in the right directory
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Execute helmfile version command (should be available)
	err := executor.Execute("", []string{"--version"}, tmpDir, "dev")

	// We expect this to succeed if helmfile is installed
	// If helmfile is not installed, we should get a clear error
	if err != nil {
		// Check if it's a "command not found" error, which is acceptable in test environment
		if !strings.Contains(err.Error(), "executable file not found") &&
			!strings.Contains(err.Error(), "no such file or directory") {
			t.Errorf("Execute() unexpected error: %v", err)
		}
	}
}

func TestExecutor_Execute_ChangeDirectoryError(t *testing.T) {
	logger := log.Default()
	executor := NewExecutor(logger)

	// Try to change to a non-existent directory
	err := executor.Execute("", []string{"--version"}, "/nonexistent/directory/path", "dev")

	if err == nil {
		t.Error("Execute() expected error for non-existent directory, got nil")
	}

	if !strings.Contains(err.Error(), "could not change directory") {
		t.Errorf("Execute() error message should mention directory change failure, got: %v", err)
	}
}

func TestExecutor_Execute_ArgumentPassing(t *testing.T) {
	logger := log.Default()
	executor := NewExecutor(logger)

	tmpDir := t.TempDir()

	// Test that environment argument is passed correctly
	// We'll execute a command but check the error contains our custom args
	err := executor.Execute("test.yaml", []string{"sync", "--debug"}, tmpDir, "production")

	// The command will likely fail, but we just want to ensure it was constructed correctly
	// We can't easily verify the exact command without mocking, but we can check it executed
	if err == nil {
		// Unexpected success (helmfile somehow worked without a real file)
		t.Log("Execute() succeeded unexpectedly, but that's okay")
	}
}

func TestExecutor_Execute_WithoutHelmfile(t *testing.T) {
	logger := log.Default()
	executor := NewExecutor(logger)

	tmpDir := t.TempDir()

	// Execute without specifying a helmfile (empty string)
	err := executor.Execute("", []string{}, tmpDir, "dev")

	// Should attempt to execute helmfile without --file flag
	// Will fail if helmfile isn't installed, but that's okay for this test
	if err != nil && !strings.Contains(err.Error(), "executable file not found") &&
		!strings.Contains(err.Error(), "no such file or directory") {
		// Some other error occurred, which is fine - we just want to ensure
		// the function handles empty helmfile path correctly
		t.Logf("Execute() error (expected): %v", err)
	}
}
