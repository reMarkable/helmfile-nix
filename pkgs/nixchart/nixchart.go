// Package nixchart provides functionality to render Helm charts using Nix expressions.
package nixchart

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"reflect"

	_ "embed"

	"github.com/reMarkable/helmfile-nix/pkgs/nixeval"
	"github.com/reMarkable/helmfile-nix/pkgs/utils"
)

//go:embed eval.nix
var eval string

// RenderCharts takes a map of chart objects and a base path, renders the charts,
// and returns a slice of file paths to the rendered charts or an error.
func RenderCharts(obj map[string]any, base string) ([]string, error) {
	releasesValue := reflect.ValueOf(obj["releases"])
	var cleanup []string
	if releasesValue.Kind() != reflect.Slice {
		return nil, errors.New("releases is not a slice")
	}

	for i := 0; i < releasesValue.Len(); i++ {
		element := releasesValue.Index(i)
		log.Printf("Processing release at index %d: %v\n", i, element)
		if element.Kind() != reflect.Map {
			chart, ok := element.Interface().(map[string]any)
			if !ok {
				return nil, fmt.Errorf("release at index %d is not a hash: %v", i, element)
			}

			nixChart := chart["nixChart"]
			if nixChart != nil {
				if renderedChart, err := evalChart(chart, base); err == nil {
					chart["chart"] = renderedChart
					cleanup = append(cleanup, renderedChart)
					delete(chart, "nixChart")
				} else {
					return nil, fmt.Errorf("failed to evaluate chart %s: %w", chart["name"], err)
				}

				log.Printf("Rendering chart: %v\n", nixChart)
			}
		}
	}
	return cleanup, nil
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

var evalChart = func(chart map[string]any, hfbase string) (string, error) {
	nixChart, ok := chart["nixChart"].(string)
	if !ok {
		return "", fmt.Errorf("expected 'nixChart' to be a string, but got %T", chart["nixChart"])
	}

	fileName, base, err := utils.FindFileNameAndBase(path.Join(hfbase, nixChart), []string{"chart.nix"})
	if err != nil {
		return "", fmt.Errorf("failed to find chart file: %w", err)
	}

	f, err := utils.WriteEvalNix(eval)
	if err != nil {
		return "", fmt.Errorf("could not write eval.nix: %s", err)
	}

	defer func() {
		if err := os.Remove(f.Name()); err != nil {
			log.Fatalf("could not remove eval.nix: %s", err)
		}
	}()

	val, err := os.CreateTemp("", "val.*.json")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary file for values: %s", err)
	}

	defer func() {
		if err := os.Remove(val.Name()); err != nil {
			panic(fmt.Sprintf("unable to remove %s: %s", val.Name(), err))
		}
	}()
	v := chart["values"]
	if v == nil {
		v = map[string]any{}
	}
	delete(chart, "values") // Remove values from chart to avoid duplication in the rendered chart
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
		log.Fatalln("Failed to create temporary file for values:", err)
	}
	expr := fmt.Sprintf(`(import %s).render "%s" "%s" "%s"`, f.Name(), fileName, base, val.Name())
	ne := nixeval.NewNixEval(expr)
	cmd := ne.Args(false)
	json, err := ne.Eval(cmd)
	if err != nil {
		log.Fatalln("Failed to evaluate chart:", chart, " : ", err)
	}
	yaml, err := utils.JSONToYAMLs(json, func(v any) {})
	if err != nil {
		log.Fatalln("Failed to convert JSON to YAML for chart:", chart, " : ", err)
	}
	chartDir, err := os.MkdirTemp(os.TempDir(), "nixchart-")
	if err != nil {
		log.Fatalln("Failed to create temporary directory for chart:", chart, " : ", err)
	}
	if err = os.WriteFile(chartDir+"/resources.yaml", yaml, 0644); err != nil {
		log.Fatalln("Failed to write resources.yaml for chart:", chart, " : ", err)
	}

	return chartDir, nil
}
