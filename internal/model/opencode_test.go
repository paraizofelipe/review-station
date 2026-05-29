package model

import (
	"testing"

	"paraizofelipe/review-station/internal/config"
	"paraizofelipe/review-station/internal/gitlab"
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

func TestBuildOpenCodeEnvSetsMRVars(t *testing.T) {
	repo := config.Repo{Path: "org/app", Name: "app", Local: "~/projects/app"}
	mr := &gitlab.MergeRequest{IID: 7, SourceBranch: "feat", WebURL: "https://x/mr/7"}

	env := buildOpenCodeEnv(repo, mr)

	if env["RS_MR_IID"] != "7" {
		t.Errorf("RS_MR_IID = %q, want 7", env["RS_MR_IID"])
	}
	if env["RS_MR_SOURCE_BRANCH"] != "feat" {
		t.Errorf("RS_MR_SOURCE_BRANCH = %q, want feat", env["RS_MR_SOURCE_BRANCH"])
	}
	if env["RS_PROJECT_PATH"] != "org/app" {
		t.Errorf("RS_PROJECT_PATH = %q, want org/app", env["RS_PROJECT_PATH"])
	}
}
