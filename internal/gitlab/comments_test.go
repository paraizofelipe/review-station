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

func TestFetchDiscussions(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	response := []map[string]any{
		{
			"id": "abc123",
			"notes": []map[string]any{
				{
					"id":         float64(1),
					"body":       "Please fix the timeout",
					"author":     map[string]any{"username": "joao"},
					"created_at": now.Format(time.RFC3339),
					"system":     false,
					"resolvable": true,
					"resolved":   false,
				},
				{
					"id":         float64(2),
					"body":       "Agreed, will fix",
					"author":     map[string]any{"username": "maria"},
					"created_at": now.Format(time.RFC3339),
					"system":     false,
					"resolvable": false,
					"resolved":   false,
				},
			},
		},
		{
			"id": "sys1",
			"notes": []map[string]any{
				{
					"id":         float64(3),
					"body":       "added 2 commits",
					"author":     map[string]any{"username": "joao"},
					"created_at": now.Format(time.RFC3339),
					"system":     true,
					"resolvable": false,
					"resolved":   false,
				},
			},
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

	discussions, err := client.FetchDiscussions(context.Background(), repo, 42)
	if err != nil {
		t.Fatalf("FetchDiscussions() error = %v", err)
	}
	if len(discussions) != 2 {
		t.Fatalf("len(discussions) = %d, want 2", len(discussions))
	}

	d := discussions[0]
	if d.ID != "abc123" {
		t.Errorf("ID = %q, want %q", d.ID, "abc123")
	}
	if len(d.Notes) != 2 {
		t.Fatalf("len(Notes) = %d, want 2", len(d.Notes))
	}
	if d.Notes[0].Body != "Please fix the timeout" {
		t.Errorf("Notes[0].Body = %q", d.Notes[0].Body)
	}
	if d.Notes[0].Author != "joao" {
		t.Errorf("Notes[0].Author = %q", d.Notes[0].Author)
	}
	if d.Notes[0].System {
		t.Error("Notes[0].System should be false")
	}
	if !d.Notes[0].Resolvable {
		t.Error("Notes[0].Resolvable should be true")
	}
	if d.Notes[1].Author != "maria" {
		t.Errorf("Notes[1].Author = %q", d.Notes[1].Author)
	}

	sys := discussions[1]
	if !sys.Notes[0].System {
		t.Error("system Notes[0].System should be true")
	}
}

func TestFetchDiscussionsUnauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	defer srv.Close()

	client := gitlab.NewClient(srv.URL, "bad-token")
	repo := config.Repo{Name: "repo", Path: "org/repo"}

	_, err := client.FetchDiscussions(context.Background(), repo, 42)
	if err == nil {
		t.Error("FetchDiscussions() expected error for 401, got nil")
	}
}
