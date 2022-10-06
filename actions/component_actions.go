package actions

import (
	"errors"
	"fmt"
	"github.com/madridianfox/elc/core"
	"path"
)

func resolveCompNames(ws *core.Workspace, options *core.GlobalOptions, namesFromArgs []string) ([]string, error) {
	var compNames []string

	if options.Tag != "" {
		compNames = ws.FindComponentNamesByTag(options.Tag)
		if len(compNames) == 0 {
			return nil, errors.New(fmt.Sprintf("components with tag %s not found", options.Tag))
		}
	} else if options.ComponentName != "" {
		compNames = []string{options.ComponentName}
	} else if len(namesFromArgs) > 0 {
		compNames = namesFromArgs
	} else {
		currentCompName, err := ws.ComponentNameByPath()
		if err != nil {
			return nil, err
		}
		compNames = []string{currentCompName}
	}

	return compNames, nil
}

func StartServiceAction(options *core.GlobalOptions, svcNames []string) error {
	ws, err := core.GetWorkspaceConfig(options.WorkspaceName)
	if err != nil {
		return err
	}

	compNames, err := resolveCompNames(ws, options, svcNames)
	if err != nil {
		return err
	}

	for _, compName := range compNames {
		comp, err := ws.ComponentByName(compName)
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
	ws, err := core.GetWorkspaceConfig(options.WorkspaceName)
	if err != nil {
		return err
	}

	var compNames []string

	if stopAll {
		compNames = ws.GetComponentNames()
	} else {
		compNames, err = resolveCompNames(ws, options, svcNames)
		if err != nil {
			return err
		}
	}

	for _, compName := range compNames {
		comp, err := ws.ComponentByName(compName)
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
	ws, err := core.GetWorkspaceConfig(options.WorkspaceName)
	if err != nil {
		return err
	}

	compNames, err := resolveCompNames(ws, options, svcNames)
	if err != nil {
		return err
	}

	for _, compName := range compNames {
		comp, err := ws.ComponentByName(compName)
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

func PrintVarsAction(options *core.GlobalOptions, svcNames []string) error {
	ws, err := core.GetWorkspaceConfig(options.WorkspaceName)
	if err != nil {
		return err
	}

	compNames, err := resolveCompNames(ws, options, svcNames)
	if err != nil {
		return err
	}

	if len(compNames) > 1 {
		return errors.New("too many components for show")
	}

	comp, err := ws.ComponentByName(compNames[0])
	if err != nil {
		return err
	}

	err = comp.DumpVars()
	if err != nil {
		return err
	}

	return nil
}

func ComposeCommandAction(options *core.GlobalOptions, args []string) error {
	ws, err := core.GetWorkspaceConfig(options.WorkspaceName)
	if err != nil {
		return err
	}

	compNames, err := resolveCompNames(ws, options, []string{})
	if err != nil {
		return err
	}

	if len(compNames) > 1 {
		return errors.New("too many components")
	}

	comp, err := ws.ComponentByName(compNames[0])
	if err != nil {
		return err
	}

	options.Cmd = args

	_, err = comp.Compose(options)
	if err != nil {
		return err
	}

	return nil
}

func WrapCommandAction(options *core.GlobalOptions, command []string) error {
	ws, err := core.GetWorkspaceConfig(options.WorkspaceName)
	if err != nil {
		return err
	}

	compNames, err := resolveCompNames(ws, options, []string{})
	if err != nil {
		return err
	}

	if len(compNames) > 1 {
		return errors.New("too many components")
	}

	comp, err := ws.ComponentByName(compNames[0])
	if err != nil {
		return err
	}

	var hostName string

	if comp.Config.HostedIn != "" {
		hostName = comp.Config.HostedIn
	} else {
		hostName = comp.Name
	}

	hostComp, err := ws.ComponentByName(hostName)
	if err != nil {
		return err
	}

	_, err = hostComp.Wrap(command, options)
	if err != nil {
		return err
	}

	return nil
}

func ExecAction(options *core.GlobalOptions) error {
	ws, err := core.GetWorkspaceConfig(options.WorkspaceName)
	if err != nil {
		return err
	}

	compNames, err := resolveCompNames(ws, options, []string{})
	if err != nil {
		return err
	}

	if len(compNames) > 1 {
		return errors.New("too many components")
	}

	comp, err := ws.ComponentByName(compNames[0])
	if err != nil {
		return err
	}

	var hostName string

	if comp.Config.HostedIn != "" {
		hostName = comp.Config.HostedIn
	} else {
		hostName = comp.Name
	}

	hostComp, err := ws.ComponentByName(hostName)
	if err != nil {
		return err
	}

	if comp.Config.ExecPath != "" {
		options.WorkingDir, err = ws.Context.RenderString(comp.Config.ExecPath)
		if err != nil {
			return err
		}
	}

	_, err = hostComp.Exec(options)
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

func CloneComponentAction(options *core.GlobalOptions, svcNames []string, noHook bool) error {
	ws, err := core.GetWorkspaceConfig(options.WorkspaceName)
	if err != nil {
		return err
	}

	compNames, err := resolveCompNames(ws, options, svcNames)
	if err != nil {
		return err
	}

	for _, compName := range compNames {
		comp, err := ws.ComponentByName(compName)
		if err != nil {
			return err
		}

		err = comp.Clone(options, noHook)
		if err != nil {
			return err
		}
	}

	return nil
}
