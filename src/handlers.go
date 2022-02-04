package src

import (
	"errors"
)

func RunAction(st *State, args []string) (int, error) {
	var code int
	var err error
	var action string

	action, err = getAction(args)
	if err != nil {
		return 0, err
	}

	switch action {
	case "start":
		code, err = actionStart(st, args[1:])
	case "stop":
		code, err = actionStop(st, args[1:])
	case "destroy":
		code, err = actionDestroy(st, args[1:])
	case "compose":
		code, err = actionCompose(st, args[1:])
	default:
		code, err = actionExec(st, args)
	}

	if err != nil {
		return 0, err
	}

	return code, nil
}

func actionStart(st *State, args []string) (int, error) {
	if contains(args, "--help") || contains(args, "-h") {
		return 0, errors.New("Usage: elc start [service]")
	}

	svcNames, err := getServiceNames(st, args)
	if err != nil {
		return 0, err
	}

	var code int

	for _, svcName := range svcNames {
		running, err := checkIsRunning(st, svcName)
		if err != nil {
			return 0, err
		}
		if !running {
			err = startDependencies(st, svcName)
			if err != nil {
				return 0, err
			}
			code, err = execComposeCommandInteractive(st, svcName, []string{"up", "-d"})
			if err != nil {
				return 0, err
			}
		}
	}

	return code, nil
}

func actionStop(st *State, args []string) (int, error) {
	if contains(args, "--help") || contains(args, "-h") {
		return 0, errors.New("Usage: elc stop [service]")
	}

	svcNames, err := getServiceNames(st, args)
	if err != nil {
		return 0, err
	}

	var code int

	for _, svcName := range svcNames {
		code, err = execComposeCommandInteractive(st, svcName, []string{"stop"})
		if err != nil {
			return 0, err
		}
	}

	return code, nil
}

func actionDestroy(st *State, args []string) (int, error) {
	if contains(args, "--help") || contains(args, "-h") {
		return 0, errors.New("Usage: elc down [service]")
	}

	svcNames, err := getServiceNames(st, args)
	if err != nil {
		return 0, err
	}

	var code int

	for _, svcName := range svcNames {
		code, err = execComposeCommandInteractive(st, svcName, []string{"down"})
		if err != nil {
			return 0, err
		}
	}

	return code, nil
}

func actionCompose(st *State, args []string) (int, error) {
	svcName, err := st.FindServiceByPath()
	if err != nil {
		return 0, err
	}

	return execComposeCommandInteractive(st, svcName, args)
}

func actionExec(st *State, args []string) (int, error) {
	svcName, err := st.FindServiceByPath()
	if err != nil {
		return 0, err
	}

	_, err = actionStart(st, []string{})
	if err != nil {
		return 0, err
	}

	command := append([]string{"exec", "app"}, args...)

	return execComposeCommandInteractive(st, svcName, command)
}

func checkIsRunning(st *State, svcName string) (bool, error) {
	_, out, err := execComposeCommandToString(st, svcName, []string{"ps", "--status=running", "-q"})
	if err != nil {
		return false, err
	}

	running := out != ""

	return running, nil
}

func startDependencies(st *State, svcName string) error {
	svc, err := FindServicebyName(st.Config.Services, svcName)
	if err != nil {
		return nil
	}
	for _, depName := range svc.Dependencies {
		_, err = actionStart(st, []string{depName})
		if err != nil {
			return nil
		}
	}

	return nil
}
