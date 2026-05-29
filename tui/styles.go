package tui

import "github.com/charmbracelet/lipgloss"

const logo = `
   █████████               █████             ██████   ██████            █████
  ███░░░░░███             ░░███             ░░██████ ██████            ░░███
 ░███    ░███  █████ ████ ███████    ██████  ░███░█████░███   ██████   ███████    ██████
 ░███████████ ░░███ ░███ ░░░███░    ███░░███ ░███░░███ ░███  ░░░░░███ ░░░███░    ███░░███
 ░███░░░░░███  ░███ ░███   ░███    ░███ ░███ ░███ ░░░  ░███   ███████   ░███    ░███████
 ░███    ░███  ░███ ░███   ░███ ███░███ ░███ ░███      ░███  ███░░███   ░███ ███░███░░░
 █████   █████ ░░████████  ░░█████ ░░██████  █████     █████░░████████  ░░█████ ░░██████
░░░░░   ░░░░░   ░░░░░░░░    ░░░░░   ░░░░░░  ░░░░░     ░░░░░  ░░░░░░░░    ░░░░░   ░░░░░░
`

// TokyoNight-inspired color palette
const (
	colBg          = "#1a1b26"
	colBgPanel     = "#24283b"
	colBgElement   = "#2f3346"
	colBorder      = "#3b4261"
	colBorderAct   = "#7aa2f7"
	colText        = "#c0caf5"
	colTextMuted   = "#565f89"
	colBlue        = "#7aa2f7"
	colCyan        = "#7dcfff"
	colGreen       = "#9ece6a"
	colOrange      = "#e0af68"
	colRed         = "#f7768e"
	colPurple      = "#bb9af7"
	colYellow      = "#e0af68"
	colWhite       = "#ffffff"
)

var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(colBlue)).
			Padding(0, 1).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(colBlue)).
			BorderBottom(true).
			Width(72)

	statusBarLeft = lipgloss.NewStyle().
			Background(lipgloss.Color(colBgPanel)).
			Foreground(lipgloss.Color(colTextMuted)).
			Padding(0, 1).
			Width(60).
			Align(lipgloss.Left)

	statusBarRight = lipgloss.NewStyle().
			Background(lipgloss.Color(colBgPanel)).
			Foreground(lipgloss.Color(colTextMuted)).
			Padding(0, 1).
			Align(lipgloss.Right)

	sideTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(colBlue)).
			Padding(0, 1).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(colBorder)).
			BorderBottom(true).
			Width(24)

	sideItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colText)).
			Padding(0, 1).
			Width(22)

	sideActiveStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colBlue)).
			Bold(true).
			Padding(0, 1).
			Width(22)

	sideCountStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colTextMuted)).
			Padding(0, 0)

	itemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colText)).
			Padding(0, 1)

	itemSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(colBlue)).
				Bold(true).
				Padding(0, 1)

	itemInstalledStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(colGreen)).
				Bold(true)

	itemNotInstalledStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(colRed))

	itemDescStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colTextMuted))

	detailStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(colBorder)).
			Padding(1, 2).
			Width(56)

	detailTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color(colBlue)).
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(colBlue)).
				BorderBottom(true).
				Padding(0, 1).
				Width(60)

	detailLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(colTextMuted)).
				Width(14).
				Align(lipgloss.Right)

	detailValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(colText)).
				Padding(0, 0)

	detailSectionStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color(colCyan)).
				Padding(0, 0)

	paletteStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(colBlue)).
			Background(lipgloss.Color(colBgPanel)).
			Padding(1, 2).
			Width(60)

	paletteInputStyle = lipgloss.NewStyle().
				Background(lipgloss.Color(colBgElement)).
				Foreground(lipgloss.Color(colText)).
				Padding(0, 1).
				Width(56)

	paletteItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(colText)).
				Padding(0, 1).
				Width(54)

	paletteActiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(colBlue)).
				Bold(true).
				Padding(0, 1).
				Width(54)

	paletteHintStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(colTextMuted)).
				Padding(0, 1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colTextMuted)).
			Padding(0, 1)

	helpHeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(colBlue)).
			Padding(0, 1)

	installStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(colOrange)).
			Padding(1, 2).
			Width(60)

	runStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(colPurple)).
			Padding(1, 2).
			Width(60)

	logoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colCyan)).
			Bold(true)

	bootTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(colBlue)).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(colBlue)).
			BorderBottom(true).
			Padding(0, 1).
			Width(72)

	infoKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colTextMuted))

	infoValStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colText))

	searchStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(colBgElement)).
			Foreground(lipgloss.Color(colText)).
			Padding(0, 1).
			Width(40)

	terminalStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(colGreen)).
			Padding(1, 2).
			Width(60)
)
