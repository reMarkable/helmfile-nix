// Package helmfile provides operations for working with helmfile.
package helmfile

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

// Executor handles execution of the helmfile binary.
type Executor struct {
	logger *log.Logger
}

// NewExecutor creates a new helmfile executor.
func NewExecutor(logger *log.Logger) *Executor {
	return &Executor{
		logger: logger,
	}
}

// Execute calls helmfile with the given arguments.
func (e *Executor) Execute(hfFile string, args []string, base string, env string) error {
	baseArgs := []string{"-e", env}
	if len(hfFile) > 0 {
		baseArgs = append(baseArgs, "--file", hfFile)
	}
	err := os.Chdir(base)
	if err != nil {
		return fmt.Errorf("could not change directory to %s: %w", base, err)
	}

	finalArgs := append(baseArgs, args...)
	fmt.Printf("calling helmfile %s\n", strings.Join(finalArgs[1:], " "))
	cmd := exec.Command("helmfile", finalArgs...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	return cmd.Run()
}
