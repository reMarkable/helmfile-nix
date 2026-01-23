package helmfile

import (
	"fmt"
	"log"
	"os"
	"reflect"

	"github.com/reMarkable/helmfile-nix/pkgs/nixchart"
	"github.com/reMarkable/helmfile-nix/pkgs/nixeval"
	"github.com/reMarkable/helmfile-nix/pkgs/tempfiles"
	"github.com/reMarkable/helmfile-nix/pkgs/transform"
)

// Renderer handles rendering helmfile configurations via Nix evaluation.
type Renderer struct {
	evalNix        string
	showTrace      bool
	stateValuesSet []string
	logger         *log.Logger
}

// NewRenderer creates a new helmfile renderer.
func NewRenderer(evalNix string, showTrace bool, stateValuesSet []string, logger *log.Logger) *Renderer {
	return &Renderer{
		evalNix:        evalNix,
		showTrace:      showTrace,
		stateValuesSet: stateValuesSet,
		logger:         logger,
	}
}

// Render renders the helmfile using Nix evaluation.
// Returns the rendered YAML content and a slice of temporary chart directories that need cleanup.
func (r *Renderer) Render(fileName, base, env, valuesJSONPath string) ([]byte, []string, error) {
	f, err := tempfiles.WriteEvalNix(r.evalNix)
	if err != nil {
		return nil, nil, fmt.Errorf("could not write eval.nix: %w", err)
	}

	defer func() {
		if err := os.Remove(f.Name()); err != nil {
			r.logger.Fatalf("Could not remove eval.nix: %s", err)
		}
	}()

	expr := fmt.Sprintf(`(import %s).render "%s" "%s" "%s" "%s"`, f.Name(), fileName, base, env, valuesJSONPath)
	ne := nixeval.NewNixEval(expr)
	cmd := ne.Args(r.showTrace)
	json, err := ne.Eval(cmd)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to eval nix: %w\n%s", err, json)
	}

	var cleanup []string
	yaml, err := transform.JSONToYAMLs(json, func(v any) {
		if reflect.TypeOf(v).Kind() == reflect.Map {
			// Check if map has a list of releases
			if _, ok := v.(map[string]any)["releases"]; ok {
				var err error
				cleanup, err = nixchart.RenderCharts(v.(map[string]any), base)
				if err != nil {
					r.logger.Fatalf("Failed to render charts: %s", err)
				}
			}
		}
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to convert JSON to YAML: %w\n%s", err, json)
	}

	return yaml, cleanup, nil
}
