package core

import (
	"path"
	"testing"

	"github.com/golang/mock/gomock"
)

func TestSetNativeHooksPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPc := NewMockPC(ctrl)
	Pc = mockPc

	svcPath := "/tmp/workspaces/project1/apps/test"
	hooksDir := "githooks"
	hooksPath := path.Join(svcPath, hooksDir)

	mockPc.EXPECT().FileExists(path.Join(svcPath, ".git")).Return(true)
	mockPc.EXPECT().FileExists(hooksPath).Return(true)
	mockPc.EXPECT().ExecToString(
		[]string{"git", "-C", svcPath, "config", "core.hooksPath", hooksDir},
		nil,
	).Return(0, "", nil)
	mockPc.EXPECT().Println("\033[0;32mcore.hooksPath set to githooks in /tmp/workspaces/project1/apps/test/.git.\033[0m")

	if err := SetNativeHooksPath(&GlobalOptions{}, svcPath, "./githooks/"); err != nil {
		t.Fatal(err)
	}
}

func TestSetNativeHooksPathSkipsMissingRepo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPc := NewMockPC(ctrl)
	Pc = mockPc

	svcPath := "/tmp/workspaces/project1/apps/test"
	mockPc.EXPECT().FileExists(path.Join(svcPath, ".git")).Return(false)
	mockPc.EXPECT().Println("\033[0;33mRepository /tmp/workspaces/project1/apps/test/.git is not exists, skip hooks installation.\033[0m")

	if err := SetNativeHooksPath(&GlobalOptions{}, svcPath, "githooks"); err != nil {
		t.Fatal(err)
	}
}
