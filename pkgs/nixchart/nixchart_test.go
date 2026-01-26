package nixchart

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestRenderCharts_Success(t *testing.T) {
	t.Parallel()
	origEvalChart := evalChart
	evalChart = func(_ context.Context, chart map[string]any, base string) (string, error) {
		tmpDir := t.TempDir()
		resourcesPath := filepath.Join(tmpDir, "resources.yaml")
		if err := os.WriteFile(resourcesPath, []byte("mocked: true\n"), 0o600); err != nil {
			return "", err
		}

		return tmpDir, nil
	}
	defer func() { evalChart = origEvalChart }()

	chartPath := filepath.Join("..", "..", "testData", "nixChart", "chart.nix")
	obj := map[string]any{
		"releases": []any{
			map[string]any{
				"name":     "test-chart",
				"nixChart": chartPath,
				"values": map[string]any{
					"namespace":                       "test-ns",
					"karpenter_instance_profile_role": "test-role",
					"cluster_name":                    "test-cluster",
				},
			},
		},
	}
	cleanup, err := RenderCharts(t.Context(), obj, ".")
	if err != nil {
		t.Fatalf("RenderCharts failed: %v", err)
	}

	if len(cleanup) != 1 {
		t.Fatalf("Expected 1 chart rendered, got %d", len(cleanup))
	}

	// Check that the rendered chart directory exists and contains resources.yaml
	resourcesPath := filepath.Join(cleanup[0], "resources.yaml")
	if _, err := os.Stat(resourcesPath); err != nil {
		t.Fatalf("resources.yaml not found: %v", err)
	}

	CleanupCharts(cleanup)
}

func TestPrepareChartValues(t *testing.T) {
	t.Parallel()
	// Test with a map[string]any
	chartMap := map[string]any{
		"values": map[string]any{
			"a": 1,
			"b": "two",
		},
	}
	vals := prepareChartValues(chartMap)
	if vals["a"] != 1 || vals["b"] != "two" {
		t.Errorf("Expected map values, got: %#v", vals)
	}

	// Test with a []map[string]any (should merge)
	chartList := map[string]any{
		"values": []map[string]any{
			{"a": 1, "b": "two"},
			{"b": "overwritten", "c": 3},
		},
	}
	vals = prepareChartValues(chartList)
	if vals["a"] != 1 || vals["b"] != "overwritten" || vals["c"] != 3 {
		t.Errorf("Expected merged values, got: %#v", vals)
	}

	// Test with no values
	chartNil := map[string]any{}
	vals = prepareChartValues(chartNil)
	if len(vals) != 1 {
		t.Errorf("Expected only release meta, got: %#v", vals)
	}

	// Test that namespace and release metadata is copied into values.
	chartMap = map[string]any{
		"namespace": "test-ns",
		"installed": true,
		"values": map[string]any{
			"a":         1,
			"b":         "two",
			"namespace": "should-be-overwritten",
		},
	}
	vals = prepareChartValues(chartMap)
	if vals["namespace"] != "test-ns" || vals["release"] == nil {
		t.Errorf("Expected copied values, got: %#v", vals)
	}
	releaseMap, ok := vals["release"].(map[string]any)
	if !ok || releaseMap["values"] != nil {
		t.Errorf("Expected release metadata without values, got: %#v", releaseMap)
	}
}

func TestCleanupCharts_NonExistent(t *testing.T) {
	t.Parallel()
	// Test that CleanupCharts handles non-existent directories gracefully
	cleanup := []string{"/nonexistent/directory/path"}

	// Should not panic, just log error
	CleanupCharts(cleanup)
}

func TestCleanupCharts_Mixed(t *testing.T) {
	t.Parallel()
	// Create a real directory
	tmpDir := t.TempDir()

	// Mix of existing and non-existent directories
	cleanup := []string{tmpDir, "/nonexistent/path"}

	// Should clean up what it can without panicking
	CleanupCharts(cleanup)

	// Verify the temp directory was removed
	if _, err := os.Stat(tmpDir); !os.IsNotExist(err) {
		t.Error("CleanupCharts() should have removed temp directory")
	}
}

func TestCleanupCharts_Empty(t *testing.T) {
	t.Parallel()
	// Test with empty cleanup list
	CleanupCharts([]string{})
	// Should not panic
}
