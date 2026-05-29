package model

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/ansi"
	"github.com/charmbracelet/glamour/styles"
	"github.com/charmbracelet/lipgloss"

	"paraizofelipe/review-station/internal/gitlab"
	"paraizofelipe/review-station/internal/ui"
)

func (m Model) View() string {
	if !m.Ready {
		return "inicializando..."
	}

	if m.Screen == ScreenReply {
		return m.renderReplyView()
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

	if m.ShowOwnerFilter {
		filterArea := m.renderOwnerFilterOverlay()
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
	return lipgloss.NewStyle().Width(m.Width).Render(row)
}

const keyBarHeight = 2

func (m Model) renderKeyBar(lines []string) string {
	for len(lines) < keyBarHeight {
		lines = append(lines, "")
	}
	if len(lines) > keyBarHeight {
		lines = lines[:keyBarHeight]
	}
	for i, line := range lines {
		lines[i] = ui.StyleStatusBar.Width(m.Width).Render(line)
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderStatusbar() string {
	switch {
	case m.ShowFilter:
		return m.renderKeyBar([]string{
			"j/k ou ↑/↓ navegar opções  enter aplicar",
			"esc ou f fechar sem aplicar",
		})
	case m.FilterChordPending:
		return m.renderKeyBar([]string{
			"f+s ou f+f filtrar status  f+o filtrar owner",
			"esc cancelar chord de filtro",
		})
	case m.ShowOwnerFilter:
		return m.renderKeyBar([]string{
			"digite owner  enter aplicar filtro",
			"esc limpar filtro e fechar",
		})
	default:
		owner := ""
		if m.OwnerFilter != "" {
			owner = "  filtro atual: @" + m.OwnerFilter
		}
		return m.renderKeyBar([]string{
			"j/k ou ↑/↓ navegar  enter abrir MR  f+s status  f+o owner" + owner,
			"r atualizar  q ou ctrl+c sair",
		})
	}
}

func (m Model) renderList() string {
	if m.anyLoading() && len(m.Projects) == 0 {
		return "\n  Carregando MRs..."
	}
	if len(m.Items) == 0 {
		return "\n  Nenhum MR encontrado."
	}

	visible := filterItemsByOwner(m.Items, m.OwnerFilter)

	hasMR := false
	for _, item := range visible {
		if item.Kind == ItemMR {
			hasMR = true
			break
		}
	}
	if !hasMR && m.OwnerFilter != "" {
		return "\n  Nenhum MR de @" + m.OwnerFilter + "."
	}

	var sb strings.Builder
	for i, item := range visible {
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
	parts := strings.Split(item.Repo.Path, "/")
	label := item.Repo.Path
	if len(parts) > 2 {
		label = strings.Join(parts[len(parts)-2:], "/")
	}
	count := len(item.Project.MRs)
	noun := "MRs"
	if count == 1 {
		noun = "MR"
	}
	labelStyled := ui.StyleProjectHeader.Render(fmt.Sprintf("%s (%d %s)", label, count, noun))
	dashes := ui.StyleSectionDivider.Render(strings.Repeat("─", 15))
	return "\n" + dashes + " " + labelStyled + " " + dashes
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

func (m Model) renderOwnerFilterOverlay() string {
	label := ui.StylePopoverItem.Render("Owner: ")
	input := m.OwnerFilterInput.View()
	inner := "── Filtrar por owner ──\n" + label + input
	box := ui.StylePopoverBorder.Render(inner)
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
	return lipgloss.NewStyle().Width(m.Width).Render(row)
}

func (m Model) renderCommentsStatusbar() string {
	return m.renderKeyBar([]string{
		"j/k ou ↑/↓ scroll  ctrl+d/u página  tab/shift+tab comentário",
		"r responder  c colapsar  backspace ou esc voltar  q ou ctrl+c sair",
	})
}

// replyTextareaHeight returns the inner content line count for the reply
// textarea, reserving the bottom half of the screen for the markdown preview.
func replyTextareaHeight(totalH int) int {
	h := (totalH - 7 - keyBarHeight) / 2
	if h < 3 {
		h = 3
	}
	return h
}

func (m Model) renderReplyView() string {
	header := m.renderReplyHeader()
	statusbar := m.renderReplyStatusbar()
	previewLines := max(m.Height-7-keyBarHeight-replyTextareaHeight(m.Height), 2)
	divider := ui.StyleSectionDivider.Render("── Preview " + strings.Repeat("─", max(m.Width-12, 4)))
	var sb strings.Builder
	sb.WriteString(header)
	sb.WriteString("\n")
	sb.WriteString(m.ReplyInput.View())
	sb.WriteString(divider + "\n")
	sb.WriteString(m.renderReplyPreview(previewLines))
	sb.WriteString(statusbar)
	return sb.String()
}

// renderReplyPreview renders up to maxLines of the current textarea content
// using glamour (markdown + syntax highlighting), or a placeholder if empty.
func (m Model) renderReplyPreview(maxLines int) string {
	body := m.ReplyInput.Value()
	renderWidth := max(m.Width-2, 10)
	if strings.TrimSpace(body) == "" {
		return ui.StyleMeta.Render("  (preview vazio)") + "\n"
	}
	r := newDescriptionRenderer(renderWidth)
	text := strings.Trim(renderMarkdownPadded(r, body, renderWidth), "\n")
	lines := strings.Split(text, "\n")
	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}
	return strings.Join(lines, "\n") + "\n"
}

func (m Model) renderReplyHeader() string {
	right := ui.StyleMeta.Render("esc cancelar")
	raw := "↳ Respondendo"
	idx := 0
	for _, d := range m.Discussions {
		if len(d.Notes) == 0 || d.Notes[0].System {
			continue
		}
		if idx == m.CommentCursor {
			raw = fmt.Sprintf("↳ Respondendo a @%s", d.Notes[0].Author)
			break
		}
		idx++
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
	return lipgloss.NewStyle().Width(m.Width).Render(left + strings.Repeat(" ", gap) + right)
}

func (m Model) renderReplyStatusbar() string {
	if m.ReplySending {
		return m.renderKeyBar([]string{
			ui.StyleMeta.Render("[enviando...]"),
			"esc cancelar",
		})
	}
	if m.ReplyError != nil {
		return m.renderKeyBar([]string{
			ui.StyleError.Render("Erro: " + m.ReplyError.Error()),
			"ctrl+s tentar enviar novamente  esc cancelar",
		})
	}
	return m.renderKeyBar([]string{
		"digite/edite resposta no textarea",
		"ctrl+s enviar  esc cancelar",
	})
}

// newReplyInput cria e estiliza o textarea de resposta.
func newReplyInput() textarea.Model {
	ta := textarea.New()
	ta.Placeholder = "Escreva sua resposta..."
	ta.FocusedStyle.Base = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ui.ColorOrange)
	ta.BlurredStyle.Base = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ui.ColorBorder)
	ta.FocusedStyle.Text = lipgloss.NewStyle().Foreground(ui.ColorFg)
	ta.BlurredStyle.Text = lipgloss.NewStyle().Foreground(ui.ColorMuted)
	ta.FocusedStyle.Placeholder = lipgloss.NewStyle().Foreground(ui.ColorGray)
	ta.BlurredStyle.Placeholder = lipgloss.NewStyle().Foreground(ui.ColorGray)
	return ta
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

// boxChrome é o espaço consumido pela borda (2) + padding horizontal (2) de
// uma caixa lipgloss; usado para derivar a largura interna de conteúdo.
const boxChrome = 4

// commentGlamourStyle clona o estilo "dark" do glamour zerando a margem do
// documento e definindo o fundo para corresponder ao background da caixa que
// envolve o comentário, prevenindo sangramento de ANSI resets.
func commentGlamourStyle(bg string) ansi.StyleConfig {
	cfg := styles.DarkStyleConfig
	zero := uint(0)
	black := "#000000"
	cfg.Document.Margin = &zero
	cfg.Document.StylePrimitive.BackgroundColor = &bg
	// Código inline herda o fundo da caixa.
	cfg.Code.StylePrimitive.BackgroundColor = &bg
	// Blocos de código: fundo preto, sem margem lateral (o preenchimento de
	// espaços no pré-processamento estende o fundo até a largura do conteúdo).
	cfg.CodeBlock.Margin = &zero
	cfg.CodeBlock.StylePrimitive.BackgroundColor = &black
	return cfg
}

func newCommentRenderer(wrap int, bg string) *glamour.TermRenderer {
	if wrap < 10 {
		wrap = 10
	}
	r, err := glamour.NewTermRenderer(
		glamour.WithStyles(commentGlamourStyle(bg)),
		glamour.WithWordWrap(wrap),
	)
	if err != nil {
		return nil
	}
	return r
}

// newDescriptionRenderer cria um renderer sem background definido, para que
// o conteúdo herde o fundo padrão do terminal.
func newDescriptionRenderer(wrap int) *glamour.TermRenderer {
	if wrap < 10 {
		wrap = 10
	}
	cfg := styles.DarkStyleConfig
	zero := uint(0)
	black := "#000000"
	cfg.Document.Margin = &zero
	cfg.CodeBlock.Margin = &zero
	cfg.CodeBlock.StylePrimitive.BackgroundColor = &black
	r, err := glamour.NewTermRenderer(
		glamour.WithStyles(cfg),
		glamour.WithWordWrap(wrap),
	)
	if err != nil {
		return nil
	}
	return r
}

// buildRenderedDiscussions renderiza todos os comentários e notas de sistema.
// selectedComment é o índice (base-0) do comentário de usuário selecionado,
// ou -1 para nenhum. collapsed mapeia índices de comentários colapsados.
// Retorna o conteúdo renderizado e os offsets em linhas de cada comentário
// de usuário dentro do conteúdo (para scroll por Tab).
func buildRenderedDiscussions(mr *gitlab.MergeRequest, discussions []gitlab.Discussion, diffs []gitlab.FileDiff, width int, selectedComment int, collapsed map[int]bool) (string, []int) {
	// Gutter de 2 colunas (◯/◉/│ + espaço) reservado à esquerda das caixas.
	const cursorGutter = 2
	parentContent := max(width-boxChrome-cursorGutter, 14)
	descRenderer := newDescriptionRenderer(parentContent)
	parentRenderer := newCommentRenderer(parentContent, string(ui.ColorBg))
	replyRenderer := newCommentRenderer(parentContent, string(ui.ColorBg1))

	var sb strings.Builder
	var offsets []int
	lineCount := 0

	addStr := func(s string) {
		sb.WriteString(s)
		lineCount += strings.Count(s, "\n")
	}

	if mr != nil && strings.TrimSpace(mr.Description) != "" {
		addStr(renderMRDescription(mr, descRenderer, parentContent))
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

	if len(userDiscussions) > 0 {
		divLen := max(width-22, 4)
		addStr(ui.StyleSectionDivider.Render(
			"\n── Comentários " + strings.Repeat("─", divLen),
		))
		addStr("\n\n")
	}

	for i, d := range userDiscussions {
		if i > 0 {
			addStr(ui.StyleTimeline.Render("│") + "\n")
		}
		offsets = append(offsets, lineCount)
		addStr(renderDiscussion(d, diffs, parentRenderer, replyRenderer, parentContent, i == selectedComment, collapsed[i]))
	}

	if len(systemNotes) > 0 {
		divLen := max(width-26, 4)
		addStr(ui.StyleSectionDivider.Render(
			"\n── Atividade do sistema " + strings.Repeat("─", divLen) + "\n",
		))
		for _, n := range systemNotes {
			addStr(renderSystemNote(n, width))
		}
	}
	return sb.String(), offsets
}

func renderMRDescription(mr *gitlab.MergeRequest, r *glamour.TermRenderer, codeWidth int) string {
	header := ui.StyleCommentAuthor.Render("@"+mr.Author) +
		ui.StyleMeta.Render("  •  "+renderAge(mr.CreatedAt))
	body := strings.Trim(renderMarkdownPadded(r, mr.Description, codeWidth), "\n")
	content := header + "\n\n" + body
	return "\n\n" + indentLines(content, 2) + "\n"
}

func renderDiscussion(d gitlab.Discussion, diffs []gitlab.FileDiff, parentRenderer, replyRenderer *glamour.TermRenderer, parentContent int, selected, collapsed bool) string {
	if collapsed {
		return renderCollapsedDiscussion(d, parentContent, selected)
	}
	var sb strings.Builder
	isFirst := true
	for _, note := range d.Notes {
		if note.System {
			continue
		}
		if isFirst {
			isFirst = false
			header := ui.StyleCommentAuthor.Render("@"+note.Author) +
				ui.StyleMetaOnComment.Render("  •  "+renderAge(note.CreatedAt))
			if note.Resolvable && note.Resolved {
				header += ui.StyleMetaOnComment.Render("  ") +
					ui.StyleResolvedBadgeOnComment.Render("[✓ resolvido]")
			}
			divider := ui.StyleCommentDivider.Render(strings.Repeat("─", parentContent))
			body := strings.Trim(renderMarkdownPadded(parentRenderer, note.Body, parentContent), "\n")

			var inner string
			if ctx := renderDiffContext(note.Position, diffs, parentContent); ctx != "" {
				ctxDivider := ui.StyleCommentDivider.Render(strings.Repeat("─", parentContent))
				inner = ctx + "\n" + ctxDivider + "\n" + header + "\n" + divider + "\n" + body
			} else {
				inner = header + "\n" + divider + "\n" + body
			}

			box := ui.StyleCommentBox.Width(parentContent + 2).Render(inner)
			sb.WriteString(applyCommentCursor(box, selected) + "\n")
		} else {
			header := ui.StyleReplyArrow.Render("↳ ") +
				ui.StyleReplyAuthor.Render("@"+note.Author) +
				ui.StyleMetaOnReply.Render("  •  "+renderAge(note.CreatedAt))
			body := strings.Trim(renderMarkdownPadded(replyRenderer, note.Body, parentContent), "\n")
			inner := header + "\n" + body
			box := ui.StyleReplyBox.Width(parentContent + 2).Render(inner)
			sb.WriteString(applyReplyConnector(box) + "\n")
		}
	}
	return sb.String()
}

// renderCollapsedDiscussion renderiza uma linha compacta para um comentário
// colapsado: apenas cabeçalho (autor, tempo, badge) e contagem de notas.
func renderCollapsedDiscussion(d gitlab.Discussion, parentContent int, selected bool) string {
	var first *gitlab.Note
	noteCount := 0
	for i := range d.Notes {
		if d.Notes[i].System {
			continue
		}
		if first == nil {
			first = &d.Notes[i]
		}
		noteCount++
	}
	if first == nil {
		return ""
	}

	header := ui.StyleCommentAuthor.Render("@"+first.Author) +
		ui.StyleMetaOnComment.Render("  •  "+renderAge(first.CreatedAt))
	if first.Resolvable && first.Resolved {
		header += ui.StyleMetaOnComment.Render("  ") +
			ui.StyleResolvedBadgeOnComment.Render("[✓ resolvido]")
	}
	noun := "nota"
	if noteCount != 1 {
		noun = "notas"
	}
	header += ui.StyleMetaOnComment.Render(fmt.Sprintf("  ▸ %d %s", noteCount, noun))

	box := ui.StyleCommentBox.Width(parentContent + 2).Render(header)
	return applyCommentCursor(box, selected) + "\n"
}

// applyCommentCursor prefixa cada linha do box pai com o gutter de timeline:
// primeira linha usa ◯ (não selecionado) ou ◉ (selecionado); demais usam │.
func applyCommentCursor(box string, selected bool) string {
	bar := ui.StyleTimeline.Render("│") + " "
	firstPrefix := ui.StyleTimeline.Render("◯") + " "
	if selected {
		firstPrefix = ui.StyleCursor.Render("◉") + " "
	}
	lines := strings.Split(box, "\n")
	for i, line := range lines {
		if i == 0 {
			lines[i] = firstPrefix + line
		} else if line != "" {
			lines[i] = bar + line
		}
	}
	return strings.Join(lines, "\n")
}

// applyReplyConnector prefixa o box de resposta com "│ " no gutter,
// mantendo visualmente a continuidade da thread.
func applyReplyConnector(box string) string {
	bar := ui.StyleTimeline.Render("│") + " "
	lines := strings.Split(box, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = bar + line
		}
	}
	return strings.Join(lines, "\n")
}

// renderDiffContext retorna o bloco de diff com syntax highlighting e
// backgrounds coloridos (delta-style) para a posição de uma note inline,
// ou string vazia se não houver contexto disponível.
func renderDiffContext(pos *gitlab.Position, diffs []gitlab.FileDiff, width int) string {
	if pos == nil || len(diffs) == 0 {
		return ""
	}

	var fd *gitlab.FileDiff
	for i := range diffs {
		if diffs[i].NewPath == pos.NewPath || diffs[i].OldPath == pos.OldPath {
			fd = &diffs[i]
			break
		}
	}
	if fd == nil {
		return ""
	}

	lines := gitlab.ExtractDiffContext(fd.Diff, pos.NewLine, pos.OldLine, 3)
	if len(lines) == 0 {
		return ""
	}

	filePath := pos.NewPath
	if filePath == "" {
		filePath = pos.OldPath
	}
	targetLine := pos.NewLine
	if targetLine == 0 {
		targetLine = pos.OldLine
	}

	var sb strings.Builder
	sb.WriteString(ui.StyleDiffPath.Render(fmt.Sprintf("· %s:%d", filePath, targetLine)))
	sb.WriteString("\n")
	sb.WriteString(renderDiffBlock(lines, filePath, width))

	return sb.String()
}

// indentLines prefixa cada linha de s com n espaços.
func indentLines(s string, n int) string {
	pad := strings.Repeat(" ", n)
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = pad + line
	}
	return strings.Join(lines, "\n")
}

var (
	htmlTagRe   = regexp.MustCompile(`<[^>]+>`)
	mdEmphRe    = regexp.MustCompile("[*_`]+")
	liRe        = regexp.MustCompile(`(?s)<li>(.*?)</li>`)
	commitRe    = regexp.MustCompile(`(?i)^([0-9a-f]{7,40})\s*[-–]\s*(.+)$`)
	compareLine = "[compare with previous version]"
)

// cleanSystemText remove tags HTML e marcadores markdown de ênfase, deixando
// o texto plano legível.
func cleanSystemText(s string) string {
	s = htmlTagRe.ReplaceAllString(s, "")
	s = mdEmphRe.ReplaceAllString(s, "")
	return strings.TrimSpace(s)
}

type sysCommit struct {
	sha string
	msg string
}

// parseSystemNote separa a ação (primeira linha) dos commits embutidos em
// notas "added N commits", descartando o link "Compare with previous version"
// e o resumo de range do GitLab.
func parseSystemNote(body string) (action string, commits []sysCommit) {
	lines := strings.Split(body, "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if action == "" && !strings.Contains(line, "<li>") {
			if strings.HasPrefix(strings.ToLower(strings.TrimSpace(line)), compareLine) {
				continue
			}
			action = cleanSystemText(line)
		}
	}

	for _, m := range liRe.FindAllStringSubmatch(body, -1) {
		inner := cleanSystemText(m[1])
		cm := commitRe.FindStringSubmatch(inner)
		if cm == nil {
			continue // pula resumo de range (ex.: "a7cf...822c - 2 commits from branch")
		}
		commits = append(commits, sysCommit{sha: cm[1], msg: strings.TrimSpace(cm[2])})
	}
	return action, commits
}

func renderSystemNote(n gitlab.Note, width int) string {
	action, commits := parseSystemNote(n.Body)
	if action == "" {
		action = cleanSystemText(n.Body)
	}

	var sb strings.Builder
	line := ui.StyleSystemNote.Render("⚙  ") +
		ui.StyleSystemNoteAuthor.Render("@"+n.Author) +
		ui.StyleSystemNote.Render(" "+action+"  ") +
		ui.StyleSystemNoteTime.Render("• "+renderAge(n.CreatedAt))
	sb.WriteString(line)
	sb.WriteString("\n")

	const maxCommits = 2
	msgWidth := max(width-20, 16)
	for i, c := range commits {
		if i >= maxCommits {
			sb.WriteString(ui.StyleMeta.Render(
				fmt.Sprintf("    ↳ … +%d commits", len(commits)-maxCommits),
			))
			sb.WriteString("\n")
			break
		}
		sha := c.sha
		if len(sha) > 8 {
			sha = sha[:8]
		}
		msg := c.msg
		if len([]rune(msg)) > msgWidth {
			msg = string([]rune(msg)[:msgWidth-1]) + "…"
		}
		sb.WriteString(ui.StyleMeta.Render(fmt.Sprintf("    ↳ %s  %s", sha, msg)))
		sb.WriteString("\n")
	}
	return sb.String()
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

// padCodeBlocksInMarkdown preenchee cada linha dentro de blocos de código
// cercados (```…```) com espaços até `width` colunas. Isso faz com que o
// renderer de fallback do glamour aplique o BackgroundColor ao longo de toda
// a largura da linha, e não apenas ao texto.
func padCodeBlocksInMarkdown(md string, width int) string {
	lines := strings.Split(md, "\n")
	inCode := false
	for i, line := range lines {
		trimmed := strings.TrimLeft(line, " \t")
		if !inCode {
			if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
				inCode = true
			}
		} else {
			// Linha de fechamento: só caracteres de cerca, sem info-string.
			allFence := strings.TrimRight(trimmed, "`") == "" || strings.TrimRight(trimmed, "~") == ""
			if allFence && len(trimmed) >= 3 {
				inCode = false
			} else {
				runes := []rune(line)
				if len(runes) < width {
					lines[i] = line + strings.Repeat(" ", width-len(runes))
				}
			}
		}
	}
	return strings.Join(lines, "\n")
}

// renderMarkdownPadded aplica padCodeBlocksInMarkdown antes de renderizar,
// garantindo background de largura total nos blocos de código.
func renderMarkdownPadded(r *glamour.TermRenderer, body string, codeWidth int) string {
	return renderMarkdown(r, padCodeBlocksInMarkdown(body, codeWidth))
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
