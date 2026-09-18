package core

import (
	"strings"
	"testing"

	yaml "go.yaml.in/yaml/v3"
)

func TestOrderedVarsCoercesScalars(t *testing.T) {
	const raw = `
USER_ID: 1000
RATIO: 1.5
ENABLED: true
NAME: alice
EMPTY: null
`
	var ov OrderedVars
	if err := yaml.Unmarshal([]byte(raw), &ov); err != nil {
		t.Fatal(err)
	}

	want := map[string]string{
		"USER_ID": "1000",
		"RATIO":   "1.5",
		"ENABLED": "true",
		"NAME":    "alice",
		"EMPTY":   "",
	}
	if len(ov) != len(want) {
		t.Fatalf("got %d vars, want %d: %+v", len(ov), len(want), ov)
	}
	for _, item := range ov {
		if got := want[item.Key]; got != item.Value {
			t.Errorf("%s: got %q, want %q", item.Key, item.Value, got)
		}
	}
}

func TestOrderedVarsPreservesOrder(t *testing.T) {
	const raw = `
A: 1
B: 2
C: 3
`
	var ov OrderedVars
	if err := yaml.Unmarshal([]byte(raw), &ov); err != nil {
		t.Fatal(err)
	}
	keys := []string{ov[0].Key, ov[1].Key, ov[2].Key}
	if keys[0] != "A" || keys[1] != "B" || keys[2] != "C" {
		t.Fatalf("order broken: %v", keys)
	}
}

func TestOrderedVarsErrorIncludesVariableName(t *testing.T) {
	const raw = `
OK: 1
BAD:
  nested: true
`
	var ov OrderedVars
	err := yaml.Unmarshal([]byte(raw), &ov)
	if err == nil {
		t.Fatal("expected error for nested value")
	}
	if !strings.Contains(err.Error(), `variable "BAD"`) {
		t.Fatalf("error should name variable, got: %v", err)
	}
}
