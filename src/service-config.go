package src

import (
	"gopkg.in/yaml.v2"
)

type TemplateConfig struct {
	Path        string        `yaml:"path"`
	ComposeFile string        `yaml:"compose_file"`
	Variables   yaml.MapSlice `yaml:"variables"`
}

type ServiceConfig struct {
	TemplateConfig `yaml:",inline"`
	Extends        string              `yaml:"extends"`
	Dependencies   map[string][]string `yaml:"dependencies"`
}

type ModuleConfig struct {
	Path     string `yaml:"path"`
	HostedIn string `yaml:"hosted_in"`
	ExecPath string `yaml:"exec_path"`
}

func (svcCfg *ServiceConfig) GetDeps(mode string) []string {
	var result []string
	for key, modes := range svcCfg.Dependencies {
		if contains(modes, mode) {
			result = append(result, key)
		}
	}

	return result
}
