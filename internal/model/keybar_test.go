package model

import (
	"strings"
	"testing"
)

func stripANSI(s string) string {
	return ansiRe.ReplaceAllString(s, "")
}

func TestRenderKeyBarAlwaysUsesTwoFixedFooterLines(t *testing.T) {
	m := Model{Width: 80}
	got := stripANSI(m.renderKeyBar([]string{
		"j/k navegar  enter abrir  f filtros",
		"r atualizar  q sair",
	}))

	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("renderKeyBar() deve ocupar exatamente 2 linhas, got %d: %q", len(lines), got)
	}
	if !strings.Contains(lines[0], "j/k navegar") || !strings.Contains(lines[0], "enter abrir") || !strings.Contains(lines[0], "f filtros") {
		t.Errorf("primeira linha nao contem todos os binds esperados: %q", lines[0])
	}
	if !strings.Contains(lines[1], "r atualizar") || !strings.Contains(lines[1], "q sair") {
		t.Errorf("segunda linha nao contem todos os binds esperados: %q", lines[1])
	}
}

func TestListStatusbarShowsAllCurrentScreenBindKeys(t *testing.T) {
	m := Model{Width: 120}
	got := stripANSI(m.renderStatusbar())

	for _, want := range []string{"j/k", "enter", "f+s", "f+o", "r", "q", "ctrl+c"} {
		if !strings.Contains(got, want) {
			t.Errorf("renderStatusbar() deveria conter %q; got %q", want, got)
		}
	}
}

func TestCommentsStatusbarShowsAllCurrentScreenBindKeys(t *testing.T) {
	m := Model{Width: 140}
	got := stripANSI(m.renderCommentsStatusbar())

	for _, want := range []string{"j/k", "ctrl+d/u", "tab", "shift+tab", "r", "c", "backspace", "esc", "q", "ctrl+c"} {
		if !strings.Contains(got, want) {
			t.Errorf("renderCommentsStatusbar() deveria conter %q; got %q", want, got)
		}
	}
}
