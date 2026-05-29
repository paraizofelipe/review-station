package ui

import "testing"

func TestStatusBarHasDedicatedBackgroundColor(t *testing.T) {
	if StyleStatusBar.GetBackground() != ColorStatusBarBg {
		t.Fatalf("StyleStatusBar background = %v, want %v", StyleStatusBar.GetBackground(), ColorStatusBarBg)
	}
}
