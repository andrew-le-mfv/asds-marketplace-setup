package plugins

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/your-org/asds-marketplace-setup/internal/tui/styles"
)

// View renders the plugin browser.
func (m Model) View() string {
	title := styles.TitleStyle.Render("📦 Available Plugins")
	subtitle := styles.SubtitleStyle.Render("Browse all plugins from the marketplace")

	var lines []string
	lines = append(lines, "", title, subtitle, "")

	for i, item := range m.items {
		cursor := "  "
		style := styles.NormalStyle
		if i == m.cursor {
			cursor = "▸ "
			style = styles.SelectedStyle
		}

		req := ""
		if item.Required {
			req = styles.WarningStyle.Render(" [required]")
		}

		line := fmt.Sprintf("%s%s%s", cursor, item.Name, req)
		lines = append(lines, style.Render(line))
		if i == m.cursor {
			lines = append(lines, styles.SubtitleStyle.Render(fmt.Sprintf("    Source: %s | Role: %s", item.Source, item.RoleName)))
		}
	}

	lines = append(lines, "", styles.HelpStyle.Render("↑↓ navigate"))

	return styles.BoxStyle.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}
