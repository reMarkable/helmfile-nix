package main

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/andreyvit/diff"
)

var cwd, _ = os.Getwd()

func TestBase(t *testing.T) {
	inputs := []string{"testData/helm", "testData/helm/helmfile.nix", "./testData/helm/helmfile.nix", cwd + "/testData/helm/helmfile.nix", "testData/helm/helmfile.nix"}
	for _, input := range inputs {
		opts = Options{File: input, Env: "test"}
		base, err := findBase()
		if err != nil {
			t.Error("full path failed:", err)
		}
		if base != cwd+"/testData/helm" {
			t.Error("Base not matched: ", base, " != ", cwd+"testData/helm")
		}
	}
}

var output = `environments:
    dev:
        values: []
---
releases:
    - chart: grafana/grafana
      name: grafana
repositories:
    - name: grafana
      url: https://grafana.github.io/helm-charts
`

func TestRender(t *testing.T) {
	hf, err := renderHelmfile(cwd+"/testData/helm", "dev")
	if err != nil {
		t.Error("Failed to parse helmfile: ", err)
	}
	if string(hf) != output {
		t.Errorf("Result not as expected:\n%v", diff.LineDiff(string(hf), output))
	}
}

func TestTemplate(t *testing.T) {
	storeStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	callHelmfile([]byte(output), []string{"lint"}, cwd+"/testData/helm", "dev")
	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = storeStdout

	// restore the stdout
	if !strings.Contains(string(out), "1 chart(s) linted, 0 chart(s) failed\n") {
		t.Error("Output not matched: ::", string(out), "::")
	}
}
