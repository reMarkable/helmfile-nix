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
		if err := os.WriteFile(resourcesPath, []byte("mocked: true\n"), 0644); err != nil {
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
