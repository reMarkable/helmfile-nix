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
var eval string

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
	retcode := 0
	defer func() {
		os.Exit(retcode)
	}()

	args, err := parseArgs()
	if err != nil {
		l.Println("Could not parse args: ", err)
		retcode = 1
		return
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
		err := callHelmfile("", []string{}, ".", opts.Env)
		if err != nil {
			l.Println("Running helmfile failed: ", err)
		}
		retcode = 1
		return
	}
	l.Printf("Args: %v\n", args)
	l.Printf("file: %v env: %v\n", opts.Env, opts.File)

	hfFileName, base, err := findFileNameAndBase()
	if err != nil {
		l.Fatalln("Could not find helmfile.nix or helmfile.gotmpl.nix: ", err)
	}
	hfContent, err := renderHelmfile(hfFileName, base, opts.Env)
	if err != nil {
		l.Fatalln("Failed to render helmfile: ", err)
	}

	if args[len(args)-1] == "render" {
		fmt.Println(string(hfContent))
		return
	}

	hfFile, err := writeHelmfileYaml(hfFileName, base, hfContent)
	if err != nil {
		log.Fatalf("Could not write helmfile YAML: %s", err)
	}

	defer func() {
		if err := os.Remove(hfFile.Name()); err != nil {
			panic(fmt.Sprintf("unable to remove %s: %s", hfFile.Name(), err))
		}
	}()

	callErr := callHelmfile(hfFile.Name(), args[1:], base, opts.Env)

	if callErr != nil {
		log.Println("Running helmfile failed: ", err)
		retcode = 1
	}
}

// Call helmfile with the given arguments. Pipe the rendered helmfile into stdin.
func callHelmfile(hf string, args []string, base string, env string) error {
	baseArgs := []string{"-e", env}
	if len(hf) > 0 {
		baseArgs = append(baseArgs, "--file", hf)
	}
	err := os.Chdir(base)
	if err != nil {
		log.Fatalf("Could not change directory to %s: %s\n", base, err)
	}
	finalArgs := append(baseArgs, args...)
	fmt.Printf("calling helmfile %s\n", strings.Join(finalArgs[1:], " "))
	cmd := exec.Command("helmfile", finalArgs...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	return cmd.Run()
}

// Find the filename and base directory of the helmfile.nix
func findFileNameAndBase() (string, string, error) {
	path, err := filepath.Abs(opts.File)
	if err != nil {
		return "", "", err
	}

	fileInfo, err := os.Stat(path)
	if err != nil {
		return "", "", err
	}

	// Check if the file is a directory
	if fileInfo.IsDir() {
		// Read the contents of the directory
		files, err := os.ReadDir(path)
		if err != nil {
			return "", "", err
		}

		// Check if the desired file exists in the directory
		for _, file := range files {
			if file.Name() == "helmfile.nix" {
				return "helmfile.nix", path, nil
			} else if file.Name() == "helmfile.gotmpl.nix" {
				return "helmfile.gotmpl.nix", path, nil
			}
		}

		l.Fatalln("No helmfile.nix or helmfile.gotmpl.nix found in: ", path)
	}

	if filepath.Base(path) != "helmfile.nix" && filepath.Base(path) != "helmfile.gotmpl.nix" {
		l.Fatalln("Trying to use a file that is not helmfile.nix or helmfile.gotmpl.nix: ", path)
	}

	return filepath.Base(path), filepath.Dir(path), nil
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
func renderHelmfile(fileName, base string, env string) ([]byte, error) {
	f, err := writeEvalNix()
	if err != nil {
		log.Fatalf("Could not write eval.nix: %s", err)
	}
	defer func() {
		if err := os.Remove(f.Name()); err != nil {
			l.Fatalf("Could not remove eval.nix: %s", err)
		}
	}()
	val, err := writeValJson(base, opts.Env, opts.StateValuesSet)
	if err != nil {
		log.Fatalf("Could not write values.json: %s", err)
	}
	defer func() {
		if err := os.Remove(val.Name()); err != nil {
			l.Fatalf("Could not remove values.json: %s", err)
		}
	}()

	expr := fmt.Sprintf(`(import %s).render "%s" "%s" "%s" "%s"`, f.Name(), fileName, base, env, val.Name())
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
	defer func() {
		if err := f.Close(); err != nil {
			log.Fatalf("Could not close eval.nix: %s", err)
		}
	}()
	if _, err := f.WriteString(eval); err != nil {
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
	defer func() {
		if err := f.Close(); err != nil {
			log.Fatalf("Could not close values.json: %s", err)
		}
	}()
	// Get defaults
	defaultVal, err := os.ReadFile(state + "/env/defaults.yaml")
	if err != nil {
		if os.IsNotExist(err) {
			defaultVal = []byte("{}")
		} else {
			return nil, err
		}
	}
	var m, n map[string]interface{}
	err = yaml.Unmarshal(defaultVal, &m)
	if err != nil {
		return nil, err
	}
	// Get env specific values
	envVal, err := os.ReadFile(state + "/env/" + env + ".yaml")
	if err != nil {
		if os.IsNotExist(err) {
			envVal = []byte("{}")
		} else {
			return nil, err
		}
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
						mref[key], err = unmarshalOption(kv[1])
						if err != nil {
							return nil, err
						}
					} else {
						if _, ok := m[key]; !ok {
							mref[key] = make(map[string]interface{})
						}
						mref = m[key].(map[string]interface{})
					}
				}
			} else {
				m[kv[0]], err = unmarshalOption(kv[1])
				if err != nil {
					return nil, err
				}
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
		return nil, err
	}
	return f, nil
}

// writeHelmfileYaml writes the helmfile.yaml to a temporary file in the state base dir. Must be manually removed.
func writeHelmfileYaml(fileName, base string, hf []byte) (*os.File, error) {
	extension := "yaml"
	if strings.HasSuffix(fileName, ".gotmpl.nix") {
		extension = "yaml.gotmpl"
	}

	f, err := os.CreateTemp(base, fmt.Sprintf("helmfile.*.%s", extension))
	defer func() {
		if err := f.Close(); err != nil {
			log.Fatalf("Could not close helmfile.yaml: %s", err)
		}
	}()
	if err != nil {
		return nil, err
	}

	if _, err := f.Write(hf); err != nil {
		return nil, err
	}

	return f, nil
}

func unmarshalOption(val string) (interface{}, error) {
	var v interface{}
	err := yaml.Unmarshal([]byte(val), &v)
	if err != nil {
		return nil, err
	}
	return v, nil
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
