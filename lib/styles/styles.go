package styles

import "github.com/charmbracelet/lipgloss"

var (
	HeaderBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("6")). // cyan
			Width(30).
			Align(lipgloss.Center).
			PaddingLeft(1).
			PaddingRight(1)

	Cursor = lipgloss.NewStyle().
		Foreground(lipgloss.Color("2")). // green
		Bold(true)

	SectionTitle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("4")). // blue
			Bold(true)

	ItemName = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")) // white

	ItemNameDisabled = lipgloss.NewStyle().
				Foreground(lipgloss.Color("8")) // gray

	StatusInstalled = lipgloss.NewStyle().
			Foreground(lipgloss.Color("2")) // green

	StatusNotInstalled = lipgloss.NewStyle().
				Foreground(lipgloss.Color("1")) // red

	SuccessBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("2")). // green
			Foreground(lipgloss.Color("2")).
			PaddingLeft(1).
			PaddingRight(1)

	WarningBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("1")). // red
			Foreground(lipgloss.Color("1")).
			PaddingLeft(1).
			PaddingRight(1)

	GuideBullet = lipgloss.NewStyle().
			Foreground(lipgloss.Color("6")) // cyan

	GuideText = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")) // white

	ErrorText = lipgloss.NewStyle().
			Foreground(lipgloss.Color("1")). // red
			Bold(true)

	LoadingText = lipgloss.NewStyle().
			Foreground(lipgloss.Color("2")). // green
			Italic(true)
)
