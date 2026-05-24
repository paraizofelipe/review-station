package gitlab

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"time"

	"paraizofelipe/review-station/internal/config"
)

type glabMRResponse struct {
	IID    int    `json:"iid"`
	Title  string `json:"title"`
	Author struct {
		Username string `json:"username"`
	} `json:"author"`
	SourceBranch string `json:"source_branch"`
	TargetBranch string `json:"target_branch"`
	CreatedAt    string `json:"created_at"`
	WebURL       string `json:"web_url"`
	State        string `json:"state"`
}

func FetchMRsFallback(repo config.Repo, state string) ([]MergeRequest, error) {
	glabPath, err := exec.LookPath("glab")
	if err != nil {
		return nil, fmt.Errorf("glab not found in PATH: %w", err)
	}

	out, err := exec.Command(glabPath,
		"mr", "list",
		"--repo", repo.Path,
		"--state", state,
		"--output", "json",
	).Output()
	if err != nil {
		return nil, fmt.Errorf("glab mr list failed for %s: %w", repo.Path, err)
	}

	var raw []glabMRResponse
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, fmt.Errorf("glab output parse error for %s: %w", repo.Path, err)
	}

	mrs := make([]MergeRequest, 0, len(raw))
	for _, r := range raw {
		created, _ := time.Parse(time.RFC3339, r.CreatedAt)
		mrs = append(mrs, MergeRequest{
			IID:          r.IID,
			Title:        r.Title,
			Author:       r.Author.Username,
			SourceBranch: r.SourceBranch,
			TargetBranch: r.TargetBranch,
			CreatedAt:    created,
			WebURL:       r.WebURL,
			State:        r.State,
		})
	}
	return mrs, nil
}
