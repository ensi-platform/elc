package src

import (
	"bytes"
	"os"
	"os/exec"
)

func contains(list []string, item string) bool {
	for _, value := range list {
		if value == item {
			return true
		}
	}
	return false
}

func execIntercative(command []string, env []string) (int, error) {
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

func execComposeCommandInteractive(st *State, name string, composeCommand []string) (int, error) {
	env, err := st.GetEnv(name)
	if err != nil {
		return 0, err
	}

	composeFile, err := st.GetComposeFile(name)
	if err != nil {
		return 0, err
	}

	command := append([]string{"docker", "compose", "-f", composeFile}, composeCommand...)

	code, err := execIntercative(command, env)
	if err != nil {
		return 0, err
	}

	return code, nil
}

func execComposeCommandToString(st *State, name string, composeCommand []string) (int, string, error) {
	env, err := st.GetEnv(name)
	if err != nil {
		return 0, "", err
	}

	composeFile, err := st.GetComposeFile(name)
	if err != nil {
		return 0, "", err
	}

	command := append([]string{"docker", "compose", "-f", composeFile}, composeCommand...)

	code, out, err := execToString(command, env)
	if err != nil {
		return 0, "", err
	}

	return code, out, nil
}
