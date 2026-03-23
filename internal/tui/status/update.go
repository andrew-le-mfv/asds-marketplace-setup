package status

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/andrew-le-mfv/asds-marketplace-setup/internal/claude"
	"github.com/andrew-le-mfv/asds-marketplace-setup/internal/config"
	"github.com/andrew-le-mfv/asds-marketplace-setup/internal/installer"
)

// UninstallCompleteMsg is sent when uninstallation finishes.
type UninstallCompleteMsg struct {
	Results []installer.InstallResult
	Error   error
}

// Update handles messages for the status tab.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch m.step {
		case StepView:
			return m.updateView(msg)
		case StepConfirm:
			return m.updateConfirm(msg)
		case StepResult:
			if msg.String() == "enter" || msg.String() == "esc" {
				m.refresh()
				m.step = StepView
			}
		}

	case UninstallCompleteMsg:
		if msg.Error != nil {
			m.step = StepResult
			m.errorMsg = msg.Error.Error()
		} else {
			m.step = StepResult
			m.uninstResults = msg.Results
		}
	}

	return m, nil
}

func (m Model) updateView(msg tea.KeyMsg) (Model, tea.Cmd) {
	installed := m.installedScopes()

	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(installed)-1 {
			m.cursor++
		}
	case "d", "delete", "backspace":
		if len(installed) > 0 {
			m.step = StepConfirm
		}
	}
	return m, nil
}

func (m Model) updateConfirm(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "y", "enter":
		m.step = StepUninstalling
		return m, m.doUninstall()
	case "n", "esc":
		m.step = StepView
	}
	return m, nil
}

func (m Model) doUninstall() tea.Cmd {
	return func() tea.Msg {
		idx := m.selectedScopeIndex()
		if idx < 0 {
			return UninstallCompleteMsg{Error: fmt.Errorf("no scope selected")}
		}

		si := m.scopes[idx]
		manifest := si.Manifest

		pluginRefs := make([]string, len(manifest.Plugins))
		for i, p := range manifest.Plugins {
			pluginRefs[i] = p.FullRef
		}

		inst := installer.NewInstaller(true)
		results, err := inst.Uninstall(pluginRefs, si.Scope, m.projectRoot)
		if err != nil {
			return UninstallCompleteMsg{Error: err}
		}

		// Remove CLAUDE.md marker block
		if si.Scope != config.ScopeUser {
			claudeMDPath, pathErr := claude.ClaudeMDPath(si.Scope, m.projectRoot)
			if pathErr == nil {
				claude.RemoveMarkerBlock(claudeMDPath)
			}
		}

		// Remove manifest file
		manifestPath := claude.ManifestPath(si.Scope, m.projectRoot)
		os.Remove(manifestPath)

		return UninstallCompleteMsg{Results: results}
	}
}
