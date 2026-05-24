package ui

import "github.com/charmbracelet/lipgloss"

var (
	ColorBg       = lipgloss.Color("#282828")
	ColorFg       = lipgloss.Color("#ebdbb2")
	ColorYellow   = lipgloss.Color("#d79921")
	ColorOrange   = lipgloss.Color("#fe8019")
	ColorBlue     = lipgloss.Color("#83a598")
	ColorGreen    = lipgloss.Color("#b8bb26")
	ColorPurple   = lipgloss.Color("#d3869b")
	ColorGray     = lipgloss.Color("#928374")
	ColorRed      = lipgloss.Color("#cc241d")
	ColorStatusBg = lipgloss.Color("#3c3836")
	ColorStatusFg = lipgloss.Color("#a89984")
	ColorBorder   = lipgloss.Color("#504945")

	StyleProjectHeader = lipgloss.NewStyle().
				Foreground(ColorYellow).
				Background(ColorBg).
				Bold(true)

	StyleCursor = lipgloss.NewStyle().
			Foreground(ColorOrange).
			Background(ColorBg).
			Bold(true)

	StyleMRNumber = lipgloss.NewStyle().
			Foreground(ColorBlue).
			Background(ColorBg)

	StyleAuthor = lipgloss.NewStyle().
			Foreground(ColorGreen).
			Background(ColorBg)

	StyleBranch = lipgloss.NewStyle().
			Foreground(ColorPurple).
			Background(ColorBg)

	StyleMeta = lipgloss.NewStyle().
			Foreground(ColorGray).
			Background(ColorBg)

	StyleCIPassed = lipgloss.NewStyle().
			Foreground(ColorGreen).
			Background(ColorBg)

	StyleCIFailed = lipgloss.NewStyle().
			Foreground(ColorRed).
			Background(ColorBg)

	StyleCINone = lipgloss.NewStyle().
			Foreground(ColorGray).
			Background(ColorBg)

	StyleStatusBar = lipgloss.NewStyle().
			Background(ColorStatusBg).
			Foreground(ColorStatusFg).
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

	StyleError = lipgloss.NewStyle().
			Foreground(ColorRed).
			Background(ColorBg)
)
