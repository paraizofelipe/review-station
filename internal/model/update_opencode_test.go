package model

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"paraizofelipe/review-station/internal/config"
	"paraizofelipe/review-station/internal/gitlab"
)

func keyA() tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")}
}

func TestUpdateCommentsAStartsInlineWhenConfigured(t *testing.T) {
	m := makeTestModel()
	m.Screen = ScreenComments
	m.Config = config.Config{OpenCode: config.OpenCodeConfig{Command: "opencode run 'MR {{.IID}}'"}}
	m.ActiveRepo = config.Repo{Path: "org/app", Name: "app", Local: "~/projects/app"}
	m.ActiveMR = &gitlab.MergeRequest{IID: 9}

	got, cmd := m.updateComments(keyA())
	result := got.(Model)

	if cmd == nil {
		t.Fatal("esperava tea.Cmd (ExecProcess) para o takeover inline")
	}
	if result.OpenCodeStatus != "" {
		t.Errorf("OpenCodeStatus = %q, want vazio no takeover inline", result.OpenCodeStatus)
	}
}

func TestUpdateCommentsAWarnsWhenUnconfigured(t *testing.T) {
	m := makeTestModel()
	m.Screen = ScreenComments
	m.ActiveRepo = config.Repo{Path: "org/app", Local: "~/projects/app"}
	m.ActiveMR = &gitlab.MergeRequest{IID: 9}

	got, cmd := m.updateComments(keyA())
	result := got.(Model)

	if cmd != nil {
		t.Error("não deveria disparar cmd sem comando configurado")
	}
	if result.OpenCodeStatus != "opencode não configurado" {
		t.Errorf("OpenCodeStatus = %q, want %q", result.OpenCodeStatus, "opencode não configurado")
	}
}

func TestUpdateCommentsAWarnsWhenNoLocalPath(t *testing.T) {
	m := makeTestModel()
	m.Screen = ScreenComments
	m.Config = config.Config{OpenCode: config.OpenCodeConfig{Command: "opencode run"}}
	m.ActiveRepo = config.Repo{Path: "org/app", Local: ""}
	m.ActiveMR = &gitlab.MergeRequest{IID: 9}

	got, cmd := m.updateComments(keyA())
	result := got.(Model)

	if cmd != nil {
		t.Error("não deveria disparar cmd sem path local")
	}
	if result.OpenCodeStatus != "repo sem path local" {
		t.Errorf("OpenCodeStatus = %q, want %q", result.OpenCodeStatus, "repo sem path local")
	}
}

func TestOpenCodeLaunchedMsgErrorSetsStatus(t *testing.T) {
	m := makeTestModel()
	got, _ := m.Update(OpenCodeLaunchedMsg{Err: errFake})
	result := got.(Model)
	if result.OpenCodeStatus == "" {
		t.Error("erro de launch deveria preencher OpenCodeStatus")
	}
}

var errFake = fakeErr("falhou")

type fakeErr string

func (e fakeErr) Error() string { return string(e) }

func keyShiftA() tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("A")}
}

func TestUpdateCommentsCapitalAStartsWindowWhenConfigured(t *testing.T) {
	m := makeTestModel()
	m.Screen = ScreenComments
	m.Config = config.Config{OpenCode: config.OpenCodeConfig{Command: "opencode run 'MR {{.IID}}'"}}
	m.ActiveRepo = config.Repo{Path: "org/app", Name: "app", Local: "~/projects/app"}
	m.ActiveMR = &gitlab.MergeRequest{IID: 9}

	got, cmd := m.updateComments(keyShiftA())
	result := got.(Model)

	if cmd == nil {
		t.Fatal("esperava um tea.Cmd para abrir o opencode em janela externa")
	}
	if result.OpenCodeStatus != "opencode iniciado" {
		t.Errorf("OpenCodeStatus = %q, want %q", result.OpenCodeStatus, "opencode iniciado")
	}
}

func TestUpdateCommentsCapitalAWarnsWhenUnconfigured(t *testing.T) {
	m := makeTestModel()
	m.Screen = ScreenComments
	m.ActiveRepo = config.Repo{Path: "org/app", Local: "~/projects/app"}
	m.ActiveMR = &gitlab.MergeRequest{IID: 9}

	got, cmd := m.updateComments(keyShiftA())
	result := got.(Model)

	if cmd != nil {
		t.Error("não deveria disparar cmd sem comando configurado")
	}
	if result.OpenCodeStatus != "opencode não configurado" {
		t.Errorf("OpenCodeStatus = %q, want %q", result.OpenCodeStatus, "opencode não configurado")
	}
}
