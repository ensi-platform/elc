package src

import (
	"bytes"
	"os"
	"os/exec"
)

func execInteractive(command []string, env []string) (int, error) {
	cmd := exec.Command(command[0], command[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = env

	err := cmd.Run()

	return cmd.ProcessState.ExitCode(), err
}

func execToString(command []string, env []string) (int, string, error) {
	var buff bytes.Buffer
	cmd := exec.Command(command[0], command[1:]...)
	cmd.Stdout = &buff
	cmd.Env = env

	err := cmd.Run()
	return cmd.ProcessState.ExitCode(), buff.String(), err
}
