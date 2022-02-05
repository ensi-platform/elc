package src

import (
	"fmt"
)

type TemplateConfig struct {
	Name        string            `yaml:"name"`
	Path        string            `yaml:"path"`
	ComposeFile string            `yaml:"compose_file"`
	Variables   map[string]string `yaml:"variables"`
}

type ServiceConfig struct {
	TemplateConfig `yaml:",inline"`
	Extends        string              `yaml:"extends"`
	Dependencies   map[string][]string `yaml:"dependencies"`
}

type ModuleConfig struct {
	Name     string `yaml:"name"`
	Path     string `yaml:"path"`
	HostedIn string `yaml:"hosted_in"`
	ExecPath string `yaml:"exec_path"`
}

func (svcCfg *TemplateConfig) GetEnv() []string {
	var env []string
	for key, value := range svcCfg.Variables {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}

	return env
}

func (svcCfg *ServiceConfig) GetDeps(tag string) []string {
	var result []string
	for key, tags := range svcCfg.Dependencies {
		if tag == "" || contains(tags, tag) {
			result = append(result, key)
		}
	}

	return result
}
