package gitlab

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"paraizofelipe/review-station/internal/config"
)

type FileDiff struct {
	OldPath string
	NewPath string
	Diff    string
}

type DiffLineKind int

const (
	DiffLineKindContext DiffLineKind = iota
	DiffLineKindAdded
	DiffLineKindRemoved
)

type DiffLine struct {
	Kind    DiffLineKind
	Content string
	OldLine int // 0 = não aplicável
	NewLine int // 0 = não aplicável
}

type fileDiffResponse struct {
	OldPath string `json:"old_path"`
	NewPath string `json:"new_path"`
	Diff    string `json:"diff"`
}

func (c Client) FetchDiffs(ctx context.Context, repo config.Repo, mrIID int) ([]FileDiff, error) {
	encoded := url.PathEscape(repo.Path)
	apiURL := fmt.Sprintf("%s/api/v4/projects/%s/merge_requests/%d/diffs?per_page=30",
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
		return nil, fmt.Errorf("gitlab API returned %d for diffs of MR !%d in %s", resp.StatusCode, mrIID, repo.Path)
	}

	var raw []fileDiffResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	return parseDiffs(raw), nil
}

func FetchDiffsFallback(ctx context.Context, repo config.Repo, mrIID int) ([]FileDiff, error) {
	glabPath, err := exec.LookPath("glab")
	if err != nil {
		return nil, fmt.Errorf("glab not found in PATH: %w", err)
	}

	apiPath := fmt.Sprintf("projects/%s/merge_requests/%d/diffs?per_page=30",
		url.PathEscape(repo.Path), mrIID)

	out, err := exec.CommandContext(ctx, glabPath, "api", apiPath).Output()
	if err != nil {
		return nil, fmt.Errorf("glab api failed: %w", err)
	}

	var raw []fileDiffResponse
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, fmt.Errorf("glab output parse error: %w", err)
	}

	return parseDiffs(raw), nil
}

func parseDiffs(raw []fileDiffResponse) []FileDiff {
	diffs := make([]FileDiff, 0, len(raw))
	for _, r := range raw {
		diffs = append(diffs, FileDiff{
			OldPath: r.OldPath,
			NewPath: r.NewPath,
			Diff:    r.Diff,
		})
	}
	return diffs
}

var hunkHeaderRe = regexp.MustCompile(`^@@ -(\d+)(?:,\d+)? \+(\d+)(?:,\d+)? @@`)

// parseDiffLines converte o texto de unified diff em uma lista de DiffLine
// com números de linha absolutos calculados a partir dos headers de hunk.
func parseDiffLines(diffText string) []DiffLine {
	var lines []DiffLine
	oldLine, newLine := 0, 0

	for _, raw := range strings.Split(diffText, "\n") {
		if m := hunkHeaderRe.FindStringSubmatch(raw); m != nil {
			old, _ := strconv.Atoi(m[1])
			neu, _ := strconv.Atoi(m[2])
			oldLine = old
			newLine = neu
			continue
		}
		if len(raw) == 0 {
			continue
		}
		switch raw[0] {
		case ' ':
			lines = append(lines, DiffLine{Kind: DiffLineKindContext, Content: raw[1:], OldLine: oldLine, NewLine: newLine})
			oldLine++
			newLine++
		case '-':
			lines = append(lines, DiffLine{Kind: DiffLineKindRemoved, Content: raw[1:], OldLine: oldLine})
			oldLine++
		case '+':
			lines = append(lines, DiffLine{Kind: DiffLineKindAdded, Content: raw[1:], NewLine: newLine})
			newLine++
		}
	}
	return lines
}

// ExtractDiffContext retorna até `before` linhas de contexto terminando na
// linha alvo. targetNewLine e targetOldLine são mutuamente exclusivos: passe
// 0 no que não se aplicar.
func ExtractDiffContext(diffText string, targetNewLine, targetOldLine, before int) []DiffLine {
	all := parseDiffLines(diffText)

	targetIdx := -1
	for i, l := range all {
		if targetNewLine > 0 && l.NewLine == targetNewLine {
			targetIdx = i
			break
		}
		if targetOldLine > 0 && targetNewLine == 0 && l.OldLine == targetOldLine {
			targetIdx = i
			break
		}
	}
	if targetIdx == -1 {
		return nil
	}

	start := targetIdx - before
	if start < 0 {
		start = 0
	}
	return all[start : targetIdx+1]
}
