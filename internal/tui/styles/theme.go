package styles

import (
	"github.com/charmbracelet/lipgloss"
)

// Colors — a cohesive palette for the ASDS TUI.
var (
	Primary    = lipgloss.Color("#7C3AED") // Purple
	Secondary  = lipgloss.Color("#06B6D4") // Cyan
	Success    = lipgloss.Color("#22C55E") // Green
	Warning    = lipgloss.Color("#F59E0B") // Amber
	Danger     = lipgloss.Color("#EF4444") // Red
	Muted      = lipgloss.Color("#6B7280") // Gray
	Text       = lipgloss.Color("#F9FAFB") // Almost white
	TextDim    = lipgloss.Color("#9CA3AF") // Light gray
	Background = lipgloss.Color("#111827") // Dark
	Surface    = lipgloss.Color("#1F2937") // Slightly lighter
	Border     = lipgloss.Color("#374151") // Gray border
)

// Layout styles
var (
	AppStyle = lipgloss.NewStyle().
			Padding(0, 1)

	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Primary).
			Padding(1, 2)

	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Text)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(TextDim)

	FooterStyle = lipgloss.NewStyle().
			Foreground(Muted).
			Padding(0, 1)
)

// Tab styles
var (
	ActiveTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Primary).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(Primary).
			Padding(0, 2)

	InactiveTabStyle = lipgloss.NewStyle().
				Foreground(TextDim).
				Padding(0, 2)

	TabGapStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(Border)
)

// Content styles
var (
	SelectedStyle = lipgloss.NewStyle().
			Foreground(Primary).
			Bold(true)

	NormalStyle = lipgloss.NewStyle().
			Foreground(Text)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(Success)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(Danger)

	WarningStyle = lipgloss.NewStyle().
			Foreground(Warning)

	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Border).
			Padding(1, 2)

	HelpStyle = lipgloss.NewStyle().
			Foreground(Muted)
)
