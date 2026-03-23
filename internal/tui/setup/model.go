package setup

import (
	"github.com/andrew-le-mfv/asds-marketplace-setup/internal/claude"
	"github.com/andrew-le-mfv/asds-marketplace-setup/internal/config"
	"github.com/andrew-le-mfv/asds-marketplace-setup/internal/installer"
)

// Step tracks the wizard's current position.
type Step int

const (
	StepRoleSelect Step = iota
	StepScopeSelect
	StepConfirm
	StepInstalling
	StepComplete
	StepError
	StepRoleDetail
	StepUninstallConfirm
	StepUninstalling
	StepUninstallComplete
)

// installedInfo tracks an installed role's manifest and scope.
type installedInfo struct {
	Scope    config.Scope
	Manifest *config.Manifest
}

// Model holds the setup wizard state.
type Model struct {
	step            Step
	roles           []roleItem
	selectedRole    int
	scopes          []scopeItem
	selectedScope   int
	marketplaceCfgs []*config.MarketplaceConfig
	projectRoot     string
	installResults  []installer.InstallResult
	uninstResults   []installer.InstallResult
	errorMsg        string
	width           int
	height          int

	// installedRoles maps roleID to its installation info.
	installedRoles map[string]installedInfo
}

type roleItem struct {
	ID              string
	DisplayName     string
	Description     string
	PluginCount     int
	MarketplaceName string
}

type scopeItem struct {
	Scope       config.Scope
	Label       string
	Description string
}

// New creates a new setup wizard model.
func New(cfgs []*config.MarketplaceConfig, projectRoot string) Model {
	var roles []roleItem
	seen := make(map[string]bool)

	for _, cfg := range cfgs {
		for _, name := range cfg.RoleNames() {
			key := cfg.Marketplace.Name + ":" + name
			if seen[key] {
				continue
			}
			seen[key] = true
			r := cfg.Roles[name]
			roles = append(roles, roleItem{
				ID:              name,
				DisplayName:     r.DisplayName,
				Description:     r.Description,
				PluginCount:     len(r.Plugins),
				MarketplaceName: cfg.Marketplace.Name,
			})
		}
	}

	scopes := []scopeItem{
		{Scope: config.ScopeUser, Label: "User (global)", Description: "Install for you — ~/.claude/settings.json"},
		{Scope: config.ScopeProject, Label: "Project (shared)", Description: "Install for this project — .claude/settings.json"},
		{Scope: config.ScopeLocal, Label: "Local (private)", Description: "Install locally — .claude/settings.local.json"},
	}

	m := Model{
		step:            StepRoleSelect,
		roles:           roles,
		scopes:          scopes,
		marketplaceCfgs: cfgs,
		projectRoot:     projectRoot,
		installedRoles:  make(map[string]installedInfo),
	}
	m.refreshInstalled()
	return m
}

// SelectedRoleID returns the currently selected role ID.
func (m Model) SelectedRoleID() string {
	if m.selectedRole < len(m.roles) {
		return m.roles[m.selectedRole].ID
	}
	return ""
}

// SelectedScope returns the currently selected scope.
func (m Model) SelectedScope() config.Scope {
	if m.selectedScope < len(m.scopes) {
		return m.scopes[m.selectedScope].Scope
	}
	return config.ScopeProject
}

// RefreshMarketplaces reloads marketplace configs and rebuilds the roles list.
func (m *Model) RefreshMarketplaces(cfgs []*config.MarketplaceConfig) {
	m.marketplaceCfgs = cfgs

	seen := make(map[string]bool)
	var roles []roleItem
	for _, cfg := range cfgs {
		for _, name := range cfg.RoleNames() {
			key := cfg.Marketplace.Name + ":" + name
			if seen[key] {
				continue
			}
			seen[key] = true
			r := cfg.Roles[name]
			roles = append(roles, roleItem{
				ID:              name,
				DisplayName:     r.DisplayName,
				Description:     r.Description,
				PluginCount:     len(r.Plugins),
				MarketplaceName: cfg.Marketplace.Name,
			})
		}
	}
	m.roles = roles
	if m.selectedRole >= len(m.roles) {
		m.selectedRole = max(0, len(m.roles)-1)
	}
	m.refreshInstalled()
}

// refreshInstalled scans all scopes for manifests and updates installedRoles.
func (m *Model) refreshInstalled() {
	m.installedRoles = make(map[string]installedInfo)
	for _, s := range []config.Scope{config.ScopeUser, config.ScopeProject, config.ScopeLocal} {
		mp := claude.ManifestPath(s, m.projectRoot)
		manifest, err := config.ReadManifest(mp)
		if err == nil && manifest.Role != "" {
			m.installedRoles[manifest.Role] = installedInfo{
				Scope:    s,
				Manifest: manifest,
			}
		}
	}
}
