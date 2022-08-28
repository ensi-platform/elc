package actions

import (
	"fmt"
	"github.com/madridianfox/elc/core"
	"path"
)

func StartServiceAction(options *core.GlobalOptions, svcNames []string) error {
	ws, err := core.GetWorkspaceConfig()
	if err != nil {
		return err
	}

	if options.ComponentName != "" {
		svcNames = append(svcNames, options.ComponentName)
	}

	if len(svcNames) > 0 {
		for _, svcName := range svcNames {
			comp, err := ws.ComponentByName(svcName)
			if err != nil {
				return err
			}

			err = comp.Start(options)
			if err != nil {
				return err
			}
		}
	} else {
		comp, err := ws.ComponentByPath()
		if err != nil {
			return err
		}

		err = comp.Start(options)
		if err != nil {
			return err
		}
	}

	return nil
}

func StopServiceAction(stopAll bool, svcNames []string, destroy bool, options *core.GlobalOptions) error {
	ws, err := core.GetWorkspaceConfig()
	if err != nil {
		return err
	}

	if options.ComponentName != "" {
		svcNames = append(svcNames, options.ComponentName)
	}

	if stopAll {
		svcNames = ws.GetComponentNames()
	}

	if len(svcNames) > 0 {
		for _, svcName := range svcNames {
			comp, err := ws.ComponentByName(svcName)
			if err != nil {
				return err
			}
			if destroy {
				err = comp.Destroy(options)
			} else {
				err = comp.Stop(options)
			}
			if err != nil {
				return err
			}
		}
	} else {
		comp, err := ws.ComponentByPath()
		if err != nil {
			return err
		}

		if destroy {
			err = comp.Destroy(options)
		} else {
			err = comp.Stop(options)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func RestartServiceAction(hardRestart bool, svcNames []string, options *core.GlobalOptions) error {
	ws, err := core.GetWorkspaceConfig()
	if err != nil {
		return err
	}

	if len(svcNames) > 0 {
		for _, svcName := range svcNames {
			comp, err := ws.ComponentByName(svcName)
			if err != nil {
				return err
			}

			err = comp.Restart(hardRestart, options)
			if err != nil {
				return err
			}
		}
	} else {
		comp, err := ws.ComponentByPath()
		if err != nil {
			return err
		}

		err = comp.Restart(hardRestart, options)
		if err != nil {
			return err
		}
	}

	return nil
}

func PrintVarsAction(svcNames []string) error {
	ws, err := core.GetWorkspaceConfig()
	if err != nil {
		return err
	}

	var comp *core.Component

	if len(svcNames) > 0 {
		comp, err = ws.ComponentByName(svcNames[0])
		if err != nil {
			return err
		}
	} else {
		comp, err = ws.ComponentByPath()
		if err != nil {
			return err
		}
	}

	err = comp.DumpVars()
	if err != nil {
		return err
	}

	return nil
}

func ComposeCommandAction(args []string, composeParams core.GlobalOptions) error {
	ws, err := core.GetWorkspaceConfig()
	if err != nil {
		return err
	}

	composeParams.Cmd = args

	if composeParams.ComponentName == "" {
		composeParams.ComponentName, err = ws.ComponentNameByPath()
		if err != nil {
			return err
		}
	}

	comp, err := ws.ComponentByName(composeParams.ComponentName)
	if err != nil {
		return err
	}

	_, err = comp.Compose(&composeParams)
	if err != nil {
		return err
	}

	return nil
}

func WrapCommandAction(globalOptions core.GlobalOptions, command []string) error {
	ws, err := core.GetWorkspaceConfig()
	if err != nil {
		return err
	}

	var comp *core.Component

	svcName := globalOptions.ComponentName

	if svcName == "" {
		comp, err = ws.ComponentByPath()
	} else {
		comp, err = ws.ComponentByName(svcName)
	}
	if err != nil {
		return err
	}

	if comp.Config.HostedIn != "" {
		svcName = comp.Config.HostedIn
	} else {
		svcName = comp.Name
	}

	hostComp, err := ws.ComponentByName(svcName)
	if err != nil {
		return err
	}

	_, err = hostComp.Wrap(command)
	if err != nil {
		return err
	}

	return nil
}

func ExecAction(options core.GlobalOptions) error {
	ws, err := core.GetWorkspaceConfig()
	if err != nil {
		return err
	}

	var comp *core.Component

	if options.ComponentName == "" {
		comp, err = ws.ComponentByPath()
	} else {
		comp, err = ws.ComponentByName(options.ComponentName)
	}
	if err != nil {
		return err
	}

	if comp.Config.HostedIn != "" {
		options.ComponentName = comp.Config.HostedIn
	} else {
		options.ComponentName = comp.Name
	}

	if comp.Config.ExecPath != "" {
		options.WorkingDir, err = ws.Context.RenderString(comp.Config.ExecPath)
		if err != nil {
			return err
		}
	}

	hostComp, err := ws.ComponentByName(options.ComponentName)
	if err != nil {
		return err
	}

	_, err = hostComp.Exec(&options)
	if err != nil {
		return err
	}

	return nil
}

func SetGitHooksAction(scriptsFolder string, elcBinary string) error {
	folders, err := core.Pc.ReadDir(scriptsFolder)
	if err != nil {
		return err
	}
	for _, folder := range folders {
		if !folder.IsDir() {
			continue
		}
		files, err := core.Pc.ReadDir(path.Join(scriptsFolder, folder.Name()))
		if err != nil {
			return err
		}
		hookScripts := make([]string, 0)
		for _, file := range files {
			hookScripts = append(hookScripts, path.Join(scriptsFolder, folder.Name(), file.Name()))
		}
		script := core.GenerateHookScript(hookScripts, elcBinary)
		err = core.Pc.WriteFile(fmt.Sprintf(".git/hooks/%s", folder.Name()), []byte(script), 0755)
		if err != nil {
			return err
		}
	}

	return nil
}
