package model

import (
	tea "github.com/charmbracelet/bubbletea"
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
		m.Items = RebuildItems(m.Projects)
		return m, nil

	case FetchErrorMsg:
		m.Loading[msg.Repo.Path] = false
		m.Errors[msg.Repo.Path] = msg.Err
		return m, nil

	case tea.KeyMsg:
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
		if m.Cursor < len(m.Items)-1 {
			m.Cursor++
			m.syncViewport()
		}

	case "k", "up":
		if m.Cursor > 0 {
			m.Cursor--
			m.syncViewport()
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

func (m *Model) syncViewport() {
	if !m.Ready {
		return
	}
	lineHeight := 3
	cursorLine := m.Cursor * lineHeight
	if cursorLine < m.Viewport.YOffset {
		m.Viewport.YOffset = cursorLine
	} else if cursorLine >= m.Viewport.YOffset+m.Viewport.Height {
		m.Viewport.YOffset = cursorLine - m.Viewport.Height + lineHeight
	}
}
