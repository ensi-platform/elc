package src

import (
	"fmt"
	"path"
)

func StartServiceAction(svcNames []string) error {
	ws, err := getWorkspaceConfig()
	if err != nil {
		return err
	}

	if len(svcNames) > 0 {
		for _, svcName := range svcNames {
			comp, err := ws.componentByName(svcName)
			if err != nil {
				return err
			}

			err = comp.Start(&startParams)
			if err != nil {
				return err
			}
		}
	} else {
		comp, err := ws.componentByPath()
		if err != nil {
			return err
		}

		err = comp.Start(&startParams)
		if err != nil {
			return err
		}
	}

	return nil
}

func StopServiceAction(stopAll bool, svcNames []string, destroy bool) error {
	ws, err := getWorkspaceConfig()
	if err != nil {
		return err
	}

	if stopAll {
		svcNames = ws.getComponentNames()
	}

	if len(svcNames) > 0 {
		for _, svcName := range svcNames {
			comp, err := ws.componentByName(svcName)
			if err != nil {
				return err
			}
			if destroy {
				err = comp.Destroy()
			} else {
				err = comp.Stop()
			}
			if err != nil {
				return err
			}
		}
	} else {
		comp, err := ws.componentByPath()
		if err != nil {
			return err
		}

		if destroy {
			err = comp.Destroy()
		} else {
			err = comp.Stop()
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func RestartServiceAction(hardRestart bool, svcNames []string) error {
	ws, err := getWorkspaceConfig()
	if err != nil {
		return err
	}

	if len(svcNames) > 0 {
		for _, svcName := range svcNames {
			comp, err := ws.componentByName(svcName)
			if err != nil {
				return err
			}

			err = comp.Restart(hardRestart)
			if err != nil {
				return err
			}
		}
	} else {
		comp, err := ws.componentByPath()
		if err != nil {
			return err
		}

		err = comp.Restart(hardRestart)
		if err != nil {
			return err
		}
	}

	return nil
}

func PrintVarsAction(svcNames []string) error {
	ws, err := getWorkspaceConfig()
	if err != nil {
		return err
	}

	var comp *Component

	if len(svcNames) > 0 {
		comp, err = ws.componentByName(svcNames[0])
		if err != nil {
			return err
		}
	} else {
		comp, err = ws.componentByPath()
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

func ComposeCommandAction(composeParams SvcComposeParams) error {
	ws, err := getWorkspaceConfig()
	if err != nil {
		return err
	}

	if composeParams.SvcName == "" {
		composeParams.SvcName, err = ws.componentNameByPath()
		if err != nil {
			return err
		}
	}

	comp, err := ws.componentByName(composeParams.SvcName)
	if err != nil {
		return err
	}

	_, err = comp.Compose(&composeParams)
	if err != nil {
		return err
	}

	return nil
}

func WrapCommandAction(globalOptions DefaultOptions, command []string) error {
	ws, err := getWorkspaceConfig()
	if err != nil {
		return err
	}

	var comp *Component

	svcName := globalOptions.ComponentName

	if svcName == "" {
		comp, err = ws.componentByPath()
	} else {
		comp, err = ws.componentByName(svcName)
	}
	if err != nil {
		return err
	}

	if comp.Config.HostedIn != "" {
		svcName = comp.Config.HostedIn
	} else {
		svcName = comp.Name
	}

	hostComp, err := ws.componentByName(svcName)
	if err != nil {
		return err
	}

	_, err = hostComp.Wrap(command)
	if err != nil {
		return err
	}

	return nil
}

func ExecAction(execParams SvcExecParams) error {
	ws, err := getWorkspaceConfig()
	if err != nil {
		return err
	}

	var comp *Component

	if execParams.SvcName == "" {
		comp, err = ws.componentByPath()
	} else {
		comp, err = ws.componentByName(execParams.SvcName)
	}
	if err != nil {
		return err
	}

	if comp.Config.HostedIn != "" {
		execParams.SvcName = comp.Config.HostedIn
	} else {
		execParams.SvcName = comp.Name
	}

	if comp.Config.ExecPath != "" {
		execParams.WorkingDir, err = ws.Context.renderString(comp.Config.ExecPath)
		if err != nil {
			return err
		}
	}

	hostComp, err := ws.componentByName(execParams.SvcName)
	if err != nil {
		return err
	}

	_, err = hostComp.Exec(&execParams)
	if err != nil {
		return err
	}

	return nil
}

func SetGitHooksAction(scriptsFolder string, elcBinary string) error {
	folders, err := Pc.ReadDir(scriptsFolder)
	if err != nil {
		return err
	}
	for _, folder := range folders {
		if !folder.IsDir() {
			continue
		}
		files, err := Pc.ReadDir(path.Join(scriptsFolder, folder.Name()))
		if err != nil {
			return err
		}
		hookScripts := make([]string, 0)
		for _, file := range files {
			hookScripts = append(hookScripts, path.Join(scriptsFolder, folder.Name(), file.Name()))
		}
		script := generateHookScript(hookScripts, elcBinary)
		err = Pc.WriteFile(fmt.Sprintf(".git/hooks/%s", folder.Name()), []byte(script), 0755)
		if err != nil {
			return err
		}
	}

	return nil
}
