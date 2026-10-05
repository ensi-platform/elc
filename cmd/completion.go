package cmd

import (
	"sort"
	"strings"

	"github.com/ensi-platform/elc/core"
	"github.com/spf13/cobra"
)

var composeSubcommands = []string{
	"build", "config", "create", "down", "events", "exec", "images", "kill",
	"logs", "pause", "port", "ps", "pull", "push", "restart", "rm", "run",
	"start", "stop", "top", "unpause", "up", "version",
}

func ensurePC() {
	if core.Pc == nil {
		core.Pc = &core.RealPC{}
	}
}

func completionWorkspace() (*core.Workspace, error) {
	ensurePC()
	return core.LoadWorkspaceConfigForCompletion(globalOptions.WorkspaceName)
}

func uniqueSorted(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, v := range values {
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	sort.Strings(out)
	return out
}

func filterAlreadyUsed(candidates []string, used []string) []string {
	usedSet := make(map[string]struct{}, len(used))
	for _, u := range used {
		usedSet[u] = struct{}{}
	}
	out := make([]string, 0, len(candidates))
	for _, c := range candidates {
		if _, ok := usedSet[c]; !ok {
			out = append(out, c)
		}
	}
	return out
}

func listWorkspaceNames() []string {
	ensurePC()
	hc, err := core.CheckAndLoadHC()
	if err != nil {
		return nil
	}
	names := make([]string, 0, len(hc.Workspaces)+1)
	for _, ws := range hc.Workspaces {
		names = append(names, ws.Name)
	}
	return uniqueSorted(names)
}

func listComponentNames() []string {
	ws, err := completionWorkspace()
	if err != nil {
		return nil
	}
	names := ws.GetComponentNamesList()
	for alias := range ws.Aliases {
		names = append(names, alias)
	}
	return uniqueSorted(names)
}

func listTagNames() []string {
	ws, err := completionWorkspace()
	if err != nil {
		return nil
	}
	var tags []string
	for _, comp := range ws.Components {
		if comp.Config.IsTemplate {
			continue
		}
		tags = append(tags, comp.Config.Tags...)
	}
	return uniqueSorted(tags)
}

func listModeNames() []string {
	ws, err := completionWorkspace()
	modes := []string{"default", "hook"}
	if err != nil {
		return uniqueSorted(modes)
	}
	for _, comp := range ws.Components {
		for _, modeList := range comp.Config.Dependencies {
			modes = append(modes, modeList...)
		}
	}
	return uniqueSorted(modes)
}

func resolveCompletionComponent(ws *core.Workspace) (*core.Component, error) {
	if globalOptions.ComponentName != "" {
		return ws.ComponentByName(globalOptions.ComponentName)
	}
	return ws.ComponentByPath()
}

func listWorktreeBranches() []string {
	ws, err := completionWorkspace()
	if err != nil {
		return nil
	}
	filter := globalOptions.ComponentName
	if filter == "" {
		if name, err := ws.ComponentNameByPath(); err == nil {
			filter = name
		}
	}
	infos, err := ws.ListWorktrees(filter)
	if err != nil {
		return nil
	}
	branches := make([]string, 0, len(infos))
	for _, info := range infos {
		branches = append(branches, info.Branch)
	}
	return uniqueSorted(branches)
}

func listGitBranches() []string {
	ws, err := completionWorkspace()
	if err != nil {
		return nil
	}
	comp, err := resolveCompletionComponent(ws)
	if err != nil {
		return nil
	}
	svcPath, found := comp.Context.Find("SVC_PATH")
	if !found || svcPath == "" {
		return nil
	}
	code, out, err := core.Pc.ExecToString(
		[]string{"git", "-C", svcPath, "for-each-ref", "--format=%(refname:short)", "refs/heads", "refs/remotes/origin"},
		nil,
	)
	if err != nil || code != 0 || out == "" {
		return nil
	}

	var branches []string
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || line == "origin" || line == "origin/HEAD" {
			continue
		}
		line = strings.TrimPrefix(line, "origin/")
		branches = append(branches, line)
	}
	return uniqueSorted(branches)
}

func completeNoFile(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return nil, cobra.ShellCompDirectiveNoFileComp
}

func completeComponentsFlag(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return listComponentNames(), cobra.ShellCompDirectiveNoFileComp
}

func completeWorkspacesFlag(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return listWorkspaceNames(), cobra.ShellCompDirectiveNoFileComp
}

func completeWorkspaceSelectArgs(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	names := append(listWorkspaceNames(), "auto")
	return uniqueSorted(names), cobra.ShellCompDirectiveNoFileComp
}

func completeWorkspaceNamesArgs(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return listWorkspaceNames(), cobra.ShellCompDirectiveNoFileComp
}

func completeWorkspaceSetRootArgs(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
	switch len(args) {
	case 0:
		return listWorkspaceNames(), cobra.ShellCompDirectiveNoFileComp
	case 1:
		return nil, cobra.ShellCompDirectiveFilterDirs
	default:
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
}

func completeWorkspaceAddArgs(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
	switch len(args) {
	case 0:
		return nil, cobra.ShellCompDirectiveNoFileComp
	case 1:
		return nil, cobra.ShellCompDirectiveFilterDirs
	default:
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
}

func completeTagsFlag(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return listTagNames(), cobra.ShellCompDirectiveNoFileComp
}

func completeModesFlag(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return listModeNames(), cobra.ShellCompDirectiveNoFileComp
}

func completeWorktreeBranchesFlag(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return listWorktreeBranches(), cobra.ShellCompDirectiveNoFileComp
}

func completeSourceFlag(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	branches := append([]string{"HEAD"}, listGitBranches()...)
	return uniqueSorted(branches), cobra.ShellCompDirectiveNoFileComp
}

func completeComponentArgs(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
	return filterAlreadyUsed(listComponentNames(), args), cobra.ShellCompDirectiveNoFileComp
}

func completeHookNames(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return core.ComponentHookNames(), cobra.ShellCompDirectiveNoFileComp
}

func completeComposeArgs(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveDefault
	}
	return composeSubcommands, cobra.ShellCompDirectiveNoFileComp
}

func completeDirectoryArgs(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return nil, cobra.ShellCompDirectiveFilterDirs
}

func completeGitBranchArgs(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return listGitBranches(), cobra.ShellCompDirectiveNoFileComp
}

func completeWorktreeRemoveArgs(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return listWorktreeBranches(), cobra.ShellCompDirectiveNoFileComp
}

func registerGlobalFlagCompletions(root *cobra.Command) {
	_ = root.RegisterFlagCompletionFunc("component", completeComponentsFlag)
	_ = root.RegisterFlagCompletionFunc("svc", completeComponentsFlag)
	_ = root.RegisterFlagCompletionFunc("workspace", completeWorkspacesFlag)
	_ = root.RegisterFlagCompletionFunc("tag", completeTagsFlag)
	_ = root.RegisterFlagCompletionFunc("branch", completeWorktreeBranchesFlag)
	_ = root.RegisterFlagCompletionFunc("source", completeSourceFlag)
	_ = root.RegisterFlagCompletionFunc("mode", completeModesFlag)
}

func registerModeFlagCompletion(command *cobra.Command) {
	if command.Flags().Lookup("mode") != nil {
		_ = command.RegisterFlagCompletionFunc("mode", completeModesFlag)
	}
}
