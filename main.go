package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"

	_ "embed"

	flags "github.com/jessevdk/go-flags"
	"github.com/reMarkable/helmfile-nix/pkgs/environment"
	"github.com/reMarkable/helmfile-nix/pkgs/filesystem"
	"github.com/reMarkable/helmfile-nix/pkgs/nixchart"
	"github.com/reMarkable/helmfile-nix/pkgs/nixeval"
	"github.com/reMarkable/helmfile-nix/pkgs/tempfiles"
	"github.com/reMarkable/helmfile-nix/pkgs/transform"
)

//go:embed eval.nix
var eval string

var version = "dev"

// List of temporary directories that need to be cleaned up after use.
var cleanup []string

// Options - We only care about these settings, the remaining are passed through unharmed to helmfile
type Options struct {
	File           string   `short:"f" long:"file" description:"helmfile.nix to use" default:"."`
	Env            string   `short:"e" long:"environment" description:"Environment to deploy to" default:"dev"`
	ShowTrace      []bool   `long:"show-trace" description:"Enable stacktraces"`
	StateValuesSet []string `long:"state-values-set" description:"Set state values"`
	Version        bool     `short:"v" long:"version" description:"Print version and exit"`
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
	if opts.Version {
		fmt.Printf("helmfile-nix version %s\n", version)
		cmd := exec.Command("helmfile", "--version")
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		callErr := cmd.Run()
		if callErr != nil {
			log.Println("Running helmfile failed: ", err)
			retcode = 1
		}
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

	hfFileName, base, err := filesystem.FindFileNameAndBase(opts.File, []string{"helmfile.nix", "helmfile.gotmpl.nix"})
	if err != nil {
		l.Fatalln("Could not find helmfile: ", err)
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

	nixchart.CleanupCharts(cleanup)
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

// Parse the command line arguments, return remaining arguments.
func parseArgs() ([]string, error) {
	args, err := parser.ParseArgs(os.Args)
	if err != nil {
		return nil, err
	}

	return args, nil
}

// Render the helmfile using the nix command.
func renderHelmfile(fileName, base string, env string) ([]byte, error) {
	f, err := tempfiles.WriteEvalNix(eval)
	if err != nil {
		log.Fatalf("Could not write eval.nix: %s", err)
	}

	defer func() {
		if err := os.Remove(f.Name()); err != nil {
			l.Fatalf("Could not remove eval.nix: %s", err)
		}
	}()

	val, err := writeValJSON(base, opts.Env, opts.StateValuesSet)
	if err != nil {
		log.Fatalf("Could not write values.json: %s", err)
	}

	defer func() {
		if err := os.Remove(val.Name()); err != nil {
			l.Fatalf("Could not remove values.json: %s", err)
		}
	}()

	expr := fmt.Sprintf(`(import %s).render "%s" "%s" "%s" "%s"`, f.Name(), fileName, base, env, val.Name())
	ne := nixeval.NewNixEval(expr)
	cmd := ne.Args(len(opts.ShowTrace) > 0)
	json, err := ne.Eval(cmd)
	if err != nil {
		l.Fatalf("Failed to eval nix: %s\n%s", err, json)
	}

	yaml, err := transform.JSONToYAMLs(json, func(v any) {
		if reflect.TypeOf(v).Kind() == reflect.Map {
			// Check if map has a list of releases
			if _, ok := v.(map[string]any)["releases"]; ok {
				var err error
				cleanup, err = nixchart.RenderCharts(v.(map[string]any), base)
				if err != nil {
					log.Fatalf("Failed to render charts: %s", err)
				}
			}
		}
	})
	if err != nil {
		l.Fatalf("Failed to convert JSON to YAML: %s\n%s", err, json)
	}

	return yaml, nil
}

// Merge the yaml values to a json file for nix.
func writeValJSON(state string, env string, overrides []string) (*os.File, error) {
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
	defaultsPath := filepath.Join(state, "env", "defaults.yaml")
	envPath := filepath.Join(state, "env", env+".yaml")

	m, err := environment.LoadYamlFile(defaultsPath)
	if err != nil {
		return nil, err
	}

	n, err := environment.LoadYamlFile(envPath)
	if err != nil {
		return nil, err
	}

	m = environment.MergeMaps(m, n)
	// Handle state overrides
	for _, v := range overrides {
		vals := strings.SplitSeq(v, ",")
		for val := range vals {
			kv := strings.Split(val, "=")
			if len(kv) != 2 {
				return nil, fmt.Errorf("invalid state value: %s", val)
			}

			if err := environment.SetNestedMapValue(m, kv[0], kv[1]); err != nil {
				return nil, fmt.Errorf("could not set nested map value %s: %w", kv[0], err)
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
