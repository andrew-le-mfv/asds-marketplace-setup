package status

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/andrew-le-mfv/asds-marketplace-setup/internal/tui/styles"
)

// View renders the status dashboard.
func (m Model) View() string {
	title := styles.TitleStyle.Render("📊 Status Dashboard")

	var lines []string
	lines = append(lines, "", title, "")

	if m.claudeDetected {
		lines = append(lines, styles.SuccessStyle.Render(fmt.Sprintf("  Claude Code: ✓ %s", m.claudePath)))
	} else {
		lines = append(lines, styles.WarningStyle.Render("  Claude Code: ✗ not detected"))
	}

	lines = append(lines, styles.NormalStyle.Render(fmt.Sprintf("  Project root: %s", m.projectRoot)))
	lines = append(lines, "")

	for _, si := range m.scopes {
		if si.Found {
			lines = append(lines, styles.SuccessStyle.Render(
				fmt.Sprintf("  [%s] Role: %s | Plugins: %d | Method: %s | Installed: %s",
					si.Scope, si.Manifest.Role, len(si.Manifest.Plugins),
					si.Manifest.InstallMethod, si.Manifest.InstalledAt.Format("2006-01-02")),
			))
		} else {
			lines = append(lines, styles.SubtitleStyle.Render(
				fmt.Sprintf("  [%s] Not installed", si.Scope),
			))
		}
	}

	return styles.BoxStyle.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}
