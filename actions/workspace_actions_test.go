package actions

import (
	"github.com/madridianfox/elc/core"
	"os"
	"testing"
)

func TestWorkspaceShow(t *testing.T) {
	mockPc := setupMockPc(t)
	expectReadHomeConfig(mockPc)

	mockPc.EXPECT().Println("project1")

	_ = ShowCurrentWorkspaceAction(&core.GlobalOptions{})
}

func TestWorkspaceList(t *testing.T) {
	mockPc := setupMockPc(t)
	expectReadHomeConfig(mockPc)

	mockPc.EXPECT().Printf("%-10s %s\n", "project1", "/tmp/workspaces/project1")
	mockPc.EXPECT().Printf("%-10s %s\n", "project2", "/tmp/workspaces/project2")

	_ = ListWorkspacesAction()
}

func TestWorkspaceAdd(t *testing.T) {
	mockPc := setupMockPc(t)
	expectReadHomeConfig(mockPc)

	const homeConfigForAdd = `current_workspace: project1
update_command: update
workspaces:
- name: project1
  path: /tmp/workspaces/project1
  root_path: ""
- name: project2
  path: /tmp/workspaces/project2
  root_path: ""
- name: project3
  path: /tmp/workspaces/project3
  root_path: ""
`

	mockPc.EXPECT().WriteFile(fakeHomeConfigPath, []byte(homeConfigForAdd), os.FileMode(0644))
	mockPc.EXPECT().Printf("workspace '%s' is added\n", "project3")

	_ = AddWorkspaceAction("project3", "/tmp/workspaces/project3")
}

func TestWorkspaceSelect(t *testing.T) {
	mockPc := setupMockPc(t)
	expectReadHomeConfig(mockPc)

	const homeConfigForSelect = `current_workspace: project2
update_command: update
workspaces:
- name: project1
  path: /tmp/workspaces/project1
  root_path: ""
- name: project2
  path: /tmp/workspaces/project2
  root_path: ""
`

	mockPc.EXPECT().WriteFile(fakeHomeConfigPath, []byte(homeConfigForSelect), os.FileMode(0644))
	mockPc.EXPECT().Printf("active workspace changed to '%s'\n", "project2")

	_ = SelectWorkspaceAction("project2")
}
