package ui

import "github.com/charmbracelet/lipgloss"

var (
	ColorFg      = lipgloss.Color("#ebdbb2")
	ColorYellow  = lipgloss.Color("#d79921")
	ColorOrange  = lipgloss.Color("#fe8019")
	ColorBlue    = lipgloss.Color("#83a598")
	ColorGreen   = lipgloss.Color("#b8bb26")
	ColorPurple  = lipgloss.Color("#d3869b")
	ColorGray    = lipgloss.Color("#928374")
	ColorRed     = lipgloss.Color("#cc241d")
	ColorMuted   = lipgloss.Color("#a89984")
	ColorBorder  = lipgloss.Color("#504945")
	ColorSurface = lipgloss.Color("#3c3836")
	ColorBg      = lipgloss.Color("#282828") // gruvbox bg0 — fundo dos comment boxes
	ColorBg1     = lipgloss.Color("#32302f") // gruvbox bg1 — fundo dos reply boxes

	StyleProjectHeader = lipgloss.NewStyle().
				Foreground(ColorYellow).
				Bold(true)

	StyleCursor = lipgloss.NewStyle().
			Foreground(ColorOrange).
			Bold(true)

	StyleMRNumber = lipgloss.NewStyle().
			Foreground(ColorBlue)

	StyleAuthor = lipgloss.NewStyle().
			Foreground(ColorGreen)

	StyleBranch = lipgloss.NewStyle().
			Foreground(ColorPurple)

	StyleMeta = lipgloss.NewStyle().
			Foreground(ColorGray)

	StyleCIPassed = lipgloss.NewStyle().
			Foreground(ColorGreen)

	StyleCIFailed = lipgloss.NewStyle().
			Foreground(ColorRed)

	StyleCINone = lipgloss.NewStyle().
			Foreground(ColorGray)

	StyleStatusBar = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Padding(0, 1)

	StylePopoverBorder = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorBorder).
				Padding(0, 1)

	StylePopoverSelected = lipgloss.NewStyle().
				Foreground(ColorOrange).
				Bold(true)

	StylePopoverItem = lipgloss.NewStyle().
				Foreground(ColorFg)

	StyleTitleBar = lipgloss.NewStyle().
			Foreground(ColorOrange).
			Bold(true)

	// StyleCommentBox envolve cada comentário raiz. Fundo #282828 (gruvbox bg0).
	StyleCommentBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Background(ColorBg).
			Padding(0, 1)

	// StyleReplyBox é a caixa das respostas. Fundo #32302f (gruvbox bg1) para
	// diferenciar visualmente do comentário pai.
	StyleReplyBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorSurface).
			Background(ColorBg1).
			Padding(0, 1)

	// Estilos usados DENTRO das caixas. O Background deve corresponder ao da
	// caixa em que o estilo será aplicado para evitar sangramento de ANSI resets.

	StyleCommentAuthor = lipgloss.NewStyle().
				Foreground(ColorBlue).
				Background(ColorBg).
				Bold(true)

	StyleReplyAuthor = lipgloss.NewStyle().
				Foreground(ColorPurple).
				Background(ColorBg1)

	StyleReplyArrow = lipgloss.NewStyle().
			Foreground(ColorGray).
			Background(ColorBg1)

	StyleCommentDivider = lipgloss.NewStyle().
				Foreground(ColorBorder).
				Background(ColorBg)

	StyleSystemNote = lipgloss.NewStyle().
			Foreground(ColorGray).
			Italic(true)

	StyleSystemNoteAuthor = lipgloss.NewStyle().
				Foreground(ColorBlue).
				Italic(true)

	StyleSystemNoteTime = lipgloss.NewStyle().
				Foreground(ColorBorder). // #504945 — bem discreto
				Italic(true)

	StyleSectionDivider = lipgloss.NewStyle().
				Foreground(ColorGray)

	StyleResolvedBadge = lipgloss.NewStyle().
				Foreground(ColorGreen)

	// Estilos de meta e badge para dentro de cada tipo de caixa.
	StyleMetaOnComment = lipgloss.NewStyle().
				Foreground(ColorGray).
				Background(ColorBg)

	StyleMetaOnReply = lipgloss.NewStyle().
				Foreground(ColorGray).
				Background(ColorBg1)

	StyleResolvedBadgeOnComment = lipgloss.NewStyle().
					Foreground(ColorGreen).
					Background(ColorBg)

	// StyleTimeline é usado para a barra │ e os conectores de árvore ├─/└─
	// na lateral esquerda dos comentários e respostas.
	StyleTimeline = lipgloss.NewStyle().
			Foreground(ColorBorder)

	StyleError = lipgloss.NewStyle().
			Foreground(ColorRed)

	StyleDiffPath = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Italic(true)
)
