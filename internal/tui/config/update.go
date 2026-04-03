package config

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	appconfig "github.com/andrew-le-mfv/asds-marketplace-setup/internal/config"
	"github.com/andrew-le-mfv/asds-marketplace-setup/pkg/registry"
)

// MarketplacesChangedMsg is sent when marketplaces are added, removed, or toggled.
// The app intercepts this to refresh other tabs.
type MarketplacesChangedMsg struct{}

// DiscoverCompleteMsg is sent when marketplace discovery finishes.
type DiscoverCompleteMsg struct {
	Error error
}

// Update handles messages for the config tab.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case DiscoverCompleteMsg:
		if msg.Error != nil {
			m.discoverOK = false
			m.discoverErr = msg.Error.Error()
		} else {
			m.discoverOK = true
			m.discoverErr = ""
		}
		m.step = StepDiscoverResult
		return m, func() tea.Msg { return MarketplacesChangedMsg{} }

	case tea.KeyMsg:
		switch m.step {
		case StepList:
			return m.updateList(msg)
		case StepAdd, StepEdit:
			return m.updateForm(msg)
		case StepRemoveConfirm:
			return m.updateRemoveConfirm(msg)
		case StepDiscovering:
			// Waiting, ignore keys
		case StepDiscoverResult:
			if msg.String() == "enter" || msg.String() == "esc" {
				m.reload()
				m.step = StepList
			}
		case StepError:
			if msg.String() == "enter" || msg.String() == "esc" {
				m.step = StepList
			}
		}
	}

	return m, nil
}

func (m Model) updateList(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		// Last two items are "+ Add Marketplace" and "Load Defaults" toggle
		if m.cursor < len(m.mktsCfg.Marketplaces)+1 {
			m.cursor++
		}
	case "enter":
		if m.cursor == len(m.mktsCfg.Marketplaces) {
			// "+ Add Marketplace" selected
			m.nameInput.Reset()
			m.urlInput.Reset()
			m.nameInput.Focus()
			m.activeField = fieldName
			m.step = StepAdd
		}
	case "a":
		m.nameInput.Reset()
		m.urlInput.Reset()
		m.nameInput.Focus()
		m.activeField = fieldName
		m.step = StepAdd
	case "e":
		if m.cursor < len(m.mktsCfg.Marketplaces) && len(m.mktsCfg.Marketplaces) > 0 {
			entry := m.mktsCfg.Marketplaces[m.cursor]
			m.nameInput.SetValue(entry.Name)
			m.urlInput.SetValue(entry.URL)
			m.nameInput.Focus()
			m.activeField = fieldName
			m.step = StepEdit
		}
	case "d", "delete", "backspace":
		if m.cursor < len(m.mktsCfg.Marketplaces) && len(m.mktsCfg.Marketplaces) > 0 {
			m.step = StepRemoveConfirm
		}
	case " ":
		if m.cursor == len(m.mktsCfg.Marketplaces)+1 {
			// Toggle "Load Defaults"
			m.mktsCfg.LoadDefaults = !m.mktsCfg.LoadDefaults
			m.save()
			return m, func() tea.Msg { return MarketplacesChangedMsg{} }
		}
		if m.cursor < len(m.mktsCfg.Marketplaces) && len(m.mktsCfg.Marketplaces) > 0 {
			m.mktsCfg.Marketplaces[m.cursor].Enabled = !m.mktsCfg.Marketplaces[m.cursor].Enabled
			m.save()
			return m, func() tea.Msg { return MarketplacesChangedMsg{} }
		}
	}
	return m, nil
}

func (m Model) updateForm(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "up", "down":
		if m.activeField == fieldName {
			m.nameInput.Blur()
			m.urlInput.Focus()
			m.activeField = fieldURL
		} else {
			m.urlInput.Blur()
			m.nameInput.Focus()
			m.activeField = fieldName
		}
		return m, nil
	case "enter":
		name := m.nameInput.Value()
		url := m.urlInput.Value()
		if name == "" || url == "" {
			m.errorMsg = "name and URL are required"
			m.step = StepError
			return m, nil
		}

		entry := appconfig.MarketplaceEntry{Name: name, URL: url, Enabled: true}

		if m.step == StepAdd {
			if err := m.mktsCfg.AddMarketplace(entry); err != nil {
				m.errorMsg = err.Error()
				m.step = StepError
				return m, nil
			}
		} else {
			oldName := m.mktsCfg.Marketplaces[m.cursor].Name
			if err := m.mktsCfg.UpdateMarketplace(oldName, entry); err != nil {
				m.errorMsg = err.Error()
				m.step = StepError
				return m, nil
			}
		}

		m.save()
		m.step = StepDiscovering
		return m, m.doDiscover(name, url)
	case "esc":
		m.step = StepList
		return m, nil
	}

	// Forward to active text input
	var cmd tea.Cmd
	if m.activeField == fieldName {
		m.nameInput, cmd = m.nameInput.Update(msg)
	} else {
		m.urlInput, cmd = m.urlInput.Update(msg)
	}
	return m, cmd
}

func (m Model) updateRemoveConfirm(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "y", "enter":
		name := m.mktsCfg.Marketplaces[m.cursor].Name
		m.mktsCfg.RemoveMarketplace(name)
		m.save()
		if m.cursor >= len(m.mktsCfg.Marketplaces) && m.cursor > 0 {
			m.cursor--
		}
		m.step = StepList
		return m, func() tea.Msg { return MarketplacesChangedMsg{} }
	case "n", "esc":
		m.step = StepList
	}
	return m, nil
}

func (m Model) doDiscover(name, url string) tea.Cmd {
	return func() tea.Msg {
		_, err := registry.DiscoverMarketplace(url, name)
		if err != nil {
			return DiscoverCompleteMsg{Error: fmt.Errorf("discovery failed: %w", err)}
		}
		return DiscoverCompleteMsg{}
	}
}
