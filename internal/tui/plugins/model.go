package plugins

import (
	"github.com/andrew-le-mfv/asds-marketplace-setup/internal/config"
)

// PluginItem represents a plugin in the browser list.
type PluginItem struct {
	Name     string
	Source   string
	Required bool
	RoleName string
}

// Model holds the plugin browser state.
type Model struct {
	items  []PluginItem
	cursor int
	width  int
	height int
}

// New creates a plugin browser from a marketplace config.
func New(cfg *config.MarketplaceConfig) Model {
	seen := make(map[string]bool)
	var items []PluginItem

	for _, roleName := range cfg.RoleNames() {
		role := cfg.Roles[roleName]
		for _, p := range role.Plugins {
			if seen[p.Source] {
				continue
			}
			seen[p.Source] = true
			items = append(items, PluginItem{
				Name:     p.Name,
				Source:   p.Source,
				Required: p.Required,
				RoleName: roleName,
			})
		}
	}

	return Model{items: items}
}
