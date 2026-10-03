package actions

import (
	"errors"
	"fmt"

	"github.com/ensi-platform/elc/core"
)

func resolveSingleMainComponent(options *core.GlobalOptions) (*core.Workspace, *core.Component, error) {
	if options.Tag != "" {
		return nil, nil, errors.New("worktree commands do not support --tag")
	}

	ws, err := core.GetWorkspaceConfig(options.WorkspaceName)
	if err != nil {
		return nil, nil, err
	}

	compNames, err := resolveCompNames(ws, options, nil)
	if err != nil {
		return nil, nil, err
	}
	if len(compNames) != 1 {
		return nil, nil, errors.New("worktree command requires a single component")
	}

	comp, err := ws.ComponentByName(compNames[0])
	if err != nil {
		return nil, nil, err
	}
	return ws, comp, nil
}

func LaunchAction(options *core.GlobalOptions, command []string) error {
	if len(command) == 0 {
		return errors.New("command is required")
	}
	if options.Tag != "" {
		return errors.New("launch does not support --tag")
	}

	ws, err := core.GetWorkspaceConfig(options.WorkspaceName)
	if err != nil {
		return err
	}

	compNames, err := resolveCompNames(ws, options, nil)
	if err != nil {
		return err
	}
	if len(compNames) != 1 {
		return errors.New("launch requires a single component")
	}

	comp, err := ws.ComponentByName(compNames[0])
	if err != nil {
		return err
	}

	branch := options.Branch
	cwdWorktreePath := ""
	if branch == "" {
		cwdName, cwdBranch, matchErr := ws.MatchCwd()
		if matchErr == nil && cwdName == comp.Name {
			branch = cwdBranch
			cwdWorktreePath = ws.CwdWorktreePath()
		}
	}

	if branch != "" {
		if cwdWorktreePath != "" {
			comp, err = comp.ForWorktreeAt(branch, cwdWorktreePath)
			if err != nil {
				return err
			}
			cloned, err := comp.IsCloned()
			if err != nil {
				return err
			}
			if !cloned {
				return errors.New(fmt.Sprintf("worktree path %s does not exist", cwdWorktreePath))
			}
		} else {
			comp, err = comp.EnsureWorktree(options, branch, false)
			if err != nil {
				return err
			}
		}
	}

	_, err = comp.Launch(command, options)
	return err
}

func AddWorktreeAction(options *core.GlobalOptions, branch string, hookArgs []string, noHook bool) error {
	_, comp, err := resolveSingleMainComponent(options)
	if err != nil {
		return err
	}

	return comp.AddWorktree(options, branch, hookArgs, noHook)
}

func RemoveWorktreeAction(options *core.GlobalOptions, branch string, force bool, noHook bool) error {
	ws, comp, err := resolveSingleMainComponent(options)
	if err != nil {
		return err
	}

	if branch == "" {
		branch = options.Branch
	}
	if branch == "" {
		cwdName, cwdBranch, matchErr := ws.MatchCwd()
		if matchErr == nil && cwdName == comp.Name {
			branch = cwdBranch
		}
	}
	if branch == "" {
		return errors.New("branch is required")
	}

	return comp.RemoveWorktree(options, branch, force, noHook)
}

func ListWorktreesAction(options *core.GlobalOptions) error {
	if options.Tag != "" {
		return errors.New("worktree list does not support --tag")
	}

	ws, err := core.GetWorkspaceConfig(options.WorkspaceName)
	if err != nil {
		return err
	}

	items, err := ws.ListWorktrees(options.ComponentName)
	if err != nil {
		return err
	}

	for _, item := range items {
		_, _ = core.Pc.Println(item.String())
	}

	return nil
}
