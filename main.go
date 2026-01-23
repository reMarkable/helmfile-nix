package main

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"os/exec"

	flags "github.com/jessevdk/go-flags"

	"github.com/reMarkable/helmfile-nix/pkgs/environment"
	"github.com/reMarkable/helmfile-nix/pkgs/filesystem"
	"github.com/reMarkable/helmfile-nix/pkgs/helmfile"
	"github.com/reMarkable/helmfile-nix/pkgs/nixchart"
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
			l.Println("Running helmfile failed: ", callErr)
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

	executor := helmfile.NewExecutor(l)

	if !seen {
		l.Println("No command provided. Call 'render' to see the rendered helmfile.")
		l.Println("forwarding to helmfile help:")
		err := executor.Execute("", []string{}, ".", opts.Env)
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

	// Write environment values JSON
	valuesWriter := environment.NewValuesWriter(l)
	valJSON, err := valuesWriter.WriteJSON(base, opts.Env, opts.StateValuesSet)
	if err != nil {
		l.Fatalln("Could not write values.json: ", err)
	}

	defer func() {
		if err := os.Remove(valJSON.Name()); err != nil {
			l.Fatalf("Could not remove values.json: %s", err)
		}
	}()

	// Render helmfile
	renderer := helmfile.NewRenderer(eval, len(opts.ShowTrace) > 0, opts.StateValuesSet, l)
	hfContent, chartCleanup, err := renderer.Render(hfFileName, base, opts.Env, valJSON.Name())
	if err != nil {
		l.Fatalln("Failed to render helmfile: ", err)
	}
	cleanup = chartCleanup

	if args[len(args)-1] == "render" {
		fmt.Println(string(hfContent))
		return
	}

	writer := helmfile.NewWriter()
	hfFile, err := writer.WriteYAML(hfFileName, base, hfContent)
	if err != nil {
		l.Fatalf("Could not write helmfile YAML: %s", err)
	}

	defer func() {
		if err := os.Remove(hfFile.Name()); err != nil {
			panic(fmt.Sprintf("unable to remove %s: %s", hfFile.Name(), err))
		}
	}()

	callErr := executor.Execute(hfFile.Name(), args[1:], base, opts.Env)

	nixchart.CleanupCharts(cleanup)
	if callErr != nil {
		l.Println("Running helmfile failed: ", callErr)
		retcode = 1
	}
}

// Parse the command line arguments, return remaining arguments.
func parseArgs() ([]string, error) {
	args, err := parser.ParseArgs(os.Args)
	if err != nil {
		return nil, err
	}

	return args, nil
}
