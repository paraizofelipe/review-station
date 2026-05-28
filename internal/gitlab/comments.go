package gitlab

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"time"

	"paraizofelipe/review-station/internal/config"
)

// Position descreve a localização no diff onde uma note inline foi criada.
type Position struct {
	NewPath string
	OldPath string
	NewLine int // 0 = não aplicável (linha removida)
	OldLine int // 0 = não aplicável (linha adicionada)
}

type Note struct {
	ID         int
	Body       string
	Author     string
	CreatedAt  time.Time
	System     bool
	Resolvable bool
	Resolved   bool
	Position   *Position // nil = comentário geral, não-inline
}

type Discussion struct {
	ID    string
	Notes []Note
}

type positionResponse struct {
	NewPath string `json:"new_path"`
	OldPath string `json:"old_path"`
	NewLine *int   `json:"new_line"`
	OldLine *int   `json:"old_line"`
}

type noteResponse struct {
	ID     int    `json:"id"`
	Body   string `json:"body"`
	Author struct {
		Username string `json:"username"`
	} `json:"author"`
	CreatedAt  string            `json:"created_at"`
	System     bool              `json:"system"`
	Resolvable bool              `json:"resolvable"`
	Resolved   bool              `json:"resolved"`
	Position   *positionResponse `json:"position"`
}

type discussionResponse struct {
	ID    string         `json:"id"`
	Notes []noteResponse `json:"notes"`
}

func (c Client) FetchDiscussions(ctx context.Context, repo config.Repo, mrIID int) ([]Discussion, error) {
	encoded := url.PathEscape(repo.Path)
	apiURL := fmt.Sprintf("%s/api/v4/projects/%s/merge_requests/%d/discussions?per_page=100",
		c.BaseURL, encoded, mrIID)

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
		return nil, fmt.Errorf("gitlab API returned %d for discussions of MR !%d in %s", resp.StatusCode, mrIID, repo.Path)
	}

	var raw []discussionResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	return parseDiscussions(raw), nil
}

func FetchDiscussionsFallback(ctx context.Context, repo config.Repo, mrIID int) ([]Discussion, error) {
	glabPath, err := exec.LookPath("glab")
	if err != nil {
		return nil, fmt.Errorf("glab not found in PATH: %w", err)
	}

	apiPath := fmt.Sprintf("projects/%s/merge_requests/%d/discussions?per_page=100",
		url.PathEscape(repo.Path), mrIID)

	out, err := exec.CommandContext(ctx, glabPath, "api", apiPath).Output()
	if err != nil {
		return nil, fmt.Errorf("glab api failed: %w", err)
	}

	var raw []discussionResponse
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, fmt.Errorf("glab output parse error: %w", err)
	}

	return parseDiscussions(raw), nil
}

// ReplyToDiscussion posta uma nova nota em uma discussion existente.
func (c Client) ReplyToDiscussion(ctx context.Context, repo config.Repo, mrIID int, discussionID, body string) error {
	encoded := url.PathEscape(repo.Path)
	apiURL := fmt.Sprintf("%s/api/v4/projects/%s/merge_requests/%d/discussions/%s/notes",
		c.BaseURL, encoded, mrIID, url.PathEscape(discussionID))

	payload, err := json.Marshal(map[string]string{"body": body})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("gitlab retornou %d ao responder à discussion %s", resp.StatusCode, discussionID)
	}
	return nil
}

func parseDiscussions(raw []discussionResponse) []Discussion {
	discussions := make([]Discussion, 0, len(raw))
	for _, d := range raw {
		notes := make([]Note, 0, len(d.Notes))
		for _, n := range d.Notes {
			created, _ := time.Parse(time.RFC3339, n.CreatedAt)
			var pos *Position
			if n.Position != nil && (n.Position.NewPath != "" || n.Position.OldPath != "") {
				pos = &Position{
					NewPath: n.Position.NewPath,
					OldPath: n.Position.OldPath,
				}
				if n.Position.NewLine != nil {
					pos.NewLine = *n.Position.NewLine
				}
				if n.Position.OldLine != nil {
					pos.OldLine = *n.Position.OldLine
				}
			}
			notes = append(notes, Note{
				ID:         n.ID,
				Body:       n.Body,
				Author:     n.Author.Username,
				CreatedAt:  created,
				System:     n.System,
				Resolvable: n.Resolvable,
				Resolved:   n.Resolved,
				Position:   pos,
			})
		}
		discussions = append(discussions, Discussion{ID: d.ID, Notes: notes})
	}
	return discussions
}
