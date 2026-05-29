package ui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestStatusBarHasNoBackground(t *testing.T) {
	// A linha da legenda não deve ter background próprio: ela herda o fundo do
	// terminal, evitando manchas pretas desencontradas com o resto da tela.
	if StyleStatusBar.GetBackground() != (lipgloss.NoColor{}) {
		t.Fatalf("StyleStatusBar não deveria ter background; got %#v", StyleStatusBar.GetBackground())
	}
}

func TestBindKeyUsesSubtleGruvboxColors(t *testing.T) {
	wantBg := lipgloss.Color("#665c54") // gruvbox bg3 — discreto, mas ainda visível
	if ColorBindKeyBg != wantBg {
		t.Fatalf("ColorBindKeyBg = %v, want subtle gruvbox color %v", ColorBindKeyBg, wantBg)
	}
	if StyleBindKey.GetBackground() != wantBg {
		t.Fatalf("StyleBindKey background = %v, want %v", StyleBindKey.GetBackground(), wantBg)
	}
	if StyleBindKey.GetForeground() != ColorFg {
		t.Fatalf("StyleBindKey foreground = %v, want %v", StyleBindKey.GetForeground(), ColorFg)
	}
}
