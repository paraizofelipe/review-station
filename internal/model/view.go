package model

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"paraizofelipe/review-station/internal/gitlab"
	"paraizofelipe/review-station/internal/ui"
)

func (m Model) View() string {
	if !m.Ready {
		return "inicializando..."
	}

	header := m.renderTitleBar()
	statusbar := m.renderStatusbar()

	listContent := m.renderList()
	m.Viewport.SetContent(listContent)

	var sb strings.Builder
	sb.WriteString(header)
	sb.WriteString("\n")
	sb.WriteString(m.Viewport.View())
	sb.WriteString("\n")
	sb.WriteString(statusbar)

	if m.ShowFilter {
		return m.renderFilterOverlay()
	}

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
	return title + strings.Repeat(" ", gap) + statusStr
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
		return fmt.Sprintf("há %d horas", int(d.Hours()))
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
		m.Width, m.Height,
		lipgloss.Center, lipgloss.Center,
		box,
	)
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
