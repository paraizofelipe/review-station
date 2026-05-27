package ui

import "github.com/charmbracelet/lipgloss"

var (
	ColorFg     = lipgloss.Color("#ebdbb2")
	ColorYellow = lipgloss.Color("#d79921")
	ColorOrange = lipgloss.Color("#fe8019")
	ColorBlue   = lipgloss.Color("#83a598")
	ColorGreen  = lipgloss.Color("#b8bb26")
	ColorPurple = lipgloss.Color("#d3869b")
	ColorGray   = lipgloss.Color("#928374")
	ColorRed    = lipgloss.Color("#cc241d")
	ColorMuted   = lipgloss.Color("#a89984")
	ColorBorder  = lipgloss.Color("#504945")
	ColorSurface = lipgloss.Color("#3c3836")

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

	// StyleCommentBox envolve cada comentário (nota pai) em uma caixa de
	// largura total, com fundo de superfície para destacá-la do terminal.
	StyleCommentBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Background(ColorSurface).
			Padding(0, 1)

	// StyleReplyBox é a caixa das respostas; mesmo visual da caixa do pai,
	// renderizada com largura menor e indentada pelo chamador.
	StyleReplyBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Background(ColorSurface).
			Padding(0, 1)

	// Os estilos abaixo são usados DENTRO das caixas. Todos carregam
	// Background(ColorSurface) para que cada span ANSI preencha o fundo e o
	// fundo do terminal não vaze através dos resets do glamour.
	StyleCommentAuthor = lipgloss.NewStyle().
				Foreground(ColorBlue).
				Background(ColorSurface).
				Bold(true)

	StyleReplyAuthor = lipgloss.NewStyle().
				Foreground(ColorPurple).
				Background(ColorSurface)

	StyleReplyArrow = lipgloss.NewStyle().
			Foreground(ColorGray).
			Background(ColorSurface)

	StyleCommentDivider = lipgloss.NewStyle().
				Foreground(ColorBorder).
				Background(ColorSurface)

	StyleSystemNote = lipgloss.NewStyle().
			Foreground(ColorGray).
			Italic(true)

	StyleSectionDivider = lipgloss.NewStyle().
				Foreground(ColorGray)

	StyleResolvedBadge = lipgloss.NewStyle().
				Foreground(ColorGreen)

	// Variantes com fundo de superfície, para uso dentro das caixas de comentário.
	StyleMetaOnSurface = lipgloss.NewStyle().
				Foreground(ColorGray).
				Background(ColorSurface)

	StyleResolvedBadgeOnSurface = lipgloss.NewStyle().
					Foreground(ColorGreen).
					Background(ColorSurface)

	StyleError = lipgloss.NewStyle().
			Foreground(ColorRed)
)
