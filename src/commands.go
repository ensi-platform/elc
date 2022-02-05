package src

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
)

func checkAndLoadHC(homeConfigPath string) (*HomeConfig, error) {
	err := CheckHomeConfigIsEmpty(homeConfigPath)
	if err != nil {
		return nil, err
	}
	hc, err := LoadHomeConfig(homeConfigPath)
	if err != nil {
		return nil, err
	}

	return hc, nil
}

func getWorkspaceConfig(homeConfigPath string) (*MainConfig, error) {
	hc, err := checkAndLoadHC(homeConfigPath)
	if err != nil {
		return nil, err
	}

	wsPath, err := hc.GetCurrentWsPath()
	if err != nil {
		return nil, err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	cfg := NewConfig(wsPath, cwd)
	err = cfg.LoadFromFile()
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func addStartFlags(fs *flag.FlagSet, params *SvcStartParams) {
	fs.StringVar(&params.Tag, "tag", "", "tag for dependencies selecting")
	fs.BoolVar(&params.Force, "force", false, "force start dependencies")
}

func addComposeFlags(fs *flag.FlagSet, params *SvcComposeParams) {
	fs.StringVar(&params.SvcName, "svc", "", "name of service")
}

func CmdWorkspaceList(homeConfigPath string) error {
	hc, err := checkAndLoadHC(homeConfigPath)
	if err != nil {
		return err
	}

	for _, workspace := range hc.Workspaces {
		fmt.Printf("%-10s %s\n", workspace.Name, workspace.Path)
	}

	return nil
}

func CmdWorkspaceAdd(homeConfigPath string, args []string) error {
	hc, err := checkAndLoadHC(homeConfigPath)
	if err != nil {
		return err
	}

	if len(args) != 2 {
		return errors.New("command requires exactly 2 arguments")
	}

	name := args[0]
	wsPath := args[1]

	ws := hc.findWorkspace(name)
	if ws != nil {
		return errors.New(fmt.Sprintf("workspace with name '%s' already exists", name))
	}

	err = hc.AddWorkspace(name, wsPath)
	if err != nil {
		return err
	}

	fmt.Printf("workspace '%s' is added\n", name)
	return nil
}

func CmdWorkspaceSelect(homeConfigPath string, args []string) error {
	hc, err := checkAndLoadHC(homeConfigPath)
	if err != nil {
		return err
	}

	if len(args) != 1 {
		return errors.New("command requires exactly 1 argument")
	}

	name := args[0]

	ws := hc.findWorkspace(name)
	if ws != nil {
		return errors.New(fmt.Sprintf("workspace with name '%s' is not defined", name))
	}

	hc.CurrentWorkspace = name
	err = SaveHomeConfig(hc)
	if err != nil {
		return err
	}

	fmt.Printf("active workspace changed to '%s'\n", name)
	return nil
}

func CmdWorkspaceShow(homeConfigPath string) error {
	hc, err := checkAndLoadHC(homeConfigPath)
	if err != nil {
		return err
	}
	fmt.Println(hc.CurrentWorkspace)

	return nil
}

func CmdWorkspaceHelp() error {
	fmt.Println(
		strings.Join([]string{
			"Usage: elc workspace [command] [args]",
			"",
			"Available commands:",
			"  ls, list          - list available workspaces",
			"  show              - show current workspace name",
			"  add <name> <path> - add new workspace",
			"  select <name>     - select workspace as current",
		}, "\n"))
	return nil
}

func CmdServiceStart(homeConfigPath string, args []string) error {
	fs := flag.NewFlagSet("start", flag.ContinueOnError)
	startParams := &SvcStartParams{}
	addStartFlags(fs, startParams)
	err := fs.Parse(args)
	if err != nil {
		return err
	}

	cfg, err := getWorkspaceConfig(homeConfigPath)
	if err != nil {
		return err
	}

	svcNames := fs.Args()
	if len(svcNames) > 0 {
		for _, svcName := range svcNames {
			svc, err := CreateFromSvcName(cfg, svcName)
			if err != nil {
				return err
			}

			err = svc.Start(startParams)
			if err != nil {
				return err
			}
		}
	} else {
		svcName, err := cfg.FindServiceByPath()
		if err != nil {
			return err
		}

		svc, err := CreateFromSvcName(cfg, svcName)
		if err != nil {
			return err
		}

		err = svc.Start(startParams)
		if err != nil {
			return err
		}
	}

	return nil
}

func CmdServiceStop(homeConfigPath string, args []string) error {
	cfg, err := getWorkspaceConfig(homeConfigPath)
	if err != nil {
		return err
	}

	svcNames := args
	if len(svcNames) > 0 {
		for _, svcName := range svcNames {
			svc, err := CreateFromSvcName(cfg, svcName)
			if err != nil {
				return err
			}

			err = svc.Stop()
			if err != nil {
				return err
			}
		}
	} else {
		svcName, err := cfg.FindServiceByPath()
		if err != nil {
			return err
		}

		svc, err := CreateFromSvcName(cfg, svcName)
		if err != nil {
			return err
		}

		err = svc.Stop()
		if err != nil {
			return err
		}
	}

	return nil
}

func CmdServiceDestroy(homeConfigPath string, args []string) error {
	cfg, err := getWorkspaceConfig(homeConfigPath)
	if err != nil {
		return err
	}

	svcNames := args
	if len(svcNames) > 0 {
		for _, svcName := range svcNames {
			svc, err := CreateFromSvcName(cfg, svcName)
			if err != nil {
				return err
			}

			err = svc.Destroy()
			if err != nil {
				return err
			}
		}
	} else {
		svcName, err := cfg.FindServiceByPath()
		if err != nil {
			return err
		}

		svc, err := CreateFromSvcName(cfg, svcName)
		if err != nil {
			return err
		}

		err = svc.Destroy()
		if err != nil {
			return err
		}
	}

	return nil
}

func CmdServiceCompose(homeConfigPath string, args []string) (int, error) {
	fs := flag.NewFlagSet("compose", flag.ContinueOnError)
	composeParams := &SvcComposeParams{}
	addComposeFlags(fs, composeParams)
	err := fs.Parse(args)
	if err != nil {
		return 0, err
	}

	composeParams.Cmd = fs.Args()

	cfg, err := getWorkspaceConfig(homeConfigPath)
	if err != nil {
		return 0, err
	}

	if composeParams.SvcName == "" {
		composeParams.SvcName, err = cfg.FindServiceByPath()
		if err != nil {
			return 0, err
		}
	}

	svc, err := CreateFromSvcName(cfg, composeParams.SvcName)
	if err != nil {
		return 0, err
	}

	returnCode, err := svc.Compose(composeParams)
	if err != nil {
		return 0, err
	}

	return returnCode, nil
}

func CmdServiceExec(homeConfigPath string, args []string) (int, error) {
	fs := flag.NewFlagSet("compose", flag.ContinueOnError)
	execParams := &SvcExecParams{}
	addComposeFlags(fs, &execParams.SvcComposeParams)
	addStartFlags(fs, &execParams.SvcStartParams)
	err := fs.Parse(args)
	if err != nil {
		return 0, err
	}

	execParams.Cmd = fs.Args()

	cfg, err := getWorkspaceConfig(homeConfigPath)
	if err != nil {
		return 0, err
	}

	if execParams.SvcName == "" {
		execParams.SvcName, err = cfg.FindServiceByPath()
		if err != nil {
			return 0, err
		}
	}

	svc, err := CreateFromSvcName(cfg, execParams.SvcName)
	if err != nil {
		return 0, err
	}

	returnCode, err := svc.Exec(execParams)
	if err != nil {
		return 0, err
	}

	return returnCode, nil
}
