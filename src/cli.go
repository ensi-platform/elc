package src

import (
	"errors"
	"fmt"
	"os"
)

func getAction(args []string) (string, error) {
	if len(args) < 1 {
		return "", errors.New("Too few arguments")
	}

	return args[0], nil
}

func getServiceNames(st *State, args []string) ([]string, error) {
	var svcNames []string

	if len(args) > 0 {
		svcNames = args
	} else {
		svcNames = make([]string, 0)
		svcName, err := st.FindServiceByPath()
		if err != nil {
			return nil, err
		}
		svcNames = append(svcNames, svcName)
	}

	return svcNames, nil
}

func CheckRootError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
