package main

import (
	"fmt"
	elc "github.com/madridianfox/elc/src"
	"os"
	"os/user"
	"path"
	"strings"
)

func main() {
	currentUser, err := user.Current()
	elc.CheckRootError(err)

	homeConfigPath := path.Join(currentUser.HomeDir, ".elc.yaml")
	err = elc.CheckHomeConfigIsEmpty(homeConfigPath)
	elc.CheckRootError(err)

	hc, err := elc.LoadHomeConfig(homeConfigPath)
	elc.CheckRootError(err)

	args := os.Args[1:]
	if len(args) == 0 {
		printHelp()
	}

	firstArg := args[0]
	if firstArg == "-h" || firstArg == "--help" {
		printHelp()
	} else if firstArg == "workspace" {
		code, err := elc.RunWorkspaceAction(hc, args[1:])
		elc.CheckRootError(err)

		os.Exit(code)
	} else {
		workdir, err := hc.GetCurrentWsPath()
		cwd, err := os.Getwd()
		elc.CheckRootError(err)

		st := elc.NewState(workdir, cwd)
		err = st.LoadConfig()
		elc.CheckRootError(err)

		code, err := elc.RunAction(st, os.Args[1:])
		elc.CheckRootError(err)

		os.Exit(code)
	}
}

func printHelp() {
	fmt.Println(
		strings.Join([]string{
			"Usage: elc [options] [command] [args]",
			"",
			"Available commands:",
			"  workspace",
			"  compose",
			"  start",
			"  stop",
			"  destroy",
			"",
			"Global options:",
			"  -h, --help",
		}, "\n"))
	os.Exit(1)
}
