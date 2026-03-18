package setup

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/andrew-le-mfv/asds-marketplace-setup/internal/claude"
	"github.com/andrew-le-mfv/asds-marketplace-setup/internal/config"
	"github.com/andrew-le-mfv/asds-marketplace-setup/internal/installer"
)

// InstallCompleteMsg is sent when installation finishes.
type InstallCompleteMsg struct {
	Results []installer.InstallResult
	Error   error
}

// Update handles key events for the setup wizard.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch m.step {
		case StepRoleSelect:
			return m.updateRoleSelect(msg)
		case StepScopeSelect:
			return m.updateScopeSelect(msg)
		case StepConfirm:
			return m.updateConfirm(msg)
		case StepComplete, StepError:
			if msg.String() == "enter" || msg.String() == "esc" {
				m.step = StepRoleSelect
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
	}

	return m, nil
}

func (m Model) updateRoleSelect(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.selectedRole > 0 {
			m.selectedRole--
		}
	case "down", "j":
		if m.selectedRole < len(m.roles)-1 {
			m.selectedRole++
		}
	case "enter":
		m.step = StepScopeSelect
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
		m.step = StepRoleSelect
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

func (m Model) doInstall() tea.Cmd {
	return func() tea.Msg {
		roleID := m.SelectedRoleID()
		scope := m.SelectedScope()
		roleConfig := m.marketplaceCfg.Roles[roleID]

		inst := installer.NewInstaller(true)

		// Register marketplace (non-fatal if it fails)
		inst.RegisterMarketplace(
			m.marketplaceCfg.Marketplace.Name,
			m.marketplaceCfg.Marketplace.RegistryURL,
		)

		// Install plugins
		results, err := inst.Install(roleConfig.Plugins, scope, m.projectRoot)
		if err != nil {
			return InstallCompleteMsg{Error: err}
		}

		// Scaffold CLAUDE.md
		if scope != config.ScopeUser && len(roleConfig.ClaudeMDSnippets) > 0 {
			claudeMDPath, pathErr := claude.ClaudeMDPath(scope, m.projectRoot)
			if pathErr == nil {
				claude.UpsertMarkerBlock(claudeMDPath, roleID, roleConfig.ClaudeMDSnippets)
			}
		}

		// Write manifest
		now := time.Now().UTC()
		manifestPlugins := make([]config.ManifestPlugin, len(roleConfig.Plugins))
		for i, p := range roleConfig.Plugins {
			manifestPlugins[i] = config.ManifestPlugin{
				Name:        p.Name,
				FullRef:     p.Source,
				Required:    p.Required,
				InstalledAt: now,
			}
		}

		manifest := &config.Manifest{
			SchemaVersion:      1,
			ASDSVersion:        "0.1.0",
			InstalledAt:        now,
			UpdatedAt:          now,
			Role:               roleID,
			Scope:              scope,
			MarketplaceSource:  m.marketplaceCfg.Marketplace.RegistryURL,
			InstallMethod:      inst.Method(),
			ClaudeCodeDetected: installer.DetectClaudeCode().Found,
			Plugins:            manifestPlugins,
			ClaudeMDModified:   scope != config.ScopeUser,
		}

		manifestPath := claude.ManifestPath(scope, m.projectRoot)
		config.WriteManifest(manifestPath, manifest)

		// Gitignore for local scope
		if scope == config.ScopeLocal {
			claudeDir := claude.SettingsPath(scope, m.projectRoot)
			// Extract directory part
			for i := len(claudeDir) - 1; i >= 0; i-- {
				if claudeDir[i] == '/' {
					claude.EnsureGitignore(claudeDir[:i], ".asds-manifest.local.json")
					break
				}
			}
		}

		return InstallCompleteMsg{Results: results}
	}
}
