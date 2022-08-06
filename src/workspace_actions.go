package src

import (
	"errors"
	"fmt"
)

func ListWorkspacesAction() error {
	hc, err := checkAndLoadHC()
	if err != nil {
		return err
	}
	for _, workspace := range hc.Workspaces {
		_, _ = Pc.Printf("%-10s %s\n", workspace.Name, workspace.Path)
	}
	return nil
}

func AddWorkspaceAction(name string, wsPath string) error {
	hc, err := checkAndLoadHC()
	if err != nil {
		return err
	}

	ws := hc.findWorkspace(name)
	if ws != nil {
		return errors.New(fmt.Sprintf("workspace with name '%s' already exists", name))
	}

	err = hc.AddWorkspace(name, wsPath)
	if err != nil {
		return err
	}

	_, _ = Pc.Printf("workspace '%s' is added\n", name)

	if hc.CurrentWorkspace == "" {
		hc.CurrentWorkspace = name
		err = SaveHomeConfig(hc)
		if err != nil {
			return err
		}

		_, _ = Pc.Printf("active workspace changed to '%s'\n", name)
	}

	return nil
}

func ShowCurrentWorkspaceAction() error {
	hc, err := checkAndLoadHC()
	if err != nil {
		return err
	}
	_, _ = Pc.Println(hc.CurrentWorkspace)
	return nil
}

func SelectWorkspaceAction(name string) error {
	hc, err := checkAndLoadHC()
	if err != nil {
		return err
	}

	ws := hc.findWorkspace(name)
	if ws == nil {
		return errors.New(fmt.Sprintf("workspace with name '%s' is not defined", name))
	}

	hc.CurrentWorkspace = name
	err = SaveHomeConfig(hc)
	if err != nil {
		return err
	}

	_, _ = Pc.Printf("active workspace changed to '%s'\n", name)

	return nil
}
