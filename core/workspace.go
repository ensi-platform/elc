package core

import (
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/hashicorp/go-version"
)

type Workspace struct {
	Aliases    map[string]string
	ConfigPath string
	Config     *WorkspaceConfig
	Cwd        string
	WillStart  []string
	Context    *Context
	Components map[string]*Component
}

func NewWorkspace(wsPath string, cwd string) *Workspace {
	ws := Workspace{
		Aliases:    make(map[string]string, 0),
		ConfigPath: wsPath,
		Cwd:        cwd,
	}

	return &ws
}

func (ws *Workspace) LoadConfig() error {
	wsc := *NewWorkspaceConfig()
	err := wsc.loadFromFile(path.Join(ws.ConfigPath, "workspace.yaml"))
	if err != nil {
		return err
	}

	envPath := path.Join(ws.ConfigPath, "env.yaml")
	if Pc.FileExists(envPath) {
		envWsc := *NewWorkspaceConfig()
		err := envWsc.loadFromFile(envPath)
		if err != nil {
			return err
		}

		wsc = wsc.merge(envWsc)
	}

	ws.Config = &wsc

	return nil
}

func (ws *Workspace) init() error {
	ctx, err := ws.createContext()
	if err != nil {
		return err
	}

	ws.Context = ctx
	ws.Components = make(map[string]*Component)
	for compName := range ws.Config.Components {
		compCfg, _ := ws.Config.Components[compName]
		ws.Components[compName] = NewComponent(compName, &compCfg, ws)
	}

	for name, realName := range ws.Config.Aliases {
		ws.Aliases[name] = realName
	}

	for _, comp := range ws.Components {
		err := comp.init()
		if err != nil {
			return err
		}
	}

	return nil
}

func (ws *Workspace) checkVersion() error {
	if ws.Config.ElcMinVersion == "" {
		return nil
	}
	vCfg, err := version.NewVersion(ws.Config.ElcMinVersion)
	if err != nil {
		return err
	}
	vElc, err := version.NewVersion(Version)
	if err != nil {
		return err
	}

	if vElc.LessThanOrEqual(vCfg) {
		return errors.New(fmt.Sprintf("This workspace requires elc version %s. Please, update elc or use another binary.", ws.Config.ElcMinVersion))
	}

	return nil
}

func (ws *Workspace) createContext() (*Context, error) {
	ctx := make(Context, 0)

	ctx = ctx.add("WORKSPACE_PATH", strings.TrimRight(ws.ConfigPath, "/"))
	ctx = ctx.add("WORKSPACE_NAME", ws.Config.Name)

	for _, pair := range ws.Config.Variables {
		value, err := substVars(pair.Value.(string), &ctx)
		if err != nil {
			return nil, err
		}
		ctx = ctx.add(pair.Key.(string), value)
	}

	return &ctx, nil
}

func (ws *Workspace) ComponentByName(name string) (*Component, error) {
	realName, found := ws.Aliases[name]
	if found {
		name = realName
	}
	comp, found := ws.Components[name]
	if !found {
		return nil, errors.New(fmt.Sprintf("service '%s' not found", name))
	}
	return comp, nil
}

func (ws *Workspace) ComponentByPath() (*Component, error) {
	for _, comp := range ws.Components {
		compPath, found := comp.Context.find("SVC_PATH")
		if found {
			if strings.HasPrefix(compPath, ws.Cwd) {
				return comp, nil
			}
		}
	}
	return nil, errors.New(fmt.Sprintf("you are not in component folder"))
}

func (ws *Workspace) ComponentNameByPath() (string, error) {
	for name, comp := range ws.Components {
		compPath, found := comp.Context.find("SVC_PATH")
		if found {
			if strings.HasPrefix(compPath, ws.Cwd) {
				return name, nil
			}
		}
	}
	return "", errors.New(fmt.Sprintf("you are not in component folder"))
}

func (ws *Workspace) GetComponentNames() []string {
	result := make([]string, 0)
	for name, comp := range ws.Components {
		if !comp.Config.IsTemplate && comp.Config.HostedIn == "" {
			result = append(result, name)
		}
	}

	return result
}

func (ws *Workspace) FindComponentNamesByTag(tag string) []string {
	result := make([]string, 0)
	for name, comp := range ws.Components {
		if !comp.Config.IsTemplate {
			for _, compTag := range comp.Config.Tags {
				if compTag == tag {
					result = append(result, name)
				}
			}
		}
	}

	return result
}
