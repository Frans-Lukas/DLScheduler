package helperFunctions

import (
	"bytes"
	"os/exec"
)

func ExecuteFunction(name string, args ...string) (bytes.Buffer, bytes.Buffer, error) {
	cmd := exec.Command(name, args...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()

	return out, stderr, err
}