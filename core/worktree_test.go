package core

import (
	"testing"
)

func TestSanitizeInstanceName(t *testing.T) {
	cases := map[string]string{
		"feature/foo": "feature-foo",
		"Feature/FOO": "feature-foo",
		"feat--bar":   "feat-bar",
		"my_branch":   "my-branch",
		"already-ok":  "already-ok",
		"///":         "",
	}
	for in, want := range cases {
		if got := SanitizeInstanceName(in); got != want {
			t.Errorf("SanitizeInstanceName(%q) = %q, want %q", in, got, want)
		}
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
