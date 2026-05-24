package gitlab_test

import (
	"os/exec"
	"testing"

	"paraizofelipe/review-station/internal/config"
	"paraizofelipe/review-station/internal/gitlab"
)

func TestFetchMRsFallback_GlabNotFound(t *testing.T) {
	if _, err := exec.LookPath("glab"); err == nil {
		t.Skip("glab disponível — teste só cobre caminho de erro sem glab")
	}
	repo := config.Repo{Name: "repo", Path: "org/repo"}
	_, err := gitlab.FetchMRsFallback(repo, "opened")
	if err == nil {
		t.Error("FetchMRsFallback() expected error when glab not found, got nil")
	}
}
