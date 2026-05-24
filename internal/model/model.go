package model

import (
	"context"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"paraizofelipe/review-station/internal/config"
	"paraizofelipe/review-station/internal/gitlab"
	"paraizofelipe/review-station/internal/ui"
)

type FilterState string

const (
	FilterOpened FilterState = "opened"
	FilterClosed FilterState = "closed"
	FilterMerged FilterState = "merged"
	FilterAll    FilterState = "all"
)

type ListItemKind int

const (
	ItemHeader ListItemKind = iota
	ItemMR
)

type ListItem struct {
	Kind    ListItemKind
	Project *ProjectGroup
	MR      *gitlab.MergeRequest
	Repo    config.Repo
}

type ProjectGroup struct {
	Repo config.Repo
	MRs  []gitlab.MergeRequest
}

type FilterMenu struct {
	Options  []FilterState
	Selected int
}

type Model struct {
	Config     config.Config
	Client     gitlab.Client
	Projects   []ProjectGroup
	Loading    map[string]bool
	Errors     map[string]error
	Cursor     int
	Items      []ListItem
	Filter     FilterState
	ShowFilter bool
	FilterMenu FilterMenu
	Viewport   viewport.Model
	Width      int
	Height     int
	Ready      bool
}

// Bubbletea messages

type MRsLoadedMsg struct {
	Repo config.Repo
	MRs  []gitlab.MergeRequest
}

type FetchErrorMsg struct {
	Repo config.Repo
	Err  error
}

func New(cfg config.Config, client gitlab.Client) Model {
	loading := make(map[string]bool, len(cfg.Repos))
	errors := make(map[string]error, len(cfg.Repos))
	for _, r := range cfg.Repos {
		loading[r.Path] = true
	}
	return Model{
		Config:  cfg,
		Client:  client,
		Loading: loading,
		Errors:  errors,
		Filter:  FilterOpened,
		FilterMenu: FilterMenu{
			Options: []FilterState{FilterOpened, FilterClosed, FilterMerged, FilterAll},
		},
	}
}

func (m Model) Init() tea.Cmd {
	cmds := make([]tea.Cmd, 0, len(m.Config.Repos))
	for _, repo := range m.Config.Repos {
		repo := repo
		cmds = append(cmds, fetchMRsCmd(m.Client, repo, string(m.Filter)))
	}
	return tea.Batch(cmds...)
}

func fetchMRsCmd(client gitlab.Client, repo config.Repo, state string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		mrs, err := client.FetchMRs(ctx, repo, state)
		if err != nil {
			mrs, err = gitlab.FetchMRsFallback(ctx, repo, state)
			if err != nil {
				return FetchErrorMsg{Repo: repo, Err: err}
			}
		}
		return MRsLoadedMsg{Repo: repo, MRs: mrs}
	}
}

func RebuildItems(projects []ProjectGroup) []ListItem {
	var items []ListItem
	for i := range projects {
		pg := &projects[i]
		items = append(items, ListItem{Kind: ItemHeader, Project: pg, Repo: pg.Repo})
		for j := range pg.MRs {
			items = append(items, ListItem{Kind: ItemMR, MR: &pg.MRs[j], Repo: pg.Repo})
		}
	}
	return items
}

func newViewport(w, h int) viewport.Model {
	vp := viewport.New(w, h)
	vp.Style = lipgloss.NewStyle().Background(ui.ColorBg)
	return vp
}

