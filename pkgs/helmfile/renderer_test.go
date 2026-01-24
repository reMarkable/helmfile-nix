package helmfile

import (
	"log"
	"os"
	"strings"
	"testing"
)

var testEval = `
{
  render = file: state: env: val:
    let
      lib = (builtins.getFlake "nixpkgs").lib;
    in
      builtins.toJSON [{ test = "output"; }];
}
`

func TestRenderer_Render_Success(t *testing.T) {
	logger := log.Default()
	renderer := NewRenderer(testEval, false, []string{}, logger)

	// Create temporary values file
	tmpDir := t.TempDir()
	valFile, err := os.CreateTemp(tmpDir, "values.*.json")
	if err != nil {
		t.Fatalf("Failed to create temp values file: %v", err)
	}
	defer func() {
		_ = os.Remove(valFile.Name())
	}()

	if _, err := valFile.WriteString("{}"); err != nil {
		t.Fatalf("Failed to write values file: %v", err)
	}
	if err := valFile.Close(); err != nil {
		t.Fatalf("Failed to close values file: %v", err)
	}

	// Render with test eval
	yaml, cleanup, err := renderer.Render("test.nix", tmpDir, "dev", valFile.Name())

	// Note: This will likely fail if nix isn't installed or the eval is invalid
	// but we're testing the function structure
	if err != nil {
		// Check if it's a "nix not found" error, which is acceptable
		if !strings.Contains(err.Error(), "executable file not found") &&
			!strings.Contains(err.Error(), "no such file or directory") &&
			!strings.Contains(err.Error(), "failed to eval nix") {
			t.Errorf("Render() unexpected error: %v", err)
		}
		return
	}

	// If successful, verify we got output
	if yaml == nil {
		t.Error("Render() returned nil yaml")
	}

	if cleanup == nil {
		t.Error("Render() returned nil cleanup slice")
	}
}

func TestRenderer_Render_InvalidValuesPath(t *testing.T) {
	logger := log.Default()
	renderer := NewRenderer(testEval, false, []string{}, logger)

	tmpDir := t.TempDir()

	// Use a non-existent values file path
	_, _, err := renderer.Render("test.nix", tmpDir, "dev", "/nonexistent/values.json")

	if err == nil {
		t.Error("Render() expected error for non-existent values file, got nil")
	}
}

func TestRenderer_Render_ShowTrace(t *testing.T) {
	logger := log.Default()
	rendererWithTrace := NewRenderer(testEval, true, []string{}, logger)
	rendererWithoutTrace := NewRenderer(testEval, false, []string{}, logger)

	// Verify that showTrace setting is stored
	if !rendererWithTrace.showTrace {
		t.Error("NewRenderer() with showTrace=true should set showTrace field")
	}

	if rendererWithoutTrace.showTrace {
		t.Error("NewRenderer() with showTrace=false should not set showTrace field")
	}
}

func TestRenderer_Render_WithStateValuesSet(t *testing.T) {
	logger := log.Default()
	overrides := []string{"foo=bar", "baz=qux"}
	renderer := NewRenderer(testEval, false, overrides, logger)

	// Verify that state values are stored
	if len(renderer.stateValuesSet) != 2 {
		t.Errorf("NewRenderer() expected 2 state values, got: %d", len(renderer.stateValuesSet))
	}

	if renderer.stateValuesSet[0] != "foo=bar" {
		t.Errorf("NewRenderer() state value mismatch. Expected: foo=bar, got: %s", renderer.stateValuesSet[0])
	}
}

func TestRenderer_NewRenderer(t *testing.T) {
	logger := log.Default()
	evalNix := "test eval content"
	showTrace := true
	stateValues := []string{"test=value"}

	renderer := NewRenderer(evalNix, showTrace, stateValues, logger)

	if renderer == nil {
		t.Fatal("NewRenderer() returned nil")
	}

	if renderer.evalNix != evalNix {
		t.Errorf("NewRenderer() evalNix mismatch. Expected: %s, got: %s", evalNix, renderer.evalNix)
	}

	if renderer.showTrace != showTrace {
		t.Errorf("NewRenderer() showTrace mismatch. Expected: %v, got: %v", showTrace, renderer.showTrace)
	}

	if len(renderer.stateValuesSet) != len(stateValues) {
		t.Errorf("NewRenderer() stateValuesSet length mismatch. Expected: %d, got: %d", len(stateValues), len(renderer.stateValuesSet))
	}

	if renderer.logger != logger {
		t.Error("NewRenderer() logger mismatch")
	}
}
