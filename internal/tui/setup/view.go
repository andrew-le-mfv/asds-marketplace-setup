package setup

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/your-org/asds-marketplace-setup/internal/tui/styles"
)

// View renders the setup wizard.
func (m Model) View() string {
	switch m.step {
	case StepRoleSelect:
		return m.viewRoleSelect()
	case StepScopeSelect:
		return m.viewScopeSelect()
	case StepConfirm:
		return m.viewConfirm()
	case StepInstalling:
		return m.viewInstalling()
	case StepComplete:
		return m.viewComplete()
	case StepError:
		return m.viewError()
	default:
		return ""
	}
}

func (m Model) viewRoleSelect() string {
	title := styles.TitleStyle.Render("Select your role")
	subtitle := styles.SubtitleStyle.Render("Choose the role that best matches your work")

	var items []string
	for i, r := range m.roles {
		cursor := "  "
		style := styles.NormalStyle
		if i == m.selectedRole {
			cursor = "▸ "
			style = styles.SelectedStyle
		}
		line := fmt.Sprintf("%s%s — %s (%d plugins)", cursor, r.DisplayName, r.Description, r.PluginCount)
		items = append(items, style.Render(line))
	}

	help := styles.HelpStyle.Render("↑↓ navigate  enter select")

	content := lipgloss.JoinVertical(lipgloss.Left,
		append([]string{"", title, subtitle, ""}, append(items, "", help)...)...,
	)

	return styles.BoxStyle.Render(content)
}

func (m Model) viewScopeSelect() string {
	title := styles.TitleStyle.Render("Select installation scope")
	roleLine := styles.SubtitleStyle.Render(fmt.Sprintf("Role: %s", m.roles[m.selectedRole].DisplayName))

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

	content := lipgloss.JoinVertical(lipgloss.Left,
		append([]string{"", title, roleLine, ""}, append(items, "", help)...)...,
	)

	return styles.BoxStyle.Render(content)
}

func (m Model) viewConfirm() string {
	roleID := m.SelectedRoleID()
	role := m.marketplaceCfg.Roles[roleID]
	scope := m.SelectedScope()

	title := styles.TitleStyle.Render("Confirm installation")

	var lines []string
	lines = append(lines, "")
	lines = append(lines, title)
	lines = append(lines, "")
	lines = append(lines, styles.NormalStyle.Render(fmt.Sprintf("  Role:  %s (%s)", role.DisplayName, roleID)))
	lines = append(lines, styles.NormalStyle.Render(fmt.Sprintf("  Scope: %s", scope)))
	lines = append(lines, "")
	lines = append(lines, styles.NormalStyle.Render("  Plugins:"))

	for _, p := range role.Plugins {
		req := ""
		if p.Required {
			req = styles.WarningStyle.Render(" (required)")
		}
		lines = append(lines, styles.NormalStyle.Render(fmt.Sprintf("    • %s%s", p.Name, req)))
	}

	lines = append(lines, "")
	lines = append(lines, styles.HelpStyle.Render("enter/y confirm  esc/n go back"))

	return styles.BoxStyle.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

func (m Model) viewInstalling() string {
	content := lipgloss.JoinVertical(lipgloss.Left,
		"",
		styles.TitleStyle.Render("Installing..."),
		"",
		styles.SubtitleStyle.Render("Setting up ASDS plugins for "+m.roles[m.selectedRole].DisplayName),
		"",
	)
	return styles.BoxStyle.Render(content)
}

func (m Model) viewComplete() string {
	title := styles.SuccessStyle.Render("✅ Installation Complete!")

	var lines []string
	lines = append(lines, "", title, "")

	for _, r := range m.installResults {
		if r.Success {
			lines = append(lines, styles.SuccessStyle.Render(fmt.Sprintf("  ✓ %s", r.PluginRef)))
		} else {
			lines = append(lines, styles.ErrorStyle.Render(fmt.Sprintf("  ✗ %s: %v", r.PluginRef, r.Error)))
		}
	}

	lines = append(lines, "")
	lines = append(lines, styles.HelpStyle.Render("enter to continue"))

	return styles.BoxStyle.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

func (m Model) viewError() string {
	title := styles.ErrorStyle.Render("❌ Installation Failed")

	content := lipgloss.JoinVertical(lipgloss.Left,
		"",
		title,
		"",
		styles.NormalStyle.Render(m.errorMsg),
		"",
		styles.HelpStyle.Render("enter to go back"),
	)

	return styles.BoxStyle.Render(content)
}
