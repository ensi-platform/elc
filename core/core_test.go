package core

import "testing"

func TestCleanConfigPath(t *testing.T) {
	cases := map[string]string{
		"/tmp/ws/apps/../worktrees": "/tmp/ws/worktrees",
		"/tmp/ws/apps/../apps/test": "/tmp/ws/apps/test",
		"/tmp/ws/./apps/test":       "/tmp/ws/apps/test",
		"/tmp/ws/apps/test":         "/tmp/ws/apps/test",
		"/tmp/ws/apps/test/":        "/tmp/ws/apps/test/",
		"/api/v1/":                  "/api/v1/",
		"fpm-8.1:latest":            "fpm-8.1:latest",
		"relative/foo/../bar":       "relative/foo/../bar",
		"":                          "",
	}
	for in, want := range cases {
		if got := cleanConfigPath(in); got != want {
			t.Errorf("cleanConfigPath(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSubstVarsCleansAbsolutePaths(t *testing.T) {
	ctx := Context{
		{"WORKSPACE_PATH", "/tmp/ws"},
		{"APPS_ROOT", "/tmp/ws/apps"},
	}

	got, err := substVars("${WORKSPACE_PATH}/apps/../apps/test", &ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got != "/tmp/ws/apps/test" {
		t.Fatalf("got %q", got)
	}

	got, err = substVars("${APPS_ROOT}/../worktrees", &ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got != "/tmp/ws/worktrees" {
		t.Fatalf("got %q", got)
	}

	got, err = substVars("fpm-8.1:latest", &ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got != "fpm-8.1:latest" {
		t.Fatalf("non-path value changed: %q", got)
	}

	got, err = substVars("/api/v1/", &ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got != "/api/v1/" {
		t.Fatalf("API prefix should keep trailing slash, got %q", got)
	}
}
