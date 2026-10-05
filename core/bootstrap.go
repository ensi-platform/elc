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
	return loadWorkspaceConfig(wsName, true)
}

// LoadWorkspaceConfigForCompletion loads workspace config without the elc version gate.
// Used by shell completion where binary may be unversioned (local go build).
func LoadWorkspaceConfigForCompletion(wsName string) (*Workspace, error) {
	return loadWorkspaceConfig(wsName, false)
}

func loadWorkspaceConfig(wsName string, checkVer bool) (*Workspace, error) {
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

	if checkVer {
		err = ws.checkVersion()
		if err != nil {
			return nil, err
		}
	}

	err = ws.init()
	if err != nil {
		return nil, err
	}

	return ws, nil
}
