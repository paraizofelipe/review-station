package gitlab

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"paraizofelipe/review-station/internal/config"
)

type PipelineStatus struct {
	Status string // "success", "failed", "running", "pending", "canceled", "skipped"
}

type MergeRequest struct {
	IID          int
	Title        string
	Description  string
	Author       string
	SourceBranch string
	TargetBranch string
	CreatedAt    time.Time
	Pipeline     *PipelineStatus
	WebURL       string
	State        string
}

type mrResponse struct {
	IID         int    `json:"iid"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Author      struct {
		Username string `json:"username"`
	} `json:"author"`
	SourceBranch string `json:"source_branch"`
	TargetBranch string `json:"target_branch"`
	CreatedAt    string `json:"created_at"`
	WebURL       string `json:"web_url"`
	State        string `json:"state"`
	HeadPipeline *struct {
		Status string `json:"status"`
	} `json:"head_pipeline"`
}

func (c Client) FetchMRs(ctx context.Context, repo config.Repo, state string) ([]MergeRequest, error) {
	encoded := url.PathEscape(repo.Path)
	apiURL := fmt.Sprintf("%s/api/v4/projects/%s/merge_requests?state=%s&per_page=100",
		c.BaseURL, encoded, url.QueryEscape(state))

	req, err := c.newRequest(ctx, http.MethodGet, apiURL)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		io.Copy(io.Discard, resp.Body)
		return nil, fmt.Errorf("gitlab API returned %d for %s", resp.StatusCode, repo.Path)
	}

	var raw []mrResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	mrs := make([]MergeRequest, 0, len(raw))
	for _, r := range raw {
		created, _ := time.Parse(time.RFC3339, r.CreatedAt)
		var pipeline *PipelineStatus
		if r.HeadPipeline != nil {
			pipeline = &PipelineStatus{Status: r.HeadPipeline.Status}
		}
		mrs = append(mrs, MergeRequest{
			IID:          r.IID,
			Title:        r.Title,
			Description:  r.Description,
			Author:       r.Author.Username,
			SourceBranch: r.SourceBranch,
			TargetBranch: r.TargetBranch,
			CreatedAt:    created,
			Pipeline:     pipeline,
			WebURL:       r.WebURL,
			State:        r.State,
		})
	}
	return mrs, nil
}
