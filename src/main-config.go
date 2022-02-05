package src

import (
	"bytes"
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path"
	"strings"
	"text/template"
)

type MainConfig struct {
	WorkspacePath string            `yaml:"-"`
	Cwd           string            `yaml:"-"`
	WillStart     []string          `yaml:"-"`
	Name          string            `yaml:"name"`
	BaseDomain    string            `yaml:"base_domain"`
	Templates     []TemplateConfig  `yaml:"templates"`
	Services      []ServiceConfig   `yaml:"services"`
	Variables     map[string]string `yaml:"variables"`
}

func NewConfig(workspacePath string, cwd string) *MainConfig {
	cfg := MainConfig{
		WorkspacePath: workspacePath,
		Cwd:           cwd,
	}

	return &cfg
}

func (cfg *MainConfig) LoadFromFile() error {
	yamlFile, err := ioutil.ReadFile(path.Join(cfg.WorkspacePath, "workspace.yaml"))
	if err != nil {
		return err
	}

	tmpl, err := template.New("config").Parse(string(yamlFile))
	if err != nil {
		return err
	}

	var buff bytes.Buffer

	err = tmpl.Execute(&buff, cfg)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(buff.Bytes(), cfg)
	if err != nil {
		return err
	}

	return nil
}

func (cfg *MainConfig) FindServiceByPath() (string, error) {
	for _, svc := range cfg.Services {
		if strings.HasPrefix(cfg.Cwd, svc.Path) {
			return svc.Name, nil
		}
	}

	return "", errors.New("you are not in service folder")
}

func (cfg *MainConfig) FindServiceByName(name string) (*ServiceConfig, error) {
	for _, svc := range cfg.Services {
		if svc.Name == name {
			return &svc, nil
		}
	}

	return nil, errors.New(fmt.Sprintf("service %s not found", name))
}

func (cfg *MainConfig) FindTemplateByName(name string) (*TemplateConfig, error) {
	for _, tpl := range cfg.Templates {
		if tpl.Name == name {
			return &tpl, nil
		}
	}

	return nil, errors.New(fmt.Sprintf("template %s not found", name))
}
