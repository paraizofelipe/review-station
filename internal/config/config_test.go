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
