package model

import (
	tea "github.com/charmbracelet/bubbletea"

	"paraizofelipe/review-station/internal/config"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		headerHeight := 2
		statusHeight := 1
		vpHeight := m.Height - headerHeight - statusHeight
		if vpHeight < 1 {
			vpHeight = 1
		}
		if !m.Ready {
			m.Viewport = newViewport(m.Width, vpHeight)
			m.Ready = true
		} else {
			m.Viewport.Width = m.Width
			m.Viewport.Height = vpHeight
		}
		if m.Screen == ScreenComments && len(m.Discussions) > 0 {
			m.RenderedDiscussions = buildRenderedDiscussions(m.Discussions, m.Width)
			m.Viewport.SetContent(m.RenderedDiscussions)
		}
		return m, nil

	case MRsLoadedMsg:
		m.Loading[msg.Repo.Path] = false
		updated := false
		for i, pg := range m.Projects {
			if pg.Repo.Path == msg.Repo.Path {
				m.Projects[i].MRs = msg.MRs
				updated = true
				break
			}
		}
		if !updated {
			m.Projects = append(m.Projects, ProjectGroup{Repo: msg.Repo, MRs: msg.MRs})
		}
		m.Projects = sortProjectsByConfig(m.Projects, m.Config.Repos)
		m.Items = RebuildItems(m.Projects)
		return m, nil

	case FetchErrorMsg:
		m.Loading[msg.Repo.Path] = false
		m.Errors[msg.Repo.Path] = msg.Err
		return m, nil

	case DiscussionsLoadedMsg:
		if msg.Token != m.fetchToken {
			return m, nil
		}
		m.CommentsLoading = false
		m.Discussions = msg.Discussions
		m.RenderedDiscussions = buildRenderedDiscussions(m.Discussions, m.Width)
		m.Viewport.SetContent(m.RenderedDiscussions)
		return m, nil

	case FetchDiscussionsErrorMsg:
		if msg.Token != m.fetchToken {
			return m, nil
		}
		m.CommentsLoading = false
		m.CommentsError = msg.Err
		return m, nil

	case tea.KeyMsg:
		if m.Screen == ScreenComments {
			return m.updateComments(msg)
		}
		if m.ShowFilter {
			return m.updateFilter(msg)
		}
		return m.updateList(msg)
	}

	return m, nil
}

func (m Model) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "j", "down":
		next := m.Cursor + 1
		for next < len(m.Items) && m.Items[next].Kind == ItemHeader {
			next++
		}
		if next < len(m.Items) {
			m.Cursor = next
			m.syncViewport()
		}

	case "k", "up":
		prev := m.Cursor - 1
		for prev >= 0 && m.Items[prev].Kind == ItemHeader {
			prev--
		}
		if prev >= 0 {
			m.Cursor = prev
			m.syncViewport()
		}

	case "enter":
		if m.Cursor >= 0 && m.Cursor < len(m.Items) && m.Items[m.Cursor].Kind == ItemMR {
			item := m.Items[m.Cursor]
			m.fetchToken++
			m.Screen = ScreenComments
			m.ActiveMR = item.MR
			m.ActiveRepo = item.Repo
			m.Discussions = nil
			m.CommentsLoading = true
			m.CommentsError = nil
			m.RenderedDiscussions = ""
			m.listScrollOffset = m.Viewport.YOffset
			m.Viewport.YOffset = 0
			return m, fetchDiscussionsCmd(m.Client, item.Repo, item.MR.IID, m.fetchToken)
		}

	case "f":
		m.ShowFilter = true

	case "r":
		for _, repo := range m.Config.Repos {
			m.Loading[repo.Path] = true
			delete(m.Errors, repo.Path)
		}
		m.Projects = nil
		m.Items = nil
		m.Cursor = 0
		return m, m.Init()
	}

	return m, nil
}

func (m Model) updateFilter(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		if m.FilterMenu.Selected < len(m.FilterMenu.Options)-1 {
			m.FilterMenu.Selected++
		}

	case "k", "up":
		if m.FilterMenu.Selected > 0 {
			m.FilterMenu.Selected--
		}

	case "enter":
		m.Filter = m.FilterMenu.Options[m.FilterMenu.Selected]
		m.ShowFilter = false
		for _, repo := range m.Config.Repos {
			m.Loading[repo.Path] = true
			delete(m.Errors, repo.Path)
		}
		m.Projects = nil
		m.Items = nil
		m.Cursor = 0
		return m, m.Init()

	case "esc", "f":
		m.ShowFilter = false
	}

	return m, nil
}

func (m Model) updateComments(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "backspace", "esc":
		m.Screen = ScreenList
		m.Viewport.YOffset = m.listScrollOffset
		return m, nil
	case "q", "ctrl+c":
		return m, tea.Quit
	}
	var cmd tea.Cmd
	m.Viewport, cmd = m.Viewport.Update(msg)
	return m, cmd
}

func sortProjectsByConfig(projects []ProjectGroup, repos []config.Repo) []ProjectGroup {
	order := make(map[string]int, len(repos))
	for i, r := range repos {
		order[r.Path] = i
	}
	sorted := make([]ProjectGroup, len(projects))
	copy(sorted, projects)
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if order[sorted[i].Repo.Path] > order[sorted[j].Repo.Path] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	return sorted
}

func (m *Model) syncViewport() {
	if !m.Ready {
		return
	}
	const lineHeight = 2
	cursorStart := m.Cursor * lineHeight
	cursorEnd := cursorStart + lineHeight

	scrollTop := cursorStart
	if m.Cursor > 0 && m.Items[m.Cursor-1].Kind == ItemHeader {
		scrollTop = (m.Cursor - 1) * lineHeight
	}

	if scrollTop < m.Viewport.YOffset {
		m.Viewport.YOffset = scrollTop
	} else if cursorEnd > m.Viewport.YOffset+m.Viewport.Height {
		m.Viewport.YOffset = cursorEnd - m.Viewport.Height
	}
}
