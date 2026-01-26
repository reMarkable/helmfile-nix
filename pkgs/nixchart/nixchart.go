// Package nixchart provides functionality to render Helm charts using Nix expressions.
package nixchart

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"reflect"

	"github.com/reMarkable/helmfile-nix/pkgs/environment"
	"github.com/reMarkable/helmfile-nix/pkgs/filesystem"
	"github.com/reMarkable/helmfile-nix/pkgs/nixeval"
	"github.com/reMarkable/helmfile-nix/pkgs/tempfiles"
	"github.com/reMarkable/helmfile-nix/pkgs/transform"
)

// Static errors for nixchart package.
var (
	ErrReleasesNotSlice     = errors.New("releases is not a slice")
	ErrReleaseNotHash       = errors.New("release is not a hash")
	ErrNixChartNotString    = errors.New("expected 'nixChart' to be a string")
	ErrWriteEvalNix         = errors.New("could not write eval.nix")
	ErrCreateTempValuesFile = errors.New("failed to create temporary file for values")
)

//go:embed eval.nix
var eval string

// RenderCharts takes a map of chart objects and a base path, renders the charts,
// and returns a slice of file paths to the rendered charts or an error.
func RenderCharts(ctx context.Context, obj map[string]any, base string) ([]string, error) {
	releasesValue := reflect.ValueOf(obj["releases"])
	if releasesValue.Kind() != reflect.Slice {
		return nil, ErrReleasesNotSlice
	}

	var cleanup []string
	for i := 0; i < releasesValue.Len(); i++ {
		rendered, err := processRelease(ctx, releasesValue.Index(i), i, base)
		if err != nil {
			return nil, err
		}
		if rendered != "" {
			cleanup = append(cleanup, rendered)
		}
	}
	return cleanup, nil
}

func processRelease(ctx context.Context, element reflect.Value, index int, base string) (string, error) {
	if element.Kind() == reflect.Map {
		return "", nil
	}

	chart, ok := element.Interface().(map[string]any)
	if !ok {
		return "", fmt.Errorf("%w: release at index %d: %v", ErrReleaseNotHash, index, element)
	}

	nixChart := chart["nixChart"]
	if nixChart == nil {
		return "", nil
	}

	renderedChart, err := evalChart(ctx, chart, base)
	if err != nil {
		return "", fmt.Errorf("failed to evaluate chart %s: %w", chart["name"], err)
	}

	chart["chart"] = renderedChart
	delete(chart, "nixChart")
	log.Printf("Rendering chart: %v\n", nixChart)

	return renderedChart, nil
}

// CleanupCharts removes the chart files specified in the cleanup slice.
// It is typically used to delete temporary chart files after processing.
func CleanupCharts(cleanup []string) {
	for _, chart := range cleanup {
		if err := os.RemoveAll(chart); err != nil {
			log.Printf("Failed to remove chart directory %s: %v", chart, err)
		}
	}
}

func prepareChartValues(chart map[string]any) map[string]any {
	var v map[string]any
	vl, ok := chart["values"].([]map[string]any)
	if ok {
		mergedValues := map[string]any{}
		for _, m := range vl {
			mergedValues = environment.MergeMaps(mergedValues, m)
		}
		v = mergedValues
	} else {
		v, ok = chart["values"].(map[string]any)
		if !ok || v == nil {
			v = map[string]any{}
		}
	}
	for _, key := range []string{"namespace", "release"} {
		if v[key] != nil {
			log.Printf("warning: `%s` in values is reserved and will be overwritten\n", key)
		}
	}
	delete(chart, "values") // Remove values from chart to avoid duplication in the rendered chart
	v["release"] = chart
	if ns, ok := chart["namespace"].(string); ok && ns != "" {
		v["namespace"] = ns // Add namespace to values if it exists
	}
	return v
}

var evalChart = func(ctx context.Context, chart map[string]any, hfbase string) (string, error) {
	nixChart, ok := chart["nixChart"].(string)
	if !ok {
		return "", fmt.Errorf("%w, got %T", ErrNixChartNotString, chart["nixChart"])
	}

	fileName, base, err := filesystem.FindFileNameAndBase(path.Join(hfbase, nixChart), []string{"chart.nix"})
	if err != nil {
		return "", fmt.Errorf("failed to find chart file: %w", err)
	}

	f, err := tempfiles.WriteEvalNix(eval)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrWriteEvalNix, err)
	}

	defer func() {
		if err := os.Remove(f.Name()); err != nil {
			log.Fatalf("could not remove eval.nix: %s", err)
		}
	}()

	val, err := os.CreateTemp("", "val.*.json")
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrCreateTempValuesFile, err)
	}

	defer func() {
		if err := os.Remove(val.Name()); err != nil {
			log.Println("Failed to remove temporary file for values:", val.Name())
		}
	}()

	v := prepareChartValues(chart)
	// Serialize the values
	values, err := json.Marshal(v)
	if err != nil {
		return "", err
	}

	// Write the values
	if _, err := val.Write(values); err != nil {
		return "", err
	}

	err = val.Close()
	if err != nil {
		log.Fatalln("Failed to close temporary file for values:", err)
	}
	expr := fmt.Sprintf(`(import %s).render "%s" "%s" "%s"`, f.Name(), fileName, base, val.Name())
	ne := nixeval.NewNixEval(expr)
	cmd := ne.Args(false)
	json, err := ne.Eval(ctx, cmd)
	if err != nil {
		log.Fatalln("Failed to evaluate chart:", chart, " : ", err)
	}
	yaml, err := transform.JSONToYAMLs(json, func(v any) {})
	if err != nil {
		log.Fatalln("Failed to convert JSON to YAML for chart:", chart, " : ", err)
	}
	chartDir := path.Join(os.TempDir(), fmt.Sprintf("nixChart-%s-%s", chart["namespace"], chart["name"]))
	err = os.MkdirAll(chartDir, 0o700)
	if err != nil {
		log.Fatalln("Failed to create temporary directory for chart:", chartDir, " : ", err)
	}
	if err = os.WriteFile(chartDir+"/resources.yaml", yaml, 0o600); err != nil {
		log.Fatalln("Failed to write resources.yaml for chart:", chart, " : ", err)
	}

	return chartDir, nil
}
