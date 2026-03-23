package plugins

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/andrew-le-mfv/asds-marketplace-setup/internal/tui/styles"
)

// View renders the plugin browser.
func (m Model) View() string {
	switch m.step {
	case StepBrowse:
		return m.viewBrowse()
	case StepDetail:
		return m.viewDetail()
	case StepScopeSelect:
		return m.viewScopeSelect()
	case StepConfirm:
		return m.viewConfirm()
	case StepInstalling:
		return m.viewInstalling()
	case StepComplete:
		return m.viewComplete()
	case StepUninstallConfirm:
		return m.viewUninstallConfirm()
	case StepUninstalling:
		return m.viewUninstalling()
	case StepUninstallComplete:
		return m.viewUninstallComplete()
	case StepError:
		return m.viewError()
	default:
		return ""
	}
}

func (m Model) viewBrowse() string {
	title := styles.TitleStyle.Render("📦 Available Plugins")
	subtitle := styles.SubtitleStyle.Render("Browse and manage plugins from the marketplace")

	var lines []string
	lines = append(lines, "", title, subtitle, "")

	for i, item := range m.items {
		cursor := "  "
		style := styles.NormalStyle
		if i == m.cursor {
			cursor = "▸ "
			style = styles.SelectedStyle
		}

		badge := ""
		if item.Required {
			badge += styles.WarningStyle.Render(" [required]")
		}
		if info, ok := m.installedPlugins[item.Source]; ok {
			badge += styles.SuccessStyle.Render(fmt.Sprintf(" ✓ installed [%s]", info.Scope))
		}

		line := fmt.Sprintf("%s%s%s", cursor, item.Name, badge)
		lines = append(lines, style.Render(line))
		if i == m.cursor {
			lines = append(lines, styles.SubtitleStyle.Render(fmt.Sprintf("    Source: %s | Role: %s", item.Source, item.RoleName)))
		}
	}

	lines = append(lines, "", styles.HelpStyle.Render("↑↓ navigate  enter view details"))

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m Model) viewDetail() string {
	plugin := m.SelectedPlugin()
	info, installed := m.installedPlugins[plugin.Source]

	title := styles.TitleStyle.Render(fmt.Sprintf("📦 %s", plugin.Name))

	var lines []string
	lines = append(lines, "", title, "")

	if installed {
		lines = append(lines, styles.SuccessStyle.Render(fmt.Sprintf("  ✓ Installed — scope: %s", info.Scope)))
	} else {
		lines = append(lines, styles.SubtitleStyle.Render("  Not installed"))
	}

	lines = append(lines, "")
	lines = append(lines, styles.NormalStyle.Render(fmt.Sprintf("  Source:   %s", plugin.Source)))
	lines = append(lines, styles.NormalStyle.Render(fmt.Sprintf("  Role:     %s", plugin.RoleName)))

	if plugin.Required {
		lines = append(lines, styles.WarningStyle.Render("  Required: yes"))
	}

	lines = append(lines, "")

	if installed {
		lines = append(lines, styles.HelpStyle.Render("i install to scope  d uninstall  esc back"))
	} else {
		lines = append(lines, styles.HelpStyle.Render("i install  esc back"))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m Model) viewScopeSelect() string {
	plugin := m.SelectedPlugin()
	title := styles.TitleStyle.Render("Select installation scope")
	pluginLine := styles.SubtitleStyle.Render(fmt.Sprintf("Plugin: %s", plugin.Name))

	var items []string
	for i, s := range m.scopes {
		cursor := "  "
		style := styles.NormalStyle
		if i == m.selectedScope {
			cursor = "▸ "
			style = styles.SelectedStyle
		}
		line := fmt.Sprintf("%s%s\n    %s", cursor, s.Label, styles.SubtitleStyle.Render(s.Description))
		items = append(items, style.Render(line))
	}

	help := styles.HelpStyle.Render("↑↓ navigate  enter select  esc back")

	return lipgloss.JoinVertical(lipgloss.Left,
		append([]string{"", title, pluginLine, ""}, append(items, "", help)...)...,
	)
}

func (m Model) viewConfirm() string {
	plugin := m.SelectedPlugin()
	scope := m.SelectedScope()

	title := styles.TitleStyle.Render("Confirm installation")

	var lines []string
	lines = append(lines, "", title, "")
	lines = append(lines, styles.NormalStyle.Render(fmt.Sprintf("  Plugin: %s", plugin.Name)))
	lines = append(lines, styles.NormalStyle.Render(fmt.Sprintf("  Source: %s", plugin.Source)))
	lines = append(lines, styles.NormalStyle.Render(fmt.Sprintf("  Scope:  %s", scope)))
	lines = append(lines, "")
	lines = append(lines, styles.HelpStyle.Render("enter/y confirm  esc/n go back"))

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m Model) viewInstalling() string {
	plugin := m.SelectedPlugin()
	return lipgloss.JoinVertical(lipgloss.Left,
		"",
		styles.TitleStyle.Render("Installing..."),
		"",
		styles.SubtitleStyle.Render("Setting up "+plugin.Name),
		"",
	)
}

func (m Model) viewComplete() string {
	title := styles.SuccessStyle.Render("✅ Plugin Installed!")

	var lines []string
	lines = append(lines, "", title, "")

	for _, r := range m.installResults {
		if r.Success {
			lines = append(lines, styles.SuccessStyle.Render(fmt.Sprintf("  ✓ %s", r.PluginRef)))
		} else {
			lines = append(lines, styles.ErrorStyle.Render(fmt.Sprintf("  ✗ %s: %v", r.PluginRef, r.Error)))
		}
	}

	lines = append(lines, "", styles.HelpStyle.Render("enter to continue"))

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m Model) viewUninstallConfirm() string {
	plugin := m.SelectedPlugin()
	info := m.installedPlugins[plugin.Source]

	title := styles.WarningStyle.Render("⚠ Confirm Uninstall")

	var lines []string
	lines = append(lines, "", title, "")
	lines = append(lines, styles.NormalStyle.Render(fmt.Sprintf("  Plugin: %s", plugin.Name)))
	lines = append(lines, styles.NormalStyle.Render(fmt.Sprintf("  Source: %s", plugin.Source)))
	lines = append(lines, styles.NormalStyle.Render(fmt.Sprintf("  Scope:  %s", info.Scope)))
	lines = append(lines, "")
	lines = append(lines, styles.HelpStyle.Render("y/enter confirm  n/esc cancel"))

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m Model) viewUninstalling() string {
	plugin := m.SelectedPlugin()
	return lipgloss.JoinVertical(lipgloss.Left,
		"",
		styles.TitleStyle.Render("Uninstalling..."),
		"",
		styles.SubtitleStyle.Render("Removing "+plugin.Name),
		"",
	)
}

func (m Model) viewUninstallComplete() string {
	title := styles.SuccessStyle.Render("✅ Plugin Removed!")

	var lines []string
	lines = append(lines, "", title, "")

	for _, r := range m.uninstResults {
		if r.Success {
			lines = append(lines, styles.SuccessStyle.Render(fmt.Sprintf("  ✓ removed %s", r.PluginRef)))
		} else {
			lines = append(lines, styles.ErrorStyle.Render(fmt.Sprintf("  ✗ %s: %v", r.PluginRef, r.Error)))
		}
	}

	lines = append(lines, "", styles.HelpStyle.Render("enter to continue"))

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m Model) viewError() string {
	title := styles.ErrorStyle.Render("❌ Operation Failed")

	return lipgloss.JoinVertical(lipgloss.Left,
		"",
		title,
		"",
		styles.NormalStyle.Render(m.errorMsg),
		"",
		styles.HelpStyle.Render("enter to go back"),
	)
}
