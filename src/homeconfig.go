package src

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"strings"
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

func RunWorkspaceAction(hc *HomeConfig, args []string) (int, error) {
	var code int
	var err error
	var action string

	action, err = getAction(args)
	if err != nil {
		return 0, err
	}

	switch action {
	case "-h", "--help":
		printWorkspacesHelp()
		code, err = 0, nil
	case "list", "ls":
		code, err = actionListWorkspaces(hc)
	case "add":
		code, err = actionAddWorkspace(hc, args[1:])
	case "select":
		code, err = actionSelectWorkspace(hc, args[1:])
	case "show":
		code, err = actionShowCurrentWorkspace(hc)
	default:
		printWorkspacesHelp()
	}

	if err != nil {
		return 0, err
	}

	return code, nil
}

func printWorkspacesHelp() {
	fmt.Println(
		strings.Join([]string{
			"Usage: elc workspace [command] [args]",
			"",
			"Available commands:",
			"  ls, list          - list available workspaces",
			"  show              - show current workspace name",
			"  add <name> <path> - add new workspace",
			"  select <name>     - select workspace as current",
		}, "\n"))
}

func actionShowCurrentWorkspace(hc *HomeConfig) (int, error) {
	fmt.Println(hc.CurrentWorkspace)

	return 0, nil
}

func actionSelectWorkspace(hc *HomeConfig, args []string) (int, error) {
	if len(args) != 1 {
		return 0, errors.New("command requires exactly 1 argument")
	}

	name := args[0]

	ws := hc.findWorkspace(name)
	if ws == nil {
		return 0, errors.New(fmt.Sprintf("workspace with name '%s' is not defined", name))
	}

	hc.CurrentWorkspace = name
	err := SaveHomeConfig(hc)
	if err != nil {
		return 0, err
	}
	fmt.Printf("active workspace changed to '%s'\n", name)
	return 0, nil
}

func actionAddWorkspace(hc *HomeConfig, args []string) (int, error) {
	if len(args) != 2 {
		return 0, errors.New("command requires exactly 2 arguments")
	}

	name := args[0]
	wsPath := args[1]

	ws := hc.findWorkspace(name)
	if ws != nil {
		return 0, errors.New(fmt.Sprintf("workspace with name '%s' already exists", name))
	}

	err := hc.AddWorkspace(name, wsPath)
	if err != nil {
		return 0, err
	}

	fmt.Printf("workspace '%s' is added\n", name)

	return 0, nil
}

func actionListWorkspaces(hc *HomeConfig) (int, error) {
	for _, workspace := range hc.Workspaces {
		fmt.Printf("%-10s %s\n", workspace.Name, workspace.Path)
	}
	return 0, nil
}
