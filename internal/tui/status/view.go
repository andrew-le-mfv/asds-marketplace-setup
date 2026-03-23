package status

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/andrew-le-mfv/asds-marketplace-setup/internal/tui/styles"
)

// View renders the status dashboard.
func (m Model) View() string {
	switch m.step {
	case StepView:
		return m.viewStatus()
	case StepConfirm:
		return m.viewConfirm()
	case StepUninstalling:
		return m.viewUninstalling()
	case StepResult:
		return m.viewResult()
	default:
		return ""
	}
}

func (m Model) viewStatus() string {
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

	hasInstalled := len(m.installedScopes()) > 0
	installedIdx := 0
	for _, si := range m.scopes {
		if si.Found {
			cursor := "  "
			style := styles.SuccessStyle
			if installedIdx == m.cursor {
				cursor = "▸ "
				style = styles.SelectedStyle
			}
			installedIdx++
			lines = append(lines, style.Render(
				fmt.Sprintf("%s[%s] Role: %s | Plugins: %d | Method: %s | Installed: %s",
					cursor, si.Scope, si.Manifest.Role, len(si.Manifest.Plugins),
					si.Manifest.InstallMethod, si.Manifest.InstalledAt.Format("2006-01-02")),
			))
		} else {
			lines = append(lines, styles.SubtitleStyle.Render(
				fmt.Sprintf("  [%s] Not installed", si.Scope),
			))
		}
	}

	lines = append(lines, "")
	if hasInstalled {
		lines = append(lines, styles.HelpStyle.Render("↑↓ navigate  d uninstall"))
	} else {
		lines = append(lines, styles.HelpStyle.Render("No installations found"))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m Model) viewConfirm() string {
	idx := m.selectedScopeIndex()
	si := m.scopes[idx]
	manifest := si.Manifest

	title := styles.WarningStyle.Render("⚠ Confirm Uninstall")

	var lines []string
	lines = append(lines, "", title, "")
	lines = append(lines, styles.NormalStyle.Render(fmt.Sprintf("  Scope: %s", si.Scope)))
	lines = append(lines, styles.NormalStyle.Render(fmt.Sprintf("  Role:  %s", manifest.Role)))
	lines = append(lines, "")
	lines = append(lines, styles.NormalStyle.Render("  Plugins to remove:"))

	for _, p := range manifest.Plugins {
		lines = append(lines, styles.NormalStyle.Render(fmt.Sprintf("    • %s", p.Name)))
	}

	lines = append(lines, "")
	lines = append(lines, styles.HelpStyle.Render("y/enter confirm  n/esc cancel"))

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m Model) viewUninstalling() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		"",
		styles.TitleStyle.Render("Uninstalling..."),
		"",
		styles.SubtitleStyle.Render("Removing ASDS plugins and cleaning up"),
		"",
	)
}

func (m Model) viewResult() string {
	if m.errorMsg != "" {
		title := styles.ErrorStyle.Render("❌ Uninstall Failed")
		return lipgloss.JoinVertical(lipgloss.Left,
			"",
			title,
			"",
			styles.NormalStyle.Render(m.errorMsg),
			"",
			styles.HelpStyle.Render("enter to go back"),
		)
	}

	title := styles.SuccessStyle.Render("✅ Uninstall Complete!")

	var lines []string
	lines = append(lines, "", title, "")

	for _, r := range m.uninstResults {
		if r.Success {
			lines = append(lines, styles.SuccessStyle.Render(fmt.Sprintf("  ✓ removed %s", r.PluginRef)))
		} else {
			lines = append(lines, styles.ErrorStyle.Render(fmt.Sprintf("  ✗ %s: %v", r.PluginRef, r.Error)))
		}
	}

	lines = append(lines, "")
	lines = append(lines, styles.HelpStyle.Render("enter to continue"))

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}
