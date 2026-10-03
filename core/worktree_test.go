package core

import (
	"testing"

	"github.com/golang/mock/gomock"
)

func TestSanitizeInstanceName(t *testing.T) {
	cases := map[string]string{
		"feature/foo": "feature_foo",
		"Feature/FOO": "feature_foo",
		"feature-foo": "feature-foo",
		"feature_foo": "feature__foo",
		"feat--bar":   "feat--bar",
		"my_branch":   "my__branch",
		"already-ok":  "already-ok",
		"foo@bar":     "foo-bar",
		"фича/foo":    "_foo",
		"///":         "",
	}
	for in, want := range cases {
		if got := SanitizeInstanceName(in); got != want {
			t.Errorf("SanitizeInstanceName(%q) = %q, want %q", in, got, want)
		}
	}

	// colliding separators must stay distinct
	if SanitizeInstanceName("feature/foo") == SanitizeInstanceName("feature-foo") {
		t.Fatal("slash and dash branches collided")
	}
	if SanitizeInstanceName("feature/foo") == SanitizeInstanceName("feature_foo") {
		t.Fatal("slash and underscore branches collided")
	}
	if SanitizeInstanceName("feature-foo") == SanitizeInstanceName("feature_foo") {
		t.Fatal("dash and underscore branches collided")
	}
}

func TestValidateBranchName(t *testing.T) {
	if err := ValidateBranchName("feature/foo"); err != nil {
		t.Fatalf("valid branch rejected: %v", err)
	}
	invalid := []string{"", " ", "..", "../x", "/abs", "foo/../bar", "foo//bar"}
	for _, name := range invalid {
		if err := ValidateBranchName(name); err == nil {
			t.Errorf("expected invalid branch %q", name)
		}
	}
}

func TestIsInside(t *testing.T) {
	if !isInside("/ws/apps/test", "/ws/apps/test") {
		t.Fatal("exact path should match")
	}
	if !isInside("/ws/apps/test/src", "/ws/apps/test") {
		t.Fatal("subdirectory should match")
	}
	if isInside("/ws/apps", "/ws/apps/test") {
		t.Fatal("parent should not match")
	}
	if isInside("/ws/apps/test-other", "/ws/apps/test") {
		t.Fatal("sibling prefix should not match")
	}
}

func TestResolveWorktreeStartPoint(t *testing.T) {
	got, err := resolveWorktreeStartPoint("/repo", "HEAD", nil)
	if err != nil || got != "HEAD" {
		t.Fatalf("HEAD: got %q err %v", got, err)
	}
	got, err = resolveWorktreeStartPoint("/repo", "head", nil)
	if err != nil || got != "HEAD" {
		t.Fatalf("head: got %q err %v", got, err)
	}
}

func TestInspectGitWorktreeLinked(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockPc := NewMockPC(ctrl)
	Pc = mockPc

	cwd := "/tmp/external/feat-wt/src"
	top := "/tmp/external/feat-wt"
	main := "/tmp/workspaces/project1/apps/test"

	mockPc.EXPECT().ExecToString([]string{"git", "-C", cwd, "rev-parse", "--show-toplevel"}, nil).
		Return(0, top+"\n", nil)
	mockPc.EXPECT().ExecToString([]string{"git", "-C", cwd, "rev-parse", "--git-common-dir"}, nil).
		Return(0, main+"/.git\n", nil)
	mockPc.EXPECT().ExecToString([]string{"git", "-C", cwd, "rev-parse", "--abbrev-ref", "HEAD"}, nil).
		Return(0, "feat\n", nil)

	info, ok := inspectGitWorktree(cwd)
	if !ok || !info.Linked {
		t.Fatalf("expected linked worktree, got %+v ok=%v", info, ok)
	}
	if info.TopLevel != top || info.MainPath != main || info.Branch != "feat" {
		t.Fatalf("unexpected info: %+v", info)
	}
}

func TestInspectGitWorktreeMainClone(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockPc := NewMockPC(ctrl)
	Pc = mockPc

	main := "/tmp/workspaces/project1/apps/test"
	mockPc.EXPECT().ExecToString([]string{"git", "-C", main, "rev-parse", "--show-toplevel"}, nil).
		Return(0, main+"\n", nil)
	mockPc.EXPECT().ExecToString([]string{"git", "-C", main, "rev-parse", "--git-common-dir"}, nil).
		Return(0, ".git\n", nil)

	info, ok := inspectGitWorktree(main)
	if !ok || info.Linked {
		t.Fatalf("expected main clone, got %+v ok=%v", info, ok)
	}
	if info.MainPath != main || info.TopLevel != main {
		t.Fatalf("unexpected info: %+v", info)
	}
}
