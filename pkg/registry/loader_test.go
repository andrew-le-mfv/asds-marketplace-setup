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
	pluginDir := filepath.Join(projectDir, ".claude-plugin")
	os.MkdirAll(pluginDir, 0o755)
	os.WriteFile(filepath.Join(pluginDir, "marketplace.yaml"), []byte(mktYAML), 0o644)

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

func TestLoadAllMarketplaces_AutoDiscoverLocalPlugins(t *testing.T) {
	projectDir := t.TempDir()
	// plugins/ with valid plugins (have .claude-plugin/)
	os.MkdirAll(filepath.Join(projectDir, "plugins", "alpha-plugin", ".claude-plugin"), 0o755)
	os.MkdirAll(filepath.Join(projectDir, "plugins", "beta-plugin", ".claude-plugin"), 0o755)
	// A directory without .claude-plugin/ should be ignored
	os.MkdirAll(filepath.Join(projectDir, "plugins", "not-a-plugin"), 0o755)
	// A file should be ignored
	os.WriteFile(filepath.Join(projectDir, "plugins", "README.md"), []byte("# plugins"), 0o644)

	cfgs := registry.LoadAllMarketplaces("/nonexistent", projectDir)

	dirName := filepath.Base(projectDir)
	found := false
	for _, c := range cfgs {
		if c.Marketplace.Name == dirName {
			found = true
			role, ok := c.Roles["default"]
			if !ok {
				t.Fatal("expected 'default' role in discovered config")
			}
			if len(role.Plugins) != 2 {
				t.Errorf("expected 2 plugins, got %d", len(role.Plugins))
			}
		}
	}
	if !found {
		t.Errorf("expected auto-discovered marketplace %q", dirName)
	}

	// Discovery should NOT persist a file (avoids stale cache issues).
	savedPath := filepath.Join(projectDir, ".claude-plugin", "marketplace.yaml")
	if _, err := os.Stat(savedPath); err == nil {
		t.Error("expected .claude-plugin/marketplace.yaml to NOT be created (ephemeral discovery)")
	}
}

func TestDiscoverLocalPlugins_ExternalPlugins(t *testing.T) {
	projectDir := t.TempDir()
	// external_plugins/ with a valid plugin
	os.MkdirAll(filepath.Join(projectDir, "external_plugins", "ext-plugin", ".claude-plugin"), 0o755)

	cfg, err := registry.DiscoverLocalPlugins(projectDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	role := cfg.Roles["default"]
	if len(role.Plugins) != 1 || role.Plugins[0].Name != "ext-plugin" {
		t.Errorf("expected ext-plugin, got %v", role.Plugins)
	}
}

func TestDiscoverLocalPlugins_NestedPlugins(t *testing.T) {
	projectDir := t.TempDir()
	// Plugin nested 3 levels: plugins/asds/asds-core/.claude-plugin/
	os.MkdirAll(filepath.Join(projectDir, "plugins", "asds", "asds-core", ".claude-plugin"), 0o755)
	os.MkdirAll(filepath.Join(projectDir, "plugins", "asds", "asds-deploy", ".claude-plugin"), 0o755)
	// Direct plugin at level 1
	os.MkdirAll(filepath.Join(projectDir, "plugins", "sample-plugin", ".claude-plugin"), 0o755)

	cfg, err := registry.DiscoverLocalPlugins(projectDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	role := cfg.Roles["default"]
	if len(role.Plugins) != 3 {
		t.Errorf("expected 3 plugins, got %d: %v", len(role.Plugins), role.Plugins)
	}
	names := map[string]bool{}
	for _, p := range role.Plugins {
		names[p.Name] = true
	}
	if !names["sample-plugin"] || !names["asds-core"] || !names["asds-deploy"] {
		t.Errorf("expected sample-plugin, asds-core, asds-deploy, got %v", names)
	}
}

func TestDiscoverLocalPlugins_BothDirs(t *testing.T) {
	projectDir := t.TempDir()
	os.MkdirAll(filepath.Join(projectDir, "plugins", "p1", ".claude-plugin"), 0o755)
	os.MkdirAll(filepath.Join(projectDir, "external_plugins", "p2", ".claude-plugin"), 0o755)

	cfg, err := registry.DiscoverLocalPlugins(projectDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	role := cfg.Roles["default"]
	if len(role.Plugins) != 2 {
		t.Errorf("expected 2 plugins, got %d", len(role.Plugins))
	}
}

func TestDiscoverLocalPlugins_Deduplicates(t *testing.T) {
	projectDir := t.TempDir()
	// Same plugin name in both dirs — should appear only once
	os.MkdirAll(filepath.Join(projectDir, "plugins", "shared", ".claude-plugin"), 0o755)
	os.MkdirAll(filepath.Join(projectDir, "external_plugins", "shared", ".claude-plugin"), 0o755)

	cfg, err := registry.DiscoverLocalPlugins(projectDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	role := cfg.Roles["default"]
	if len(role.Plugins) != 1 {
		t.Errorf("expected 1 plugin (deduplicated), got %d", len(role.Plugins))
	}
}

func TestDiscoverLocalPlugins_NoPluginsDir(t *testing.T) {
	projectDir := t.TempDir()
	_, err := registry.DiscoverLocalPlugins(projectDir)
	if err == nil {
		t.Error("expected error when no scan directories exist")
	}
}

func TestDiscoverLocalPlugins_NoCloudePluginMarker(t *testing.T) {
	projectDir := t.TempDir()
	// Directories without .claude-plugin/ should not be discovered
	os.MkdirAll(filepath.Join(projectDir, "plugins", "plain-dir"), 0o755)
	_, err := registry.DiscoverLocalPlugins(projectDir)
	if err == nil {
		t.Error("expected error when no dirs have .claude-plugin/")
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
