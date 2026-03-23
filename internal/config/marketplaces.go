package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// MarketplaceEntry represents a single marketplace source in the user's configuration.
type MarketplaceEntry struct {
	Name    string `yaml:"name"`
	URL     string `yaml:"url"`
	Enabled bool   `yaml:"enabled"`
}

// MarketplacesConfig stores the user's list of configured marketplaces.
// Persisted at ~/.config/asds/marketplaces.yaml.
type MarketplacesConfig struct {
	Marketplaces []MarketplaceEntry `yaml:"marketplaces"`
}

// DefaultMarketplacesConfig returns a config with the built-in official marketplace.
func DefaultMarketplacesConfig() MarketplacesConfig {
	return MarketplacesConfig{
		Marketplaces: []MarketplaceEntry{
			{
				Name:    "asds-marketplace",
				URL:     "github.com/anthropics/claude-plugins-official",
				Enabled: true,
			},
		},
	}
}

// ResolveMarketplacesConfigPath returns the path to ~/.config/asds/marketplaces.yaml.
func ResolveMarketplacesConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "~"
	}
	return filepath.Join(home, ".config", "asds", "marketplaces.yaml")
}

// ReadMarketplacesConfig reads the marketplaces config from disk.
// Returns defaults if the file does not exist.
func ReadMarketplacesConfig(path string) (*MarketplacesConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := DefaultMarketplacesConfig()
			return &cfg, nil
		}
		return nil, fmt.Errorf("reading marketplaces config: %w", err)
	}

	var cfg MarketplacesConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing marketplaces config: %w", err)
	}
	return &cfg, nil
}

// WriteMarketplacesConfig writes the marketplaces config to disk.
func WriteMarketplacesConfig(path string, cfg *MarketplacesConfig) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating marketplaces config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling marketplaces config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing marketplaces config: %w", err)
	}
	return nil
}

// AddMarketplace appends a new marketplace entry if the name doesn't already exist.
func (c *MarketplacesConfig) AddMarketplace(entry MarketplaceEntry) error {
	for _, m := range c.Marketplaces {
		if m.Name == entry.Name {
			return fmt.Errorf("marketplace %q already exists", entry.Name)
		}
	}
	c.Marketplaces = append(c.Marketplaces, entry)
	return nil
}

// RemoveMarketplace removes a marketplace by name.
func (c *MarketplacesConfig) RemoveMarketplace(name string) error {
	for i, m := range c.Marketplaces {
		if m.Name == name {
			c.Marketplaces = append(c.Marketplaces[:i], c.Marketplaces[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("marketplace %q not found", name)
}

// UpdateMarketplace updates an existing marketplace entry by name.
func (c *MarketplacesConfig) UpdateMarketplace(name string, updated MarketplaceEntry) error {
	for i, m := range c.Marketplaces {
		if m.Name == name {
			c.Marketplaces[i] = updated
			return nil
		}
	}
	return fmt.Errorf("marketplace %q not found", name)
}

// FindMarketplace returns the marketplace entry with the given name, or nil.
func (c *MarketplacesConfig) FindMarketplace(name string) *MarketplaceEntry {
	for i, m := range c.Marketplaces {
		if m.Name == name {
			return &c.Marketplaces[i]
		}
	}
	return nil
}

// EnabledMarketplaces returns only the enabled marketplace entries.
func (c *MarketplacesConfig) EnabledMarketplaces() []MarketplaceEntry {
	var result []MarketplaceEntry
	for _, m := range c.Marketplaces {
		if m.Enabled {
			result = append(result, m)
		}
	}
	return result
}
