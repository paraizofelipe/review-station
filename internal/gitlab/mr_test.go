package gitlab_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"paraizofelipe/review-station/internal/config"
	"paraizofelipe/review-station/internal/gitlab"
)

func TestFetchMRs(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	response := []map[string]any{
		{
			"iid":           float64(42),
			"title":         "Fix login bug",
			"author":        map[string]any{"username": "joao"},
			"source_branch": "feature/auth",
			"target_branch": "main",
			"created_at":    now.Format(time.RFC3339),
			"web_url":       "https://gitlab.com/org/repo/-/merge_requests/42",
			"state":         "opened",
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer srv.Close()

	client := gitlab.NewClient(srv.URL, "test-token")
	repo := config.Repo{Name: "repo", Path: "org/repo"}

	mrs, err := client.FetchMRs(context.Background(), repo, "opened")
	if err != nil {
		t.Fatalf("FetchMRs() error = %v", err)
	}
	if len(mrs) != 1 {
		t.Fatalf("len(mrs) = %d, want 1", len(mrs))
	}
	if mrs[0].IID != 42 {
		t.Errorf("IID = %d, want 42", mrs[0].IID)
	}
	if mrs[0].Title != "Fix login bug" {
		t.Errorf("Title = %q, want %q", mrs[0].Title, "Fix login bug")
	}
	if mrs[0].Author != "joao" {
		t.Errorf("Author = %q", mrs[0].Author)
	}
	if mrs[0].SourceBranch != "feature/auth" {
		t.Errorf("SourceBranch = %q", mrs[0].SourceBranch)
	}
	if mrs[0].CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
}

func TestFetchMRsUnauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	defer srv.Close()

	client := gitlab.NewClient(srv.URL, "bad-token")
	repo := config.Repo{Name: "repo", Path: "org/repo"}

	_, err := client.FetchMRs(context.Background(), repo, "opened")
	if err == nil {
		t.Error("FetchMRs() expected error for 401, got nil")
	}
}
