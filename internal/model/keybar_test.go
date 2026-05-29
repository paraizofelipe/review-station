package model

import (
	"strings"
	"testing"

	"paraizofelipe/review-station/internal/ui"
)

func stripANSI(s string) string {
	return ansiRe.ReplaceAllString(s, "")
}

func TestRenderKeyBarKeepsShortFooterOnOneLine(t *testing.T) {
	m := Model{Width: 120}
	got := stripANSI(m.renderKeyBar("j/k navegar  enter abrir  f filtros  r atualizar  q sair"))

	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	if len(lines) != 1 {
		t.Fatalf("legenda curta deve caber em 1 linha, got %d: %q", len(lines), got)
	}
	for _, want := range []string{"j/k navegar", "enter abrir", "f filtros", "r atualizar", "q sair"} {
		if !strings.Contains(lines[0], want) {
			t.Errorf("linha unica nao contem %q: %q", want, lines[0])
		}
	}
}

func TestRenderKeyBarWrapsLongFooterWithoutTruncating(t *testing.T) {
	m := Model{Width: 80}
	got := stripANSI(strings.TrimRight(m.renderCommentsStatusbar(), "\n"))

	if strings.Contains(got, "…") {
		t.Fatalf("footer não deveria truncar com reticências em 80 col; got %q", got)
	}
	for _, want := range []string{"colapsar", "backspace ou esc voltar", "q ou ctrl+c", "sair"} {
		if !strings.Contains(got, want) {
			t.Errorf("footer deveria mostrar %q por inteiro (sem truncar); got %q", want, got)
		}
	}
}

func TestRenderKeyBarWrapsToTwoLinesWhenItDoesNotFit(t *testing.T) {
	m := Model{Width: 80}
	got := stripANSI(strings.TrimRight(m.renderCommentsStatusbar(), "\n"))

	lines := strings.Split(got, "\n")
	if len(lines) != 3 {
		t.Fatalf("footer de comentários em 80 col deveria ocupar 3 linhas; got %d: %q", len(lines), lines)
	}
	for i, line := range lines {
		if w := len([]rune(line)); w > 80 {
			t.Errorf("linha %d excede a largura 80 (%d): %q", i, w, line)
		}
	}
}

func TestRenderKeyBarHighlightsBindKeyTokens(t *testing.T) {
	m := Model{Width: 120}
	got := m.renderKeyBar("j/k nav enter abrir f+s status r recar q/ctrl+c")

	for _, key := range []string{"j/k", "enter", "f+s", "r", "q/ctrl+c"} {
		want := ui.StyleBindKey.Render(key)
		if !strings.Contains(got, want) {
			t.Fatalf("renderKeyBar() deveria destacar bindkey %q com StyleBindKey; got %q; want substring %q", key, got, want)
		}
	}
}

func TestListStatusbarShowsAllCurrentScreenBindKeys(t *testing.T) {
	m := Model{Width: 120}
	got := stripANSI(m.renderStatusbar())

	for _, want := range []string{"j/k", "enter", "f+s", "f+o", "f+p", "c", "r", "q", "ctrl+c"} {
		if !strings.Contains(got, want) {
			t.Errorf("renderStatusbar() deveria conter %q; got %q", want, got)
		}
	}
}

func TestListViewClosesWithBindKeyFooter(t *testing.T) {
	m := Model{
		Ready:    true,
		Width:    80,
		Height:   10,
		Viewport: newViewport(80, 7),
		Loading:  map[string]bool{},
		Errors:   map[string]error{},
	}

	view := stripANSI(m.View())
	footer := stripANSI(m.renderStatusbar())

	// O rodapé (que pode quebrar em mais de uma linha) deve fechar a view,
	// abaixo do conteúdo da lista.
	if !strings.HasSuffix(view, footer) {
		t.Fatalf("rodape deveria fechar a view; footer %q; view %q", footer, view)
	}
	for _, want := range []string{"j/k", "enter", "f+s", "f+o", "f+p", "c", "r", "q"} {
		if !strings.Contains(footer, want) {
			t.Errorf("rodape deveria conter %q; footer %q", want, footer)
		}
	}
}

func TestCommentsStatusbarShowsAllCurrentScreenBindKeys(t *testing.T) {
	m := Model{Width: 140}
	got := stripANSI(m.renderCommentsStatusbar())

	for _, want := range []string{"j/k", "ctrl+d/u", "tab", "shift+tab", "r", "c", "a", "backspace", "esc", "q", "ctrl+c"} {
		if !strings.Contains(got, want) {
			t.Errorf("renderCommentsStatusbar() deveria conter %q; got %q", want, got)
		}
	}
}
