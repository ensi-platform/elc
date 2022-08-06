package core

import "gopkg.in/yaml.v2"

type ModeList []string

func (s ModeList) contains(v string) bool {
	for _, item := range s {
		if item == v {
			return true
		}
	}
	return false
}

type ComponentConfig struct {
	Alias        string              `yaml:"alias"`
	ComposeFile  string              `yaml:"compose_file"`
	Dependencies map[string]ModeList `yaml:"dependencies"`
	ExecPath     string              `yaml:"exec_path"`
	Extends      string              `yaml:"extends"`
	HostedIn     string              `yaml:"hosted_in"`
	Hostname     string              `yaml:"hostname"`
	IsTemplate   bool                `yaml:"is_template"`
	Path         string              `yaml:"path"`
	Replace      bool                `yaml:"replace"`
	Variables    yaml.MapSlice       `yaml:"variables"`
}

func (cc ComponentConfig) merge(cc2 ComponentConfig) ComponentConfig {
	if cc2.Replace {
		return cc2
	}

	if cc2.Path != "" {
		cc.Path = cc2.Path
	}
	if cc2.ComposeFile != "" {
		cc.ComposeFile = cc2.ComposeFile
	}
	if cc2.Extends != "" {
		cc.Extends = cc2.Extends
	}
	if cc2.HostedIn != "" {
		cc.HostedIn = cc2.HostedIn
	}
	if cc2.ExecPath != "" {
		cc.ExecPath = cc2.ExecPath
	}
	if cc2.Alias != "" {
		cc.Alias = cc2.Alias
	}
	cc.Variables = append(cc.Variables, cc2.Variables...)

	for depSvc, modes := range cc2.Dependencies {
		if cc.Dependencies[depSvc] == nil {
			cc.Dependencies[depSvc] = make([]string, 1)
		}
		for _, mode := range modes {
			if !cc.Dependencies[depSvc].contains(mode) {
				cc.Dependencies[depSvc] = append(cc.Dependencies[depSvc], mode)
			}
		}
	}

	return cc
}

func (cc *ComponentConfig) GetDeps(mode string) []string {
	var result []string
	for key, modes := range cc.Dependencies {
		if modes.contains(mode) {
			result = append(result, key)
		}
	}

	return result
}
