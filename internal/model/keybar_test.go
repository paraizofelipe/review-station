package model

import (
	"strings"
	"testing"
)

func stripANSI(s string) string {
	return ansiRe.ReplaceAllString(s, "")
}

func TestRenderKeyBarAlwaysUsesOneFixedFooterLine(t *testing.T) {
	m := Model{Width: 120}
	got := stripANSI(m.renderKeyBar("j/k navegar  enter abrir  f filtros  r atualizar  q sair"))

	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	if len(lines) != 1 {
		t.Fatalf("renderKeyBar() deve ocupar exatamente 1 linha, got %d: %q", len(lines), got)
	}
	for _, want := range []string{"j/k navegar", "enter abrir", "f filtros", "r atualizar", "q sair"} {
		if !strings.Contains(lines[0], want) {
			t.Errorf("linha unica nao contem %q: %q", want, lines[0])
		}
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
