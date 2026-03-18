package about

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/your-org/asds-marketplace-setup/internal/tui/styles"
)

// Model is the About tab model.
type Model struct {
	version string
	width   int
	height  int
}

// New creates a new About tab model.
func New(version string) Model {
	return Model{version: version}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

// View implements tea.Model.
func (m Model) View() string {
	title := styles.TitleStyle.Render("🚀 ASDS — Agentic Software Development Suite")

	info := lipgloss.JoinVertical(lipgloss.Left,
		"",
		title,
		"",
		styles.NormalStyle.Render("Version: "+m.version),
		"",
		styles.SubtitleStyle.Render("A TUI for bootstrapping developers into curated"),
		styles.SubtitleStyle.Render("Claude Code plugin sets organized by role."),
		"",
		styles.HelpStyle.Render("GitHub: github.com/your-org/asds-marketplace-setup"),
		styles.HelpStyle.Render("Docs:   https://your-org.github.io/asds-marketplace-setup"),
	)

	return styles.BoxStyle.Render(info)
}
