package nixeval

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"strings"
)

type NixEval struct {
	expr string
}

func NewNixEval(expr string) *NixEval {
	return &NixEval{expr: expr}
}

func (n *NixEval) Eval(cmd []string) ([]byte, error) {
	eval := exec.Command("nix", cmd...)
	log.Println("Running nix", strings.Join(cmd, " "))
	eval.Stderr = os.Stderr
	var out bytes.Buffer
	eval.Stdout = &out
	err := eval.Run()
	if err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func (n *NixEval) Args() []string {
	return []string{
		"--extra-experimental-features", "nix-command",
		"--extra-experimental-features", "flakes",
		"eval",
		"--json",
		"--impure",
		"--expr", n.expr,
	}
}
