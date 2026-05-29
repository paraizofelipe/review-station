package model

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"paraizofelipe/review-station/internal/config"
	"paraizofelipe/review-station/internal/gitlab"
)

func makeItemsWithProjects() []ListItem {
	backend := config.Repo{Name: "Backend", Path: "org/backend"}
	frontend := config.Repo{Name: "Frontend", Path: "org/frontend"}
	backendGroup := &ProjectGroup{Repo: backend}
	frontendGroup := &ProjectGroup{Repo: frontend}
	return []ListItem{
		{Kind: ItemHeader, Repo: backend, Project: backendGroup},
		{Kind: ItemMR, Repo: backend, MR: &gitlab.MergeRequest{IID: 1, Title: "backend fix", Author: "ana"}},
		{Kind: ItemHeader, Repo: frontend, Project: frontendGroup},
		{Kind: ItemMR, Repo: frontend, MR: &gitlab.MergeRequest{IID: 2, Title: "frontend fix", Author: "bia"}},
		{Kind: ItemMR, Repo: frontend, MR: &gitlab.MergeRequest{IID: 3, Title: "frontend polish", Author: "caio"}},
	}
}

func TestFilterItemsByProjectKeepsOnlySelectedProjectGroup(t *testing.T) {
	items := makeItemsWithProjects()

	got := filterItemsByProject(items, "org/frontend")

	if len(got) != 3 {
		t.Fatalf("esperava header + 2 MRs do projeto selecionado, got %d", len(got))
	}
	for _, item := range got {
		if item.Repo.Path != "org/frontend" {
			t.Fatalf("item de outro projeto apareceu no filtro: %+v", item.Repo)
		}
	}
	if got[0].Kind != ItemHeader {
		t.Fatalf("primeiro item filtrado deveria ser o header do projeto, got %+v", got[0])
	}
}

func TestUpdateFilterChordPOpensProjectFilterList(t *testing.T) {
	m := makeTestModel()
	m.Config.Repos = []config.Repo{{Name: "Backend", Path: "org/backend"}, {Name: "Frontend", Path: "org/frontend"}}
	m.FilterChordPending = true

	got, _ := m.updateFilterChord(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("p")})
	result := got.(Model)

	if !result.ShowProjectFilter {
		t.Fatal("pressionar 'p' no chord deveria abrir a lista de projetos")
	}
	if result.FilterChordPending {
		t.Fatal("chord deveria ser limpo apos segunda tecla")
	}
	if len(result.ProjectFilterMenu.Options) != 2 {
		t.Fatalf("lista de projetos deveria vir da configuração, got %d opções", len(result.ProjectFilterMenu.Options))
	}
}

func TestUpdateProjectFilterEnterAppliesSelectedProject(t *testing.T) {
	m := makeTestModel()
	m.ShowProjectFilter = true
	m.ProjectFilterMenu.Options = []config.Repo{{Name: "Backend", Path: "org/backend"}, {Name: "Frontend", Path: "org/frontend"}}
	m.ProjectFilterMenu.Selected = 1
	m.Cursor = 3

	got, _ := m.updateProjectFilter(tea.KeyMsg{Type: tea.KeyEnter})
	result := got.(Model)

	if result.ShowProjectFilter {
		t.Fatal("Enter deveria fechar o overlay de projeto")
	}
	if result.ProjectFilter != "org/frontend" {
		t.Fatalf("ProjectFilter deveria ser org/frontend, got %q", result.ProjectFilter)
	}
	if result.Cursor != 0 {
		t.Fatalf("Cursor deveria voltar para 0 depois de filtrar projeto, got %d", result.Cursor)
	}
}

func TestRenderProjectFilterOverlayShowsSelectableProjects(t *testing.T) {
	m := makeTestModel()
	m.Width = 80
	m.Viewport = newViewport(80, 10)
	m.ProjectFilterMenu.Options = []config.Repo{{Name: "Backend", Path: "org/backend"}, {Name: "Frontend", Path: "org/frontend"}}
	m.ProjectFilterMenu.Selected = 1

	got := stripANSI(m.renderProjectFilterOverlay())

	for _, want := range []string{"Filtrar por projeto", "Backend", "org/backend", "> Frontend", "org/frontend"} {
		if !strings.Contains(got, want) {
			t.Fatalf("overlay deveria conter %q; got %q", want, got)
		}
	}
}
