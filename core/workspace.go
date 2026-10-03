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

	cwdResolved     bool
	cwdCompName     string
	cwdBranch       string
	cwdWorktreePath string
	cwdMatchErr     error
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

	if vElc.LessThan(vCfg) {
		return errors.New(fmt.Sprintf("This workspace requires elc version %s. Please, update elc or use another binary.", ws.Config.ElcMinVersion))
	}

	return nil
}

func (ws *Workspace) createContext() (*Context, error) {
	ctx := make(Context, 0)

	ctx = ctx.add("WORKSPACE_PATH", strings.TrimRight(ws.ConfigPath, "/"))
	ctx = ctx.add("WORKSPACE_NAME", ws.Config.Name)

	for _, pair := range ws.Config.Variables {
		value, err := substVars(pair.Value, &ctx)
		if err != nil {
			return nil, err
		}
		ctx = ctx.add(pair.Key, value)
	}

	if _, found := ctx.find("WORKTREES_PATH"); !found {
		value, err := substVars("${WORKSPACE_PATH}/worktrees", &ctx)
		if err != nil {
			return nil, err
		}
		ctx = ctx.add("WORKTREES_PATH", value)
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
	name, err := ws.ComponentNameByPath()
	if err != nil {
		return nil, err
	}
	return ws.ComponentByName(name)
}

func (ws *Workspace) ComponentNameByPath() (string, error) {
	name, _, err := ws.MatchCwd()
	return name, err
}

func (ws *Workspace) MatchCwd() (string, string, error) {
	if ws.cwdResolved {
		return ws.cwdCompName, ws.cwdBranch, ws.cwdMatchErr
	}
	ws.cwdResolved = true

	if name, branch, dest, ok := ws.detectWorktreeFromCwd(); ok {
		ws.cwdCompName = name
		ws.cwdBranch = branch
		ws.cwdWorktreePath = dest
		return name, branch, nil
	}

	var bestName string
	bestLen := -1
	for name, comp := range ws.Components {
		compPath, found := comp.Context.find("SVC_PATH")
		if !found || compPath == "" {
			continue
		}
		if isInside(ws.Cwd, compPath) && len(compPath) > bestLen {
			bestLen = len(compPath)
			bestName = name
		}
	}
	if bestName == "" {
		ws.cwdMatchErr = errors.New("you are not in component folder")
		return "", "", ws.cwdMatchErr
	}
	ws.cwdCompName = bestName
	return bestName, "", nil
}

// CwdWorktreePath returns the worktree root detected from CWD, if any.
func (ws *Workspace) CwdWorktreePath() string {
	_, _, _ = ws.MatchCwd()
	return ws.cwdWorktreePath
}

func (ws *Workspace) detectWorktreeFromCwd() (string, string, string, bool) {
	if name, branch, dest, ok := ws.detectWorktreeFromPath(); ok {
		return name, branch, dest, true
	}
	return ws.detectWorktreeFromGit()
}

func (ws *Workspace) detectWorktreeFromPath() (string, string, string, bool) {
	root := WorktreesRoot(ws)
	if !isInside(ws.Cwd, root) {
		return "", "", "", false
	}

	rel := strings.TrimPrefix(pathPrefix(ws.Cwd), pathPrefix(root))
	rel = strings.Trim(rel, "/")
	if rel == "" {
		return "", "", "", false
	}

	parts := strings.Split(rel, "/")
	compName := parts[0]
	if _, err := ws.ComponentByName(compName); err != nil {
		return "", "", "", false
	}

	current := ws.Cwd
	compRoot := path.Join(root, compName)
	for isInside(current, compRoot) && path.Clean(current) != path.Clean(compRoot) {
		if Pc.FileExists(path.Join(current, ".git")) {
			branch := strings.TrimPrefix(path.Clean(current), path.Clean(compRoot)+"/")
			if branch != "" && ValidateBranchName(branch) == nil {
				return compName, branch, current, true
			}
			return "", "", "", false
		}
		next := path.Dir(current)
		if next == current {
			break
		}
		current = next
	}

	return "", "", "", false
}

func (ws *Workspace) detectWorktreeFromGit() (string, string, string, bool) {
	// CWD inside a main component checkout is never a foreign linked worktree.
	for _, comp := range ws.Components {
		svcPath, found := comp.Context.find("SVC_PATH")
		if found && svcPath != "" && isInside(ws.Cwd, svcPath) {
			return "", "", "", false
		}
	}

	if _, ok := findLinkedWorktreeRoot(ws.Cwd); !ok {
		return "", "", "", false
	}
	info, ok := inspectGitWorktree(ws.Cwd)
	if !ok || !info.Linked {
		return "", "", "", false
	}
	if ValidateBranchName(info.Branch) != nil {
		return "", "", "", false
	}
	compName, ok := ws.componentNameByMainPath(info.MainPath)
	if !ok {
		return "", "", "", false
	}
	return compName, info.Branch, info.TopLevel, true
}

func findLinkedWorktreeRoot(cwd string) (string, bool) {
	current := cwd
	for current != "" {
		gitPath := path.Join(current, ".git")
		data, err := Pc.ReadFile(gitPath)
		if err == nil {
			if strings.HasPrefix(strings.TrimSpace(string(data)), "gitdir:") {
				return current, true
			}
			return "", false
		}
		if Pc.FileExists(gitPath) {
			// Exists but not a readable file → regular .git directory (main clone).
			return "", false
		}
		next := path.Dir(current)
		if next == current {
			break
		}
		current = next
	}
	return "", false
}

func (ws *Workspace) componentNameByMainPath(mainPath string) (string, bool) {
	mainPath = path.Clean(mainPath)
	for name, comp := range ws.Components {
		svcPath, found := comp.Context.find("SVC_PATH")
		if !found || svcPath == "" {
			continue
		}
		if path.Clean(svcPath) == mainPath {
			return name, true
		}
	}
	return "", false
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

func (ws *Workspace) GetComponentNamesList() []string {
	result := make([]string, 0)
	for name, comp := range ws.Components {
		if !comp.Config.IsTemplate {
			result = append(result, name)
		}
	}

	return result
}
