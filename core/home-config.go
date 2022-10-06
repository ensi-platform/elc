package core

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"strings"
)

type HomeConfigItem struct {
	Name     string `yaml:"name"`
	Path     string `yaml:"path"`
	RootPath string `yaml:"root_path"`
}

type HomeConfig struct {
	Path             string           `yaml:"-"`
	CurrentWorkspace string           `yaml:"current_workspace"`
	UpdateCommand    string           `yaml:"update_command"`
	Workspaces       []HomeConfigItem `yaml:"workspaces"`
}

const DefaultUpdateCommand = "curl -sSL https://raw.githubusercontent.com/ensi-platform/elc/master/get.sh | sudo -E bash"

func LoadHomeConfig(configPath string) (*HomeConfig, error) {
	yamlFile, err := Pc.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	cfg := &HomeConfig{}
	err = yaml.Unmarshal(yamlFile, cfg)
	if err != nil {
		return nil, err
	}
	cfg.Path = configPath
	return cfg, nil
}

func SaveHomeConfig(homeConfig *HomeConfig) error {
	data, err := yaml.Marshal(homeConfig)
	if err != nil {
		return err
	}

	err = Pc.WriteFile(homeConfig.Path, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

func CheckHomeConfigIsEmpty(configPath string) error {
	if Pc.FileExists(configPath) {
		return nil
	}
	return SaveHomeConfig(&HomeConfig{Path: configPath, UpdateCommand: DefaultUpdateCommand})
}

func (hc *HomeConfig) AddWorkspace(name string, path string) error {
	hc.Workspaces = append(hc.Workspaces, HomeConfigItem{Name: name, Path: path})
	return SaveHomeConfig(hc)
}

func (hc *HomeConfig) RemoveWorkspace(name string) error {
	foundWsIndex := -1
	for index, ws := range hc.Workspaces {
		if ws.Name == name {
			foundWsIndex = index
		}
	}

	if foundWsIndex == -1 {
		return errors.New(fmt.Sprintf("Workspace %s doesn't exists", name))
	}

	hc.Workspaces = append(hc.Workspaces[:foundWsIndex], hc.Workspaces[foundWsIndex+1:]...)
	return SaveHomeConfig(hc)
}

func (hc *HomeConfig) GetCurrentWorkspace(wsName string) (*HomeConfigItem, error) {
	if wsName != "" {
		hci := hc.FindWorkspace(wsName)
		if hci == nil {
			return nil, errors.New("undefined workspace")
		}
		return hci, nil
	}

	if hc.CurrentWorkspace == "" {
		return nil, errors.New("current workspace is not set")
	}

	if hc.CurrentWorkspace == "auto" {
		hci, err := hc.FindWorkspaceByPath()
		if err != nil {
			return nil, err
		}

		return hci, nil
	}

	for index, hci := range hc.Workspaces {
		if hci.Name == hc.CurrentWorkspace {
			return &hc.Workspaces[index], nil
		}
	}

	return nil, errors.New("current workspace is bad")
}

func (hc *HomeConfig) GetCurrentWsPath(wsName string) (string, error) {
	hci, err := hc.GetCurrentWorkspace(wsName)
	if err != nil {
		return "", err
	}

	return hci.Path, nil
}

func (hc *HomeConfig) FindWorkspaceByPath() (*HomeConfigItem, error) {
	cwd, err := Pc.Getwd()
	if err != nil {
		return nil, err
	}
	for index, hci := range hc.Workspaces {
		if hci.RootPath != "" && strings.HasPrefix(cwd, hci.RootPath) {
			return &hc.Workspaces[index], nil
		}
	}

	return nil, errors.New("you are not in any workspace")
}

func (hc *HomeConfig) FindWorkspace(name string) *HomeConfigItem {
	for index, hci := range hc.Workspaces {
		if hci.Name == name {
			return &hc.Workspaces[index]
		}
	}

	return nil
}
