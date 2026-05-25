package model

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"

	"paraizofelipe/review-station/internal/gitlab"
	"paraizofelipe/review-station/internal/ui"
)

func (m Model) View() string {
	if !m.Ready {
		return "inicializando..."
	}

	if m.Screen == ScreenComments {
		return m.renderCommentsView()
	}

	header := m.renderTitleBar()
	statusbar := m.renderStatusbar()

	if m.ShowFilter {
		filterArea := m.renderFilterOverlay()
		return header + "\n" + filterArea + statusbar
	}

	listContent := m.renderList()
	m.Viewport.SetContent(listContent)

	var sb strings.Builder
	sb.WriteString(header)
	sb.WriteString("\n")
	sb.WriteString(m.Viewport.View())
	sb.WriteString(statusbar)

	return sb.String()
}

func (m Model) renderTitleBar() string {
	title := ui.StyleTitleBar.Render("review-station")

	var statusStr string
	if m.anyLoading() {
		statusStr = ui.StyleMeta.Render("[carregando...]")
	} else if errCount := m.countErrors(); errCount > 0 {
		statusStr = ui.StyleError.Render(fmt.Sprintf("[%d erro(s)]", errCount))
	} else {
		statusStr = ui.StyleMeta.Render("[ok]")
	}

	gap := m.Width - lipgloss.Width(title) - lipgloss.Width(statusStr) - 2
	if gap < 1 {
		gap = 1
	}
	row := title + strings.Repeat(" ", gap) + statusStr
	return lipgloss.NewStyle().Background(ui.ColorBg).Width(m.Width).Render(row)
}

func (m Model) renderStatusbar() string {
	var keys string
	if m.ShowFilter {
		keys = "j/k navegar opções  Enter aplicar  Esc fechar"
	} else {
		keys = "j/k navegar  f filtrar  r atualizar  q sair"
	}
	return ui.StyleStatusBar.Width(m.Width).Render(keys)
}

func (m Model) renderList() string {
	if m.anyLoading() && len(m.Projects) == 0 {
		return "\n  Carregando MRs..."
	}
	if len(m.Items) == 0 {
		return "\n  Nenhum MR encontrado."
	}

	var sb strings.Builder
	for i, item := range m.Items {
		switch item.Kind {
		case ItemHeader:
			sb.WriteString(renderProjectHeader(item))
		case ItemMR:
			sb.WriteString(renderMRRow(item, i == m.Cursor, m.Width))
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

func renderProjectHeader(item ListItem) string {
	count := len(item.Project.MRs)
	noun := "MRs abertos"
	if count == 1 {
		noun = "MR aberto"
	}
	text := fmt.Sprintf("▶ %s  (%d %s)", item.Repo.Path, count, noun)
	return "\n" + ui.StyleProjectHeader.Render(text)
}

func renderMRRow(item ListItem, selected bool, width int) string {
	mr := item.MR

	cursorStr := "  "
	if selected {
		cursorStr = ui.StyleCursor.Render("> ")
	}

	mrNum := ui.StyleMRNumber.Render(fmt.Sprintf("!%d", mr.IID))
	author := ui.StyleAuthor.Render("@" + mr.Author)
	branches := ui.StyleBranch.Render(fmt.Sprintf("%s ← %s", mr.TargetBranch, mr.SourceBranch))

	// reserve space for mrNum(~5) + author(~15) + branches(~30) + padding
	titleWidth := width - 55
	if titleWidth < 10 {
		titleWidth = 10
	}
	title := mr.Title
	if len([]rune(title)) > titleWidth {
		title = string([]rune(title)[:titleWidth-1]) + "…"
	}

	line1 := fmt.Sprintf("%s%s  %s  %s  %s", cursorStr, mrNum, title, author, branches)
	line2 := fmt.Sprintf("     %s  %s", ui.StyleMeta.Render(renderAge(mr.CreatedAt)), renderCI(mr.Pipeline))

	return line1 + "\n" + line2
}

func renderAge(t time.Time) string {
	if t.IsZero() {
		return "data desconhecida"
	}
	d := time.Since(t)
	switch {
	case d < time.Hour:
		mins := int(d.Minutes())
		if mins < 1 {
			mins = 1
		}
		return fmt.Sprintf("há %d min", mins)
	case d < 24*time.Hour:
		h := int(d.Hours())
		if h == 1 {
			return "há 1 hora"
		}
		return fmt.Sprintf("há %d horas", h)
	default:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "há 1 dia"
		}
		return fmt.Sprintf("há %d dias", days)
	}
}

func renderCI(p *gitlab.PipelineStatus) string {
	if p == nil {
		return ui.StyleCINone.Render("· sem pipeline")
	}
	switch p.Status {
	case "success":
		return ui.StyleCIPassed.Render("✓ pipeline passed")
	case "failed":
		return ui.StyleCIFailed.Render("✗ pipeline failed")
	case "running":
		return ui.StyleMeta.Render("⟳ running")
	default:
		return ui.StyleMeta.Render("· " + p.Status)
	}
}

func (m Model) renderFilterOverlay() string {
	var inner strings.Builder
	inner.WriteString("── Filtrar ──\n")
	for i, opt := range m.FilterMenu.Options {
		if i == m.FilterMenu.Selected {
			inner.WriteString(ui.StylePopoverSelected.Render("> " + string(opt)))
		} else {
			inner.WriteString(ui.StylePopoverItem.Render("  " + string(opt)))
		}
		inner.WriteString("\n")
	}

	box := ui.StylePopoverBorder.Render(strings.TrimRight(inner.String(), "\n"))

	return lipgloss.Place(
		m.Width, m.Viewport.Height,
		lipgloss.Center, lipgloss.Center,
		box,
	)
}

func (m Model) renderCommentsView() string {
	header := m.renderCommentsHeader()
	statusbar := m.renderCommentsStatusbar()
	content := m.renderDiscussions()
	m.Viewport.SetContent(content)
	var sb strings.Builder
	sb.WriteString(header)
	sb.WriteString("\n")
	sb.WriteString(m.Viewport.View())
	sb.WriteString(statusbar)
	return sb.String()
}

func (m Model) renderCommentsHeader() string {
	right := ui.StyleMeta.Render("← backspace voltar")
	raw := "review-station · comentários"
	if m.ActiveMR != nil {
		raw = fmt.Sprintf("!%d  %s", m.ActiveMR.IID, m.ActiveMR.Title)
	}
	maxLeft := m.Width - lipgloss.Width(right) - 2
	if maxLeft < 1 {
		maxLeft = 1
	}
	if len([]rune(raw)) > maxLeft {
		raw = string([]rune(raw)[:maxLeft-1]) + "…"
	}
	left := ui.StyleTitleBar.Render(raw)
	gap := m.Width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}
	row := left + strings.Repeat(" ", gap) + right
	return lipgloss.NewStyle().Background(ui.ColorBg).Width(m.Width).Render(row)
}

func (m Model) renderCommentsStatusbar() string {
	return ui.StyleStatusBar.Width(m.Width).Render(
		"j/k scroll  ctrl+d/u página  backspace voltar  q sair",
	)
}

func (m Model) renderDiscussions() string {
	if m.CommentsLoading {
		return "\n  Carregando comentários..."
	}
	if m.CommentsError != nil {
		return "\n  " + ui.StyleError.Render("Erro: "+m.CommentsError.Error())
	}
	if len(m.Discussions) == 0 {
		return "\n  Nenhum comentário encontrado."
	}
	return m.RenderedDiscussions
}

func buildRenderedDiscussions(discussions []gitlab.Discussion, width int) string {
	if len(discussions) == 0 {
		return ""
	}
	contentWidth := width - 4
	if contentWidth < 20 {
		contentWidth = 20
	}
	r, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(contentWidth),
	)
	if err != nil {
		r = nil
	}

	var userDiscussions []gitlab.Discussion
	var systemNotes []gitlab.Note
	for _, d := range discussions {
		if len(d.Notes) > 0 && d.Notes[0].System {
			systemNotes = append(systemNotes, d.Notes...)
		} else {
			userDiscussions = append(userDiscussions, d)
		}
	}

	var sb strings.Builder
	for i, d := range userDiscussions {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(renderDiscussion(d, r, width))
	}
	if len(systemNotes) > 0 {
		divLen := max(width-26, 4)
		sb.WriteString(ui.StyleSectionDivider.Render(
			"\n── Atividade do sistema " + strings.Repeat("─", divLen) + "\n",
		))
		for _, n := range systemNotes {
			sb.WriteString(renderSystemNote(n))
		}
	}
	return sb.String()
}

func renderDiscussion(d gitlab.Discussion, r *glamour.TermRenderer, width int) string {
	var sb strings.Builder
	isFirst := true
	for _, note := range d.Notes {
		if note.System {
			continue
		}
		if isFirst {
			isFirst = false
			header := ui.StyleCommentAuthor.Render("@"+note.Author) +
				ui.StyleMeta.Render("  •  "+renderAge(note.CreatedAt))
			if note.Resolvable && note.Resolved {
				header += "  " + ui.StyleResolvedBadge.Render("[✓ resolvido]")
			}
			divider := ui.StyleCommentDivider.Render(strings.Repeat("─", max(width-4, 4)))
			body := renderMarkdown(r, note.Body)
			sb.WriteString("  " + header + "\n")
			sb.WriteString("  " + divider + "\n")
			sb.WriteString(body)
		} else {
			header := ui.StyleReplyArrow.Render("    ↳ ") +
				ui.StyleReplyAuthor.Render("@"+note.Author) +
				ui.StyleMeta.Render("  •  "+renderAge(note.CreatedAt))
			body := renderMarkdown(r, note.Body)
			lines := strings.Split(strings.TrimRight(body, "\n"), "\n")
			var indented strings.Builder
			for _, line := range lines {
				indented.WriteString("    " + line + "\n")
			}
			sb.WriteString(header + "\n")
			sb.WriteString(indented.String())
		}
	}
	return sb.String()
}

func renderSystemNote(n gitlab.Note) string {
	return ui.StyleSystemNote.Render(
		fmt.Sprintf("⚙  @%s %s  •  %s", n.Author, n.Body, renderAge(n.CreatedAt)),
	) + "\n"
}

func renderMarkdown(r *glamour.TermRenderer, body string) string {
	if r == nil {
		return body + "\n"
	}
	rendered, err := r.Render(body)
	if err != nil {
		return body + "\n"
	}
	return rendered
}

func (m Model) anyLoading() bool {
	for _, v := range m.Loading {
		if v {
			return true
		}
	}
	return false
}

func (m Model) countErrors() int {
	n := 0
	for _, e := range m.Errors {
		if e != nil {
			n++
		}
	}
	return n
}
