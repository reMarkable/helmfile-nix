// Package filesystem provides file system utility functions for helmfile-nix.
package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
)

// FindFileNameAndBase - find the filename and base directory of the helmfile.nix
func FindFileNameAndBase(input string, wanted []string) (string, string, error) {
	path, err := filepath.Abs(input)
	if err != nil {
		return "", "", err
	}

	fileInfo, err := os.Stat(path)
	if err != nil {
		return "", "", err
	}

	// If it's a directory, we need to find the helmfile
	if fileInfo.IsDir() {
		// Read the contents of the directory
		files, err := os.ReadDir(path)
		if err != nil {
			return "", "", err
		}

		// Check if the desired file exists in the directory
		for _, file := range files {
			for _, w := range wanted {
				if file.Name() == w {
					return w, path, nil
				}
			}
		}

	}
	// Check if the file is one of the wanted files
	for _, w := range wanted {
		if filepath.Base(path) == w {
			return w, filepath.Dir(path), nil
		}
	}

	return "", "", fmt.Errorf("expected %v,  found : %s ", wanted, path)
}
