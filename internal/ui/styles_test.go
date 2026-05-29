package ui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestStatusBarHasDedicatedBackgroundColor(t *testing.T) {
	if StyleStatusBar.GetBackground() != ColorStatusBarBg {
		t.Fatalf("StyleStatusBar background = %v, want %v", StyleStatusBar.GetBackground(), ColorStatusBarBg)
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
