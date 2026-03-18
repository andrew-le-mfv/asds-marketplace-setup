package setup

import (
	"github.com/your-org/asds-marketplace-setup/internal/config"
	"github.com/your-org/asds-marketplace-setup/internal/installer"
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
)

// Model holds the setup wizard state.
type Model struct {
	step           Step
	roles          []roleItem
	selectedRole   int
	scopes         []scopeItem
	selectedScope  int
	marketplaceCfg *config.MarketplaceConfig
	projectRoot    string
	installResults []installer.InstallResult
	errorMsg       string
	width          int
	height         int
}

type roleItem struct {
	ID          string
	DisplayName string
	Description string
	PluginCount int
}

type scopeItem struct {
	Scope       config.Scope
	Label       string
	Description string
}

// New creates a new setup wizard model.
func New(cfg *config.MarketplaceConfig, projectRoot string) Model {
	roles := make([]roleItem, 0, len(cfg.Roles))
	for _, name := range cfg.RoleNames() {
		r := cfg.Roles[name]
		roles = append(roles, roleItem{
			ID:          name,
			DisplayName: r.DisplayName,
			Description: r.Description,
			PluginCount: len(r.Plugins),
		})
	}

	scopes := []scopeItem{
		{Scope: config.ScopeUser, Label: "User (global)", Description: "Install for you — ~/.claude/settings.json"},
		{Scope: config.ScopeProject, Label: "Project (shared)", Description: "Install for this project — .claude/settings.json"},
		{Scope: config.ScopeLocal, Label: "Local (private)", Description: "Install locally — .claude/settings.local.json"},
	}

	return Model{
		step:           StepRoleSelect,
		roles:          roles,
		scopes:         scopes,
		marketplaceCfg: cfg,
		projectRoot:    projectRoot,
	}
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
