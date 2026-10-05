package cmd

import (
	"strings"
	"testing"

	"github.com/ensi-platform/elc/core"
	"github.com/golang/mock/gomock"
	"github.com/spf13/cobra"
)

func TestUniqueSortedAndFilterAlreadyUsed(t *testing.T) {
	got := uniqueSorted([]string{"b", "a", "b", "", "c"})
	want := []string{"a", "b", "c"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("uniqueSorted = %v, want %v", got, want)
	}

	filtered := filterAlreadyUsed([]string{"a", "b", "c"}, []string{"b"})
	if strings.Join(filtered, ",") != "a,c" {
		t.Fatalf("filterAlreadyUsed = %v", filtered)
	}
}

func TestCompleteHookNames(t *testing.T) {
	names, directive := completeHookNames(nil, nil, "")
	if strings.Join(names, ",") != strings.Join(core.ComponentHookNames(), ",") {
		t.Fatalf("hook names = %v", names)
	}
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Fatalf("directive = %v, want NoFileComp", directive)
	}

	names, directive = completeHookNames(nil, []string{"after_clone"}, "")
	if len(names) != 0 {
		t.Fatalf("expected no more hook completions, got %v", names)
	}
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Fatalf("directive = %v, want NoFileComp", directive)
	}
}

func TestCompleteComposeArgs(t *testing.T) {
	names, directive := completeComposeArgs(nil, nil, "")
	if len(names) == 0 {
		t.Fatal("expected compose subcommands")
	}
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Fatalf("directive = %v, want NoFileComp", directive)
	}

	names, directive = completeComposeArgs(nil, []string{"logs"}, "")
	if names != nil {
		t.Fatalf("expected nil after first arg, got %v", names)
	}
	if directive != cobra.ShellCompDirectiveDefault {
		t.Fatalf("directive = %v, want Default", directive)
	}
}

func TestListWorkspaceNames(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockPc := core.NewMockPC(ctrl)
	core.Pc = mockPc

	mockPc.EXPECT().HomeDir().Return("/tmp/home", nil)
	mockPc.EXPECT().FileExists("/tmp/home/.elc.yaml").Return(true)
	mockPc.EXPECT().ReadFile("/tmp/home/.elc.yaml").Return([]byte(`
current_workspace: project1
workspaces:
- name: project1
  path: /tmp/ws1
- name: zed
  path: /tmp/ws2
`), nil)

	got := listWorkspaceNames()
	if strings.Join(got, ",") != "project1,zed" {
		t.Fatalf("listWorkspaceNames = %v", got)
	}
}
