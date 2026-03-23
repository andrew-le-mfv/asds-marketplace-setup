package plugins

import (
	"github.com/andrew-le-mfv/asds-marketplace-setup/internal/claude"
	"github.com/andrew-le-mfv/asds-marketplace-setup/internal/config"
	"github.com/andrew-le-mfv/asds-marketplace-setup/internal/installer"
)

// Step tracks the plugin browser's current state.
type Step int

const (
	StepBrowse Step = iota
	StepDetail
	StepScopeSelect
	StepConfirm
	StepInstalling
	StepComplete
	StepUninstallConfirm
	StepUninstalling
	StepUninstallComplete
	StepError
)

// PluginItem represents a plugin in the browser list.
type PluginItem struct {
	Name            string
	Source          string
	Required        bool
	RoleName        string
	MarketplaceName string
}

// installedScope tracks which scope a plugin is installed in.
type installedScope struct {
	Scope config.Scope
}

// scopeItem represents an installation scope option.
type scopeItem struct {
	Scope       config.Scope
	Label       string
	Description string
}

// Model holds the plugin browser state.
type Model struct {
	step            Step
	items           []PluginItem
	cursor          int
	selectedScope   int
	scopes          []scopeItem
	width           int
	height          int
	projectRoot     string
	marketplaceCfgs []*config.MarketplaceConfig
	installResults  []installer.InstallResult
	uninstResults   []installer.InstallResult
	errorMsg        string

	// installedPlugins maps plugin source to the scope it's installed in.
	installedPlugins map[string]installedScope
}

// New creates a plugin browser from marketplace configs.
func New(cfgs []*config.MarketplaceConfig, projectRoot string) Model {
	seen := make(map[string]bool)
	var items []PluginItem

	for _, cfg := range cfgs {
		for _, roleName := range cfg.RoleNames() {
			role := cfg.Roles[roleName]
			for _, p := range role.Plugins {
				if seen[p.Source] {
					continue
				}
				seen[p.Source] = true
				items = append(items, PluginItem{
					Name:            p.Name,
					Source:          p.Source,
					Required:        p.Required,
					RoleName:        roleName,
					MarketplaceName: cfg.Marketplace.Name,
				})
			}
		}
	}

	scopes := []scopeItem{
		{Scope: config.ScopeUser, Label: "User (global)", Description: "Install for you — ~/.claude/settings.json"},
		{Scope: config.ScopeProject, Label: "Project (shared)", Description: "Install for this project — .claude/settings.json"},
		{Scope: config.ScopeLocal, Label: "Local (private)", Description: "Install locally — .claude/settings.local.json"},
	}

	m := Model{
		step:            StepBrowse,
		items:           items,
		scopes:          scopes,
		projectRoot:     projectRoot,
		marketplaceCfgs: cfgs,
	}
	m.RefreshInstalled()
	return m
}

// SelectedPlugin returns the currently selected plugin.
func (m Model) SelectedPlugin() PluginItem {
	if m.cursor < len(m.items) {
		return m.items[m.cursor]
	}
	return PluginItem{}
}

// SelectedScope returns the currently selected scope.
func (m Model) SelectedScope() config.Scope {
	if m.selectedScope < len(m.scopes) {
		return m.scopes[m.selectedScope].Scope
	}
	return config.ScopeProject
}

// RefreshMarketplaces reloads marketplace configs and rebuilds the plugin items list.
func (m *Model) RefreshMarketplaces(cfgs []*config.MarketplaceConfig) {
	m.marketplaceCfgs = cfgs

	seen := make(map[string]bool)
	var items []PluginItem
	for _, cfg := range cfgs {
		for _, roleName := range cfg.RoleNames() {
			role := cfg.Roles[roleName]
			for _, p := range role.Plugins {
				if seen[p.Source] {
					continue
				}
				seen[p.Source] = true
				items = append(items, PluginItem{
					Name:            p.Name,
					Source:          p.Source,
					Required:        p.Required,
					RoleName:        roleName,
					MarketplaceName: cfg.Marketplace.Name,
				})
			}
		}
	}
	m.items = items
	if m.cursor >= len(m.items) {
		m.cursor = max(0, len(m.items)-1)
	}
	m.RefreshInstalled()
}

// RefreshInstalled rescans all scopes' settings files to find enabled plugins.
func (m *Model) RefreshInstalled() {
	m.installedPlugins = make(map[string]installedScope)
	for _, s := range []config.Scope{config.ScopeUser, config.ScopeProject, config.ScopeLocal} {
		settingsPath := claude.SettingsPath(s, m.projectRoot)
		settings, err := claude.ReadSettings(settingsPath)
		if err != nil {
			continue
		}
		ep, ok := settings["enabledPlugins"].(map[string]any)
		if !ok {
			continue
		}
		for source, enabled := range ep {
			if val, ok := enabled.(bool); ok && val {
				m.installedPlugins[source] = installedScope{Scope: s}
			}
		}
	}
}
