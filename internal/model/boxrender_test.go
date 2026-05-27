package model

import (
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"paraizofelipe/review-station/internal/gitlab"
)

var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func visibleWidth(s string) int {
	return len([]rune(ansiRe.ReplaceAllString(s, "")))
}

func TestBoxRenderDump(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	now := time.Now()
	discussions := []gitlab.Discussion{
		{Notes: []gitlab.Note{
			{Author: "maria", CreatedAt: now.Add(-2 * time.Hour), Resolvable: true, Resolved: true,
				Body: "Acho que esse método deveria **validar** o input antes de salvar.\n\n```go\nif x == nil {\n    return errors.New(\"vazio\")\n}\n```"},
			{Author: "joao", CreatedAt: now.Add(-time.Hour), Body: "Boa, vou adicionar a `validação`."},
		}},
		{Notes: []gitlab.Note{
			{Author: "ana", CreatedAt: now.Add(-30 * time.Minute), Body: "Pode renomear essa variável?"},
		}},
		{Notes: []gitlab.Note{
			{System: true, Author: "bot", CreatedAt: now.Add(-10 * time.Minute), Body: "changed status to merged"},
		}},
	}

	width := 60
	out := buildRenderedDiscussions(discussions, width)

	// Dump legível (ANSI removido) para inspeção visual da estrutura.
	t.Log("\n" + ansiRe.ReplaceAllString(out, ""))

	// Toda linha de borda deve ter a mesma largura visível (caixas fechadas).
	for _, line := range strings.Split(out, "\n") {
		plain := ansiRe.ReplaceAllString(line, "")
		if strings.ContainsAny(plain, "╭╰") {
			w := visibleWidth(line)
			if w > width {
				t.Errorf("linha de borda excede largura %d: %d -> %q", width, w, plain)
			}
		}
	}

	// O fundo de superfície (#3c3836) deve aparecer no output.
	if !strings.Contains(out, "48;2;60;56;54") {
		t.Errorf("esperava sequência de background da superfície no output")
	}
}

func TestParseSystemNote(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		wantAction  string
		wantCommits []sysCommit
	}{
		{
			name:       "checklist com markdown",
			body:       "marked the checklist item **Tests** as completed",
			wantAction: "marked the checklist item Tests as completed",
		},
		{
			name:       "added 1 commit com html e compare",
			body:       "added 1 commit\n\n<ul><li>63c198e6 - chore: return has children parameter</li></ul>\n\n[Compare with previous version](/luizalabs/x/-/merge_requests/439/diffs?diff_id=4474251&start_sha=67414e)",
			wantAction: "added 1 commit",
			wantCommits: []sysCommit{
				{sha: "63c198e6", msg: "chore: return has children parameter"},
			},
		},
		{
			name:       "added N commits com resumo de range",
			body:       "added 4 commits\n\n<ul><li>a7cf88c6...822056c6 - 2 commits from branch <code>main</code></li><li>04976c4f - chore: a</li><li>d4349cf6 - chore: b</li></ul>\n\n[Compare with previous version](/x)",
			wantAction: "added 4 commits",
			wantCommits: []sysCommit{
				{sha: "04976c4f", msg: "chore: a"},
				{sha: "d4349cf6", msg: "chore: b"},
			},
		},
		{
			name:       "approved",
			body:       "approved this merge request",
			wantAction: "approved this merge request",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			action, commits := parseSystemNote(tc.body)
			if action != tc.wantAction {
				t.Errorf("action = %q, quero %q", action, tc.wantAction)
			}
			if len(commits) != len(tc.wantCommits) {
				t.Fatalf("commits = %+v, quero %+v", commits, tc.wantCommits)
			}
			for i := range commits {
				if commits[i] != tc.wantCommits[i] {
					t.Errorf("commit[%d] = %+v, quero %+v", i, commits[i], tc.wantCommits[i])
				}
			}
			if strings.Contains(action, "<") || strings.Contains(action, "*") {
				t.Errorf("action ainda tem markup: %q", action)
			}
		})
	}
}
