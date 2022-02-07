package main

import (
	"fmt"
	elc "github.com/madridianfox/elc/src"
	"os"
	"os/user"
	"path"
)

func main() {
	if elc.NeedHelp(os.Args[1:], "COMMAND", []string{
		"Available commands:",
		fmt.Sprintf("  %-15s - %s", "exec", "execute command inside service's container"),
		fmt.Sprintf("  %-15s - %s", "compose", "run docker-compose command"),
		fmt.Sprintf("  %-15s - %s", "destroy", "delete service containers"),
		fmt.Sprintf("  %-15s - %s", "help", "print this help message"),
		fmt.Sprintf("  %-15s - %s", "restart", "restart service"),
		fmt.Sprintf("  %-15s - %s", "set-hooks", "install git hooks"),
		fmt.Sprintf("  %-15s - %s", "start", "start service"),
		fmt.Sprintf("  %-15s - %s", "stop", "stop service"),
		fmt.Sprintf("  %-15s - %s", "vars", "print variables"),
		fmt.Sprintf("  %-15s - %s", "workspace", "manage workspaces"),
		"Any other arguments will be used for invoke of implicit exec command.",
		"",
		"You can get help for any command invoke it with '--help' option.",
	}) {
		os.Exit(0)
	}
	var err error
	var returnCode int

	currentUser, err := user.Current()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	homeConfigPath := path.Join(currentUser.HomeDir, ".elc.yaml")

	switch os.Args[1] {
	case "workspace":
		switch os.Args[2] {
		case "list", "ls":
			err = elc.CmdWorkspaceList(homeConfigPath, os.Args[3:])
		case "add":
			err = elc.CmdWorkspaceAdd(homeConfigPath, os.Args[3:])
		case "select":
			err = elc.CmdWorkspaceSelect(homeConfigPath, os.Args[3:])
		case "show":
			err = elc.CmdWorkspaceShow(homeConfigPath, os.Args[3:])
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
	case "set-hooks":
		err = elc.CmdServiceSetHooks(os.Args[2:])
	case "exec":
		returnCode, err = elc.CmdServiceExec(homeConfigPath, os.Args[2:])
	default:
		returnCode, err = elc.CmdServiceExec(homeConfigPath, os.Args[1:])
	}

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	os.Exit(returnCode)
}
