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
	var err error
	var returnCode int

	currentUser, err := user.Current()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	homeConfigPath := path.Join(currentUser.HomeDir, ".elc.yaml")

	switch os.Args[1] {
	case "-h", "--help":
		printHelp()
	case "workspace":
		switch os.Args[2] {
		case "list", "ls":
			err = elc.CmdWorkspaceList(homeConfigPath)
		case "add":
			err = elc.CmdWorkspaceAdd(homeConfigPath, os.Args[3:])
		case "select":
			err = elc.CmdWorkspaceSelect(homeConfigPath, os.Args[3:])
		case "show":
			err = elc.CmdWorkspaceShow(homeConfigPath)
		default:
			err = elc.CmdWorkspaceHelp()
		}
	case "start":
		err = elc.CmdServiceStart(homeConfigPath, os.Args[2:])
	case "stop":
		err = elc.CmdServiceStop(homeConfigPath, os.Args[2:])
	case "restart":
		err = elc.CmdServiceRestart(homeConfigPath, os.Args[2:])
	case "destroy":
		err = elc.CmdServiceDestroy(homeConfigPath, os.Args[2:])
	case "compose":
		returnCode, err = elc.CmdServiceCompose(homeConfigPath, os.Args[2:])
	case "vars":
		err = elc.CmdServiceVars(homeConfigPath, os.Args[2:])
	default:
		returnCode, err = elc.CmdServiceExec(homeConfigPath, os.Args[1:])
	}

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	os.Exit(returnCode)
}

func printHelp() {
	fmt.Println(
		strings.Join([]string{
			"Usage: elc command [options] [args]",
			"",
			"Available commands:",
			fmt.Sprintf("  %-15s - %s", "workspace", "manage workspaces"),
			fmt.Sprintf("  %-15s - %s", "compose", "run docker-compose command"),
			fmt.Sprintf("  %-15s - %s", "start", "start service"),
			fmt.Sprintf("  %-15s - %s", "stop", "stop service"),
			fmt.Sprintf("  %-15s - %s", "restart", "restart service"),
			fmt.Sprintf("  %-15s - %s", "destroy", "delete service containers"),
			fmt.Sprintf("  %-15s - %s", "vars", "print variables"),
			fmt.Sprintf("  %-15s - %s", "help", "print this help message"),
			"",
			"Any other command will be executed as \"compose exec command\"",
		}, "\n"))
	os.Exit(1)
}
