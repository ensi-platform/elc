package src

import (
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

func execComposeCommand(st *State, name string, composeCommand []string) (int, error) {
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
