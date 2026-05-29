package model

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"paraizofelipe/review-station/internal/config"
	"paraizofelipe/review-station/internal/gitlab"
)

func makeTestModel() Model {
	m := Model{
		Filter:        FilterOpened,
		CommentCursor: -1,
		ReplyInput:    newReplyInput(),
		FilterMenu: FilterMenu{
			Options: []FilterState{FilterOpened, FilterClosed, FilterMerged, FilterAll},
		},
		Loading: map[string]bool{},
		Errors:  map[string]error{},
	}
	m.OwnerFilterInput = newOwnerFilterInput()
	return m
}

func makeItemsWithAuthors() []ListItem {
	repo := config.Repo{Path: "org/project"}
	pg := &ProjectGroup{Repo: repo}
	return []ListItem{
		{Kind: ItemHeader, Repo: repo, Project: pg},
		{Kind: ItemMR, Repo: repo, MR: &gitlab.MergeRequest{IID: 1, Author: "felipeparaizo"}},
		{Kind: ItemMR, Repo: repo, MR: &gitlab.MergeRequest{IID: 2, Author: "joaosilva"}},
		{Kind: ItemMR, Repo: repo, MR: &gitlab.MergeRequest{IID: 3, Author: "felipecosta"}},
	}
}

func TestFilterItemsByOwner_emptyFilter(t *testing.T) {
	items := makeItemsWithAuthors()
	got := filterItemsByOwner(items, "")
	if len(got) != len(items) {
		t.Errorf("filtro vazio deveria retornar todos os items: got %d, want %d", len(got), len(items))
	}
}

func TestFilterItemsByOwner_matchSubstring(t *testing.T) {
	items := makeItemsWithAuthors()
	got := filterItemsByOwner(items, "felipe")
	mrCount := 0
	for _, item := range got {
		if item.Kind == ItemMR {
			mrCount++
			if !strings.Contains(strings.ToLower(item.MR.Author), "felipe") {
				t.Errorf("item nao deveria aparecer: author=%q", item.MR.Author)
			}
		}
	}
	if mrCount != 2 {
		t.Errorf("esperava 2 MRs com 'felipe', got %d", mrCount)
	}
}

func TestFilterItemsByOwner_caseInsensitive(t *testing.T) {
	items := makeItemsWithAuthors()
	got := filterItemsByOwner(items, "FELIPE")
	mrCount := 0
	for _, item := range got {
		if item.Kind == ItemMR {
			mrCount++
		}
	}
	if mrCount != 2 {
		t.Errorf("filtro deveria ser case-insensitive: got %d MRs", mrCount)
	}
}

func TestFilterItemsByOwner_noMatch(t *testing.T) {
	items := makeItemsWithAuthors()
	got := filterItemsByOwner(items, "xyz")
	for _, item := range got {
		if item.Kind == ItemMR {
			t.Errorf("nao deveria ter MRs com filtro sem match, author=%q", item.MR.Author)
		}
	}
}

func TestUpdateList_fKeyEntersChordPending(t *testing.T) {
	m := makeTestModel()
	got, _ := m.updateList(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("f")})
	result := got.(Model)
	if !result.FilterChordPending {
		t.Error("pressionar 'f' deveria entrar no estado FilterChordPending")
	}
	if result.ShowFilter {
		t.Error("pressionar 'f' nao deveria abrir o filtro de status diretamente")
	}
}

func TestUpdateFilterChord_oOpensOwnerFilter(t *testing.T) {
	m := makeTestModel()
	m.FilterChordPending = true
	got, _ := m.updateFilterChord(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("o")})
	result := got.(Model)
	if !result.ShowOwnerFilter {
		t.Error("pressionar 'o' no chord deveria abrir o filtro de owner")
	}
	if result.FilterChordPending {
		t.Error("chord deveria ser limpo apos segunda tecla")
	}
}

func TestUpdateFilterChord_sOpensStatusFilter(t *testing.T) {
	m := makeTestModel()
	m.FilterChordPending = true
	got, _ := m.updateFilterChord(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	result := got.(Model)
	if !result.ShowFilter {
		t.Error("pressionar 's' no chord deveria abrir o filtro de status")
	}
	if result.FilterChordPending {
		t.Error("chord deveria ser limpo apos segunda tecla")
	}
}

func TestUpdateFilterChord_escCancelsChord(t *testing.T) {
	m := makeTestModel()
	m.FilterChordPending = true
	got, _ := m.updateFilterChord(tea.KeyMsg{Type: tea.KeyEsc})
	result := got.(Model)
	if result.FilterChordPending {
		t.Error("esc deveria cancelar o chord")
	}
	if result.ShowFilter || result.ShowOwnerFilter {
		t.Error("esc no chord nao deveria abrir nenhum filtro")
	}
}

func TestUpdateOwnerFilter_enterConfirmsFilter(t *testing.T) {
	m := makeTestModel()
	m.ShowOwnerFilter = true
	m.OwnerFilterInput.SetValue("paraizo")
	got, _ := m.updateOwnerFilter(tea.KeyMsg{Type: tea.KeyEnter})
	result := got.(Model)
	if result.ShowOwnerFilter {
		t.Error("Enter deveria fechar o overlay de owner filter")
	}
	if result.OwnerFilter != "paraizo" {
		t.Errorf("OwnerFilter deveria ser 'paraizo', got %q", result.OwnerFilter)
	}
}

func TestUpdateOwnerFilter_escClearsFilter(t *testing.T) {
	m := makeTestModel()
	m.ShowOwnerFilter = true
	m.OwnerFilter = "paraizo"
	m.OwnerFilterInput.SetValue("paraizo")
	got, _ := m.updateOwnerFilter(tea.KeyMsg{Type: tea.KeyEsc})
	result := got.(Model)
	if result.ShowOwnerFilter {
		t.Error("Esc deveria fechar o overlay de owner filter")
	}
	if result.OwnerFilter != "" {
		t.Errorf("Esc deveria limpar o OwnerFilter, got %q", result.OwnerFilter)
	}
}

func TestUpdateListCKeyClearsAppliedFiltersAndReloadsAllMRs(t *testing.T) {
	m := makeTestModel()
	m.Config.Repos = []config.Repo{{Path: "org/project"}}
	m.Filter = FilterClosed
	m.OwnerFilter = "paraizo"
	m.ProjectFilter = "org/project"
	m.Items = makeItemsWithAuthors()
	m.Projects = []ProjectGroup{{Repo: m.Config.Repos[0], MRs: []gitlab.MergeRequest{{IID: 1, Author: "paraizo"}}}}
	m.Cursor = 2
	m.Loading = map[string]bool{"org/project": false}
	m.Errors = map[string]error{"org/project": errors.New("erro antigo")}

	got, cmd := m.updateList(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})
	result := got.(Model)

	if result.Filter != FilterAll {
		t.Fatalf("pressionar c deveria limpar filtro de status para all, got %q", result.Filter)
	}
	if result.OwnerFilter != "" {
		t.Fatalf("pressionar c deveria limpar OwnerFilter, got %q", result.OwnerFilter)
	}
	if result.ProjectFilter != "" {
		t.Fatalf("pressionar c deveria limpar ProjectFilter, got %q", result.ProjectFilter)
	}
	if result.Cursor != 0 {
		t.Fatalf("cursor deveria voltar para 0 depois de limpar filtros, got %d", result.Cursor)
	}
	if len(result.Items) != 0 || len(result.Projects) != 0 {
		t.Fatalf("limpar filtros deveria descartar MRs filtrados antes do reload, items=%d projects=%d", len(result.Items), len(result.Projects))
	}
	if !result.Loading["org/project"] {
		t.Fatalf("limpar filtro de status deveria recarregar todos os MRs")
	}
	if _, ok := result.Errors["org/project"]; ok {
		t.Fatalf("erros antigos deveriam ser limpos ao recarregar")
	}
	if cmd == nil {
		t.Fatalf("limpar filtros deveria disparar reload com filtro all")
	}
}
