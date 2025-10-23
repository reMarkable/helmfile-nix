package nixchart

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRenderCharts_Success(t *testing.T) {
	origEvalChart := evalChart
	evalChart = func(chart map[string]any, base string) (string, error) {
		tmpDir := t.TempDir()
		resourcesPath := filepath.Join(tmpDir, "resources.yaml")
		if err := os.WriteFile(resourcesPath, []byte("mocked: true\n"), 0o644); err != nil {
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
	cleanup, err := RenderCharts(obj, ".")
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

	// Test with missing or nil values
	chartNil := map[string]any{}
	vals = prepareChartValues(chartNil)
	if len(vals) != 1 {
		t.Errorf("Expected only release meta, got: %#v", vals)
	}
}
