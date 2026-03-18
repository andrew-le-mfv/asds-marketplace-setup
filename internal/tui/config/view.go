package config

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/your-org/asds-marketplace-setup/internal/tui/styles"
)

// View renders the config viewer.
func (m Model) View() string {
	title := styles.TitleStyle.Render("⚙ Configuration")

	var lines []string
	lines = append(lines, "", title, "")

	lines = append(lines, styles.NormalStyle.Render(fmt.Sprintf("  Config file: %s", m.cfgPath)))
	lines = append(lines, "")

	if m.asdsCfg != nil {
		lines = append(lines, styles.NormalStyle.Render(fmt.Sprintf("  Marketplace URL: %s", m.asdsCfg.MarketplaceURL)))
	} else {
		lines = append(lines, styles.WarningStyle.Render("  No config file found (using defaults)"))
	}

	lines = append(lines, "")
	lines = append(lines, styles.HelpStyle.Render("Config location: ~/.config/asds/config.yaml"))
	lines = append(lines, styles.HelpStyle.Render("Precedence: CLI flags > env vars > config file > defaults"))

	return styles.BoxStyle.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}
