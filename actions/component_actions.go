package actions

import (
	"errors"
	"fmt"
	"github.com/ensi-platform/elc/core"
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

func resolveComponents(ws *core.Workspace, options *core.GlobalOptions, namesFromArgs []string) ([]*core.Component, error) {
	if options.Tag != "" && options.Branch != "" {
		return nil, errors.New("--tag and --branch cannot be used together")
	}

	compNames, err := resolveCompNames(ws, options, namesFromArgs)
	if err != nil {
		return nil, err
	}

	cwdName, cwdBranch, _ := ws.MatchCwd()

	result := make([]*core.Component, 0, len(compNames))
	for _, compName := range compNames {
		comp, err := ws.ComponentByName(compName)
		if err != nil {
			return nil, err
		}

		branch := options.Branch
		if branch == "" && cwdBranch != "" && cwdName == comp.Name {
			branch = cwdBranch
		}

		if branch != "" {
			instance, err := comp.ForWorktree(branch)
			if err != nil {
				return nil, err
			}
			cloned, err := instance.IsCloned()
			if err != nil {
				return nil, err
			}
			if !cloned {
				instance, err = comp.EnsureWorktree(options, branch, false)
				if err != nil {
					return nil, err
				}
			}
			comp = instance
		}

		result = append(result, comp)
	}

	return result, nil
}

func resolveComponent(ws *core.Workspace, options *core.GlobalOptions, namesFromArgs []string) (*core.Component, error) {
	comps, err := resolveComponents(ws, options, namesFromArgs)
	if err != nil {
		return nil, err
	}
	if len(comps) > 1 {
		return nil, errors.New("too many components")
	}
	return comps[0], nil
}

func execHostComponent(comp *core.Component, ws *core.Workspace) (*core.Component, error) {
	if comp.Config.HostedIn == "" {
		return comp, nil
	}
	return ws.ComponentByName(comp.Config.HostedIn)
}

func ListCompNames(ws *core.Workspace, options *core.GlobalOptions) ([]string, error) {
	var compNames []string
	if options.Tag == "" {
		return ws.GetComponentNamesList(), nil
	}

	compNames = ws.FindComponentNamesByTag(options.Tag)
	if len(compNames) == 0 {
		return nil, errors.New(fmt.Sprintf("components with tag %s not found", options.Tag))
	}

	return compNames, nil
}

func StartServiceAction(options *core.GlobalOptions, svcNames []string) error {
	ws, err := core.GetWorkspaceConfig(options.WorkspaceName)
	if err != nil {
		return err
	}

	comps, err := resolveComponents(ws, options, svcNames)
	if err != nil {
		return err
	}

	for _, comp := range comps {
		if err := comp.Start(options); err != nil {
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

	var comps []*core.Component

	if stopAll {
		compNames := ws.GetComponentNames()
		for _, compName := range compNames {
			comp, err := ws.ComponentByName(compName)
			if err != nil {
				return err
			}
			comps = append(comps, comp)
		}
	} else {
		comps, err = resolveComponents(ws, options, svcNames)
		if err != nil {
			return err
		}
	}

	for _, comp := range comps {
		if destroy {
			err = comp.Destroy(options)
		} else {
			err = comp.Stop(options)
		}
		if err != nil {
			fmt.Printf("Error: %s\n", err)
		}
	}

	return nil
}

func RestartServiceAction(hardRestart bool, svcNames []string, options *core.GlobalOptions) error {
	ws, err := core.GetWorkspaceConfig(options.WorkspaceName)
	if err != nil {
		return err
	}

	comps, err := resolveComponents(ws, options, svcNames)
	if err != nil {
		return err
	}

	for _, comp := range comps {
		if err := comp.Restart(hardRestart, options); err != nil {
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

	comp, err := resolveComponent(ws, options, svcNames)
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

	comp, err := resolveComponent(ws, options, []string{})
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

	comp, err := resolveComponent(ws, options, []string{})
	if err != nil {
		return err
	}

	hostComp, err := execHostComponent(comp, ws)
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

	comp, err := resolveComponent(ws, options, []string{})
	if err != nil {
		return err
	}

	hostComp, err := execHostComponent(comp, ws)
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

func RunAction(options *core.GlobalOptions) error {
	ws, err := core.GetWorkspaceConfig(options.WorkspaceName)
	if err != nil {
		return err
	}

	comp, err := resolveComponent(ws, options, []string{})
	if err != nil {
		return err
	}

	hostComp, err := execHostComponent(comp, ws)
	if err != nil {
		return err
	}

	if comp.Config.ExecPath != "" {
		options.WorkingDir, err = ws.Context.RenderString(comp.Config.ExecPath)
		if err != nil {
			return err
		}
	}

	_, err = hostComp.Run(options)
	if err != nil {
		return err
	}

	return nil
}

func SetGitHooksAction(options *core.GlobalOptions, scriptsFolder string, elcBinary string, native bool) error {
	ws, err := core.GetWorkspaceConfig(options.WorkspaceName)
	if err != nil {
		return err
	}

	compNames, err := resolveCompNames(ws, options, []string{})
	if err != nil {
		return err
	}

	for _, compName := range compNames {
		comp, err := ws.ComponentByName(compName)
		if err != nil {
			return err
		}

		err = comp.UpdateHooks(options, elcBinary, scriptsFolder, native)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
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
			fmt.Printf("Error: %s\n", err)
		}
	}

	return nil
}

func ListServicesAction(options *core.GlobalOptions) error {
	ws, err := core.GetWorkspaceConfig(options.WorkspaceName)
	if err != nil {
		return err
	}

	compNames, err := ListCompNames(ws, options)
	if err != nil {
		return err
	}

	for _, compName := range compNames {
		_, _ = core.Pc.Println(compName)
	}

	return nil
}
