package geamon

import (
	"io"
	"os"
	"os/exec"
)

func fork(stdout io.Writer, stderr io.Writer, env []string) (*exec.Cmd, error) {
	cmd := &exec.Cmd{
		Path: os.Args[0],
		Args: os.Args,
	}

	envs := os.Environ()
	cmd.Env = append(envs, env...)
	cmd.SysProcAttr = sysProcAttr()
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return cmd, nil
}
