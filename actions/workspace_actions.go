package actions

import (
	"errors"
	"fmt"
	"github.com/madridianfox/elc/core"
)

func ListWorkspacesAction() error {
	hc, err := core.CheckAndLoadHC()
	if err != nil {
		return err
	}
	for _, workspace := range hc.Workspaces {
		_, _ = core.Pc.Printf("%-10s %s\n", workspace.Name, workspace.Path)
	}
	return nil
}

func AddWorkspaceAction(name string, wsPath string) error {
	hc, err := core.CheckAndLoadHC()
	if err != nil {
		return err
	}

	ws := hc.FindWorkspace(name)
	if ws != nil {
		return errors.New(fmt.Sprintf("workspace with name '%s' already exists", name))
	}

	err = hc.AddWorkspace(name, wsPath)
	if err != nil {
		return err
	}

	_, _ = core.Pc.Printf("workspace '%s' is added\n", name)

	if hc.CurrentWorkspace == "" {
		hc.CurrentWorkspace = name
		err = core.SaveHomeConfig(hc)
		if err != nil {
			return err
		}

		_, _ = core.Pc.Printf("active workspace changed to '%s'\n", name)
	}

	return nil
}

func ShowCurrentWorkspaceAction() error {
	hc, err := core.CheckAndLoadHC()
	if err != nil {
		return err
	}
	_, _ = core.Pc.Println(hc.CurrentWorkspace)
	return nil
}

func SelectWorkspaceAction(name string) error {
	hc, err := core.CheckAndLoadHC()
	if err != nil {
		return err
	}

	ws := hc.FindWorkspace(name)
	if ws == nil {
		return errors.New(fmt.Sprintf("workspace with name '%s' is not defined", name))
	}

	hc.CurrentWorkspace = name
	err = core.SaveHomeConfig(hc)
	if err != nil {
		return err
	}

	_, _ = core.Pc.Printf("active workspace changed to '%s'\n", name)

	return nil
}
