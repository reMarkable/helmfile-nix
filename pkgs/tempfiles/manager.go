// Package tempfiles provides temporary file lifecycle management for helmfile-nix.
package tempfiles

import (
	"log"
	"os"
)

var l = log.Default()

// WriteEvalNix writes the eval.nix to a temporary file. Must be removed after use.
func WriteEvalNix(eval string) (*os.File, error) {
	f, err := os.CreateTemp("", "eval.*.nix")
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := f.Close(); err != nil {
			l.Fatalf("Could not close eval.nix: %s", err)
		}
	}()

	if _, err := f.WriteString(eval); err != nil {
		return nil, err
	}

	return f, nil
}
