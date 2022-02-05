package src

import (
	"errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

type HomeConfigItem struct {
	Name string `yaml:"name"`
	Path string `yaml:"path"`
}

type HomeConfig struct {
	Path             string           `yaml:"-"`
	CurrentWorkspace string           `yaml:"current_workspace"`
	Workspaces       []HomeConfigItem `yaml:"workspaces"`
}

func LoadHomeConfig(configPath string) (*HomeConfig, error) {
	yamlFile, err := ioutil.ReadFile(configPath)
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

	err = ioutil.WriteFile(homeConfig.Path, data, 0600)
	if err != nil {
		return err
	}

	return nil
}

func CheckHomeConfigIsEmpty(configPath string) error {
	_, err := os.Stat(configPath)
	if err == nil {
		return nil
	}
	return SaveHomeConfig(&HomeConfig{Path: configPath})
}

func (hc *HomeConfig) AddWorkspace(name string, path string) error {
	hc.Workspaces = append(hc.Workspaces, HomeConfigItem{Name: name, Path: path})
	return SaveHomeConfig(hc)
}

func (hc *HomeConfig) GetCurrentWsPath() (string, error) {
	if hc.CurrentWorkspace == "" {
		return "", errors.New("current workspace is not set")
	}

	for _, hci := range hc.Workspaces {
		if hci.Name == hc.CurrentWorkspace {
			return hci.Path, nil
		}
	}

	return "", errors.New("current workspace is bad")
}

func (hc *HomeConfig) findWorkspace(name string) *HomeConfigItem {
	for _, workspace := range hc.Workspaces {
		if workspace.Name == name {
			return &workspace
		}
	}

	return nil
}
