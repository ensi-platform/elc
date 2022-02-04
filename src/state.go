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

type State struct {
	WorkspacePath string
	Cwd           string
	Config        Config
}

type Config struct {
	Name       string
	BaseDomain string            `yaml:"base_domain"`
	Templates  []Service         `yaml:"templates"`
	Services   []Service         `yaml:"services"`
	Variables  map[string]string `yaml:"variables"`
}

type Service struct {
	Name        string
	Path        string
	Extends     string            `yaml:"extends"`
	ComposeFile string            `yaml:"compose_file"`
	Variables   map[string]string `yaml:"variables"`
}

func NewState(workspacePath string, cwd string) *State {
	st := State{
		WorkspacePath: workspacePath,
		Cwd:           cwd,
	}

	return &st
}

func (st *State) LoadConfig() error {
	yamlFile, err := ioutil.ReadFile(path.Join(st.WorkspacePath, "workspace.yaml"))
	if err != nil {
		return err
	}

	tmpl, err := template.New("config").Parse(string(yamlFile))
	if err != nil {
		return err
	}

	var buff bytes.Buffer

	err = tmpl.Execute(&buff, st)
	if err != nil {
		return err
	}

	cfg := &Config{}
	err = yaml.Unmarshal(buff.Bytes(), cfg)
	if err != nil {
		return err
	}
	st.Config = *cfg

	return nil
}

func (st *State) GetEnv(svcName string) ([]string, error) {
	env := make([]string, 5)
	for key, value := range st.Config.Variables {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}

	svc, err := FindServicebyName(st.Config.Services, svcName)
	if err != nil {
		return nil, err
	}

	if svc.Extends != "" {
		tpl, err := FindServicebyName(st.Config.Templates, svc.Extends)
		if err != nil {
			return nil, err
		}
		env = append(env, tpl.GetEnv()...)
		env = append(env, fmt.Sprintf("TPL_PATH=%s", tpl.Path))
	}

	env = append(env, svc.GetEnv()...)
	env = append(env, fmt.Sprintf("SVC_PATH=%s", svc.Path))
	env = append(env, fmt.Sprintf("APP_NAME=%s", svc.Name))
	env = append(env, fmt.Sprintf("COMPOSE_PROJECT_NAME=%s-%s", st.Config.Name, svc.Name))
	env = append(env, fmt.Sprintf("WORKSPACE_NAME=%s", st.Config.Name))

	return env, nil
}

func (svc *Service) GetEnv() []string {
	env := make([]string, 5)
	for key, value := range svc.Variables {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}

	return env
}

func (st *State) GetComposeFile(svcName string) (string, error) {
	svc, err := FindServicebyName(st.Config.Services, svcName)
	if err != nil {
		return "", err
	}

	if svc.ComposeFile != "" {
		return svc.ComposeFile, nil
	}

	if svc.Extends != "" {
		tpl, err := FindServicebyName(st.Config.Templates, svc.Extends)
		if err != nil {
			return "", err
		}
		return tpl.ComposeFile, nil
	}

	return "", errors.New("compose file is not defined in service or template")
}

func FindServicebyName(services []Service, name string) (*Service, error) {
	for _, svc := range services {
		if svc.Name == name {
			return &svc, nil
		}
	}

	return nil, errors.New(fmt.Sprintf("service or template %s not found", name))
}

func (st *State) FindServiceByPath() (string, error) {
	for _, svc := range st.Config.Services {
		if strings.HasPrefix(st.Cwd, svc.Path) {
			return svc.Name, nil
		}
	}

	return "", errors.New("you are not in service folder")
}
