package helmfile

import (
	"fmt"
	"log"
	"os"
	"strings"
)

// Writer handles writing helmfile YAML files.
type Writer struct{}

// NewWriter creates a new helmfile writer.
func NewWriter() *Writer {
	return &Writer{}
}

// WriteYAML writes the helmfile YAML to a temporary file in the base directory.
// The caller is responsible for removing the file after use.
func (w *Writer) WriteYAML(fileName, base string, content []byte) (*os.File, error) {
	extension := "yaml"
	if strings.HasSuffix(fileName, ".gotmpl.nix") {
		extension = "yaml.gotmpl"
	}

	f, err := os.CreateTemp(base, fmt.Sprintf("helmfile.*.%s", extension))
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := f.Close(); err != nil {
			log.Fatalf("Could not close helmfile.yaml: %s", err)
		}
	}()

	if _, err := f.Write(content); err != nil {
		return nil, err
	}

	return f, nil
}
