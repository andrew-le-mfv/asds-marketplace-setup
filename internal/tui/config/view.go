package config

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/andrew-le-mfv/asds-marketplace-setup/internal/tui/styles"
)

// View renders the marketplace manager.
func (m Model) View() string {
	switch m.step {
	case StepList:
		return m.viewList()
	case StepAdd:
		return m.viewForm("Add Marketplace")
	case StepEdit:
		return m.viewForm("Edit Marketplace")
	case StepRemoveConfirm:
		return m.viewRemoveConfirm()
	case StepDiscovering:
		return m.viewDiscovering()
	case StepDiscoverResult:
		return m.viewDiscoverResult()
	case StepError:
		return m.viewError()
	default:
		return ""
	}
}

func (m Model) viewList() string {
	title := styles.TitleStyle.Render("⚙ Marketplaces")
	subtitle := styles.SubtitleStyle.Render("Manage your marketplace sources")

	var lines []string
	lines = append(lines, "", title, subtitle, "")

	for i, mkt := range m.mktsCfg.Marketplaces {
		cursor := "  "
		style := styles.NormalStyle
		if i == m.cursor {
			cursor = "▸ "
			style = styles.SelectedStyle
		}

		enabledIcon := "✓"
		if !mkt.Enabled {
			enabledIcon = "✗"
		}

		line := fmt.Sprintf("%s[%s] %s", cursor, enabledIcon, mkt.Name)
		lines = append(lines, style.Render(line))
		if i == m.cursor {
			lines = append(lines, styles.SubtitleStyle.Render(fmt.Sprintf("    URL: %s", mkt.URL)))
		}
	}

	// "+ Add Marketplace" action item
	addIdx := len(m.mktsCfg.Marketplaces)
	addCursor := "  "
	addStyle := styles.NormalStyle
	if m.cursor == addIdx {
		addCursor = "▸ "
		addStyle = styles.SelectedStyle
	}
	lines = append(lines, addStyle.Render(addCursor+"+ Add Marketplace"))

	// "Load Defaults" toggle
	defaultsIdx := addIdx + 1
	defaultsCursor := "  "
	defaultsStyle := styles.NormalStyle
	if m.cursor == defaultsIdx {
		defaultsCursor = "▸ "
		defaultsStyle = styles.SelectedStyle
	}
	defaultsIcon := "✗"
	if m.mktsCfg.LoadDefaults {
		defaultsIcon = "✓"
	}
	lines = append(lines, defaultsStyle.Render(fmt.Sprintf("%s[%s] Load Defaults", defaultsCursor, defaultsIcon)))

	lines = append(lines, "")
	lines = append(lines, styles.HelpStyle.Render("  Config: "+m.cfgPath))
	lines = append(lines, "")
	lines = append(lines, styles.HelpStyle.Render("↑↓ navigate  enter select  e edit  d remove  space toggle"))

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m Model) viewForm(title string) string {
	t := styles.TitleStyle.Render(title)

	var lines []string
	lines = append(lines, "", t, "")

	nameLabel := "  Name: "
	urlLabel := "  URL:  "
	if m.activeField == fieldName {
		nameLabel = styles.SelectedStyle.Render(nameLabel)
	}
	if m.activeField == fieldURL {
		urlLabel = styles.SelectedStyle.Render(urlLabel)
	}

	lines = append(lines, nameLabel+m.nameInput.View())
	lines = append(lines, urlLabel+m.urlInput.View())
	lines = append(lines, "")
	lines = append(lines, styles.HelpStyle.Render("↑↓ switch field  enter save  esc cancel"))

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m Model) viewRemoveConfirm() string {
	mkt := m.mktsCfg.Marketplaces[m.cursor]
	title := styles.WarningStyle.Render("⚠ Remove Marketplace")

	var lines []string
	lines = append(lines, "", title, "")
	lines = append(lines, styles.NormalStyle.Render(fmt.Sprintf("  Name: %s", mkt.Name)))
	lines = append(lines, styles.NormalStyle.Render(fmt.Sprintf("  URL:  %s", mkt.URL)))
	lines = append(lines, "")
	lines = append(lines, styles.HelpStyle.Render("y/enter confirm  n/esc cancel"))

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m Model) viewDiscovering() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		"",
		styles.TitleStyle.Render("Discovering marketplace..."),
		"",
		styles.SubtitleStyle.Render("Checking if URL is a valid marketplace"),
		"",
	)
}

func (m Model) viewDiscoverResult() string {
	var lines []string
	if m.discoverOK {
		lines = append(lines, "", styles.SuccessStyle.Render("✅ Marketplace validated and saved!"), "")
	} else {
		lines = append(lines, "", styles.WarningStyle.Render("⚠ Saved but could not validate marketplace"), "")
		lines = append(lines, styles.SubtitleStyle.Render("  The URL was saved. Plugins will load if URL becomes available."))
		if m.discoverErr != "" {
			lines = append(lines, styles.SubtitleStyle.Render("  Error: "+m.discoverErr))
		}
	}
	lines = append(lines, "", styles.HelpStyle.Render("enter to continue"))

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m Model) viewError() string {
	title := styles.ErrorStyle.Render("❌ Error")

	return lipgloss.JoinVertical(lipgloss.Left,
		"",
		title,
		"",
		styles.NormalStyle.Render("  "+m.errorMsg),
		"",
		styles.HelpStyle.Render("enter to go back"),
	)
}
