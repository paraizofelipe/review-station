package model

import (
	"testing"

	"paraizofelipe/review-station/internal/config"
	"paraizofelipe/review-station/internal/gitlab"
	"paraizofelipe/review-station/internal/launcher"
)

func TestResolveOpenCodeCommandUsesRepoOverride(t *testing.T) {
	cfg := config.Config{OpenCode: config.OpenCodeConfig{Command: "global {{.IID}}"}}
	repo := config.Repo{Path: "org/app", Name: "app", Local: "~/projects/app", OpenCodeCommand: "repo {{.IID}}"}
	mr := &gitlab.MergeRequest{IID: 42}

	got, err := resolveOpenCodeCommand(repo, mr, cfg)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if got != "repo 42" {
		t.Errorf("got %q, want %q (override do repo)", got, "repo 42")
	}
}

func TestResolveOpenCodeCommandFallsBackToGlobal(t *testing.T) {
	cfg := config.Config{OpenCode: config.OpenCodeConfig{Command: "global {{.SourceBranch}}->{{.TargetBranch}}"}}
	repo := config.Repo{Path: "org/app", Local: "~/projects/app"}
	mr := &gitlab.MergeRequest{IID: 1, SourceBranch: "feat", TargetBranch: "main"}

	got, err := resolveOpenCodeCommand(repo, mr, cfg)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if got != "global feat->main" {
		t.Errorf("got %q, want %q (fallback global)", got, "global feat->main")
	}
}

func TestResolveOpenCodeCommandEmptyWhenUnconfigured(t *testing.T) {
	cfg := config.Config{}
	repo := config.Repo{Path: "org/app", Local: "~/projects/app"}
	mr := &gitlab.MergeRequest{IID: 1}

	got, err := resolveOpenCodeCommand(repo, mr, cfg)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if got != "" {
		t.Errorf("got %q, want vazio quando nada configurado", got)
	}
}

func TestResolveOpenCodeCommandInvalidTemplate(t *testing.T) {
	cfg := config.Config{OpenCode: config.OpenCodeConfig{Command: "{{.Nope"}}
	repo := config.Repo{Path: "org/app", Local: "~/projects/app"}
	mr := &gitlab.MergeRequest{IID: 1}

	if _, err := resolveOpenCodeCommand(repo, mr, cfg); err == nil {
		t.Fatal("esperava erro de template inválido")
	}
}

func TestResolveOpenCodeCommandEmptyWhenNilMR(t *testing.T) {
	cfg := config.Config{OpenCode: config.OpenCodeConfig{Command: "global {{.IID}}"}}
	repo := config.Repo{Path: "org/app", Local: "~/projects/app"}

	got, err := resolveOpenCodeCommand(repo, nil, cfg)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if got != "" {
		t.Errorf("got %q, want vazio quando mr é nil", got)
	}
}

func TestResolveOpenCodeCommandExecuteError(t *testing.T) {
	cfg := config.Config{OpenCode: config.OpenCodeConfig{Command: "{{.UnknownField}}"}}
	repo := config.Repo{Path: "org/app", Local: "~/projects/app"}
	mr := &gitlab.MergeRequest{IID: 1}

	if _, err := resolveOpenCodeCommand(repo, mr, cfg); err == nil {
		t.Fatal("esperava erro de execução do template (campo inexistente)")
	}
}

func TestBuildOpenCodeEnvSetsAllVars(t *testing.T) {
	repo := config.Repo{Path: "org/app", Name: "app", Local: "~/projects/app"}
	mr := &gitlab.MergeRequest{
		IID: 7, Title: "T", Description: "D", Author: "ana",
		SourceBranch: "feat", TargetBranch: "main",
		WebURL: "https://x/mr/7", State: "opened",
	}
	env := buildOpenCodeEnv(repo, mr)
	want := map[string]string{
		"RS_MR_IID": "7", "RS_MR_TITLE": "T", "RS_MR_DESCRIPTION": "D",
		"RS_MR_AUTHOR": "ana", "RS_MR_SOURCE_BRANCH": "feat", "RS_MR_TARGET_BRANCH": "main",
		"RS_MR_WEB_URL": "https://x/mr/7", "RS_MR_STATE": "opened",
		"RS_PROJECT_PATH": "org/app", "RS_PROJECT_NAME": "app",
		"RS_LOCAL": launcher.ExpandHome("~/projects/app"),
	}
	for k, v := range want {
		if env[k] != v {
			t.Errorf("env[%q] = %q, want %q", k, env[k], v)
		}
	}
}

func TestBuildOpenCodeEnvNilMR(t *testing.T) {
	env := buildOpenCodeEnv(config.Repo{Path: "org/app"}, nil)
	if env == nil {
		t.Fatal("esperava map não-nil para mr nil")
	}
	if len(env) != 0 {
		t.Errorf("esperava map vazio para mr nil, got %v", env)
	}
}
