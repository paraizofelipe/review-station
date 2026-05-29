package model

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"

	"paraizofelipe/review-station/internal/config"
	"paraizofelipe/review-station/internal/gitlab"
	"paraizofelipe/review-station/internal/launcher"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		headerHeight := 2
		vpHeight := m.Height - headerHeight - keyBarHeight
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
			m.RenderedDiscussions, m.CommentOffsets = buildRenderedDiscussions(m.ActiveMR, m.Discussions, m.Diffs, m.Width, m.CommentCursor, m.CollapsedComments)
			m.Viewport.SetContent(m.RenderedDiscussions)
		}
		if m.Screen == ScreenReply {
			m.ReplyInput.SetWidth(m.Width - 2)
			m.ReplyInput.SetHeight(replyTextareaHeight(m.Height))
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
		m.CollapsedComments = preCollapseResolved(msg.Discussions)
		m.RenderedDiscussions, m.CommentOffsets = buildRenderedDiscussions(m.ActiveMR, m.Discussions, m.Diffs, m.Width, m.CommentCursor, m.CollapsedComments)
		m.Viewport.SetContent(m.RenderedDiscussions)
		return m, nil

	case FetchDiscussionsErrorMsg:
		if msg.Token != m.fetchToken {
			return m, nil
		}
		m.CommentsLoading = false
		m.CommentsError = msg.Err
		return m, nil

	case DiffsLoadedMsg:
		if msg.Token != m.fetchToken {
			return m, nil
		}
		m.DiffsLoading = false
		m.Diffs = msg.Diffs
		if !m.CommentsLoading && len(m.Discussions) > 0 {
			m.RenderedDiscussions, m.CommentOffsets = buildRenderedDiscussions(m.ActiveMR, m.Discussions, m.Diffs, m.Width, m.CommentCursor, m.CollapsedComments)
			m.Viewport.SetContent(m.RenderedDiscussions)
		}
		return m, nil

	case FetchDiffsErrorMsg:
		if msg.Token != m.fetchToken {
			return m, nil
		}
		m.DiffsLoading = false
		// Diffs são opcionais: não bloqueia exibição dos comentários.
		return m, nil

	case OpenCodeLaunchedMsg:
		if msg.Err != nil {
			m.OpenCodeStatus = "opencode falhou: " + msg.Err.Error()
		}
		return m, nil

	case ReplySuccessMsg:
		if msg.Token != m.fetchToken {
			return m, nil
		}
		m.ReplySending = false
		m.Screen = ScreenComments
		m.ReplyInput.Blur()
		m.ReplyInput.Reset()
		m.CommentsLoading = true
		return m, fetchDiscussionsCmd(m.Client, m.ActiveRepo, m.ActiveMR.IID, m.fetchToken)

	case ReplyErrorMsg:
		if msg.Token != m.fetchToken {
			return m, nil
		}
		m.ReplySending = false
		m.ReplyError = msg.Err
		return m, nil

	case tea.KeyMsg:
		if m.Screen == ScreenReply {
			return m.updateReply(msg)
		}
		if m.Screen == ScreenComments {
			return m.updateComments(msg)
		}
		if m.FilterChordPending {
			return m.updateFilterChord(msg)
		}
		if m.ShowOwnerFilter {
			return m.updateOwnerFilter(msg)
		}
		if m.ShowProjectFilter {
			return m.updateProjectFilter(msg)
		}
		if m.ShowFilter {
			return m.updateFilter(msg)
		}
		return m.updateList(msg)
	}

	// Repassa mensagens não-teclado para o textarea quando na tela de reply
	// (necessário para animação do cursor e outros eventos internos do widget).
	if m.Screen == ScreenReply {
		var cmd tea.Cmd
		m.ReplyInput, cmd = m.ReplyInput.Update(msg)
		return m, cmd
	}

	if m.ShowOwnerFilter {
		var cmd tea.Cmd
		m.OwnerFilterInput, cmd = m.OwnerFilterInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	items := m.visibleItems()

	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "j", "down":
		next := m.Cursor + 1
		for next < len(items) && items[next].Kind == ItemHeader {
			next++
		}
		if next < len(items) {
			m.Cursor = next
			m.syncViewport()
		}

	case "k", "up":
		prev := m.Cursor - 1
		for prev >= 0 && items[prev].Kind == ItemHeader {
			prev--
		}
		if prev >= 0 {
			m.Cursor = prev
			m.syncViewport()
		}

	case "enter":
		if m.Cursor >= 0 && m.Cursor < len(items) && items[m.Cursor].Kind == ItemMR {
			item := items[m.Cursor]
			m.fetchToken++
			m.Screen = ScreenComments
			m.ActiveMR = item.MR
			m.ActiveRepo = item.Repo
			m.Discussions = nil
			m.Diffs = nil
			m.CommentsLoading = true
			m.CommentsError = nil
			m.DiffsLoading = true
			m.RenderedDiscussions = ""
			m.CommentCursor = -1
			m.CommentOffsets = nil
			m.CollapsedComments = nil
			m.OpenCodeStatus = ""
			m.listScrollOffset = m.Viewport.YOffset
			m.Viewport.YOffset = 0
			return m, tea.Batch(
				fetchDiscussionsCmd(m.Client, item.Repo, item.MR.IID, m.fetchToken),
				fetchDiffsCmd(m.Client, item.Repo, item.MR.IID, m.fetchToken),
			)
		}

	case "f":
		m.FilterChordPending = true

	case "c":
		m.Filter = FilterAll
		m.OwnerFilter = ""
		m.ProjectFilter = ""
		m.OwnerFilterInput.Reset()
		m.ProjectFilterMenu.Selected = 0
		for _, repo := range m.Config.Repos {
			m.Loading[repo.Path] = true
			delete(m.Errors, repo.Path)
		}
		m.Projects = nil
		m.Items = nil
		m.Cursor = 0
		m.Viewport.YOffset = 0
		return m, m.Init()

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

func (m Model) updateFilterChord(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.FilterChordPending = false
	switch msg.String() {
	case "o":
		m.ShowOwnerFilter = true
		m.OwnerFilterInput.SetValue(m.OwnerFilter)
		return m, m.OwnerFilterInput.Focus()
	case "p":
		m.ShowProjectFilter = true
		m.ProjectFilterMenu.Options = m.projectFilterOptions()
		m.ProjectFilterMenu.Selected = m.selectedProjectFilterIndex()
	case "s", "f":
		m.ShowFilter = true
	}
	return m, nil
}

func (m Model) updateProjectFilter(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if len(m.ProjectFilterMenu.Options) == 0 {
		m.ProjectFilterMenu.Options = m.projectFilterOptions()
	}
	switch msg.String() {
	case "j", "down":
		if m.ProjectFilterMenu.Selected < len(m.ProjectFilterMenu.Options)-1 {
			m.ProjectFilterMenu.Selected++
		}
	case "k", "up":
		if m.ProjectFilterMenu.Selected > 0 {
			m.ProjectFilterMenu.Selected--
		}
	case "enter":
		if len(m.ProjectFilterMenu.Options) > 0 {
			m.ProjectFilter = m.ProjectFilterMenu.Options[m.ProjectFilterMenu.Selected].Path
		}
		m.ShowProjectFilter = false
		m.Cursor = 0
		return m, nil
	case "esc":
		m.ProjectFilter = ""
		m.ShowProjectFilter = false
		m.ProjectFilterMenu.Selected = 0
		m.Cursor = 0
		return m, nil
	}
	return m, nil
}

func (m Model) updateOwnerFilter(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.OwnerFilter = m.OwnerFilterInput.Value()
		m.ShowOwnerFilter = false
		m.OwnerFilterInput.Blur()
		m.Cursor = 0
		return m, nil
	case "esc":
		m.OwnerFilter = ""
		m.ShowOwnerFilter = false
		m.OwnerFilterInput.Reset()
		m.OwnerFilterInput.Blur()
		m.Cursor = 0
		return m, nil
	}
	var cmd tea.Cmd
	m.OwnerFilterInput, cmd = m.OwnerFilterInput.Update(msg)
	return m, cmd
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
		m.CommentCursor = -1
		m.Viewport.YOffset = m.listScrollOffset
		return m, nil
	case "q", "ctrl+c":
		return m, tea.Quit
	case "tab":
		if n := countUserDiscussions(m.Discussions); n > 0 {
			if m.CommentCursor < 0 {
				m.CommentCursor = 0
			} else {
				m.CommentCursor = (m.CommentCursor + 1) % n
			}
			m.RenderedDiscussions, m.CommentOffsets = buildRenderedDiscussions(m.ActiveMR, m.Discussions, m.Diffs, m.Width, m.CommentCursor, m.CollapsedComments)
			m.Viewport.SetContent(m.RenderedDiscussions)
			if m.CommentCursor < len(m.CommentOffsets) {
				m.Viewport.YOffset = m.CommentOffsets[m.CommentCursor]
			}
		}
		return m, nil
	case "shift+tab":
		if n := countUserDiscussions(m.Discussions); n > 0 {
			if m.CommentCursor < 0 {
				m.CommentCursor = n - 1
			} else {
				m.CommentCursor = (m.CommentCursor - 1 + n) % n
			}
			m.RenderedDiscussions, m.CommentOffsets = buildRenderedDiscussions(m.ActiveMR, m.Discussions, m.Diffs, m.Width, m.CommentCursor, m.CollapsedComments)
			m.Viewport.SetContent(m.RenderedDiscussions)
			if m.CommentCursor < len(m.CommentOffsets) {
				m.Viewport.YOffset = m.CommentOffsets[m.CommentCursor]
			}
		}
		return m, nil
	case "r":
		if m.CommentCursor >= 0 && !m.CommentsLoading {
			if discID := getDiscussionID(m.Discussions, m.CommentCursor); discID != "" {
				m.ReplyDiscussionID = discID
				m.ReplyError = nil
				m.ReplySending = false
				m.Screen = ScreenReply
				m.ReplyInput.Reset()
				m.ReplyInput.SetWidth(m.Width - 2)
				m.ReplyInput.SetHeight(replyTextareaHeight(m.Height))
				m.ReplyInput.Focus()
				return m, textarea.Blink
			}
		}
		return m, nil
	case "c":
		if m.CommentCursor >= 0 && !m.CommentsLoading {
			if m.CollapsedComments == nil {
				m.CollapsedComments = make(map[int]bool)
			}
			m.CollapsedComments[m.CommentCursor] = !m.CollapsedComments[m.CommentCursor]
			m.RenderedDiscussions, m.CommentOffsets = buildRenderedDiscussions(m.ActiveMR, m.Discussions, m.Diffs, m.Width, m.CommentCursor, m.CollapsedComments)
			m.Viewport.SetContent(m.RenderedDiscussions)
			if m.CommentCursor < len(m.CommentOffsets) {
				m.Viewport.YOffset = m.CommentOffsets[m.CommentCursor]
			}
		}
		return m, nil
	case "a":
		command, err := resolveOpenCodeCommand(m.ActiveRepo, m.ActiveMR, m.Config)
		if err != nil {
			m.OpenCodeStatus = "opencode: template inválido"
			return m, nil
		}
		if command == "" {
			m.OpenCodeStatus = "opencode não configurado"
			return m, nil
		}
		if m.ActiveRepo.Local == "" {
			m.OpenCodeStatus = "repo sem path local"
			return m, nil
		}
		m.OpenCodeStatus = "opencode iniciado"
		env := buildOpenCodeEnv(m.ActiveRepo, m.ActiveMR)
		window := fmt.Sprintf("rs-review-%d", m.ActiveMR.IID)
		return m, launchOpenCodeCmd(command, m.ActiveRepo.Local, window, env)
	}
	var cmd tea.Cmd
	m.Viewport, cmd = m.Viewport.Update(msg)
	return m, cmd
}

func countUserDiscussions(discussions []gitlab.Discussion) int {
	n := 0
	for _, d := range discussions {
		if len(d.Notes) > 0 && !d.Notes[0].System {
			n++
		}
	}
	return n
}

// ReplySuccessMsg é emitido quando o envio de uma resposta tem sucesso.
type ReplySuccessMsg struct{ Token int64 }

// ReplyErrorMsg é emitido quando o envio de uma resposta falha.
type ReplyErrorMsg struct {
	Token int64
	Err   error
}

func (m Model) updateReply(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.Screen = ScreenComments
		m.ReplyInput.Blur()
		return m, nil
	case "ctrl+s":
		body := strings.TrimSpace(m.ReplyInput.Value())
		if body != "" && !m.ReplySending {
			m.ReplySending = true
			m.ReplyError = nil
			return m, sendReplyCmd(m.Client, m.ActiveRepo, m.ActiveMR.IID, m.ReplyDiscussionID, body, m.fetchToken)
		}
		return m, nil
	}
	var cmd tea.Cmd
	m.ReplyInput, cmd = m.ReplyInput.Update(msg)
	return m, cmd
}

func launchOpenCodeCmd(command, local, window string, env map[string]string) tea.Cmd {
	return func() tea.Msg {
		return OpenCodeLaunchedMsg{Err: launcher.Launch(command, local, window, env)}
	}
}

func sendReplyCmd(client gitlab.Client, repo config.Repo, mrIID int, discussionID, body string, token int64) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		err := client.ReplyToDiscussion(ctx, repo, mrIID, discussionID, body)
		if err != nil {
			return ReplyErrorMsg{Token: token, Err: err}
		}
		return ReplySuccessMsg{Token: token}
	}
}

// getDiscussionID retorna o ID da discussion no índice cursor (base-0
// nos comentários de usuário).
func getDiscussionID(discussions []gitlab.Discussion, cursor int) string {
	idx := 0
	for _, d := range discussions {
		if len(d.Notes) == 0 || d.Notes[0].System {
			continue
		}
		if idx == cursor {
			return d.ID
		}
		idx++
	}
	return ""
}

// preCollapseResolved constrói o mapa inicial de colapso com todos os
// comentários resolvidos já colapsados.
func preCollapseResolved(discussions []gitlab.Discussion) map[int]bool {
	collapsed := make(map[int]bool)
	idx := 0
	for _, d := range discussions {
		if len(d.Notes) == 0 || d.Notes[0].System {
			continue
		}
		if d.Notes[0].Resolvable && d.Notes[0].Resolved {
			collapsed[idx] = true
		}
		idx++
	}
	return collapsed
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
