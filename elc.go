package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"

	"gopkg.in/yaml.v2"
)

// ====================================

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

func newState(workspacePath string, cwd string) *State {
	st := State{
		WorkspacePath: workspacePath,
		Cwd:           cwd,
	}

	return &st
}

func (st *State) loadConfig() error {
	yamlFile, err := ioutil.ReadFile(path.Join(st.WorkspacePath, "workspace.yaml"))
	if err != nil {
		return err
	}

	cfg := &Config{}
	err = yaml.Unmarshal(yamlFile, cfg)
	if err != nil {
		return err
	}
	st.Config = *cfg

	return nil
}

func (st *State) getEnv(svcName string) ([]string, error) {
	env := make([]string, 5)
	for key, value := range st.Config.Variables {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}

	svc, err := findServicebyName(st.Config.Services, svcName)
	if err != nil {
		return nil, err
	}

	if svc.Extends != "" {
		tpl, err := findServicebyName(st.Config.Templates, svc.Extends)
		if err != nil {
			return nil, err
		}
		env = append(env, tpl.getEnv(st)...)
		env = append(env, fmt.Sprintf("TPL_PATH=%s", tpl.Path))
	}

	env = append(env, svc.getEnv(st)...)
	env = append(env, fmt.Sprintf("SVC_PATH=%s", svc.Path))
	env = append(env, fmt.Sprintf("APP_NAME=%s", svc.Name))
	env = append(env, fmt.Sprintf("COMPOSE_PROJECT_NAME=%s-%s", st.Config.Name, svc.Name))

	return env, nil
}

func (svc *Service) getEnv(st *State) []string {
	env := make([]string, 5)
	for key, value := range svc.Variables {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}

	return env
}

func (st *State) getComposeFile(svcName string) (string, error) {
	svc, err := findServicebyName(st.Config.Services, svcName)
	if err != nil {
		return "", err
	}

	if svc.ComposeFile != "" {
		return svc.ComposeFile, nil
	}

	if svc.Extends != "" {
		tpl, err := findServicebyName(st.Config.Templates, svc.Extends)
		if err != nil {
			return "", err
		}
		return tpl.ComposeFile, nil
	}

	return "", errors.New("compose file is not defined in service or template")
}

func findServicebyName(services []Service, name string) (*Service, error) {
	for _, svc := range services {
		if svc.Name == name {
			return &svc, nil
		}
	}

	return nil, errors.New(fmt.Sprintf("service or template %s not found", name))
}

func (st *State) findServiceByPath() (string, error) {
	for _, svc := range st.Config.Services {
		if strings.HasPrefix(st.Cwd, svc.Path) {
			return svc.Name, nil
		}
	}

	return "", errors.New("you are not in service folder")
}

// ================================

func contains(list []string, item string) bool {
	for _, value := range list {
		if value == item {
			return true
		}
	}
	return false
}

// ================================

func execIntercative(command []string, env []string) (int, error) {
	cmd := exec.Command(command[0], command[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = env

	err := cmd.Run()

	return cmd.ProcessState.ExitCode(), err
}

func execComposeCommand(st *State, name string, composeCommand []string) (int, error) {
	env, err := st.getEnv(name)
	if err != nil {
		return 0, err
	}

	composeFile, err := st.getComposeFile(name)
	if err != nil {
		return 0, err
	}

	command := append([]string{"docker", "compose", "-f", composeFile}, composeCommand...)

	code, err := execIntercative(command, env)
	if err != nil {
		return 0, err
	}

	return code, nil
}

// ================== HANDLERS

func actionGlobalHelp() (int, error) {
	fmt.Println("Usage: ensi init")

	return 0, nil
}

// -----------------
func actionTest(st *State) (int, error) {
	env, err := st.getEnv("crm")
	if err != nil {
		return 0, err
	}

	composeFile, err := st.getComposeFile("crm")
	if err != nil {
		return 0, err
	}

	code, err := execIntercative([]string{"docker", "compose", "-f", composeFile, "config"}, env)
	if err != nil {
		return 0, err
	}

	return code, nil
}

// -----------------
func actionStart(st *State, args []string) (int, error) {
	if contains(args, "--help") || contains(args, "-h") {
		return 0, errors.New("Usage: elc start [service]")
	}

	svcNames, err := getServiceNames(st, args)
	if err != nil {
		return 0, err
	}

	var code int

	for _, svcName := range svcNames {
		code, err = execComposeCommand(st, svcName, []string{"up", "-d"})
		if err != nil {
			return 0, err
		}
	}

	return code, nil
}

// -----------------
func actionStop(st *State, args []string) (int, error) {
	if contains(args, "--help") || contains(args, "-h") {
		return 0, errors.New("Usage: elc stop [service]")
	}

	svcNames, err := getServiceNames(st, args)
	if err != nil {
		return 0, err
	}

	var code int

	for _, svcName := range svcNames {
		code, err = execComposeCommand(st, svcName, []string{"stop"})
		if err != nil {
			return 0, err
		}
	}

	return code, nil
}

// -----------------
func actionDestroy(st *State, args []string) (int, error) {
	if contains(args, "--help") || contains(args, "-h") {
		return 0, errors.New("Usage: elc down [service]")
	}

	svcNames, err := getServiceNames(st, args)
	if err != nil {
		return 0, err
	}

	var code int

	for _, svcName := range svcNames {
		code, err = execComposeCommand(st, svcName, []string{"down"})
		if err != nil {
			return 0, err
		}
	}

	return code, nil
}

// ================= CLI

func getAction(args []string) (string, error) {
	if len(args) < 2 {
		return "", errors.New("Too few arguments")
	}

	return args[1], nil
}

func runAction(st *State, args []string) (int, error) {
	var code int
	var err error
	var action string

	action, err = getAction(os.Args)
	if err != nil {
		return 0, err
	}

	switch action {
	case "start":
		code, err = actionStart(st, args[1:])
	case "stop":
		code, err = actionStop(st, args[1:])
	case "destroy":
		code, err = actionDestroy(st, args[1:])
	default:
		code, err = actionGlobalHelp()
	}

	if err != nil {
		return 0, err
	}

	return code, nil
}

func getServiceNames(st *State, args []string) ([]string, error) {
	var svcNames []string

	if len(args) > 0 {
		svcNames = args
	} else {
		svcNames = make([]string, 1)
		svcName, err := st.findServiceByPath()
		if err != nil {
			return nil, err
		}
		svcNames = append(svcNames, svcName)
	}

	return svcNames, nil
}

// ================ Entrance

func main() {
	workdir := os.Getenv("WORKDIR")
	if workdir == "" {
		fmt.Println("variable WORKDIR is not defined")
		os.Exit(1)
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	st := newState(workdir, cwd)
	err = st.loadConfig()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	code, err := runAction(st, os.Args[1:])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	os.Exit(code)
}
