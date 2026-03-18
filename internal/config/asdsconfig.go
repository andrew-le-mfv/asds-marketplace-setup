package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const defaultMarketplaceURL = "github.com/your-org/asds-marketplace"

// ASDSConfig is the TUI's own configuration stored at ~/.config/asds/config.yaml.
type ASDSConfig struct {
	MarketplaceURL string `yaml:"marketplace_url"`
}

// DefaultASDSConfig returns the default ASDS configuration.
func DefaultASDSConfig() ASDSConfig {
	return ASDSConfig{
		MarketplaceURL: defaultMarketplaceURL,
	}
}

// ResolveASDSConfigPath returns the path to ~/.config/asds/config.yaml.
func ResolveASDSConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "~"
	}
	return filepath.Join(home, ".config", "asds", "config.yaml")
}

// ReadASDSConfig reads the ASDS config from disk.
// Returns defaults if the file does not exist.
func ReadASDSConfig(path string) (*ASDSConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := DefaultASDSConfig()
			return &cfg, nil
		}
		return nil, fmt.Errorf("reading ASDS config: %w", err)
	}

	cfg := DefaultASDSConfig()
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing ASDS config: %w", err)
	}
	return &cfg, nil
}

// WriteASDSConfig writes the ASDS config to disk.
func WriteASDSConfig(path string, cfg *ASDSConfig) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling ASDS config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing ASDS config: %w", err)
	}
	return nil
}
