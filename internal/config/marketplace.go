package config

import (
	"fmt"
	"sort"

	"gopkg.in/yaml.v3"
)

// MarketplaceConfig represents the parsed asds-marketplace.yaml.
type MarketplaceConfig struct {
	SchemaVersion int                 `yaml:"schema_version"`
	Marketplace   MarketplaceInfo     `yaml:"marketplace"`
	Roles         map[string]Role     `yaml:"roles"`
	Defaults      MarketplaceDefaults `yaml:"defaults"`
}

// MarketplaceInfo holds marketplace metadata.
type MarketplaceInfo struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Version     string `yaml:"version"`
	RegistryURL string `yaml:"registry_url"`
}

// Role defines a developer role and its associated plugins.
type Role struct {
	DisplayName      string      `yaml:"display_name"`
	Description      string      `yaml:"description"`
	Plugins          []PluginRef `yaml:"plugins"`
	ClaudeMDSnippets []string    `yaml:"claude_md_snippets"`
}

// PluginRef is a reference to a plugin in the marketplace.
type PluginRef struct {
	Name     string `yaml:"name"`
	Source   string `yaml:"source"`
	Required bool   `yaml:"required"`
}

// MarketplaceDefaults holds default values from the config.
type MarketplaceDefaults struct {
	Scope                   string `yaml:"scope"`
	AutoRegisterMarketplace bool   `yaml:"auto_register_marketplace"`
}

// ParseMarketplaceConfig parses YAML bytes into a MarketplaceConfig.
func ParseMarketplaceConfig(data []byte) (*MarketplaceConfig, error) {
	var cfg MarketplaceConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing marketplace config: %w", err)
	}
	return &cfg, nil
}

// RoleNames returns sorted role IDs.
func (c *MarketplaceConfig) RoleNames() []string {
	names := make([]string, 0, len(c.Roles))
	for k := range c.Roles {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}
