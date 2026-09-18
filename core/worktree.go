package core

import (
	"errors"
	"fmt"
	"path"
	"strings"
	"unicode"
)

func WorktreesRoot(ws *Workspace) string {
	if ws.Context != nil {
		if root, found := ws.Context.find("WORKTREES_PATH"); found && root != "" {
			return root
		}
	}
	return path.Join(strings.TrimRight(ws.ConfigPath, "/"), "worktrees")
}

func WorktreePath(ws *Workspace, componentName, branch string) string {
	return path.Join(WorktreesRoot(ws), componentName, branch)
}

func ValidateBranchName(branch string) error {
	branch = strings.TrimSpace(branch)
	if branch == "" {
		return errors.New("branch is required")
	}
	if path.IsAbs(branch) {
		return errors.New("invalid branch name")
	}
	cleaned := path.Clean("/" + branch)
	if cleaned == "/" || strings.Contains(cleaned, "/../") || strings.HasSuffix(cleaned, "/..") {
		return errors.New("invalid branch name")
	}
	for _, part := range strings.Split(branch, "/") {
		if part == "" || part == "." || part == ".." {
			return errors.New("invalid branch name")
		}
	}
	if SanitizeInstanceName(branch) == "" {
		return errors.New("invalid branch name")
	}
	return nil
}

func SanitizeInstanceName(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	prevDash := false
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			prevDash = false
			continue
		}
		if !prevDash {
			b.WriteByte('-')
			prevDash = true
		}
	}
	return strings.Trim(b.String(), "-")
}

func pathPrefix(p string) string {
	p = path.Clean(p)
	if p == "/" {
		return "/"
	}
	return p + "/"
}

func isInside(cwd, root string) bool {
	if cwd == "" || root == "" {
		return false
	}
	return strings.HasPrefix(pathPrefix(cwd), pathPrefix(root))
}

func gitC(repo string, args ...string) []string {
	return append([]string{"git", "-C", repo}, args...)
}

func gitRefExists(repo, ref string, env []string) bool {
	code, _, err := Pc.ExecToString(gitC(repo, "show-ref", "--verify", "--quiet", ref), env)
	return err == nil && code == 0
}

func resolveWorktreeStartPoint(repo, source string, env []string) (string, error) {
	source = strings.TrimSpace(source)
	if source == "" {
		return "", errors.New("source is required")
	}
	if strings.EqualFold(source, "head") {
		return "HEAD", nil
	}
	if gitRefExists(repo, "refs/heads/"+source, env) {
		return source, nil
	}
	if gitRefExists(repo, "refs/remotes/origin/"+source, env) {
		return "origin/" + source, nil
	}
	return "", errors.New(fmt.Sprintf("source '%s' not found locally or on origin", source))
}

func (comp *Component) gitEnv() []string {
	if comp.Context == nil {
		return nil
	}
	return comp.Context.renderMapToEnv()
}

func (comp *Component) ForWorktree(branch string) (*Component, error) {
	if err := ValidateBranchName(branch); err != nil {
		return nil, err
	}

	dest := WorktreePath(comp.Workspace, comp.Name, branch)
	cfg := *comp.Config
	cfg.Path = dest

	clone := NewComponent(comp.Name, &cfg, comp.Workspace)
	if err := clone.init(); err != nil {
		return nil, err
	}

	sanitized := SanitizeInstanceName(branch)
	ctx := clone.Context.add("GIT_BRANCH", branch)
	ctx = ctx.add("APP_NAME", fmt.Sprintf("%s-%s", comp.Name, sanitized))
	ctx = ctx.add("COMPOSE_PROJECT_NAME", fmt.Sprintf("%s-%s-%s", comp.Workspace.Config.Name, comp.Name, sanitized))
	clone.Context = &ctx

	return clone, nil
}

func (comp *Component) getWorktreeCreateHook() string {
	if comp.Config.Hooks.WorktreeCreate != "" {
		return comp.Config.Hooks.WorktreeCreate
	}
	if comp.Template != nil {
		return comp.Template.Hooks.WorktreeCreate
	}
	return ""
}

func (comp *Component) getWorktreeRemoveHook() string {
	if comp.Config.Hooks.WorktreeRemove != "" {
		return comp.Config.Hooks.WorktreeRemove
	}
	if comp.Template != nil {
		return comp.Template.Hooks.WorktreeRemove
	}
	return ""
}

func (comp *Component) runWorktreeHook(hook string, options *GlobalOptions, noHook bool, hookArgs []string, hookName string) error {
	if noHook {
		return nil
	}
	if hook == "" {
		if len(hookArgs) > 0 {
			return errors.New(fmt.Sprintf("%s is not defined, cannot pass extra arguments", hookName))
		}
		return nil
	}

	hook, err := comp.Context.RenderString(hook)
	if err != nil {
		return err
	}
	if hook == "" {
		if len(hookArgs) > 0 {
			return errors.New(fmt.Sprintf("%s is not defined, cannot pass extra arguments", hookName))
		}
		return nil
	}

	command := append([]string{hook}, hookArgs...)
	_, err = comp.execInteractive(command, options)
	return err
}

func (comp *Component) runWorktreeCreateHook(options *GlobalOptions, noHook bool, hookArgs []string) error {
	return comp.runWorktreeHook(comp.getWorktreeCreateHook(), options, noHook, hookArgs, "hooks.worktree_create")
}

func (comp *Component) runWorktreeRemoveHook(options *GlobalOptions, noHook bool) error {
	return comp.runWorktreeHook(comp.getWorktreeRemoveHook(), options, noHook, nil, "hooks.worktree_remove")
}

func (comp *Component) AddWorktree(options *GlobalOptions, branch string, hookArgs []string, noHook bool) error {
	if err := ValidateBranchName(branch); err != nil {
		return err
	}
	if comp.Config.IsTemplate {
		return errors.New("cannot create worktree for a template")
	}
	if comp.Config.Repository == "" {
		return errors.New(fmt.Sprintf("repository of component %s is not defined. Check workspace.yaml", comp.Name))
	}

	if !noHook && len(hookArgs) > 0 && comp.getWorktreeCreateHook() == "" {
		return errors.New("hooks.worktree_create is not defined, cannot pass extra arguments")
	}

	cloned, err := comp.IsCloned()
	if err != nil {
		return err
	}
	if !cloned {
		return errors.New(fmt.Sprintf("component %s is not cloned", comp.Name))
	}

	mainPath, found := comp.Context.find("SVC_PATH")
	if !found || mainPath == "" {
		return errors.New("path of component is not defined.Check workspace.yaml")
	}

	dest := WorktreePath(comp.Workspace, comp.Name, branch)
	if Pc.FileExists(dest) {
		return errors.New(fmt.Sprintf("worktree '%s' for component %s already exists", branch, comp.Name))
	}

	if err := Pc.MkdirAll(path.Dir(dest)); err != nil {
		return err
	}

	env := comp.gitEnv()
	if options.Debug {
		_, _ = Pc.Printf(">> %s\n", strings.Join(gitC(mainPath, "fetch", "--prune"), " "))
	}
	if !options.DryRun {
		code, _, err := Pc.ExecToString(gitC(mainPath, "fetch", "--prune"), env)
		if err != nil || code != 0 {
			if err != nil {
				return errors.New(fmt.Sprintf("git fetch failed: %s", err))
			}
			return errors.New("git fetch failed")
		}
	}

	hasLocal := gitRefExists(mainPath, "refs/heads/"+branch, env)
	if !hasLocal && gitRefExists(mainPath, "refs/remotes/origin/"+branch, env) {
		// Create a local branch that tracks origin/<branch>; user only passes <branch>.
		_, err := comp.execInteractive(gitC(mainPath, "branch", "--track", branch, "origin/"+branch), options)
		if err != nil {
			return err
		}
		hasLocal = true
	}

	var addCmd []string
	if hasLocal {
		addCmd = gitC(mainPath, "worktree", "add", dest, branch)
	} else if options.Source != "" {
		startPoint, err := resolveWorktreeStartPoint(mainPath, options.Source, env)
		if err != nil {
			return err
		}
		_, _ = Pc.Printf("branch '%s' not found, creating from %s\n", branch, startPoint)
		addCmd = gitC(mainPath, "worktree", "add", "-b", branch, dest, startPoint)
	} else {
		return errors.New(fmt.Sprintf("branch '%s' not found locally or on origin (use --source=<branch|HEAD> to create it)", branch))
	}

	if _, err := comp.execInteractive(addCmd, options); err != nil {
		return err
	}

	instance, err := comp.ForWorktree(branch)
	if err != nil {
		return err
	}

	return instance.runWorktreeCreateHook(options, noHook, hookArgs)
}

// EnsureWorktree returns an existing worktree instance or creates one (fetch + remote resolve).
func (comp *Component) EnsureWorktree(options *GlobalOptions, branch string, noHook bool) (*Component, error) {
	instance, err := comp.ForWorktree(branch)
	if err != nil {
		return nil, err
	}

	exists, err := instance.IsCloned()
	if err != nil {
		return nil, err
	}
	if exists {
		return instance, nil
	}

	if err := comp.AddWorktree(options, branch, nil, noHook); err != nil {
		return nil, err
	}

	return comp.ForWorktree(branch)
}

func (comp *Component) RemoveWorktree(options *GlobalOptions, branch string, force bool, noHook bool) error {
	if err := ValidateBranchName(branch); err != nil {
		return err
	}

	mainPath, found := comp.Context.find("SVC_PATH")
	if !found || mainPath == "" {
		return errors.New("path of component is not defined.Check workspace.yaml")
	}

	dest := WorktreePath(comp.Workspace, comp.Name, branch)
	instance, err := comp.ForWorktree(branch)
	if err != nil {
		return err
	}

	if err := instance.runWorktreeRemoveHook(options, noHook); err != nil {
		if !force {
			return err
		}
		_, _ = Pc.Printf("warning: hooks.worktree_remove failed: %s\n", err)
	}

	if instance.HasCompose() {
		if _, err := instance.execComposeInteractive([]string{"down"}, options); err != nil {
			if !force {
				return err
			}
			_, _ = Pc.Printf("warning: failed to destroy worktree containers: %s\n", err)
		}
	}

	removeCmd := gitC(mainPath, "worktree", "remove", dest)
	if force {
		removeCmd = gitC(mainPath, "worktree", "remove", "--force", dest)
	}
	if _, err := comp.execInteractive(removeCmd, options); err != nil {
		return err
	}

	_, _ = comp.execInteractive(gitC(mainPath, "worktree", "prune"), options)
	return nil
}
