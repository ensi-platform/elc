package src

import (
	"github.com/golang/mock/gomock"
	"os"
	"path"
	"testing"
)

const fakeHomeConfigPath = "/tmp/home/.elc.yaml"
const fakeWorkspacePath = "/tmp/workspaces/project1"

const baseHomeConfig = `
current_workspace: project1
update_command: update
workspaces:
- name: project1
  path: /tmp/workspaces/project1
- name: project2
  path: /tmp/workspaces/project2
`

func expectReadHomeConfig(mockPC *MockPC) {
	mockPC.EXPECT().FileExists(fakeHomeConfigPath).Return(true)
	mockPC.EXPECT().ReadFile(fakeHomeConfigPath).Return([]byte(baseHomeConfig), nil)
}

func TestWorkspaceShow(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPC := NewMockPC(ctrl)
	Pc = mockPC

	expectReadHomeConfig(mockPC)

	mockPC.EXPECT().Println("project1")

	_ = CmdWorkspaceShow(fakeHomeConfigPath, []string{})
}

func TestWorkspaceList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPC := NewMockPC(ctrl)
	Pc = mockPC

	expectReadHomeConfig(mockPC)

	mockPC.EXPECT().Printf("%-10s %s\n", "project1", "/tmp/workspaces/project1")
	mockPC.EXPECT().Printf("%-10s %s\n", "project2", "/tmp/workspaces/project2")

	_ = CmdWorkspaceList(fakeHomeConfigPath, []string{})
}

const homeConfigForAdd = `current_workspace: project1
update_command: update
workspaces:
- name: project1
  path: /tmp/workspaces/project1
- name: project2
  path: /tmp/workspaces/project2
- name: project3
  path: /tmp/workspaces/project3
`

func TestWorkspaceAdd(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPC := NewMockPC(ctrl)
	Pc = mockPC

	expectReadHomeConfig(mockPC)
	mockPC.EXPECT().WriteFile(fakeHomeConfigPath, []byte(homeConfigForAdd), os.FileMode(0644))
	mockPC.EXPECT().Printf("workspace '%s' is added\n", "project3")

	_ = CmdWorkspaceAdd(fakeHomeConfigPath, []string{"project3", "/tmp/workspaces/project3"})
}

const homeConfigForSelect = `current_workspace: project2
update_command: update
workspaces:
- name: project1
  path: /tmp/workspaces/project1
- name: project2
  path: /tmp/workspaces/project2
`

func TestWorkspaceSelect(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPC := NewMockPC(ctrl)
	Pc = mockPC

	expectReadHomeConfig(mockPC)
	mockPC.EXPECT().WriteFile(fakeHomeConfigPath, []byte(homeConfigForSelect), os.FileMode(0644))
	mockPC.EXPECT().Printf("active workspace changed to '%s'\n", "project2")

	_ = CmdWorkspaceSelect(fakeHomeConfigPath, []string{"project2"})
}

const workspaceConfig = `
name: ensi
services:
  test:
    path: "${WORKSPACE_PATH}/apps/test"
`

func expectReadWorkspaceConfig(mockPC *MockPC, workspacePath string, config string, env string) {
	configPath := path.Join(workspacePath, "workspace.yaml")
	envPath := path.Join(workspacePath, "env.yaml")
	mockPC.EXPECT().Getwd().
		Return(path.Join(workspacePath, "apps/test"), nil)
	mockPC.EXPECT().ReadFile(configPath).
		Return([]byte(config), nil)

	envExists := env != ""
	mockPC.EXPECT().FileExists(envPath).
		Return(envExists)
	if envExists {
		mockPC.EXPECT().ReadFile(envPath).
			Return([]byte(env), nil)
	}
}

func TestServiceStart(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPC := NewMockPC(ctrl)
	Pc = mockPC

	expectReadHomeConfig(mockPC)
	expectReadWorkspaceConfig(mockPC, fakeWorkspacePath, workspaceConfig, "")

	composeFilePath := path.Join(fakeWorkspacePath, "apps/test/docker-compose.yml")

	mockPC.EXPECT().
		ExecToString([]string{"docker", "compose", "-f", composeFilePath, "ps", "--status=running", "-q"}, gomock.Any()).
		Return(0, "", nil)

	mockPC.EXPECT().
		ExecInteractive([]string{"docker", "compose", "-f", composeFilePath, "up", "-d"}, gomock.Any()).
		Return(0, nil)

	_ = CmdServiceStart(fakeHomeConfigPath, []string{})
}

const workspaceConfigWithDeps = `
name: ensi
services:
  dep1:
    path: "${WORKSPACE_PATH}/apps/dep1"
  dep2:
    path: "${WORKSPACE_PATH}/apps/dep2"
  dep3:
    path: "${WORKSPACE_PATH}/apps/dep3"
  test:
    path: "${WORKSPACE_PATH}/apps/test"
    dependencies:
      dep1: [default]
      dep2: [default, hook]
      dep3: []
`

func expectStartService(mockPC *MockPC, composeFilePath string) {
	mockPC.EXPECT().
		ExecToString([]string{"docker", "compose", "-f", composeFilePath, "ps", "--status=running", "-q"}, gomock.Any()).
		Return(0, "", nil)

	mockPC.EXPECT().
		ExecInteractive([]string{"docker", "compose", "-f", composeFilePath, "up", "-d"}, gomock.Any()).
		Return(0, nil)
}

func expectStopService(mockPC *MockPC, composeFilePath string) {
	mockPC.EXPECT().
		ExecToString([]string{"docker", "compose", "-f", composeFilePath, "ps", "--status=running", "-q"}, gomock.Any()).
		Return(0, "asdasd", nil)

	mockPC.EXPECT().
		ExecInteractive([]string{"docker", "compose", "-f", composeFilePath, "stop"}, gomock.Any()).
		Return(0, nil)
}

func expectDestroyService(mockPC *MockPC, composeFilePath string) {
	mockPC.EXPECT().
		ExecToString([]string{"docker", "compose", "-f", composeFilePath, "ps", "--status=running", "-q"}, gomock.Any()).
		Return(0, "asdasd", nil)

	mockPC.EXPECT().
		ExecInteractive([]string{"docker", "compose", "-f", composeFilePath, "down"}, gomock.Any()).
		Return(0, nil)
}

func TestServiceStartWithDeps(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPC := NewMockPC(ctrl)
	Pc = mockPC

	// default mode
	expectReadHomeConfig(mockPC)
	expectReadWorkspaceConfig(mockPC, fakeWorkspacePath, workspaceConfigWithDeps, "")

	expectStartService(mockPC, path.Join(fakeWorkspacePath, "apps/dep1/docker-compose.yml"))
	expectStartService(mockPC, path.Join(fakeWorkspacePath, "apps/dep2/docker-compose.yml"))
	expectStartService(mockPC, path.Join(fakeWorkspacePath, "apps/test/docker-compose.yml"))

	_ = CmdServiceStart(fakeHomeConfigPath, []string{})

	// hook mode
	expectReadHomeConfig(mockPC)
	expectReadWorkspaceConfig(mockPC, fakeWorkspacePath, workspaceConfigWithDeps, "")

	expectStartService(mockPC, path.Join(fakeWorkspacePath, "apps/dep2/docker-compose.yml"))
	expectStartService(mockPC, path.Join(fakeWorkspacePath, "apps/test/docker-compose.yml"))

	_ = CmdServiceStart(fakeHomeConfigPath, []string{"--mode=hook"})

	// single mode
	expectReadHomeConfig(mockPC)
	expectReadWorkspaceConfig(mockPC, fakeWorkspacePath, workspaceConfigWithDeps, "")

	expectStartService(mockPC, path.Join(fakeWorkspacePath, "apps/test/docker-compose.yml"))

	_ = CmdServiceStart(fakeHomeConfigPath, []string{"--mode="})

	// by name
	expectReadHomeConfig(mockPC)
	expectReadWorkspaceConfig(mockPC, fakeWorkspacePath, workspaceConfigWithDeps, "")

	expectStartService(mockPC, path.Join(fakeWorkspacePath, "apps/dep3/docker-compose.yml"))

	_ = CmdServiceStart(fakeHomeConfigPath, []string{"dep3"})

	// by names
	expectReadHomeConfig(mockPC)
	expectReadWorkspaceConfig(mockPC, fakeWorkspacePath, workspaceConfigWithDeps, "")

	expectStartService(mockPC, path.Join(fakeWorkspacePath, "apps/dep3/docker-compose.yml"))
	expectStartService(mockPC, path.Join(fakeWorkspacePath, "apps/dep1/docker-compose.yml"))

	_ = CmdServiceStart(fakeHomeConfigPath, []string{"dep3", "dep1"})
}

func TestServiceStop(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPC := NewMockPC(ctrl)
	Pc = mockPC

	// current
	expectReadHomeConfig(mockPC)
	expectReadWorkspaceConfig(mockPC, fakeWorkspacePath, workspaceConfigWithDeps, "")

	expectStopService(mockPC, path.Join(fakeWorkspacePath, "apps/test/docker-compose.yml"))

	_ = CmdServiceStop(fakeHomeConfigPath, []string{})

	// by name
	expectReadHomeConfig(mockPC)
	expectReadWorkspaceConfig(mockPC, fakeWorkspacePath, workspaceConfigWithDeps, "")

	expectStopService(mockPC, path.Join(fakeWorkspacePath, "apps/dep1/docker-compose.yml"))

	_ = CmdServiceStop(fakeHomeConfigPath, []string{"dep1"})

	// by names
	expectReadHomeConfig(mockPC)
	expectReadWorkspaceConfig(mockPC, fakeWorkspacePath, workspaceConfigWithDeps, "")

	expectStopService(mockPC, path.Join(fakeWorkspacePath, "apps/dep1/docker-compose.yml"))
	expectStopService(mockPC, path.Join(fakeWorkspacePath, "apps/dep2/docker-compose.yml"))

	_ = CmdServiceStop(fakeHomeConfigPath, []string{"dep1", "dep2"})

	// all
	expectReadHomeConfig(mockPC)
	expectReadWorkspaceConfig(mockPC, fakeWorkspacePath, workspaceConfigWithDeps, "")

	expectStopService(mockPC, path.Join(fakeWorkspacePath, "apps/dep1/docker-compose.yml"))
	expectStopService(mockPC, path.Join(fakeWorkspacePath, "apps/dep2/docker-compose.yml"))
	expectStopService(mockPC, path.Join(fakeWorkspacePath, "apps/dep3/docker-compose.yml"))
	expectStopService(mockPC, path.Join(fakeWorkspacePath, "apps/test/docker-compose.yml"))

	_ = CmdServiceStop(fakeHomeConfigPath, []string{"--all"})
}

func TestServiceDestroy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPC := NewMockPC(ctrl)
	Pc = mockPC

	// current
	expectReadHomeConfig(mockPC)
	expectReadWorkspaceConfig(mockPC, fakeWorkspacePath, workspaceConfigWithDeps, "")

	expectDestroyService(mockPC, path.Join(fakeWorkspacePath, "apps/test/docker-compose.yml"))

	_ = CmdServiceDestroy(fakeHomeConfigPath, []string{})

	// by name
	expectReadHomeConfig(mockPC)
	expectReadWorkspaceConfig(mockPC, fakeWorkspacePath, workspaceConfigWithDeps, "")

	expectDestroyService(mockPC, path.Join(fakeWorkspacePath, "apps/dep1/docker-compose.yml"))

	_ = CmdServiceDestroy(fakeHomeConfigPath, []string{"dep1"})

	// by names
	expectReadHomeConfig(mockPC)
	expectReadWorkspaceConfig(mockPC, fakeWorkspacePath, workspaceConfigWithDeps, "")

	expectDestroyService(mockPC, path.Join(fakeWorkspacePath, "apps/dep1/docker-compose.yml"))
	expectDestroyService(mockPC, path.Join(fakeWorkspacePath, "apps/dep2/docker-compose.yml"))

	_ = CmdServiceDestroy(fakeHomeConfigPath, []string{"dep1", "dep2"})

	// all
	expectReadHomeConfig(mockPC)
	expectReadWorkspaceConfig(mockPC, fakeWorkspacePath, workspaceConfigWithDeps, "")

	expectDestroyService(mockPC, path.Join(fakeWorkspacePath, "apps/dep1/docker-compose.yml"))
	expectDestroyService(mockPC, path.Join(fakeWorkspacePath, "apps/dep2/docker-compose.yml"))
	expectDestroyService(mockPC, path.Join(fakeWorkspacePath, "apps/dep3/docker-compose.yml"))
	expectDestroyService(mockPC, path.Join(fakeWorkspacePath, "apps/test/docker-compose.yml"))

	_ = CmdServiceDestroy(fakeHomeConfigPath, []string{"--all"})
}

func TestServiceRestart(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPC := NewMockPC(ctrl)
	Pc = mockPC

	// current
	expectReadHomeConfig(mockPC)
	expectReadWorkspaceConfig(mockPC, fakeWorkspacePath, workspaceConfigWithDeps, "")

	expectStopService(mockPC, path.Join(fakeWorkspacePath, "apps/test/docker-compose.yml"))
	expectStartService(mockPC, path.Join(fakeWorkspacePath, "apps/test/docker-compose.yml"))

	_ = CmdServiceRestart(fakeHomeConfigPath, []string{})

	// by name
	expectReadHomeConfig(mockPC)
	expectReadWorkspaceConfig(mockPC, fakeWorkspacePath, workspaceConfigWithDeps, "")

	expectStopService(mockPC, path.Join(fakeWorkspacePath, "apps/dep1/docker-compose.yml"))
	expectStartService(mockPC, path.Join(fakeWorkspacePath, "apps/dep1/docker-compose.yml"))

	_ = CmdServiceRestart(fakeHomeConfigPath, []string{"dep1"})

	// hard
	expectReadHomeConfig(mockPC)
	expectReadWorkspaceConfig(mockPC, fakeWorkspacePath, workspaceConfigWithDeps, "")

	expectDestroyService(mockPC, path.Join(fakeWorkspacePath, "apps/test/docker-compose.yml"))
	expectStartService(mockPC, path.Join(fakeWorkspacePath, "apps/test/docker-compose.yml"))

	_ = CmdServiceRestart(fakeHomeConfigPath, []string{"--hard"})
}

func TestServiceCompose(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPC := NewMockPC(ctrl)
	Pc = mockPC

	// current
	expectReadHomeConfig(mockPC)
	expectReadWorkspaceConfig(mockPC, fakeWorkspacePath, workspaceConfigWithDeps, "")

	mockPC.EXPECT().
		ExecInteractive([]string{"docker", "compose", "-f", path.Join(fakeWorkspacePath, "apps/test/docker-compose.yml"), "some", "command"}, gomock.Any()).
		Return(0, nil)

	_, _ = CmdServiceCompose(fakeHomeConfigPath, []string{"some", "command"})

	// by name
	expectReadHomeConfig(mockPC)
	expectReadWorkspaceConfig(mockPC, fakeWorkspacePath, workspaceConfigWithDeps, "")

	mockPC.EXPECT().
		ExecInteractive([]string{"docker", "compose", "-f", path.Join(fakeWorkspacePath, "apps/dep1/docker-compose.yml"), "some", "command"}, gomock.Any()).
		Return(0, nil)

	_, _ = CmdServiceCompose(fakeHomeConfigPath, []string{"--svc=dep1", "some", "command"})
}

func TestServiceExec(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPC := NewMockPC(ctrl)
	Pc = mockPC

	// simple
	mockPC.EXPECT().Getuid().Return(1000)
	expectReadHomeConfig(mockPC)
	expectReadWorkspaceConfig(mockPC, fakeWorkspacePath, workspaceConfig, "")

	expectStartService(mockPC, path.Join(fakeWorkspacePath, "apps/test/docker-compose.yml"))
	mockPC.EXPECT().
		IsTerminal().
		Return(true)
	mockPC.EXPECT().
		ExecInteractive([]string{"docker", "compose", "-f", path.Join(fakeWorkspacePath, "apps/test/docker-compose.yml"), "exec", "-u", "1000", "app", "some", "command"}, gomock.Any()).
		Return(0, nil)

	_, _ = CmdServiceExec(fakeHomeConfigPath, []string{"some", "command"})

	// without tty
	mockPC.EXPECT().Getuid().Return(1000)
	expectReadHomeConfig(mockPC)
	expectReadWorkspaceConfig(mockPC, fakeWorkspacePath, workspaceConfig, "")

	expectStartService(mockPC, path.Join(fakeWorkspacePath, "apps/test/docker-compose.yml"))
	mockPC.EXPECT().
		IsTerminal().
		Return(false)
	mockPC.EXPECT().
		ExecInteractive([]string{"docker", "compose", "-f", path.Join(fakeWorkspacePath, "apps/test/docker-compose.yml"), "exec", "-u", "1000", "-T", "app", "some", "command"}, gomock.Any()).
		Return(0, nil)

	_, _ = CmdServiceExec(fakeHomeConfigPath, []string{"some", "command"})

	// with uid
	mockPC.EXPECT().Getuid().Return(1000)
	expectReadHomeConfig(mockPC)
	expectReadWorkspaceConfig(mockPC, fakeWorkspacePath, workspaceConfig, "")

	expectStartService(mockPC, path.Join(fakeWorkspacePath, "apps/test/docker-compose.yml"))
	mockPC.EXPECT().
		IsTerminal().
		Return(true)
	mockPC.EXPECT().
		ExecInteractive([]string{"docker", "compose", "-f", path.Join(fakeWorkspacePath, "apps/test/docker-compose.yml"), "exec", "-u", "0", "app", "some", "command"}, gomock.Any()).
		Return(0, nil)

	_, _ = CmdServiceExec(fakeHomeConfigPath, []string{"--uid=0", "some", "command"})
}

const workspaceConfigWithVars = `
name: ensi
variables:
  V_GL: vglobal
  V_GL_SIMPLE_VAR: ${V_GL}-a
  V_GL_WITH_DEFAULT: ${UNDEFINED:-default}
  V_GL_WITH_DEFAULT_VAR: ${UNDEFINED:-$V_GL}
services:
  test:
    path: "${WORKSPACE_PATH}/apps/test"
    variables:
      V_IN_SVC: vinsvc
  test1:
    path: "${WORKSPACE_PATH}/apps/test1"
    extends: tpl1
    variables:
      V_IN_SVC: vinsvc

templates:
  tpl1:
    path: "${WORKSPACE_PATH}/templates/tpl1"
    variables:
      V_IN_TPL: vintpl
`

func TestServiceVars(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPC := NewMockPC(ctrl)
	Pc = mockPC

	// simple
	expectReadHomeConfig(mockPC)
	expectReadWorkspaceConfig(mockPC, fakeWorkspacePath, workspaceConfigWithVars, "")

	mockPC.EXPECT().Println("WORKSPACE_PATH=/tmp/workspaces/project1")
	mockPC.EXPECT().Println("WORKSPACE_NAME=ensi")

	mockPC.EXPECT().Println("V_GL=vglobal")
	mockPC.EXPECT().Println("V_GL_SIMPLE_VAR=vglobal-a")
	mockPC.EXPECT().Println("V_GL_WITH_DEFAULT=default")
	mockPC.EXPECT().Println("V_GL_WITH_DEFAULT_VAR=vglobal")

	mockPC.EXPECT().Println("APP_NAME=test")
	mockPC.EXPECT().Println("COMPOSE_PROJECT_NAME=ensi-test")
	mockPC.EXPECT().Println("SVC_PATH=/tmp/workspaces/project1/apps/test")
	mockPC.EXPECT().Println("COMPOSE_FILE=/tmp/workspaces/project1/apps/test/docker-compose.yml")

	mockPC.EXPECT().Println("V_IN_SVC=vinsvc")

	_ = CmdServiceVars(fakeHomeConfigPath, []string{})

	// simple
	expectReadHomeConfig(mockPC)
	expectReadWorkspaceConfig(mockPC, fakeWorkspacePath, workspaceConfigWithVars, "")

	mockPC.EXPECT().Println("WORKSPACE_PATH=/tmp/workspaces/project1")
	mockPC.EXPECT().Println("WORKSPACE_NAME=ensi")

	mockPC.EXPECT().Println("V_GL=vglobal")
	mockPC.EXPECT().Println("V_GL_SIMPLE_VAR=vglobal-a")
	mockPC.EXPECT().Println("V_GL_WITH_DEFAULT=default")
	mockPC.EXPECT().Println("V_GL_WITH_DEFAULT_VAR=vglobal")

	mockPC.EXPECT().Println("V_IN_TPL=vintpl")

	mockPC.EXPECT().Println("TPL_PATH=/tmp/workspaces/project1/templates/tpl1")
	mockPC.EXPECT().Println("COMPOSE_FILE=/tmp/workspaces/project1/templates/tpl1/docker-compose.yml")
	mockPC.EXPECT().Println("APP_NAME=test1")
	mockPC.EXPECT().Println("COMPOSE_PROJECT_NAME=ensi-test1")
	mockPC.EXPECT().Println("SVC_PATH=/tmp/workspaces/project1/apps/test1")

	mockPC.EXPECT().Println("V_IN_SVC=vinsvc")

	_ = CmdServiceVars(fakeHomeConfigPath, []string{"test1"})
}
