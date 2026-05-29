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

func TestCommentsStatusbarShowsInlineBind(t *testing.T) {
	m := makeTestModel()
	m.Width = 200

	got := stripANSI(m.renderCommentsStatusbar())

	if !strings.Contains(got, "a opencode inline") {
		t.Errorf("keybar deveria conter 'a opencode inline'; got %q", got)
	}
}

func TestCommentsStatusbarShowsWindowBind(t *testing.T) {
	m := makeTestModel()
	m.Width = 200

	got := stripANSI(m.renderCommentsStatusbar())

	if !strings.Contains(got, "A janela") {
		t.Errorf("keybar deveria conter 'A janela'; got %q", got)
	}
}
