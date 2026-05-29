package model

import (
	"context"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"paraizofelipe/review-station/internal/config"
	"paraizofelipe/review-station/internal/gitlab"
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

type Screen int

const (
	ScreenList Screen = iota
	ScreenComments
	ScreenReply
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

type ProjectFilterMenu struct {
	Options  []config.Repo
	Selected int
}

type Model struct {
	Config              config.Config
	Client              gitlab.Client
	Projects            []ProjectGroup
	Loading             map[string]bool
	Errors              map[string]error
	Cursor              int
	Items               []ListItem
	Filter              FilterState
	ShowFilter          bool
	FilterMenu          FilterMenu
	FilterChordPending  bool
	ShowOwnerFilter     bool
	OwnerFilter         string
	OwnerFilterInput    textinput.Model
	ShowProjectFilter   bool
	ProjectFilter       string
	ProjectFilterMenu   ProjectFilterMenu
	Viewport            viewport.Model
	Width               int
	Height              int
	Ready               bool
	Screen              Screen
	ActiveMR            *gitlab.MergeRequest
	ActiveRepo          config.Repo
	Discussions         []gitlab.Discussion
	Diffs               []gitlab.FileDiff
	CommentsLoading     bool
	CommentsError       error
	DiffsLoading        bool
	RenderedDiscussions string
	CommentCursor       int
	CommentOffsets      []int
	CollapsedComments   map[int]bool
	ReplyInput          textarea.Model
	ReplyDiscussionID   string
	ReplySending        bool
	ReplyError          error
	fetchToken          int64
	listScrollOffset    int
	OpenCodeStatus      string
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

type DiscussionsLoadedMsg struct {
	Token       int64
	Discussions []gitlab.Discussion
}

type FetchDiscussionsErrorMsg struct {
	Token int64
	Err   error
}

type DiffsLoadedMsg struct {
	Token int64
	Diffs []gitlab.FileDiff
}

type FetchDiffsErrorMsg struct {
	Token int64
	Err   error
}

// OpenCodeLaunchedMsg é emitido após tentar disparar a sessão opencode.
type OpenCodeLaunchedMsg struct{ Err error }

func New(cfg config.Config, client gitlab.Client) Model {
	loading := make(map[string]bool, len(cfg.Repos))
	errors := make(map[string]error, len(cfg.Repos))
	for _, r := range cfg.Repos {
		loading[r.Path] = true
	}
	return Model{
		Config:           cfg,
		Client:           client,
		Loading:          loading,
		Errors:           errors,
		Filter:           FilterOpened,
		CommentCursor:    -1,
		ReplyInput:       newReplyInput(),
		OwnerFilterInput: newOwnerFilterInput(),
		FilterMenu: FilterMenu{
			Options: []FilterState{FilterOpened, FilterClosed, FilterMerged, FilterAll},
		},
		ProjectFilterMenu: ProjectFilterMenu{Options: cfg.Repos},
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

func fetchDiscussionsCmd(client gitlab.Client, repo config.Repo, mrIID int, token int64) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		discussions, err := client.FetchDiscussions(ctx, repo, mrIID)
		if err != nil {
			discussions, err = gitlab.FetchDiscussionsFallback(ctx, repo, mrIID)
			if err != nil {
				return FetchDiscussionsErrorMsg{Token: token, Err: err}
			}
		}
		return DiscussionsLoadedMsg{Token: token, Discussions: discussions}
	}
}

func fetchDiffsCmd(client gitlab.Client, repo config.Repo, mrIID int, token int64) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		diffs, err := client.FetchDiffs(ctx, repo, mrIID)
		if err != nil {
			diffs, err = gitlab.FetchDiffsFallback(ctx, repo, mrIID)
			if err != nil {
				return FetchDiffsErrorMsg{Token: token, Err: err}
			}
		}
		return DiffsLoadedMsg{Token: token, Diffs: diffs}
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
	return viewport.New(w, h)
}

func newOwnerFilterInput() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "filtrar por owner..."
	ti.CharLimit = 64
	return ti
}

func filterItemsByOwner(items []ListItem, substr string) []ListItem {
	if substr == "" {
		return items
	}
	lower := strings.ToLower(substr)
	result := make([]ListItem, 0, len(items))
	for _, item := range items {
		if item.Kind == ItemHeader {
			result = append(result, item)
			continue
		}
		if item.MR != nil && strings.Contains(strings.ToLower(item.MR.Author), lower) {
			result = append(result, item)
		}
	}
	return result
}

func filterItemsByProject(items []ListItem, projectPath string) []ListItem {
	if projectPath == "" {
		return items
	}
	result := make([]ListItem, 0, len(items))
	for _, item := range items {
		if item.Repo.Path == projectPath {
			result = append(result, item)
		}
	}
	return result
}

func (m Model) visibleItems() []ListItem {
	items := filterItemsByProject(m.Items, m.ProjectFilter)
	return filterItemsByOwner(items, m.OwnerFilter)
}
