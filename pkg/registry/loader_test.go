package registry_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/andrew-le-mfv/asds-marketplace-setup/internal/config"
	"github.com/andrew-le-mfv/asds-marketplace-setup/pkg/registry"
)

func TestLoadAllMarketplaces_DefaultsOnly(t *testing.T) {
	cfgs := registry.LoadAllMarketplaces("/nonexistent", "/nonexistent/project")
	if len(cfgs) == 0 {
		t.Fatal("expected at least 1 marketplace from defaults")
	}
	if cfgs[0].Marketplace.Name != "asds-marketplace" {
		t.Errorf("expected default marketplace, got %q", cfgs[0].Marketplace.Name)
	}
}

func TestLoadAllMarketplaces_WithProjectMarketplace(t *testing.T) {
	projectDir := t.TempDir()
	mktYAML := `
schema_version: 1
marketplace:
  name: "project-mkt"
  description: "Project local"
  version: "1.0.0"
  registry_url: "local"
roles:
  custom:
    display_name: "Custom"
    description: "Custom role"
    plugins:
      - name: "custom-plugin"
        source: "custom-plugin@local"
        required: true
defaults:
  scope: project
`
	os.WriteFile(filepath.Join(projectDir, "marketplace.yaml"), []byte(mktYAML), 0o644)

	cfgs := registry.LoadAllMarketplaces("/nonexistent", projectDir)

	found := false
	for _, c := range cfgs {
		if c.Marketplace.Name == "project-mkt" {
			found = true
		}
	}
	if !found {
		t.Error("expected project marketplace to be loaded")
	}
}

func TestLoadAllMarketplaces_WithRemoteMarketplace(t *testing.T) {
	yamlContent := `
schema_version: 1
marketplace:
  name: "remote-mkt"
  description: "Remote"
  version: "1.0.0"
  registry_url: "remote"
roles:
  dev:
    display_name: "Dev"
    description: "Dev"
    plugins:
      - name: "remote-plugin"
        source: "remote-plugin@remote"
        required: true
defaults:
  scope: project
`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(yamlContent))
	}))
	defer server.Close()

	cfgDir := t.TempDir()
	mktsCfgPath := filepath.Join(cfgDir, "marketplaces.yaml")
	mktsCfg := &config.MarketplacesConfig{
		Marketplaces: []config.MarketplaceEntry{
			{Name: "remote-mkt", URL: server.URL, Enabled: true},
		},
	}
	config.WriteMarketplacesConfig(mktsCfgPath, mktsCfg)

	cfgs := registry.LoadAllMarketplaces(mktsCfgPath, "/nonexistent/project")

	found := false
	for _, c := range cfgs {
		if c.Marketplace.Name == "remote-mkt" {
			found = true
		}
	}
	if !found {
		t.Error("expected remote marketplace to be loaded")
	}
}

func TestLoadAllMarketplaces_DeduplicatesByName(t *testing.T) {
	cfgDir := t.TempDir()
	mktsCfgPath := filepath.Join(cfgDir, "marketplaces.yaml")
	mktsCfg := &config.MarketplacesConfig{
		Marketplaces: []config.MarketplaceEntry{
			{Name: "asds-marketplace", URL: "github.com/anthropics/claude-plugins-official", Enabled: true},
		},
	}
	config.WriteMarketplacesConfig(mktsCfgPath, mktsCfg)

	cfgs := registry.LoadAllMarketplaces(mktsCfgPath, "/nonexistent/project")

	count := 0
	for _, c := range cfgs {
		if c.Marketplace.Name == "asds-marketplace" {
			count++
		}
	}
	if count > 1 {
		t.Errorf("expected 1 asds-marketplace, got %d (duplicates)", count)
	}
}
