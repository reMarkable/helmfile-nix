package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	_ "embed"

	flags "github.com/jessevdk/go-flags"
	"sigs.k8s.io/yaml"
)

//go:embed eval.nix
var eval []byte

type Options struct {
	// Slice of bool will append 'true' each time the option
	// is encountered (can be set multiple times, like -vvv)
	File string `short:"f" long:"file" description:"helmfile.nix to use" default:"."`
	Env  string `short:"e" long:"environment" description:"Environment to deploy to" default:"dev"`
}

var (
	opts   Options
	parser = flags.NewParser(&opts, flags.IgnoreUnknown|flags.PassDoubleDash)
	l      = log.Default()
)

// Start the program
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
	defer os.Remove(f.Name())
	if err != nil {
		log.Fatalf("Could not write eval.nix: %s", err)
	}
	expr := fmt.Sprintf("(import %s).render \"%s\" \"%s\"", f.Name(), base, env)
	cmd := []string{"--extra-experimental-features", "nix-command", "eval", "--json", "--impure", "--expr", expr}
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
