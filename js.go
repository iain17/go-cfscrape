package cfscrape

import (
	"bytes"
	"os/exec"
	"strings"
)

var NodeExecutor = JSExecutorFunc(NodeExecute)

// A JSExecutor executes a javascript expression, and returns the resulting
// string. Note that the javascript that it must execute can be arbitrary.
type JSExecutor interface {
	ExecuteJS(javascript string) (string, error)
}

type JSExecutorFunc func(string) (string, error)

func (f JSExecutorFunc) ExecuteJS(s string) (string, error) {
	return f(s)
}

// Executes a javascript expression with node. This executor makes no security
// guarantees.
func NodeExecute(javascript string) (string, error) {
	var stdout bytes.Buffer

	cmd := exec.Command("node", "-p")
	cmd.Stdin = strings.NewReader(javascript)
	cmd.Stdout = &stdout

	err := cmd.Run()
	if err != nil {
		return "", err
	}

	return stdout.String(), nil
}
