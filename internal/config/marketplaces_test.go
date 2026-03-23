package config_test

import (
	"path/filepath"
	"testing"

	"github.com/andrew-le-mfv/asds-marketplace-setup/internal/config"
)

func TestMarketplacesConfig_Defaults(t *testing.T) {
	cfg := config.DefaultMarketplacesConfig()
	if len(cfg.Marketplaces) != 1 {
		t.Fatalf("expected 1 default marketplace, got %d", len(cfg.Marketplaces))
	}
	if cfg.Marketplaces[0].Name != "asds-marketplace" {
		t.Errorf("default name = %q, want %q", cfg.Marketplaces[0].Name, "asds-marketplace")
	}
	if !cfg.Marketplaces[0].Enabled {
		t.Error("default marketplace should be enabled")
	}
}

func TestMarketplacesConfig_WriteAndRead(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "marketplaces.yaml")

	cfg := &config.MarketplacesConfig{
		Marketplaces: []config.MarketplaceEntry{
			{Name: "test", URL: "github.com/test/plugins", Enabled: true},
		},
	}

	if err := config.WriteMarketplacesConfig(path, cfg); err != nil {
		t.Fatalf("write error: %v", err)
	}

	loaded, err := config.ReadMarketplacesConfig(path)
	if err != nil {
		t.Fatalf("read error: %v", err)
	}

	if len(loaded.Marketplaces) != 1 {
		t.Fatalf("expected 1 marketplace, got %d", len(loaded.Marketplaces))
	}
	if loaded.Marketplaces[0].Name != "test" {
		t.Errorf("name = %q, want %q", loaded.Marketplaces[0].Name, "test")
	}
}

func TestReadMarketplacesConfig_NotFound_ReturnsDefaults(t *testing.T) {
	cfg, err := config.ReadMarketplacesConfig("/nonexistent/marketplaces.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Marketplaces) != 1 {
		t.Fatalf("expected 1 default marketplace, got %d", len(cfg.Marketplaces))
	}
}

func TestMarketplacesConfig_AddMarketplace(t *testing.T) {
	cfg := config.DefaultMarketplacesConfig()

	err := cfg.AddMarketplace(config.MarketplaceEntry{Name: "custom", URL: "github.com/custom/mkt", Enabled: true})
	if err != nil {
		t.Fatalf("add error: %v", err)
	}
	if len(cfg.Marketplaces) != 2 {
		t.Fatalf("expected 2 marketplaces, got %d", len(cfg.Marketplaces))
	}

	// Duplicate name should fail
	err = cfg.AddMarketplace(config.MarketplaceEntry{Name: "custom", URL: "github.com/other", Enabled: true})
	if err == nil {
		t.Error("expected error for duplicate name")
	}
}

func TestMarketplacesConfig_RemoveMarketplace(t *testing.T) {
	cfg := config.DefaultMarketplacesConfig()
	cfg.AddMarketplace(config.MarketplaceEntry{Name: "custom", URL: "github.com/custom/mkt", Enabled: true})

	if err := cfg.RemoveMarketplace("custom"); err != nil {
		t.Fatalf("remove error: %v", err)
	}
	if len(cfg.Marketplaces) != 1 {
		t.Fatalf("expected 1 marketplace after remove, got %d", len(cfg.Marketplaces))
	}

	if err := cfg.RemoveMarketplace("nonexistent"); err == nil {
		t.Error("expected error for nonexistent marketplace")
	}
}

func TestMarketplacesConfig_UpdateMarketplace(t *testing.T) {
	cfg := config.DefaultMarketplacesConfig()

	err := cfg.UpdateMarketplace("asds-marketplace", config.MarketplaceEntry{
		Name:    "asds-marketplace",
		URL:     "github.com/updated/url",
		Enabled: false,
	})
	if err != nil {
		t.Fatalf("update error: %v", err)
	}
	if cfg.Marketplaces[0].URL != "github.com/updated/url" {
		t.Errorf("URL not updated: %q", cfg.Marketplaces[0].URL)
	}
	if cfg.Marketplaces[0].Enabled {
		t.Error("expected disabled after update")
	}
}

func TestMarketplacesConfig_EnabledMarketplaces(t *testing.T) {
	cfg := &config.MarketplacesConfig{
		Marketplaces: []config.MarketplaceEntry{
			{Name: "a", URL: "url-a", Enabled: true},
			{Name: "b", URL: "url-b", Enabled: false},
			{Name: "c", URL: "url-c", Enabled: true},
		},
	}

	enabled := cfg.EnabledMarketplaces()
	if len(enabled) != 2 {
		t.Fatalf("expected 2 enabled, got %d", len(enabled))
	}
}

func TestMarketplacesConfig_FindMarketplace(t *testing.T) {
	cfg := config.DefaultMarketplacesConfig()

	found := cfg.FindMarketplace("asds-marketplace")
	if found == nil {
		t.Fatal("expected to find asds-marketplace")
	}

	notFound := cfg.FindMarketplace("nonexistent")
	if notFound != nil {
		t.Error("expected nil for nonexistent marketplace")
	}
}
