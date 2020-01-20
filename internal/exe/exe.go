package exe

import (
	"bytes"
	"fmt"
	"github.com/go-logr/logr"
	"os/exec"
)

// Args are the command arguments.
type Args []string

// Opt are the command options.
// See https://godoc.org/os/exec#Cmd for details.
type Opt struct {
	// Dir is the working directory.
	Dir string
	// Env is the execution environment.
	Env []string
}

// Action executes 'cmd' with 'args' and 'options'.
// Upon completion it returns stdout and stderr.
func Run(cmd string, args Args, options Opt, log logr.Logger) (string, string, error) {
	log.V(2).Info("Action", "cmd", cmd, "args", args)
	c := exec.Command(cmd, args...)
	c.Env = options.Env
	c.Dir = options.Dir
	var stdout, stderr bytes.Buffer
	c.Stdout, c.Stderr = &stdout, &stderr
	err := c.Run()
	outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())
	log.V(3).Info("Action-result", "stderr", errStr, "stdout", outStr)
	if err != nil {
		return "", "", fmt.Errorf("%s %v: %w - %s", cmd, args, err, errStr)
	}

	return outStr, errStr, nil
}
