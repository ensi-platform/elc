package actions

import (
	"errors"
	"os"
	"path"
	"testing"

	"github.com/ensi-platform/elc/core"
	"github.com/golang/mock/gomock"
)

const workspaceConfigWithRepo = `
name: ensi
services:
  test:
    path: "${WORKSPACE_PATH}/apps/test"
    repository: git@github.com:example/test.git
    hooks:
      worktree_create: ${WORKSPACE_PATH}/hooks/after-worktree.sh
      worktree_remove: ${WORKSPACE_PATH}/hooks/before-worktree-remove.sh
  dep1:
    path: "${WORKSPACE_PATH}/apps/dep1"
`

const workspaceConfigWithRepoAndDeps = `
name: ensi
services:
  dep1:
    path: "${WORKSPACE_PATH}/apps/dep1"
  test:
    path: "${WORKSPACE_PATH}/apps/test"
    repository: git@github.com:example/test.git
    dependencies:
      dep1: [default]
`

func worktreeDest(branch string) string {
	return path.Join(fakeWorkspacePath, "worktrees/test", branch)
}

func worktreeCompose(branch string) string {
	return path.Join(worktreeDest(branch), "docker-compose.yml")
}

func TestWorktreeAddLocalBranch(t *testing.T) {
	mockPc := setupMockPc(t)
	expectReadHomeConfig(mockPc)
	expectReadWorkspaceConfig(mockPc, fakeWorkspacePath, workspaceConfigWithRepo, "")

	mainPath := path.Join(fakeWorkspacePath, "apps/test")
	dest := worktreeDest("feat")
	env := gomock.Any()

	mockPc.EXPECT().FileExists(mainPath).Return(true)
	mockPc.EXPECT().FileExists(dest).Return(false)
	mockPc.EXPECT().MkdirAll(path.Join(fakeWorkspacePath, "worktrees/test")).Return(nil)
	mockPc.EXPECT().ExecToString([]string{"git", "-C", mainPath, "fetch", "--prune"}, env).Return(0, "", nil)
	mockPc.EXPECT().ExecToString([]string{"git", "-C", mainPath, "show-ref", "--verify", "--quiet", "refs/heads/feat"}, env).Return(0, "", nil)
	mockPc.EXPECT().ExecInteractive([]string{"git", "-C", mainPath, "worktree", "add", dest, "feat"}, env).Return(0, nil)
	mockPc.EXPECT().ExecInteractive([]string{path.Join(fakeWorkspacePath, "hooks/after-worktree.sh")}, env).Return(0, nil)

	err := AddWorktreeAction(&core.GlobalOptions{}, "feat", nil, false)
	if err != nil {
		t.Fatal(err)
	}
}

func TestWorktreeAddRemoteBranch(t *testing.T) {
	mockPc := setupMockPc(t)
	expectReadHomeConfig(mockPc)
	expectReadWorkspaceConfig(mockPc, fakeWorkspacePath, workspaceConfigWithRepo, "")

	mainPath := path.Join(fakeWorkspacePath, "apps/test")
	dest := worktreeDest("feat")
	env := gomock.Any()
	missing := errors.New("not found")

	mockPc.EXPECT().FileExists(mainPath).Return(true)
	mockPc.EXPECT().FileExists(dest).Return(false)
	mockPc.EXPECT().MkdirAll(path.Join(fakeWorkspacePath, "worktrees/test")).Return(nil)
	mockPc.EXPECT().ExecToString([]string{"git", "-C", mainPath, "fetch", "--prune"}, env).Return(0, "", nil)
	mockPc.EXPECT().ExecToString([]string{"git", "-C", mainPath, "show-ref", "--verify", "--quiet", "refs/heads/feat"}, env).Return(1, "", missing)
	mockPc.EXPECT().ExecToString([]string{"git", "-C", mainPath, "show-ref", "--verify", "--quiet", "refs/remotes/origin/feat"}, env).Return(0, "", nil)
	mockPc.EXPECT().ExecInteractive([]string{"git", "-C", mainPath, "branch", "--track", "feat", "origin/feat"}, env).Return(0, nil)
	mockPc.EXPECT().ExecInteractive([]string{"git", "-C", mainPath, "worktree", "add", dest, "feat"}, env).Return(0, nil)
	mockPc.EXPECT().ExecInteractive([]string{path.Join(fakeWorkspacePath, "hooks/after-worktree.sh")}, env).Return(0, nil)

	err := AddWorktreeAction(&core.GlobalOptions{}, "feat", nil, false)
	if err != nil {
		t.Fatal(err)
	}
}

func TestWorktreeAddNewBranchRequiresSource(t *testing.T) {
	mockPc := setupMockPc(t)
	expectReadHomeConfig(mockPc)
	expectReadWorkspaceConfig(mockPc, fakeWorkspacePath, workspaceConfigWithRepo, "")

	mainPath := path.Join(fakeWorkspacePath, "apps/test")
	dest := worktreeDest("feat")
	env := gomock.Any()
	missing := errors.New("not found")

	mockPc.EXPECT().FileExists(mainPath).Return(true)
	mockPc.EXPECT().FileExists(dest).Return(false)
	mockPc.EXPECT().MkdirAll(path.Join(fakeWorkspacePath, "worktrees/test")).Return(nil)
	mockPc.EXPECT().ExecToString([]string{"git", "-C", mainPath, "fetch", "--prune"}, env).Return(0, "", nil)
	mockPc.EXPECT().ExecToString([]string{"git", "-C", mainPath, "show-ref", "--verify", "--quiet", "refs/heads/feat"}, env).Return(1, "", missing)
	mockPc.EXPECT().ExecToString([]string{"git", "-C", mainPath, "show-ref", "--verify", "--quiet", "refs/remotes/origin/feat"}, env).Return(1, "", missing)

	err := AddWorktreeAction(&core.GlobalOptions{}, "feat", nil, false)
	if err == nil {
		t.Fatal("expected error when branch does not exist and --source is not set")
	}
}

func TestWorktreeAddFetchFailure(t *testing.T) {
	mockPc := setupMockPc(t)
	expectReadHomeConfig(mockPc)
	expectReadWorkspaceConfig(mockPc, fakeWorkspacePath, workspaceConfigWithRepo, "")

	mainPath := path.Join(fakeWorkspacePath, "apps/test")
	dest := worktreeDest("feat")
	env := gomock.Any()

	mockPc.EXPECT().FileExists(mainPath).Return(true)
	mockPc.EXPECT().FileExists(dest).Return(false)
	mockPc.EXPECT().MkdirAll(path.Join(fakeWorkspacePath, "worktrees/test")).Return(nil)
	mockPc.EXPECT().ExecToString([]string{"git", "-C", mainPath, "fetch", "--prune"}, env).Return(1, "", errors.New("network unreachable"))

	err := AddWorktreeAction(&core.GlobalOptions{Source: "HEAD"}, "feat", nil, true)
	if err == nil {
		t.Fatal("expected error when git fetch fails")
	}
}

func TestWorktreeAddNewBranchFromHead(t *testing.T) {
	mockPc := setupMockPc(t)
	expectReadHomeConfig(mockPc)
	expectReadWorkspaceConfig(mockPc, fakeWorkspacePath, workspaceConfigWithRepo, "")

	mainPath := path.Join(fakeWorkspacePath, "apps/test")
	dest := worktreeDest("feat")
	env := gomock.Any()
	missing := errors.New("not found")

	mockPc.EXPECT().FileExists(mainPath).Return(true)
	mockPc.EXPECT().FileExists(dest).Return(false)
	mockPc.EXPECT().MkdirAll(path.Join(fakeWorkspacePath, "worktrees/test")).Return(nil)
	mockPc.EXPECT().ExecToString([]string{"git", "-C", mainPath, "fetch", "--prune"}, env).Return(0, "", nil)
	mockPc.EXPECT().ExecToString([]string{"git", "-C", mainPath, "show-ref", "--verify", "--quiet", "refs/heads/feat"}, env).Return(1, "", missing)
	mockPc.EXPECT().ExecToString([]string{"git", "-C", mainPath, "show-ref", "--verify", "--quiet", "refs/remotes/origin/feat"}, env).Return(1, "", missing)
	mockPc.EXPECT().Printf("branch '%s' not found, creating from %s\n", "feat", "HEAD").Return(0, nil)
	mockPc.EXPECT().ExecInteractive([]string{"git", "-C", mainPath, "worktree", "add", "-b", "feat", dest, "HEAD"}, env).Return(0, nil)
	mockPc.EXPECT().ExecInteractive([]string{path.Join(fakeWorkspacePath, "hooks/after-worktree.sh")}, env).Return(0, nil)

	err := AddWorktreeAction(&core.GlobalOptions{Source: "head"}, "feat", nil, false)
	if err != nil {
		t.Fatal(err)
	}
}

func TestWorktreeAddNewBranchFromSource(t *testing.T) {
	mockPc := setupMockPc(t)
	expectReadHomeConfig(mockPc)
	expectReadWorkspaceConfig(mockPc, fakeWorkspacePath, workspaceConfigWithRepo, "")

	mainPath := path.Join(fakeWorkspacePath, "apps/test")
	dest := worktreeDest("feat")
	env := gomock.Any()
	missing := errors.New("not found")

	mockPc.EXPECT().FileExists(mainPath).Return(true)
	mockPc.EXPECT().FileExists(dest).Return(false)
	mockPc.EXPECT().MkdirAll(path.Join(fakeWorkspacePath, "worktrees/test")).Return(nil)
	mockPc.EXPECT().ExecToString([]string{"git", "-C", mainPath, "fetch", "--prune"}, env).Return(0, "", nil)
	mockPc.EXPECT().ExecToString([]string{"git", "-C", mainPath, "show-ref", "--verify", "--quiet", "refs/heads/feat"}, env).Return(1, "", missing)
	mockPc.EXPECT().ExecToString([]string{"git", "-C", mainPath, "show-ref", "--verify", "--quiet", "refs/remotes/origin/feat"}, env).Return(1, "", missing)
	mockPc.EXPECT().ExecToString([]string{"git", "-C", mainPath, "show-ref", "--verify", "--quiet", "refs/heads/main"}, env).Return(1, "", missing)
	mockPc.EXPECT().ExecToString([]string{"git", "-C", mainPath, "show-ref", "--verify", "--quiet", "refs/remotes/origin/main"}, env).Return(0, "", nil)
	mockPc.EXPECT().Printf("branch '%s' not found, creating from %s\n", "feat", "origin/main").Return(0, nil)
	mockPc.EXPECT().ExecInteractive([]string{"git", "-C", mainPath, "worktree", "add", "-b", "feat", dest, "origin/main"}, env).Return(0, nil)
	mockPc.EXPECT().ExecInteractive([]string{path.Join(fakeWorkspacePath, "hooks/after-worktree.sh")}, env).Return(0, nil)

	err := AddWorktreeAction(&core.GlobalOptions{Source: "main"}, "feat", nil, false)
	if err != nil {
		t.Fatal(err)
	}
}

func TestWorktreeAddHookArgs(t *testing.T) {
	mockPc := setupMockPc(t)
	expectReadHomeConfig(mockPc)
	expectReadWorkspaceConfig(mockPc, fakeWorkspacePath, workspaceConfigWithRepo, "")

	mainPath := path.Join(fakeWorkspacePath, "apps/test")
	dest := worktreeDest("feat")
	env := gomock.Any()

	mockPc.EXPECT().FileExists(mainPath).Return(true)
	mockPc.EXPECT().FileExists(dest).Return(false)
	mockPc.EXPECT().MkdirAll(path.Join(fakeWorkspacePath, "worktrees/test")).Return(nil)
	mockPc.EXPECT().ExecToString([]string{"git", "-C", mainPath, "fetch", "--prune"}, env).Return(0, "", nil)
	mockPc.EXPECT().ExecToString([]string{"git", "-C", mainPath, "show-ref", "--verify", "--quiet", "refs/heads/feat"}, env).Return(0, "", nil)
	mockPc.EXPECT().ExecInteractive([]string{"git", "-C", mainPath, "worktree", "add", dest, "feat"}, env).Return(0, nil)
	mockPc.EXPECT().ExecInteractive([]string{path.Join(fakeWorkspacePath, "hooks/after-worktree.sh"), "--env=staging"}, env).Return(0, nil)

	err := AddWorktreeAction(&core.GlobalOptions{}, "feat", []string{"--env=staging"}, false)
	if err != nil {
		t.Fatal(err)
	}
}

func TestWorktreeAddNoHook(t *testing.T) {
	mockPc := setupMockPc(t)
	expectReadHomeConfig(mockPc)
	expectReadWorkspaceConfig(mockPc, fakeWorkspacePath, workspaceConfigWithRepo, "")

	mainPath := path.Join(fakeWorkspacePath, "apps/test")
	dest := worktreeDest("feat")
	env := gomock.Any()

	mockPc.EXPECT().FileExists(mainPath).Return(true)
	mockPc.EXPECT().FileExists(dest).Return(false)
	mockPc.EXPECT().MkdirAll(path.Join(fakeWorkspacePath, "worktrees/test")).Return(nil)
	mockPc.EXPECT().ExecToString([]string{"git", "-C", mainPath, "fetch", "--prune"}, env).Return(0, "", nil)
	mockPc.EXPECT().ExecToString([]string{"git", "-C", mainPath, "show-ref", "--verify", "--quiet", "refs/heads/feat"}, env).Return(0, "", nil)
	mockPc.EXPECT().ExecInteractive([]string{"git", "-C", mainPath, "worktree", "add", dest, "feat"}, env).Return(0, nil)

	err := AddWorktreeAction(&core.GlobalOptions{}, "feat", nil, true)
	if err != nil {
		t.Fatal(err)
	}
}

func TestWorktreeAddHookArgsWithoutHook(t *testing.T) {
	mockPc := setupMockPc(t)
	expectReadHomeConfig(mockPc)
	expectReadWorkspaceConfig(mockPc, fakeWorkspacePath, workspaceConfigWithRepoAndDeps, "")

	err := AddWorktreeAction(&core.GlobalOptions{}, "feat", []string{"--env=staging"}, false)
	if err == nil {
		t.Fatal("expected error when hook args are passed without hooks.worktree_create")
	}
}

func TestWorktreeAddNotCloned(t *testing.T) {
	mockPc := setupMockPc(t)
	expectReadHomeConfig(mockPc)
	expectReadWorkspaceConfig(mockPc, fakeWorkspacePath, workspaceConfigWithRepo, "")

	mockPc.EXPECT().FileExists(path.Join(fakeWorkspacePath, "apps/test")).Return(false)

	err := AddWorktreeAction(&core.GlobalOptions{}, "feat", nil, false)
	if err == nil {
		t.Fatal("expected error when component is not cloned")
	}
}

func TestWorktreeRemove(t *testing.T) {
	mockPc := setupMockPc(t)
	expectReadHomeConfig(mockPc)
	expectReadWorkspaceConfig(mockPc, fakeWorkspacePath, workspaceConfigWithRepo, "")

	mainPath := path.Join(fakeWorkspacePath, "apps/test")
	dest := worktreeDest("feat")
	composeFile := worktreeCompose("feat")
	env := gomock.Any()

	mockPc.EXPECT().ExecInteractive([]string{path.Join(fakeWorkspacePath, "hooks/before-worktree-remove.sh")}, env).Return(0, nil)
	mockPc.EXPECT().ExecInteractive([]string{"docker", "compose", "-f", composeFile, "down"}, env).Return(0, nil)
	mockPc.EXPECT().ExecInteractive([]string{"git", "-C", mainPath, "worktree", "remove", dest}, env).Return(0, nil)
	mockPc.EXPECT().ExecInteractive([]string{"git", "-C", mainPath, "worktree", "prune"}, env).Return(0, nil)

	err := RemoveWorktreeAction(&core.GlobalOptions{}, "feat", false, false)
	if err != nil {
		t.Fatal(err)
	}
}

func TestWorktreeRemoveNoHook(t *testing.T) {
	mockPc := setupMockPc(t)
	expectReadHomeConfig(mockPc)
	expectReadWorkspaceConfig(mockPc, fakeWorkspacePath, workspaceConfigWithRepo, "")

	mainPath := path.Join(fakeWorkspacePath, "apps/test")
	dest := worktreeDest("feat")
	composeFile := worktreeCompose("feat")
	env := gomock.Any()

	mockPc.EXPECT().ExecInteractive([]string{"docker", "compose", "-f", composeFile, "down"}, env).Return(0, nil)
	mockPc.EXPECT().ExecInteractive([]string{"git", "-C", mainPath, "worktree", "remove", dest}, env).Return(0, nil)
	mockPc.EXPECT().ExecInteractive([]string{"git", "-C", mainPath, "worktree", "prune"}, env).Return(0, nil)

	err := RemoveWorktreeAction(&core.GlobalOptions{}, "feat", false, true)
	if err != nil {
		t.Fatal(err)
	}
}

func TestServiceStartByBranch(t *testing.T) {
	mockPc := setupMockPc(t)
	expectReadHomeConfig(mockPc)
	expectReadWorkspaceConfig(mockPc, fakeWorkspacePath, workspaceConfig, "")

	dest := worktreeDest("feat")
	mockPc.EXPECT().FileExists(dest).Return(true)
	expectStartService(mockPc, worktreeCompose("feat"))

	err := StartServiceAction(&core.GlobalOptions{Branch: "feat"}, []string{})
	if err != nil {
		t.Fatal(err)
	}
}

func TestServiceStartCreatesMissingWorktreeRunsHook(t *testing.T) {
	mockPc := setupMockPc(t)
	expectReadHomeConfig(mockPc)
	expectReadWorkspaceConfig(mockPc, fakeWorkspacePath, workspaceConfigWithRepo, "")

	mainPath := path.Join(fakeWorkspacePath, "apps/test")
	dest := worktreeDest("feat")
	env := gomock.Any()

	// resolveComponents IsCloned + EnsureWorktree IsCloned + AddWorktree exists check
	mockPc.EXPECT().FileExists(dest).Return(false)
	mockPc.EXPECT().FileExists(dest).Return(false)
	mockPc.EXPECT().FileExists(mainPath).Return(true)
	mockPc.EXPECT().FileExists(dest).Return(false)
	mockPc.EXPECT().MkdirAll(path.Join(fakeWorkspacePath, "worktrees/test")).Return(nil)
	mockPc.EXPECT().ExecToString([]string{"git", "-C", mainPath, "fetch", "--prune"}, env).Return(0, "", nil)
	mockPc.EXPECT().ExecToString([]string{"git", "-C", mainPath, "show-ref", "--verify", "--quiet", "refs/heads/feat"}, env).Return(0, "", nil)
	mockPc.EXPECT().ExecInteractive([]string{"git", "-C", mainPath, "worktree", "add", dest, "feat"}, env).Return(0, nil)
	mockPc.EXPECT().ExecInteractive([]string{path.Join(fakeWorkspacePath, "hooks/after-worktree.sh")}, env).Return(0, nil)
	expectStartService(mockPc, worktreeCompose("feat"))

	err := StartServiceAction(&core.GlobalOptions{Branch: "feat"}, []string{})
	if err != nil {
		t.Fatal(err)
	}
}

func TestServiceStartFromWorktreeCwd(t *testing.T) {
	mockPc := setupMockPc(t)
	expectReadHomeConfig(mockPc)
	cwd := worktreeDest("feat")
	expectReadWorkspaceConfigCwd(mockPc, fakeWorkspacePath, cwd, workspaceConfig, "")

	mockPc.EXPECT().FileExists(path.Join(cwd, ".git")).Return(true)
	mockPc.EXPECT().FileExists(cwd).Return(true)
	expectStartService(mockPc, worktreeCompose("feat"))

	err := StartServiceAction(&core.GlobalOptions{}, []string{})
	if err != nil {
		t.Fatal(err)
	}
}

func TestServiceStartFromWorktreeSubdir(t *testing.T) {
	mockPc := setupMockPc(t)
	expectReadHomeConfig(mockPc)
	cwd := path.Join(worktreeDest("feat"), "src")
	expectReadWorkspaceConfigCwd(mockPc, fakeWorkspacePath, cwd, workspaceConfig, "")

	mockPc.EXPECT().FileExists(path.Join(cwd, ".git")).Return(false)
	mockPc.EXPECT().FileExists(path.Join(worktreeDest("feat"), ".git")).Return(true)
	mockPc.EXPECT().FileExists(worktreeDest("feat")).Return(true)
	expectStartService(mockPc, worktreeCompose("feat"))

	err := StartServiceAction(&core.GlobalOptions{}, []string{})
	if err != nil {
		t.Fatal(err)
	}
}

func TestServiceStartFromMainSubdir(t *testing.T) {
	mockPc := setupMockPc(t)
	expectReadHomeConfig(mockPc)
	cwd := path.Join(fakeWorkspacePath, "apps/test/src")
	expectReadWorkspaceConfigCwd(mockPc, fakeWorkspacePath, cwd, workspaceConfig, "")
	expectStartService(mockPc, path.Join(fakeWorkspacePath, "apps/test/docker-compose.yml"))

	err := StartServiceAction(&core.GlobalOptions{}, []string{})
	if err != nil {
		t.Fatal(err)
	}
}

func TestServiceStartOtherComponentFromWorktree(t *testing.T) {
	mockPc := setupMockPc(t)
	expectReadHomeConfig(mockPc)
	expectReadWorkspaceConfigCwd(mockPc, fakeWorkspacePath, worktreeDest("feat"), workspaceConfigWithRepo, "")

	mockPc.EXPECT().FileExists(path.Join(worktreeDest("feat"), ".git")).Return(true)
	expectStartService(mockPc, path.Join(fakeWorkspacePath, "apps/dep1/docker-compose.yml"))

	err := StartServiceAction(&core.GlobalOptions{}, []string{"dep1"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestServiceStartWorktreeDependenciesUseMain(t *testing.T) {
	mockPc := setupMockPc(t)
	expectReadHomeConfig(mockPc)
	expectReadWorkspaceConfig(mockPc, fakeWorkspacePath, workspaceConfigWithRepoAndDeps, "")

	mockPc.EXPECT().FileExists(worktreeDest("feat")).Return(true)
	expectStartService(mockPc, path.Join(fakeWorkspacePath, "apps/dep1/docker-compose.yml"))
	expectStartService(mockPc, worktreeCompose("feat"))

	err := StartServiceAction(&core.GlobalOptions{Branch: "feat"}, []string{"test"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestServiceVarsWorktree(t *testing.T) {
	mockPc := setupMockPc(t)
	expectReadHomeConfig(mockPc)
	expectReadWorkspaceConfig(mockPc, fakeWorkspacePath, workspaceConfig, "")

	dest := worktreeDest("feat")
	mockPc.EXPECT().FileExists(dest).Return(true)

	mockPc.EXPECT().Println("WORKSPACE_PATH=/tmp/workspaces/project1")
	mockPc.EXPECT().Println("WORKSPACE_NAME=ensi")
	mockPc.EXPECT().Println("WORKTREES_PATH=/tmp/workspaces/project1/worktrees")
	mockPc.EXPECT().Println("SVC_PATH=/tmp/workspaces/project1/worktrees/test/feat")
	mockPc.EXPECT().Println("COMPOSE_FILE=/tmp/workspaces/project1/worktrees/test/feat/docker-compose.yml")
	mockPc.EXPECT().Println("GIT_BRANCH=feat")
	mockPc.EXPECT().Println("APP_NAME=test-feat")
	mockPc.EXPECT().Println("COMPOSE_PROJECT_NAME=ensi-test-feat")

	err := PrintVarsAction(&core.GlobalOptions{Branch: "feat"}, []string{})
	if err != nil {
		t.Fatal(err)
	}
}

func TestWorktreeCustomPath(t *testing.T) {
	mockPc := setupMockPc(t)
	expectReadHomeConfig(mockPc)

	config := `
name: ensi
variables:
  WORKTREES_PATH: ${WORKSPACE_PATH}/apps/../wt
services:
  test:
    path: "${WORKSPACE_PATH}/apps/test"
    repository: git@github.com:example/test.git
`
	expectReadWorkspaceConfig(mockPc, fakeWorkspacePath, config, "")

	mainPath := path.Join(fakeWorkspacePath, "apps/test")
	dest := path.Join(fakeWorkspacePath, "wt/test/feat")
	env := gomock.Any()

	mockPc.EXPECT().FileExists(mainPath).Return(true)
	mockPc.EXPECT().FileExists(dest).Return(false)
	mockPc.EXPECT().MkdirAll(path.Join(fakeWorkspacePath, "wt/test")).Return(nil)
	mockPc.EXPECT().ExecToString([]string{"git", "-C", mainPath, "fetch", "--prune"}, env).Return(0, "", nil)
	mockPc.EXPECT().ExecToString([]string{"git", "-C", mainPath, "show-ref", "--verify", "--quiet", "refs/heads/feat"}, env).Return(0, "", nil)
	mockPc.EXPECT().ExecInteractive([]string{"git", "-C", mainPath, "worktree", "add", dest, "feat"}, env).Return(0, nil)

	err := AddWorktreeAction(&core.GlobalOptions{}, "feat", nil, true)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLaunchMainClone(t *testing.T) {
	mockPc := setupMockPc(t)
	expectReadHomeConfig(mockPc)
	expectReadWorkspaceConfig(mockPc, fakeWorkspacePath, workspaceConfigWithRepo, "")

	mainPath := path.Join(fakeWorkspacePath, "apps/test")
	mockPc.EXPECT().ExecInteractiveInDir([]string{"code", "."}, gomock.Any(), mainPath).Return(0, nil)

	err := LaunchAction(&core.GlobalOptions{}, []string{"code", "."})
	if err != nil {
		t.Fatal(err)
	}
}

func TestLaunchCreatesWorktree(t *testing.T) {
	mockPc := setupMockPc(t)
	expectReadHomeConfig(mockPc)
	expectReadWorkspaceConfig(mockPc, fakeWorkspacePath, workspaceConfigWithRepo, "")

	mainPath := path.Join(fakeWorkspacePath, "apps/test")
	dest := worktreeDest("feat")
	env := gomock.Any()

	mockPc.EXPECT().FileExists(dest).Return(false)
	mockPc.EXPECT().FileExists(mainPath).Return(true)
	mockPc.EXPECT().FileExists(dest).Return(false)
	mockPc.EXPECT().MkdirAll(path.Join(fakeWorkspacePath, "worktrees/test")).Return(nil)
	mockPc.EXPECT().ExecToString([]string{"git", "-C", mainPath, "fetch", "--prune"}, env).Return(0, "", nil)
	mockPc.EXPECT().ExecToString([]string{"git", "-C", mainPath, "show-ref", "--verify", "--quiet", "refs/heads/feat"}, env).Return(0, "", nil)
	mockPc.EXPECT().ExecInteractive([]string{"git", "-C", mainPath, "worktree", "add", dest, "feat"}, env).Return(0, nil)
	mockPc.EXPECT().ExecInteractive([]string{path.Join(fakeWorkspacePath, "hooks/after-worktree.sh")}, env).Return(0, nil)
	mockPc.EXPECT().ExecInteractiveInDir([]string{"code", "."}, env, dest).Return(0, nil)

	err := LaunchAction(&core.GlobalOptions{Branch: "feat"}, []string{"code", "."})
	if err != nil {
		t.Fatal(err)
	}
}

func TestLaunchExistingWorktree(t *testing.T) {
	mockPc := setupMockPc(t)
	expectReadHomeConfig(mockPc)
	expectReadWorkspaceConfig(mockPc, fakeWorkspacePath, workspaceConfigWithRepo, "")

	dest := worktreeDest("feat")
	mockPc.EXPECT().FileExists(dest).Return(true)
	mockPc.EXPECT().ExecInteractiveInDir([]string{"code", "."}, gomock.Any(), dest).Return(0, nil)

	err := LaunchAction(&core.GlobalOptions{Branch: "feat"}, []string{"code", "."})
	if err != nil {
		t.Fatal(err)
	}
}

func TestListWorktreesEmpty(t *testing.T) {
	mockPc := setupMockPc(t)
	expectReadHomeConfig(mockPc)
	expectReadWorkspaceConfig(mockPc, fakeWorkspacePath, workspaceConfigWithRepo, "")

	mockPc.EXPECT().FileExists(path.Join(fakeWorkspacePath, "worktrees")).Return(false)

	err := ListWorktreesAction(&core.GlobalOptions{})
	if err != nil {
		t.Fatal(err)
	}
}

func TestListWorktrees(t *testing.T) {
	mockPc := setupMockPc(t)
	expectReadHomeConfig(mockPc)
	expectReadWorkspaceConfig(mockPc, fakeWorkspacePath, workspaceConfigWithRepo, "")

	root := path.Join(fakeWorkspacePath, "worktrees")
	testRoot := path.Join(root, "test")
	featRoot := path.Join(testRoot, "feature")
	fooRoot := path.Join(featRoot, "foo")
	mainRoot := path.Join(testRoot, "main")

	mockPc.EXPECT().FileExists(root).Return(true)
	mockPc.EXPECT().ReadDir(root).Return([]os.FileInfo{core.DirInfo("test"), core.DirInfo("unknown")}, nil)

	mockPc.EXPECT().FileExists(path.Join(testRoot, ".git")).Return(false)
	mockPc.EXPECT().ReadDir(testRoot).Return([]os.FileInfo{core.DirInfo("feature"), core.DirInfo("main")}, nil)

	mockPc.EXPECT().FileExists(path.Join(featRoot, ".git")).Return(false)
	mockPc.EXPECT().ReadDir(featRoot).Return([]os.FileInfo{core.DirInfo("foo")}, nil)
	mockPc.EXPECT().FileExists(path.Join(fooRoot, ".git")).Return(true)

	mockPc.EXPECT().FileExists(path.Join(mainRoot, ".git")).Return(true)

	mockPc.EXPECT().Println("test\tfeature/foo\t" + fooRoot)
	mockPc.EXPECT().Println("test\tmain\t" + mainRoot)

	err := ListWorktreesAction(&core.GlobalOptions{})
	if err != nil {
		t.Fatal(err)
	}
}

func TestListWorktreesFilterComponent(t *testing.T) {
	mockPc := setupMockPc(t)
	expectReadHomeConfig(mockPc)
	expectReadWorkspaceConfig(mockPc, fakeWorkspacePath, workspaceConfigWithRepo, "")

	root := path.Join(fakeWorkspacePath, "worktrees")
	testRoot := path.Join(root, "test")
	featRoot := path.Join(testRoot, "feat")

	mockPc.EXPECT().FileExists(root).Return(true)
	mockPc.EXPECT().ReadDir(root).Return([]os.FileInfo{core.DirInfo("test"), core.DirInfo("dep1")}, nil)

	mockPc.EXPECT().FileExists(path.Join(testRoot, ".git")).Return(false)
	mockPc.EXPECT().ReadDir(testRoot).Return([]os.FileInfo{core.DirInfo("feat")}, nil)
	mockPc.EXPECT().FileExists(path.Join(featRoot, ".git")).Return(true)

	mockPc.EXPECT().Println("test\tfeat\t" + featRoot)

	err := ListWorktreesAction(&core.GlobalOptions{ComponentName: "test"})
	if err != nil {
		t.Fatal(err)
	}
}
