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
		fmt.Sprintf("  %-20s - %s", elc.Color("exec", elc.CYellow), "execute command inside service's container"),
		fmt.Sprintf("  %-20s - %s", elc.Color("compose", elc.CYellow), "run docker-compose command"),
		fmt.Sprintf("  %-20s - %s", elc.Color("destroy", elc.CYellow), "delete service containers"),
		fmt.Sprintf("  %-20s - %s", elc.Color("help", elc.CYellow), "print this help message"),
		fmt.Sprintf("  %-20s - %s", elc.Color("restart", elc.CYellow), "restart service"),
		fmt.Sprintf("  %-20s - %s", elc.Color("set-hooks", elc.CYellow), "install git hooks"),
		fmt.Sprintf("  %-20s - %s", elc.Color("start", elc.CYellow), "start service"),
		fmt.Sprintf("  %-20s - %s", elc.Color("stop", elc.CYellow), "stop service"),
		fmt.Sprintf("  %-20s - %s", elc.Color("vars", elc.CYellow), "print variables"),
		fmt.Sprintf("  %-20s - %s", elc.Color("workspace", elc.CYellow), "manage workspaces"),
		fmt.Sprintf("  %-20s - %s", elc.Color("update", elc.CYellow), "download new version of elc and replace current binary"),
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
	case "update":
		err = elc.CmdUpdate(homeConfigPath, os.Args[2:])
	default:
		returnCode, err = elc.CmdServiceExec(homeConfigPath, os.Args[1:])
	}

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	os.Exit(returnCode)
}
