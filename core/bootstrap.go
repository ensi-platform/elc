package core

import (
	"path"
)

func CheckAndLoadHC() (*HomeConfig, error) {
	homeDir, err := Pc.HomeDir()
	if err != nil {
		return nil, err
	}
	homeConfigPath := path.Join(homeDir, ".elc.yaml")
	err = CheckHomeConfigIsEmpty(homeConfigPath)
	if err != nil {
		return nil, err
	}
	hc, err := LoadHomeConfig(homeConfigPath)
	if err != nil {
		return nil, err
	}

	return hc, nil
}

func GetWorkspaceConfig(wsName string) (*Workspace, error) {
	hc, err := CheckAndLoadHC()
	if err != nil {
		return nil, err
	}

	wsPath, err := hc.GetCurrentWsPath(wsName)
	if err != nil {
		return nil, err
	}

	cwd, err := Pc.Getwd()
	if err != nil {
		return nil, err
	}
	ws := NewWorkspace(wsPath, cwd)

	err = ws.LoadConfig()
	if err != nil {
		return nil, err
	}

	err = ws.checkVersion()
	if err != nil {
		return nil, err
	}

	err = ws.init()
	if err != nil {
		return nil, err
	}

	return ws, nil
}
