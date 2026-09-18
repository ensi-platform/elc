package core

import (
	"testing"

	yaml "go.yaml.in/yaml/v3"
)

func TestResolvedAfterCloneHookPrefersNested(t *testing.T) {
	cc := ComponentConfig{
		Hooks:          ComponentHooks{AfterClone: "hooks/after-clone.sh"},
		AfterCloneHook: "legacy/after-clone.sh",
	}
	if got := cc.ResolvedAfterCloneHook(); got != "hooks/after-clone.sh" {
		t.Fatalf("expected nested hook, got %q", got)
	}
}

func TestResolvedAfterCloneHookLegacyFallback(t *testing.T) {
	cc := ComponentConfig{AfterCloneHook: "legacy/after-clone.sh"}
	if got := cc.ResolvedAfterCloneHook(); got != "legacy/after-clone.sh" {
		t.Fatalf("expected legacy hook, got %q", got)
	}
}

func TestComponentHooksYAMLUnmarshal(t *testing.T) {
	const raw = `
hooks:
  after_clone: a.sh
  worktree_create: b.sh
  worktree_remove: c.sh
after_clone_hook: legacy.sh
`
	var cc ComponentConfig
	if err := yaml.Unmarshal([]byte(raw), &cc); err != nil {
		t.Fatal(err)
	}
	if cc.Hooks.AfterClone != "a.sh" || cc.Hooks.WorktreeCreate != "b.sh" || cc.Hooks.WorktreeRemove != "c.sh" {
		t.Fatalf("unexpected hooks: %+v", cc.Hooks)
	}
	if cc.AfterCloneHook != "legacy.sh" {
		t.Fatalf("unexpected legacy after_clone_hook: %q", cc.AfterCloneHook)
	}
	if got := cc.ResolvedAfterCloneHook(); got != "a.sh" {
		t.Fatalf("ResolvedAfterCloneHook = %q, want a.sh", got)
	}
}

func TestOptionalComposeFileNullVsOmit(t *testing.T) {
	var omitted ComponentConfig
	if err := yaml.Unmarshal([]byte("path: /x\n"), &omitted); err != nil {
		t.Fatal(err)
	}
	if cf := omitted.ComposeFile(); cf.Present || cf.Disabled {
		t.Fatalf("omitted compose_file should be unset, got %+v", cf)
	}

	var disabled ComponentConfig
	if err := yaml.Unmarshal([]byte("path: /x\ncompose_file: null\n"), &disabled); err != nil {
		t.Fatal(err)
	}
	if cf := disabled.ComposeFile(); !cf.Present || !cf.Disabled {
		t.Fatalf("compose_file: null should disable, got %+v", cf)
	}

	var path ComponentConfig
	if err := yaml.Unmarshal([]byte("compose_file: ${SVC_PATH}/docker-compose.yml\n"), &path); err != nil {
		t.Fatal(err)
	}
	if cf := path.ComposeFile(); !cf.Present || cf.Disabled || cf.Value != "${SVC_PATH}/docker-compose.yml" {
		t.Fatalf("unexpected compose_file: %+v", cf)
	}
}

func TestComponentHooksMerge(t *testing.T) {
	base := ComponentConfig{
		Hooks: ComponentHooks{AfterClone: "base-clone.sh", WorktreeCreate: "base-wt.sh"},
	}
	over := ComponentConfig{
		Hooks:          ComponentHooks{WorktreeRemove: "over-rm.sh"},
		AfterCloneHook: "legacy.sh",
	}
	merged := base.merge(over)
	if merged.Hooks.AfterClone != "base-clone.sh" {
		t.Fatalf("AfterClone overwritten unexpectedly: %q", merged.Hooks.AfterClone)
	}
	if merged.Hooks.WorktreeCreate != "base-wt.sh" {
		t.Fatalf("WorktreeCreate lost: %q", merged.Hooks.WorktreeCreate)
	}
	if merged.Hooks.WorktreeRemove != "over-rm.sh" {
		t.Fatalf("WorktreeRemove not merged: %q", merged.Hooks.WorktreeRemove)
	}
	if merged.AfterCloneHook != "legacy.sh" {
		t.Fatalf("legacy AfterCloneHook not merged: %q", merged.AfterCloneHook)
	}
}
