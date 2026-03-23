package setup

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/andrew-le-mfv/asds-marketplace-setup/internal/config"
	"github.com/andrew-le-mfv/asds-marketplace-setup/internal/tui/styles"
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
	case StepRoleDetail:
		return m.viewRoleDetail()
	case StepUninstallConfirm:
		return m.viewUninstallConfirm()
	case StepUninstalling:
		return m.viewUninstalling()
	case StepUninstallComplete:
		return m.viewUninstallComplete()
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

		badge := ""
		if info, ok := m.installedRoles[r.ID]; ok {
			badge = styles.SuccessStyle.Render(fmt.Sprintf(" ✓ installed [%s]", info.Scope))
		}

		line := fmt.Sprintf("%s%s — %s (%d plugins) [%s]%s", cursor, r.DisplayName, r.Description, r.PluginCount, r.MarketplaceName, badge)
		items = append(items, style.Render(line))
	}

	help := styles.HelpStyle.Render("↑↓ navigate  enter select/view")

	return lipgloss.JoinVertical(lipgloss.Left,
		append([]string{"", title, subtitle, ""}, append(items, "", help)...)...,
	)
}

func (m Model) viewRoleDetail() string {
	roleID := m.SelectedRoleID()
	var role config.Role
	for _, cfg := range m.marketplaceCfgs {
		if cfg.Marketplace.Name == m.roles[m.selectedRole].MarketplaceName {
			role = cfg.Roles[roleID]
			break
		}
	}
	info := m.installedRoles[roleID]
	manifest := info.Manifest

	title := styles.TitleStyle.Render(fmt.Sprintf("📋 %s", role.DisplayName))
	statusLine := styles.SuccessStyle.Render(fmt.Sprintf("  ✓ Installed — scope: %s | method: %s | %s",
		info.Scope, manifest.InstallMethod, manifest.InstalledAt.Format("2006-01-02")))

	var lines []string
	lines = append(lines, "", title, "", statusLine, "")
	lines = append(lines, styles.NormalStyle.Render("  Plugins:"))

	for _, p := range manifest.Plugins {
		lines = append(lines, styles.NormalStyle.Render(fmt.Sprintf("    • %s", p.Name)))
	}

	lines = append(lines, "")
	lines = append(lines, styles.HelpStyle.Render("d uninstall  i reinstall  esc back"))

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m Model) viewUninstallConfirm() string {
	roleID := m.SelectedRoleID()
	var role config.Role
	for _, cfg := range m.marketplaceCfgs {
		if cfg.Marketplace.Name == m.roles[m.selectedRole].MarketplaceName {
			role = cfg.Roles[roleID]
			break
		}
	}
	info := m.installedRoles[roleID]

	title := styles.WarningStyle.Render("⚠ Confirm Uninstall")

	var lines []string
	lines = append(lines, "", title, "")
	lines = append(lines, styles.NormalStyle.Render(fmt.Sprintf("  Role:  %s", role.DisplayName)))
	lines = append(lines, styles.NormalStyle.Render(fmt.Sprintf("  Scope: %s", info.Scope)))
	lines = append(lines, "")
	lines = append(lines, styles.NormalStyle.Render("  Plugins to remove:"))

	for _, p := range info.Manifest.Plugins {
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
		styles.SubtitleStyle.Render("Removing plugins for "+m.roles[m.selectedRole].DisplayName),
		"",
	)
}

func (m Model) viewUninstallComplete() string {
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

	return lipgloss.JoinVertical(lipgloss.Left,
		append([]string{"", title, roleLine, ""}, append(items, "", help)...)...,
	)
}

func (m Model) viewConfirm() string {
	roleID := m.SelectedRoleID()
	var role config.Role
	for _, cfg := range m.marketplaceCfgs {
		if cfg.Marketplace.Name == m.roles[m.selectedRole].MarketplaceName {
			role = cfg.Roles[roleID]
			break
		}
	}
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

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m Model) viewInstalling() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		"",
		styles.TitleStyle.Render("Installing..."),
		"",
		styles.SubtitleStyle.Render("Setting up ASDS plugins for "+m.roles[m.selectedRole].DisplayName),
		"",
	)
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
