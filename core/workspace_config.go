package core

import (
	"gopkg.in/yaml.v2"
)

type WorkspaceConfig struct {
	Name          string                     `yaml:"name"`
	ElcMinVersion string                     `yaml:"elc_min_version"`
	Components    map[string]ComponentConfig `yaml:"components"`
	Variables     yaml.MapSlice              `yaml:"variables"`

	// deprecated
	Aliases map[string]string `yaml:"aliases"`
	// deprecated
	Templates map[string]ComponentConfig `yaml:"templates"`
	// deprecated
	Services map[string]ComponentConfig `yaml:"services"`
	// deprecated
	Modules map[string]ComponentConfig `yaml:"modules"`
}

func NewWorkspaceConfig() *WorkspaceConfig {
	return &WorkspaceConfig{
		Aliases:    make(map[string]string, 0),
		Components: make(map[string]ComponentConfig, 0),
		Templates:  make(map[string]ComponentConfig, 0),
		Services:   make(map[string]ComponentConfig, 0),
		Modules:    make(map[string]ComponentConfig, 0),
	}
}

func (wsc *WorkspaceConfig) normalize() {
	for k, v := range wsc.Templates {
		v.IsTemplate = true
		wsc.Components[k] = v
	}
	wsc.Templates = nil

	for k, v := range wsc.Services {
		wsc.Components[k] = v
	}
	wsc.Services = nil

	for k, v := range wsc.Modules {
		wsc.Components[k] = v
	}
	wsc.Modules = nil
}

func (wsc WorkspaceConfig) merge(wsc2 WorkspaceConfig) WorkspaceConfig {
	for name, cc := range wsc2.Components {
		if _, exists := wsc.Components[name]; !exists {
			wsc.Components[name] = cc
		} else {
			wsc.Components[name] = wsc.Components[name].merge(cc)
		}
	}

	for alias, ccName := range wsc2.Aliases {
		wsc.Aliases[alias] = ccName
	}

	for name, comp := range wsc.Components {
		if comp.Alias != "" {
			wsc.Aliases[comp.Alias] = name
		}
	}

	wsc.Variables = append(wsc2.Variables, wsc.Variables...)

	return wsc
}

func (wsc *WorkspaceConfig) loadFromFile(wscPath string) error {
	yamlFile, err := Pc.ReadFile(wscPath)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(yamlFile, wsc)
	if err != nil {
		return err
	}

	wsc.normalize()

	return nil
}
