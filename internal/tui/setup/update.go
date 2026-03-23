package setup

import (
	"fmt"
	"os"
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

// UninstallCompleteMsg is sent when uninstallation finishes.
type UninstallCompleteMsg struct {
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
		case StepRoleDetail:
			return m.updateRoleDetail(msg)
		case StepUninstallConfirm:
			return m.updateUninstallConfirm(msg)
		case StepComplete, StepError, StepUninstallComplete:
			if msg.String() == "enter" || msg.String() == "esc" {
				m.refreshInstalled()
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
		m.refreshInstalled()

	case UninstallCompleteMsg:
		if msg.Error != nil {
			m.step = StepError
			m.errorMsg = msg.Error.Error()
		} else {
			m.step = StepUninstallComplete
			m.uninstResults = msg.Results
		}
		m.refreshInstalled()
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
		roleID := m.SelectedRoleID()
		if _, installed := m.installedRoles[roleID]; installed {
			m.step = StepRoleDetail
		} else {
			m.step = StepScopeSelect
		}
	}
	return m, nil
}

func (m Model) updateRoleDetail(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "d":
		m.step = StepUninstallConfirm
	case "i":
		m.step = StepScopeSelect
	case "esc":
		m.step = StepRoleSelect
	}
	return m, nil
}

func (m Model) updateUninstallConfirm(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "y", "enter":
		m.step = StepUninstalling
		return m, m.doUninstall()
	case "n", "esc":
		m.step = StepRoleDetail
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
		// Find the correct marketplace for this role
		var mktCfg *config.MarketplaceConfig
		for _, cfg := range m.marketplaceCfgs {
			if cfg.Marketplace.Name == m.roles[m.selectedRole].MarketplaceName {
				mktCfg = cfg
				break
			}
		}
		if mktCfg == nil {
			return InstallCompleteMsg{Error: fmt.Errorf("marketplace not found")}
		}
		roleConfig := mktCfg.Roles[roleID]

		inst := installer.NewInstaller(true)

		// Register marketplace (non-fatal if it fails)
		inst.RegisterMarketplace(
			mktCfg.Marketplace.Name,
			mktCfg.Marketplace.RegistryURL,
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
			ASDSVersion:        "0.1.1",
			InstalledAt:        now,
			UpdatedAt:          now,
			Role:               roleID,
			Scope:              scope,
			MarketplaceSource:  mktCfg.Marketplace.RegistryURL,
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

func (m Model) doUninstall() tea.Cmd {
	return func() tea.Msg {
		roleID := m.SelectedRoleID()
		info, ok := m.installedRoles[roleID]
		if !ok {
			return UninstallCompleteMsg{Error: fmt.Errorf("role %q is not installed", roleID)}
		}

		manifest := info.Manifest
		pluginRefs := make([]string, len(manifest.Plugins))
		for i, p := range manifest.Plugins {
			pluginRefs[i] = p.FullRef
		}

		inst := installer.NewInstaller(true)
		results, err := inst.Uninstall(pluginRefs, info.Scope, m.projectRoot)
		if err != nil {
			return UninstallCompleteMsg{Error: err}
		}

		// Remove CLAUDE.md marker block
		if info.Scope != config.ScopeUser {
			claudeMDPath, pathErr := claude.ClaudeMDPath(info.Scope, m.projectRoot)
			if pathErr == nil {
				claude.RemoveMarkerBlock(claudeMDPath)
			}
		}

		// Remove manifest file
		manifestPath := claude.ManifestPath(info.Scope, m.projectRoot)
		os.Remove(manifestPath)

		return UninstallCompleteMsg{Results: results}
	}
}
