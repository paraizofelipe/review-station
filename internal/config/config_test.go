package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"paraizofelipe/review-station/internal/config"
)

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	content := `
[gitlab]
base_url = "https://gitlab.example.com"
token = "glpat-abc123"

[[repo]]
name = "my-app"
path = "org/my-app"
local = "~/projects/my-app"
`
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.GitLab.BaseURL != "https://gitlab.example.com" {
		t.Errorf("BaseURL = %q, want %q", cfg.GitLab.BaseURL, "https://gitlab.example.com")
	}
	if cfg.GitLab.Token != "glpat-abc123" {
		t.Errorf("Token = %q, want %q", cfg.GitLab.Token, "glpat-abc123")
	}
	if len(cfg.Repos) != 1 {
		t.Fatalf("len(Repos) = %d, want 1", len(cfg.Repos))
	}
	if cfg.Repos[0].Path != "org/my-app" {
		t.Errorf("Repos[0].Path = %q, want %q", cfg.Repos[0].Path, "org/my-app")
	}
}

func TestLoadMissingFile(t *testing.T) {
	_, err := config.Load("/nonexistent/path/config.toml")
	if err == nil {
		t.Error("Load() expected error for missing file, got nil")
	}
}

func TestLoadExpandsEnvVars(t *testing.T) {
	t.Setenv("RS_TEST_TOKEN", "glpat-from-env")
	t.Setenv("RS_TEST_HOST", "gitlab.internal")
	t.Setenv("RS_TEST_HOME", "/home/tester")

	dir := t.TempDir()
	content := `
[gitlab]
base_url = "https://${RS_TEST_HOST}"
token = "${RS_TEST_TOKEN}"

[[repo]]
name = "app"
path = "org/app"
local = "${RS_TEST_HOME}/projects/app"

[[repo]]
name = "missing-var"
path = "org/${RS_TEST_UNSET}"
local = "literal-$RS_TEST_HOME"
`
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if got, want := cfg.GitLab.BaseURL, "https://gitlab.internal"; got != want {
		t.Errorf("BaseURL = %q, want %q", got, want)
	}
	if got, want := cfg.GitLab.Token, "glpat-from-env"; got != want {
		t.Errorf("Token = %q, want %q", got, want)
	}
	if got, want := cfg.Repos[0].Local, "/home/tester/projects/app"; got != want {
		t.Errorf("Repos[0].Local = %q, want %q", got, want)
	}
	if got, want := cfg.Repos[1].Path, "org/"; got != want {
		t.Errorf("Repos[1].Path = %q, want %q (env var ausente deve virar string vazia)", got, want)
	}
	if got, want := cfg.Repos[1].Local, "literal-$RS_TEST_HOME"; got != want {
		t.Errorf("Repos[1].Local = %q, want %q ($VAR sem chaves deve permanecer literal)", got, want)
	}
}
