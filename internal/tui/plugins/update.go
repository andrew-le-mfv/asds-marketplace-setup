package plugins

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/andrew-le-mfv/asds-marketplace-setup/internal/config"
	"github.com/andrew-le-mfv/asds-marketplace-setup/internal/installer"
)

// InstallCompleteMsg is sent when a plugin install finishes.
type InstallCompleteMsg struct {
	Results []installer.InstallResult
	Error   error
}

// UninstallCompleteMsg is sent when a plugin uninstall finishes.
type UninstallCompleteMsg struct {
	Results []installer.InstallResult
	Error   error
}

// Update handles messages for the plugins tab.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch m.step {
		case StepBrowse:
			return m.updateBrowse(msg)
		case StepDetail:
			return m.updateDetail(msg)
		case StepScopeSelect:
			return m.updateScopeSelect(msg)
		case StepConfirm:
			return m.updateConfirm(msg)
		case StepUninstallConfirm:
			return m.updateUninstallConfirm(msg)
		case StepComplete, StepError, StepUninstallComplete:
			if msg.String() == "enter" || msg.String() == "esc" {
				m.RefreshInstalled()
				m.step = StepBrowse
			}
		}

	case InstallCompleteMsg:
		if msg.Error != nil {
			m.step = StepError
			m.errorMsg = msg.Error.Error()
		} else {
			m.step = StepComplete
			m.installResults = msg.Results
		}
		m.RefreshInstalled()

	case UninstallCompleteMsg:
		if msg.Error != nil {
			m.step = StepError
			m.errorMsg = msg.Error.Error()
		} else {
			m.step = StepUninstallComplete
			m.uninstResults = msg.Results
		}
		m.RefreshInstalled()
	}

	return m, nil
}

func (m Model) updateBrowse(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.items)-1 {
			m.cursor++
		}
	case "enter":
		m.step = StepDetail
	}
	return m, nil
}

func (m Model) updateDetail(msg tea.KeyMsg) (Model, tea.Cmd) {
	plugin := m.SelectedPlugin()
	_, installed := m.installedPlugins[plugin.Source]

	switch msg.String() {
	case "enter":
		m.selectedScope = 0
		m.step = StepScopeSelect
	case "d":
		if installed {
			m.step = StepUninstallConfirm
		}
	case "esc":
		m.step = StepBrowse
	}
	return m, nil
}

func (m Model) updateScopeSelect(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.selectedScope > 0 {
			m.selectedScope--
		}
	case "down", "j":
		if m.selectedScope < len(m.scopes)-1 {
			m.selectedScope++
		}
	case "enter":
		m.step = StepConfirm
	case "esc":
		m.step = StepDetail
	}
	return m, nil
}

func (m Model) updateConfirm(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "y":
		m.step = StepInstalling
		return m, m.doInstall()
	case "esc", "n":
		m.step = StepScopeSelect
	}
	return m, nil
}

func (m Model) updateUninstallConfirm(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "y", "enter":
		m.step = StepUninstalling
		return m, m.doUninstall()
	case "n", "esc":
		m.step = StepDetail
	}
	return m, nil
}

func (m Model) doInstall() tea.Cmd {
	return func() tea.Msg {
		plugin := m.SelectedPlugin()
		scope := m.SelectedScope()

		inst := installer.NewInstaller(true)

		inst.RegisterMarketplace(
			m.marketplaceCfg.Marketplace.Name,
			m.marketplaceCfg.Marketplace.RegistryURL,
		)

		pluginRef := config.PluginRef{
			Name:     plugin.Name,
			Source:   plugin.Source,
			Required: plugin.Required,
		}

		results, err := inst.Install([]config.PluginRef{pluginRef}, scope, m.projectRoot)
		if err != nil {
			return InstallCompleteMsg{Error: err}
		}

		return InstallCompleteMsg{Results: results}
	}
}

func (m Model) doUninstall() tea.Cmd {
	return func() tea.Msg {
		plugin := m.SelectedPlugin()
		info, ok := m.installedPlugins[plugin.Source]
		if !ok {
			return UninstallCompleteMsg{Error: fmt.Errorf("plugin %q is not installed", plugin.Name)}
		}

		inst := installer.NewInstaller(true)
		results, err := inst.Uninstall([]string{plugin.Source}, info.Scope, m.projectRoot)
		if err != nil {
			return UninstallCompleteMsg{Error: err}
		}

		return UninstallCompleteMsg{Results: results}
	}
}
