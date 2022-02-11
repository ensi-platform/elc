package src

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

type MainConfig struct {
	WorkspacePath  string            `yaml:"-"`
	Cwd            string            `yaml:"-"`
	WillStart      []string          `yaml:"-"`
	LocalVariables map[string]string `yaml:"-"`
	Name           string            `yaml:"name"`
	Templates      []TemplateConfig  `yaml:"templates"`
	Services       []ServiceConfig   `yaml:"services"`
	Modules        []ModuleConfig    `yaml:"modules"`
	Variables      yaml.MapSlice     `yaml:"variables"`
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

	err = yaml.Unmarshal(yamlFile, cfg)
	if err != nil {
		return err
	}

	_, err = os.Stat(path.Join(cfg.WorkspacePath, "env.yaml"))
	if err == nil {
		yamlFile, err = ioutil.ReadFile(path.Join(cfg.WorkspacePath, "env.yaml"))
		if err != nil {
			return err
		}

		err = yaml.Unmarshal(yamlFile, &cfg.LocalVariables)
		if err != nil {
			return err
		}
	}

	return nil
}

func (cfg *MainConfig) makeGlobalEnv() (Context, error) {
	ctx := make(Context, 0)

	ctx = ctx.add("WORKSPACE_PATH", strings.TrimRight(cfg.WorkspacePath, "/"))
	ctx = ctx.add("WORKSPACE_NAME", cfg.Name)

	for key, rawValue := range cfg.LocalVariables {
		value, err := substVars(rawValue, ctx)
		if err != nil {
			return nil, err
		}
		ctx = ctx.add(key, value)
	}

	for _, pair := range cfg.Variables {
		value, err := substVars(pair.Value.(string), ctx)
		if err != nil {
			return nil, err
		}
		ctx = ctx.add(pair.Key.(string), value)
	}

	return ctx, nil
}

func (cfg *MainConfig) renderPath(path string) (string, error) {
	env, err := cfg.makeGlobalEnv()
	if err != nil {
		return "", err
	}
	return substVars(path, env)
}

func (cfg *MainConfig) FindServiceByPath() (string, error) {
	for _, svc := range cfg.Services {
		svcPath, err := cfg.renderPath(svc.Path)
		if err != nil {
			return "", err
		}
		if strings.HasPrefix(cfg.Cwd, svcPath) {
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

func (cfg *MainConfig) FindModuleByName(name string) (*ModuleConfig, error) {
	for _, mdl := range cfg.Modules {
		if mdl.Name == name {
			return &mdl, nil
		}
	}

	return nil, errors.New(fmt.Sprintf("module %s not found", name))
}

func (cfg *MainConfig) FindModuleByPath() (*ModuleConfig, error) {
	for _, mdl := range cfg.Modules {
		mdlPath, err := cfg.renderPath(mdl.Path)
		if err != nil {
			return nil, err
		}
		if strings.HasPrefix(cfg.Cwd, mdlPath) {
			return &mdl, nil
		}
	}

	return nil, errors.New("you are not in module folder")
}

func (cfg *MainConfig) GetAllSvcNames() []string {
	result := make([]string, 0)
	for _, svc := range cfg.Services {
		result = append(result, svc.Name)
	}

	return result
}
