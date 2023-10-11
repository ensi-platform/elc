package actions

import (
	"errors"
	"fmt"
	"github.com/ensi-platform/elc/core"
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

func RemoveWorkspaceAction(name string) error {
	hc, err := core.CheckAndLoadHC()
	if err != nil {
		return err
	}

	_, _ = core.Pc.Printf("workspace '%s' is removed\n", name)

	return hc.RemoveWorkspace(name)
}

func ShowCurrentWorkspaceAction(options *core.GlobalOptions) error {
	hc, err := core.CheckAndLoadHC()
	if err != nil {
		return err
	}
	hci, err := hc.GetCurrentWorkspace(options.WorkspaceName)
	if err != nil {
		return err
	}

	_, _ = core.Pc.Println(hci.Name)
	return nil
}

func SelectWorkspaceAction(name string) error {
	hc, err := core.CheckAndLoadHC()
	if err != nil {
		return err
	}

	if name != "auto" {
		ws := hc.FindWorkspace(name)
		if ws == nil {
			return errors.New(fmt.Sprintf("workspace with name '%s' is not defined", name))
		}
	}

	hc.CurrentWorkspace = name
	err = core.SaveHomeConfig(hc)
	if err != nil {
		return err
	}

	_, _ = core.Pc.Printf("active workspace changed to '%s'\n", name)

	return nil
}

func SetRootPathAction(name string, rootPath string) error {
	hc, err := core.CheckAndLoadHC()
	if err != nil {
		return err
	}

	ws := hc.FindWorkspace(name)
	if ws == nil {
		return errors.New(fmt.Sprintf("workspace with name '%s' is not defined", name))
	}

	ws.RootPath = rootPath
	err = core.SaveHomeConfig(hc)
	if err != nil {
		return err
	}

	_, _ = core.Pc.Printf("path saved\n")

	return nil
}
