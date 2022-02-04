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
		return 0, nil
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
		code, err = execComposeCommand(st, svcName, []string{"up", "-d"})
		if err != nil {
			return 0, err
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
		code, err = execComposeCommand(st, svcName, []string{"stop"})
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
		code, err = execComposeCommand(st, svcName, []string{"down"})
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

	return execComposeCommand(st, svcName, args)
}
