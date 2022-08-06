package main

import elc "github.com/madridianfox/elc/src"

func main() {
	command := elc.InitCobra()
	err := command.Execute()
	if err != nil {
		panic(err)
	}
}

//func main() {
//	elc.Pc = &elc.RealPC{}
//	args := elc.Pc.Args()
//
//	if elc.NeedHelp(args[1:], "COMMAND", []string{
//		"Available commands:",
//		fmt.Sprintf("  %-20s - %s", elc.Color("exec", elc.CYellow), "execute command inside service's container"),
//		fmt.Sprintf("  %-20s - %s", elc.Color("compose", elc.CYellow), "run docker-compose command"),
//		fmt.Sprintf("  %-20s - %s", elc.Color("destroy", elc.CYellow), "delete service containers"),
//		fmt.Sprintf("  %-20s - %s", elc.Color("help", elc.CYellow), "print this help message"),
//		fmt.Sprintf("  %-20s - %s", elc.Color("restart", elc.CYellow), "restart service"),
//		fmt.Sprintf("  %-20s - %s", elc.Color("set-hooks", elc.CYellow), "install git hooks"),
//		fmt.Sprintf("  %-20s - %s", elc.Color("start", elc.CYellow), "start service"),
//		fmt.Sprintf("  %-20s - %s", elc.Color("stop", elc.CYellow), "stop service"),
//		fmt.Sprintf("  %-20s - %s", elc.Color("vars", elc.CYellow), "print variables"),
//		fmt.Sprintf("  %-20s - %s", elc.Color("workspace", elc.CYellow), "manage workspaces"),
//		fmt.Sprintf("  %-20s - %s", elc.Color("update", elc.CYellow), "download new version of elc and replace current binary"),
//		fmt.Sprintf("  %-20s - %s", elc.Color("version", elc.CYellow), "print version"),
//		"Any other arguments will be used for invoke of implicit exec command.",
//		"",
//		"You can get help for any command invoke it with '--help' option.",
//	}) {
//		elc.Pc.Exit(0)
//	}
//
//	if len(args) < 2 {
//		fmt.Println("At least one argument is needed. Use -h option for help.")
//		elc.Pc.Exit(1)
//	}
//
//	var err error
//	var returnCode int
//
//	homeDir, err := elc.Pc.HomeDir()
//	if err != nil {
//		fmt.Println(err)
//		elc.Pc.Exit(1)
//	}
//
//	homeConfigPath := path.Join(homeDir, ".elc.yaml")
//
//	switch args[1] {
//	case "workspace":
//		switch args[2] {
//		case "list", "ls":
//			err = elc.CmdWorkspaceList(homeConfigPath, args[3:])
//		case "add":
//			err = elc.CmdWorkspaceAdd(homeConfigPath, args[3:])
//		case "select":
//			err = elc.CmdWorkspaceSelect(homeConfigPath, args[3:])
//		case "show":
//			err = elc.CmdWorkspaceShow(homeConfigPath, args[3:])
//		default:
//			err = elc.CmdWorkspaceHelp()
//		}
//	case "start":
//		err = elc.CmdServiceStart(homeConfigPath, args[2:])
//	case "stop":
//		err = elc.CmdServiceStop(homeConfigPath, args[2:])
//	case "restart":
//		err = elc.CmdServiceRestart(homeConfigPath, args[2:])
//	case "destroy":
//		err = elc.CmdServiceDestroy(homeConfigPath, args[2:])
//	case "compose":
//		returnCode, err = elc.CmdServiceCompose(homeConfigPath, args[2:])
//	case "vars":
//		err = elc.CmdServiceVars(homeConfigPath, args[2:])
//	case "set-hooks":
//		err = elc.CmdServiceSetHooks(args[2:])
//	case "exec":
//		returnCode, err = elc.CmdServiceExec(homeConfigPath, args[2:])
//	case "update":
//		err = elc.CmdUpdate(homeConfigPath, args[2:])
//	case "fix-update-command":
//		err = elc.CmdFixUpdateCommand(homeConfigPath, args[2:])
//	case "version":
//		elc.CmdVersion()
//	case "wrap":
//		returnCode, err = elc.CmdWrap(homeConfigPath, args[2:])
//	default:
//		returnCode, err = elc.CmdServiceExec(homeConfigPath, args[1:])
//	}
//
//	if err != nil {
//		fmt.Println(err)
//		elc.Pc.Exit(1)
//	}
//
//	elc.Pc.Exit(returnCode)
//}
