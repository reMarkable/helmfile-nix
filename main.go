package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	_ "embed"

	flags "github.com/jessevdk/go-flags"
	"gopkg.in/yaml.v3"
)

//go:embed eval.nix
var eval []byte

// We only care about these two, the remaining are passed through unharmed to helmfile
type Options struct {
	File           string   `short:"f" long:"file" description:"helmfile.nix to use" default:"."`
	Env            string   `short:"e" long:"environment" description:"Environment to deploy to" default:"dev"`
	ShowTrace      []bool   `long:"show-trace" description:"Enable stacktraces"`
	StateValuesSet []string `long:"state-values-set" description:"Set state values"`
}

var (
	opts   Options
	parser = flags.NewParser(&opts, flags.IgnoreUnknown|flags.PassDoubleDash)
	l      = log.Default()
)

// Main app flow.
func main() {
	args, err := parseArgs()
	if err != nil {
		l.Fatalln("Could not parse args: ", err)
	}
	seen := false
	for _, v := range args[1:] {
		if v[0] != '-' {
			seen = true
		}
	}

	if !seen {
		l.Println("No command provided. Call 'render' to see the rendered helmfile.")
		l.Println("forwarding to helmfile help:")
		callHelmfile([]byte(""), []string{}, ".", opts.Env)
		os.Exit(1)
	}
	l.Printf("Args: %v\n", args)
	l.Printf("file: %v env: %v\n", opts.Env, opts.File)

	base, err := findBase()
	if err != nil {
		l.Fatalln("Could not find helmfile.nix: ", err)
	}
	hf, err := renderHelmfile(base, opts.Env)
	if err != nil {
		l.Fatalln("Failed to render helmfile: ", err)
	}
	if args[len(args)-1] == "render" {
		fmt.Println(string(hf))
		os.Exit(0)
	}
	callHelmfile(hf, args[1:], base, opts.Env)
}

// Call helmfile with the given arguments. Pipe the rendered helmfile into stdin.
func callHelmfile(hf []byte, args []string, base string, env string) {
	baseArgs := []string{"--file", "-", "-e", env}
	err := os.Chdir(base)
	if err != nil {
		log.Fatalf("Could not change directory to %s: %s\n", base, err)
	}
	finalArgs := append(baseArgs, args...)
	cmd := exec.Command("helmfile", finalArgs...)
	cmd.Stdin = strings.NewReader(string(hf))
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	if err != nil {
		log.Fatalln("Running helmfile failed: ", err)
	}
}

// Find the base directory of the helmfile.nix
func findBase() (string, error) {
	path, err := filepath.Abs(opts.File)
	if err != nil {
		return "", err
	}
	fileInfo, err := os.Stat(path)
	if err != nil {
		return "", err
	}

	// Check if the file is a directory
	if fileInfo.IsDir() {

		// Read the contents of the directory
		files, err := os.ReadDir(path)
		if err != nil {
			return "", err
		}

		// Check if the desired file exists in the directory
		for _, file := range files {
			if file.Name() == "helmfile.nix" {
				return path, nil
			}
		}
		l.Fatalln("No helmfile.nix found in: ", path)
	}
	if filepath.Base(path) != "helmfile.nix" {
		l.Fatalln("Trying to use a file that is not helmfile.nix: ", path)
	}
	return filepath.Dir(path), nil
}

// Parse the command line arguments, return remaining arguments.
func parseArgs() ([]string, error) {
	args, err := parser.ParseArgs(os.Args)
	if err != nil {
		return nil, err
	}
	return args, nil
}

// Convert a JSON list to YAML documents.
func JSONToYAMLs(j []byte) ([]byte, error) {
	// Convert the JSON to a list of object.
	var jsonObj []interface{}
	// We are using yaml.Unmarshal here (instead of json.Unmarshal) because the
	// Go JSON library doesn't try to pick the right number type (int, float,
	// etc.) when unmarshalling to interface{}, it just picks float64
	// universally. go-yaml does go through the effort of picking the right
	// number type, so we can preserve number type throughout this process.
	err := yaml.Unmarshal(j, &jsonObj)
	if err != nil {
		return nil, err
	}

	var y []byte
	// Marshal this object into YAML.
	for _, v := range jsonObj {
		res, err := yaml.Marshal(v)
		if err != nil {
			return nil, err
		}
		if len(y) > 0 {
			y = append(y, []byte("---\n")...)
		}
		y = append(y, res...)

	}
	return y, nil
}

// Render the helmfile using the nix command.
func renderHelmfile(base string, env string) ([]byte, error) {
	f, err := writeEvalNix()
	if err != nil {
		log.Fatalf("Could not write eval.nix: %s", err)
	}
	defer os.Remove(f.Name())
	val, err := writeValJson(base, opts.Env, opts.StateValuesSet)
	if err != nil {
		log.Fatalf("Could not write values.json: %s", err)
	}
	defer os.Remove(val.Name())
	expr := fmt.Sprintf("(import %s).render \"%s\" \"%s\" \"%s\"", f.Name(), base, env, val.Name())
	cmd := []string{
		"--extra-experimental-features", "nix-command",
		"--extra-experimental-features", "flakes",
		"eval",
		"--json",
		"--impure",
		"--expr", expr,
	}
	if len(opts.ShowTrace) > 0 {
		cmd = append(cmd, "--show-trace")
	}
	eval := exec.Command("nix", cmd...)
	l.Println("Running nix", strings.Join(cmd, " "))
	eval.Stderr = os.Stderr
	var out bytes.Buffer
	eval.Stdout = &out
	err = eval.Run()
	if err != nil {
		l.Fatalf("error: %s\n%s", err, out.String())
	}
	json := out.Bytes()
	yaml, err := JSONToYAMLs(json)
	if err != nil {
		l.Fatalf("Failed to convert JSON to YAML: %s\n%s", err, json)
	}
	return yaml, nil
}

// Write the eval.nix to a temporary file. Must be removed after use.
func writeEvalNix() (*os.File, error) {
	f, err := os.CreateTemp("", "eval.*.nix")
	if err != nil {
		return nil, err
	}
	if _, err := f.Write(eval); err != nil {
		f.Close()
		return nil, err
	}
	if err := f.Close(); err != nil {
		return nil, err
	}
	return f, nil
}

// Merge the yaml values to a json file for nix.
// FIXME: This could probably be simplified
func writeValJson(state string, env string, overrides []string) (*os.File, error) {
	f, err := os.CreateTemp("", "val.*.json")
	if err != nil {
		return nil, err
	}
	// Get defaults
	defaultVal, err := os.ReadFile(state + "/env/defaults.yaml")
	if err != nil {
		return nil, err
	}
	var m, n map[string]interface{}
	err = yaml.Unmarshal(defaultVal, &m)
	if err != nil {
		return nil, err
	}
	// Get env specific values
	envVal, err := os.ReadFile(state + "/env/" + env + ".yaml")
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(envVal, &n)
	if err != nil {
		return nil, err
	}
	m = mergeMaps(m, n)
	// Handle state overrides
	for _, v := range overrides {
		vals := strings.Split(v, ",")
		for _, val := range vals {
			kv := strings.Split(val, "=")
			if len(kv) != 2 {
				return nil, fmt.Errorf("invalid state value: %s", val)
			}
			if strings.Contains(kv[0], ".") {
				// Nested value
				// Split the key into parts
				mref := m
				keys := strings.Split(kv[0], ".")
				for i, key := range keys {
					if i == len(keys)-1 {
						mref[key] = kv[1]
					} else {
						if _, ok := m[key]; !ok {
							mref[key] = make(map[string]interface{})
						}
						mref = m[key].(map[string]interface{})
					}
				}
			} else {
				var val interface{}
				err := json.Unmarshal([]byte(kv[1]), &val)
				if err != nil {
					return nil, fmt.Errorf("failed to marshal %s: %s", kv[1], err)
				}
				m[kv[0]] = val
			}
		}
	}
	// Serialize the values
	envStr, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	// Write the values
	if _, err := f.Write(envStr); err != nil {
		f.Close()
		return nil, err
	}
	if err := f.Close(); err != nil {
		return nil, err
	}
	return f, nil
}

func mergeMaps(a, b map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(a))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		if v, ok := v.(map[string]interface{}); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(map[string]interface{}); ok {
					out[k] = mergeMaps(bv, v)
					continue
				}
			}
		}
		out[k] = v
	}
	return out
}
