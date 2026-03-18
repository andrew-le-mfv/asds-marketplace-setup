package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/your-org/asds-marketplace-setup/internal/config"
)

func TestASDSConfig_Defaults(t *testing.T) {
	cfg := config.DefaultASDSConfig()
	if cfg.MarketplaceURL != "github.com/your-org/asds-marketplace" {
		t.Errorf("marketplace_url = %q, want default", cfg.MarketplaceURL)
	}
}

func TestASDSConfig_WriteAndRead(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	cfg := config.ASDSConfig{
		MarketplaceURL: "github.com/custom/marketplace",
	}

	if err := config.WriteASDSConfig(path, &cfg); err != nil {
		t.Fatalf("WriteASDSConfig error: %v", err)
	}

	loaded, err := config.ReadASDSConfig(path)
	if err != nil {
		t.Fatalf("ReadASDSConfig error: %v", err)
	}

	if loaded.MarketplaceURL != "github.com/custom/marketplace" {
		t.Errorf("marketplace_url = %q, want %q", loaded.MarketplaceURL, "github.com/custom/marketplace")
	}
}

func TestReadASDSConfig_NotFound_ReturnsDefaults(t *testing.T) {
	cfg, err := config.ReadASDSConfig("/nonexistent/config.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.MarketplaceURL != "github.com/your-org/asds-marketplace" {
		t.Errorf("expected default marketplace_url, got %q", cfg.MarketplaceURL)
	}
}

func TestResolveASDSConfigPath(t *testing.T) {
	path := config.ResolveASDSConfigPath()
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".config", "asds", "config.yaml")
	if path != expected {
		t.Errorf("config path = %q, want %q", path, expected)
	}
}
