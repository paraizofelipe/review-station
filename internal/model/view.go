package model

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

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
	return lipgloss.NewStyle().Width(m.Width).Render(row)
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

// replyIndent é o recuo (em colunas) das caixas de resposta em relação à
// caixa do comentário pai.
const replyIndent = 3

// boxChrome é o espaço consumido pela borda (2) + padding horizontal (2) de
// uma caixa lipgloss; usado para derivar a largura interna de conteúdo.
const boxChrome = 4

// commentGlamourStyle clona o estilo "dark" do glamour zerando a margem do
// documento (a caixa já fornece o padding) e pintando o fundo do documento
// com a cor de superfície, para que o fundo do terminal não vaze pelos resets
// ANSI dentro da caixa.
func commentGlamourStyle() ansi.StyleConfig {
	cfg := styles.DarkStyleConfig
	zero := uint(0)
	bg := string(ui.ColorSurface)
	cfg.Document.Margin = &zero
	cfg.Document.StylePrimitive.BackgroundColor = &bg

	// O code block é destacado pelo chroma (formatter terminal256): tokens sem
	// BackgroundColor resetam para o fundo do terminal, vazando dentro da caixa.
	// Pintamos TODOS os tokens com a cor da superfície para que o bloco use o
	// mesmo fundo do texto normal — sem virar um retângulo de cor distinta,
	// mantendo só o syntax highlighting nas cores do texto.
	// O Chroma é um ponteiro compartilhado com o estilo global — copiamos antes
	// de mutar.
	if cfg.CodeBlock.Chroma != nil {
		chromaCopy := *cfg.CodeBlock.Chroma
		setChromaBackground(&chromaCopy, bg)
		cfg.CodeBlock.Chroma = &chromaCopy
	}
	cfg.CodeBlock.StylePrimitive.BackgroundColor = &bg
	// Código inline (`x`) também herda o fundo da superfície.
	cfg.Code.StylePrimitive.BackgroundColor = &bg
	return cfg
}

// setChromaBackground define BackgroundColor em todos os tokens StylePrimitive
// do Chroma, garantindo que cada span emitido pelo chroma preencha o fundo.
func setChromaBackground(c *ansi.Chroma, color string) {
	v := reflect.ValueOf(c).Elem()
	for i := 0; i < v.NumField(); i++ {
		bg := v.Field(i).FieldByName("BackgroundColor")
		if bg.IsValid() && bg.CanSet() {
			bg.Set(reflect.ValueOf(&color))
		}
	}
}

func newCommentRenderer(wrap int) *glamour.TermRenderer {
	if wrap < 10 {
		wrap = 10
	}
	r, err := glamour.NewTermRenderer(
		glamour.WithStyles(commentGlamourStyle()),
		glamour.WithWordWrap(wrap),
		// terminal16m = truecolor; sem isso o chroma usa terminal256 e quantiza
		// o fundo do código para uma cor 256 levemente diferente da superfície
		// (truecolor), criando um retângulo visível dentro da caixa.
		glamour.WithChromaFormatter("terminal16m"),
	)
	if err != nil {
		return nil
	}
	return r
}

func buildRenderedDiscussions(discussions []gitlab.Discussion, width int) string {
	if len(discussions) == 0 {
		return ""
	}

	parentContent := max(width-boxChrome, 16)
	replyContent := max(width-replyIndent-boxChrome, 12)
	parentRenderer := newCommentRenderer(parentContent)
	replyRenderer := newCommentRenderer(replyContent)

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
		sb.WriteString(renderDiscussion(d, parentRenderer, replyRenderer, parentContent, replyContent))
	}
	if len(systemNotes) > 0 {
		divLen := max(width-26, 4)
		sb.WriteString(ui.StyleSectionDivider.Render(
			"\n── Atividade do sistema " + strings.Repeat("─", divLen) + "\n",
		))
		for _, n := range systemNotes {
			sb.WriteString(renderSystemNote(n, width))
		}
	}
	return sb.String()
}

func renderDiscussion(d gitlab.Discussion, parentRenderer, replyRenderer *glamour.TermRenderer, parentContent, replyContent int) string {
	var sb strings.Builder
	isFirst := true
	for _, note := range d.Notes {
		if note.System {
			continue
		}
		if isFirst {
			isFirst = false
			header := ui.StyleCommentAuthor.Render("@"+note.Author) +
				ui.StyleMetaOnSurface.Render("  •  "+renderAge(note.CreatedAt))
			if note.Resolvable && note.Resolved {
				header += ui.StyleMetaOnSurface.Render("  ") +
					ui.StyleResolvedBadgeOnSurface.Render("[✓ resolvido]")
			}
			divider := ui.StyleCommentDivider.Render(strings.Repeat("─", parentContent))
			body := strings.Trim(renderMarkdown(parentRenderer, note.Body), "\n")
			inner := header + "\n" + divider + "\n" + body
			box := ui.StyleCommentBox.Width(parentContent + 2).Render(inner)
			sb.WriteString(box + "\n")
		} else {
			header := ui.StyleReplyArrow.Render("↳ ") +
				ui.StyleReplyAuthor.Render("@"+note.Author) +
				ui.StyleMetaOnSurface.Render("  •  "+renderAge(note.CreatedAt))
			body := strings.Trim(renderMarkdown(replyRenderer, note.Body), "\n")
			inner := header + "\n" + body
			box := ui.StyleReplyBox.Width(replyContent + 2).Render(inner)
			sb.WriteString(indentLines(box, replyIndent) + "\n")
		}
	}
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
	sb.WriteString(ui.StyleSystemNote.Render(
		fmt.Sprintf("⚙  @%s %s  •  %s", n.Author, action, renderAge(n.CreatedAt)),
	))
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
