package main

import (
	"io"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/andreyvit/diff"
	"github.com/reMarkable/helmfile-nix/pkgs/environment"
	"github.com/reMarkable/helmfile-nix/pkgs/helmfile"
)

var cwd, _ = os.Getwd()

var output = `environments:
    dev:
        values: []
---
releases:
    - chart: ../chart/
      name: test
`

var outputTemplated = `environments:
    dev:
        values: []
---
releases:
    - chart: ../chart/
      hooks:
        - args:
            - --environment
            - '{{"{{ .Environment | toJson }}"}}'
            - --release
            - '{{"{{ .Release | toJson }}"}}'
            - --event
            - '{{"{{ .Event | toJson }}"}}'
          command: echo
          events:
            - presync
            - prepare
          showlogs: true
      name: test
`

func TestRender(t *testing.T) {
	logger := log.Default()
	valuesWriter := environment.NewValuesWriter(logger)
	renderer := helmfile.NewRenderer(eval, false, []string{}, logger)

	valJSON, err := valuesWriter.WriteJSON(cwd+"/testData/helm", "dev", []string{})
	if err != nil {
		t.Error("Failed to write values JSON: ", err)
	}
	defer func() {
		if err := os.Remove(valJSON.Name()); err != nil {
			t.Error("Failed to remove temp values file: ", err)
		}
	}()

	hf, _, err := renderer.Render("helmfile.nix", cwd+"/testData/helm", "dev", valJSON.Name())
	if err != nil {
		t.Error("Failed to parse helmfile: ", err)
	}
	if string(hf) != output {
		t.Errorf("Result not as expected:\n%v", diff.LineDiff(string(hf), output))
	}
}

func TestRenderTemplated(t *testing.T) {
	logger := log.Default()
	valuesWriter := environment.NewValuesWriter(logger)
	renderer := helmfile.NewRenderer(eval, false, []string{}, logger)

	valJSON, err := valuesWriter.WriteJSON(cwd+"/testData/helm-templated", "dev", []string{})
	if err != nil {
		t.Error("Failed to write values JSON: ", err)
	}
	defer func() {
		if err := os.Remove(valJSON.Name()); err != nil {
			t.Error("Failed to remove temp values file: ", err)
		}
	}()

	hf, _, err := renderer.Render("helmfile.gotmpl.nix", cwd+"/testData/helm-templated", "dev", valJSON.Name())
	if err != nil {
		t.Error("Failed to parse helmfile: ", err)
	}
	if string(hf) != outputTemplated {
		t.Errorf("Result not as expected:\n%v", diff.LineDiff(string(hf), outputTemplated))
	}
}

func TestTemplate(t *testing.T) {
	logger := log.Default()
	writer := helmfile.NewWriter()
	executor := helmfile.NewExecutor(logger)

	storeStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	hfFile, _ := writer.WriteYAML("helmfile.nix", cwd+"/testData/helm", []byte(output))
	defer func() {
		if err := os.Remove(hfFile.Name()); err != nil {
			panic("Failed to remove helmfile: " + err.Error())
		}
	}()

	err := executor.Execute(hfFile.Name(), []string{"lint"}, cwd+"/testData/helm", "dev")
	if err != nil {
		t.Error("Failed to call helmfile: ", err)
	}

	if err := w.Close(); err != nil {
		panic("Failed to close pipe: " + err.Error())
	}
	out, _ := io.ReadAll(r)
	os.Stdout = storeStdout

	// restore stdout
	if !strings.Contains(string(out), "1 chart(s) linted, 0 chart(s) failed\n") {
		t.Error("Output not matched: ::", string(out), "::")
	}
}

var vals = `{"bad":123,"bar":"true","foo":{"bad":"hello","bar":false,"baz":true,"foo":true}}`

func TestWriteValJson(t *testing.T) {
	logger := log.Default()
	valuesWriter := environment.NewValuesWriter(logger)

	f, err := valuesWriter.WriteJSON(cwd+"/testData/helm", "test", []string{"foo.bar=false", "bad=123", "foo.bad=hello"})
	if err != nil {
		t.Error("Failed to write values file: ", err)
	}
	defer func() {
		if err := os.Remove(f.Name()); err != nil {
			panic("Failed to remove values file: " + err.Error())
		}
	}()
	res, err := os.ReadFile(f.Name())
	if err != nil {
		t.Error("Failed to read file: ", err)
	}
	if string(res) != vals {
		t.Errorf("Result not as expected:\n%v", diff.LineDiff(string(res), vals))
	}
}
