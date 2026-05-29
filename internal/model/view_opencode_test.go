package model

import (
	"strings"
	"testing"
)

func TestCommentsStatusbarShowsOpenCodeBind(t *testing.T) {
	m := makeTestModel()
	m.Width = 200

	got := stripANSI(m.renderCommentsStatusbar())

	if !strings.Contains(got, "a opencode") {
		t.Errorf("keybar dos comentários deveria conter 'a opencode'; got %q", got)
	}
}

func TestCommentsStatusbarShowsOpenCodeStatus(t *testing.T) {
	m := makeTestModel()
	m.Width = 200
	m.OpenCodeStatus = "opencode iniciado"

	got := stripANSI(m.renderCommentsStatusbar())

	if !strings.Contains(got, "opencode iniciado") {
		t.Errorf("keybar deveria mostrar o status do opencode; got %q", got)
	}
}
