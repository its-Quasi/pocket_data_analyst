package tui

import "github.com/charmbracelet/lipgloss"

var (
	TitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1)

	SubtitleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A0A0A0"))

	SessionItemStyle = lipgloss.NewStyle().
		PaddingLeft(1).
		Foreground(lipgloss.Color("#FAFAFA"))

	SelectedSessionStyle = lipgloss.NewStyle().
		PaddingLeft(0).
		Foreground(lipgloss.Color("#7D56F4")).
		Bold(true).
		Background(lipgloss.Color("#2A2A2A"))

	UserMsgStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7D56F4")).
		Bold(true)

	AssistantMsgStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA"))

	ErrorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF5F5F"))

	BoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#555555")).
		Padding(1)

	LeftPaneStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#555555")).
		Padding(1, 1)

	RightPaneStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#555555")).
		Padding(1, 1)
)
