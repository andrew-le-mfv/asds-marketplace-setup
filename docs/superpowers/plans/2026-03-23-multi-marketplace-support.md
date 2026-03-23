# Multi-Marketplace Support Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the single-marketplace model with a multi-marketplace system that supports layered fallback config loading, per-marketplace plugin discovery, persistent marketplace storage, and full CLI + TUI management.

**Architecture:** Introduce a `MarketplacesConfig` type stored at `~/.config/asds/marketplaces.yaml` that holds a list of marketplace entries (name, URL, enabled flag). Config loading follows a 4-layer fallback chain: embedded default → user config file → remote URL → local repo `marketplace.yaml`. The Plugins tab aggregates plugins across all active marketplaces. The Config tab becomes a marketplace manager with add/edit/remove/list.

**Tech Stack:** Go 1.26, Bubble Tea, Cobra, lipgloss, charmbracelet/bubbles (textinput), gopkg.in/yaml.v3

---

## File Structure

### New files
- `internal/config/marketplaces.go` — `MarketplacesConfig` type, `MarketplaceEntry` type, read/write/add/remove/list operations for `~/.config/asds/marketplaces.yaml`
- `internal/config/marketplaces_test.go` — tests for marketplaces CRUD
- `internal/config/loader.go` — `LoadAllMarketplaces()` orchestrator implementing the 4-layer fallback chain, returns `[]*MarketplaceConfig`
- `internal/config/loader_test.go` — tests for fallback chain logic

### Modified files
- `internal/config/asdsconfig.go` — remove `MarketplaceURL` field (moved to marketplaces.yaml)
- `internal/config/asdsconfig_test.go` — update tests to match new ASDSConfig shape
- `internal/config/defaults_test.go` — fix existing test failure (12 roles not 9)
- `pkg/registry/fetch.go` — add `DiscoverMarketplace(url)` that checks for `marketplace.yaml` at a URL and returns whether it's valid; update `BuildRawURL` to try `marketplace.yaml` filename
- `pkg/registry/fetch_test.go` — tests for DiscoverMarketplace
- `internal/commands/root.go` — use `LoadAllMarketplaces()`, pass `[]*MarketplaceConfig` to TUI
- `internal/commands/install.go` — use multi-marketplace loader
- `internal/commands/update.go` — use multi-marketplace loader
- `internal/commands/root.go` — add `marketplace` parent command with subcommands
- `internal/commands/marketplace.go` — new file for `asds marketplace list|add|remove|update` CLI subcommands
- `internal/tui/app.go` — accept `[]*MarketplaceConfig`, pass to tabs
- `internal/tui/plugins/model.go` — accept `[]*MarketplaceConfig`, aggregate plugins from all marketplaces
- `internal/tui/plugins/view.go` — show marketplace source per plugin
- `internal/tui/config/model.go` — full rewrite to marketplace manager
- `internal/tui/config/update.go` — handle add/edit/remove flows, text input
- `internal/tui/config/view.go` — render marketplace list, add form, edit form

---

## Task 1: Fix existing test failure & define MarketplaceEntry type

**Files:**
- Modify: `internal/config/defaults_test.go`
- Create: `internal/config/marketplaces.go`
- Create: `internal/config/marketplaces_test.go`

- [ ] **Step 1: Fix defaults_test.go to match current 12 roles**

Update the `expectedRoles` slice in `TestDefaultMarketplaceConfig` to include all 12 roles from `default-marketplace.yaml`:
```go
expectedRoles := []string{
    "backend", "data-engineer", "developer", "devops",
    "frontend", "go-developer", "plugin-developer", "pm",
    "rust-developer", "security", "techlead", "tester",
}
```

- [ ] **Step 2: Run test to verify fix**

Run: `go test ./internal/config/ -run TestDefaultMarketplaceConfig -v`
Expected: PASS

- [ ] **Step 3: Create `internal/config/marketplaces.go`**

Define the `MarketplacesConfig` and `MarketplaceEntry` types, plus read/write/CRUD functions:

```go
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
```

- [ ] **Step 4: Create `internal/config/marketplaces_test.go`**

Write tests for all CRUD operations:

```go
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
```

- [ ] **Step 5: Run all config tests**

Run: `go test ./internal/config/ -v`
Expected: ALL PASS

- [ ] **Step 6: Commit**

```bash
git add internal/config/defaults_test.go internal/config/marketplaces.go internal/config/marketplaces_test.go
git commit -m "feat: add MarketplacesConfig type with CRUD operations"
```

---

## Task 2: Add marketplace discovery to registry package

**Files:**
- Modify: `pkg/registry/fetch.go`
- Modify: `pkg/registry/fetch_test.go`

- [ ] **Step 1: Write test for DiscoverMarketplace**

Add to `pkg/registry/fetch_test.go`:

```go
func TestDiscoverMarketplace(t *testing.T) {
	yamlContent := `
schema_version: 1
marketplace:
  name: "discovered"
  description: "Test"
  version: "1.0.0"
  registry_url: "github.com/test"
roles:
  dev:
    display_name: "Dev"
    description: "Dev"
    plugins:
      - name: "p"
        source: "p@test"
        required: true
defaults:
  scope: project
`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/yaml")
		w.Write([]byte(yamlContent))
	}))
	defer server.Close()

	cfg, err := registry.DiscoverMarketplace(server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Marketplace.Name != "discovered" {
		t.Errorf("name = %q, want %q", cfg.Marketplace.Name, "discovered")
	}
}

func TestDiscoverMarketplace_InvalidURL(t *testing.T) {
	_, err := registry.DiscoverMarketplace("http://localhost:1")
	if err == nil {
		t.Error("expected error for unreachable URL")
	}
}

func TestDiscoverMarketplace_InvalidResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not yaml at all {{{"))
	}))
	defer server.Close()

	_, err := registry.DiscoverMarketplace(server.URL)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./pkg/registry/ -run TestDiscoverMarketplace -v`
Expected: FAIL (function doesn't exist yet)

- [ ] **Step 3: Implement DiscoverMarketplace in `pkg/registry/fetch.go`**

Add the `DiscoverMarketplace` function which tries to fetch and validate a marketplace config from a URL. It tries the direct URL first, then appends `marketplace.yaml` if it looks like a repo path:

```go
// DiscoverMarketplace attempts to fetch a marketplace config from a URL.
// It tries the URL directly, then tries appending /marketplace.yaml for repo-style URLs.
// Returns the parsed config and validates it.
func DiscoverMarketplace(url string) (*config.MarketplaceConfig, error) {
	rawURL := BuildRawURL(url)
	cfg, err := FetchMarketplaceConfig(rawURL)
	if err != nil {
		return nil, fmt.Errorf("discovering marketplace at %q: %w", url, err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid marketplace at %q: %w", url, err)
	}

	return cfg, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./pkg/registry/ -v`
Expected: ALL PASS

- [ ] **Step 5: Commit**

```bash
git add pkg/registry/fetch.go pkg/registry/fetch_test.go
git commit -m "feat: add DiscoverMarketplace for URL validation"
```

---

## Task 3: Create config loader with 4-layer fallback chain

**Files:**
- Create: `internal/config/loader.go`
- Create: `internal/config/loader_test.go`

- [ ] **Step 1: Write tests for LoadAllMarketplaces**

Create `internal/config/loader_test.go`:

```go
package config_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/andrew-le-mfv/asds-marketplace-setup/internal/config"
)

func TestLoadAllMarketplaces_DefaultsOnly(t *testing.T) {
	// No user config, no project marketplace.yaml
	cfgs := config.LoadAllMarketplaces("/nonexistent", "/nonexistent/project")
	if len(cfgs) == 0 {
		t.Fatal("expected at least 1 marketplace from defaults")
	}
	if cfgs[0].Marketplace.Name != "asds-marketplace" {
		t.Errorf("expected default marketplace, got %q", cfgs[0].Marketplace.Name)
	}
}

func TestLoadAllMarketplaces_WithProjectMarketplace(t *testing.T) {
	// Create a project marketplace.yaml
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

	cfgs := config.LoadAllMarketplaces("/nonexistent", projectDir)

	// Should have default + project marketplace
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

	// Create marketplaces.yaml pointing to the test server
	cfgDir := t.TempDir()
	mktsCfgPath := filepath.Join(cfgDir, "marketplaces.yaml")
	mktsCfg := &config.MarketplacesConfig{
		Marketplaces: []config.MarketplaceEntry{
			{Name: "remote-mkt", URL: server.URL, Enabled: true},
		},
	}
	config.WriteMarketplacesConfig(mktsCfgPath, mktsCfg)

	cfgs := config.LoadAllMarketplaces(mktsCfgPath, "/nonexistent/project")

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
	// If default and user config both have "asds-marketplace", only keep one (user wins)
	cfgDir := t.TempDir()
	mktsCfgPath := filepath.Join(cfgDir, "marketplaces.yaml")
	mktsCfg := &config.MarketplacesConfig{
		Marketplaces: []config.MarketplaceEntry{
			{Name: "asds-marketplace", URL: "github.com/anthropics/claude-plugins-official", Enabled: true},
		},
	}
	config.WriteMarketplacesConfig(mktsCfgPath, mktsCfg)

	cfgs := config.LoadAllMarketplaces(mktsCfgPath, "/nonexistent/project")

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
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/config/ -run TestLoadAllMarketplaces -v`
Expected: FAIL

- [ ] **Step 3: Implement `internal/config/loader.go`**

```go
package config

import (
	"os"
	"path/filepath"

	"github.com/andrew-le-mfv/asds-marketplace-setup/pkg/registry"
)

// LoadAllMarketplaces loads marketplace configs using the 4-layer fallback chain:
//   1. Embedded default (always loaded as baseline)
//   2. User marketplaces from marketplacesConfigPath (~/.config/asds/marketplaces.yaml)
//   3. Remote fetch for each enabled marketplace URL
//   4. Project-local marketplace.yaml at projectRoot/marketplace.yaml
//
// Later layers override earlier ones by marketplace name.
// Returns a deduplicated slice of parsed MarketplaceConfig.
func LoadAllMarketplaces(marketplacesConfigPath string, projectRoot string) []*MarketplaceConfig {
	seen := make(map[string]int) // name -> index in result
	var result []*MarketplaceConfig

	addOrReplace := func(cfg *MarketplaceConfig) {
		if idx, ok := seen[cfg.Marketplace.Name]; ok {
			result[idx] = cfg
		} else {
			seen[cfg.Marketplace.Name] = len(result)
			result = append(result, cfg)
		}
	}

	// Layer 1: Embedded default
	if defaultCfg, err := DefaultMarketplaceConfig(); err == nil {
		addOrReplace(defaultCfg)
	}

	// Layer 2+3: User-configured marketplaces (read config, then fetch each)
	mktsCfg, err := ReadMarketplacesConfig(marketplacesConfigPath)
	if err == nil {
		for _, entry := range mktsCfg.EnabledMarketplaces() {
			rawURL := registry.BuildRawURL(entry.URL)
			if fetched, fetchErr := registry.FetchMarketplaceConfig(rawURL); fetchErr == nil {
				addOrReplace(fetched)
			}
			// If fetch fails, the marketplace is skipped (default still available)
		}
	}

	// Layer 4: Project-local marketplace.yaml
	if projectRoot != "" {
		localPath := filepath.Join(projectRoot, "marketplace.yaml")
		if data, readErr := os.ReadFile(localPath); readErr == nil {
			if localCfg, parseErr := ParseMarketplaceConfig(data); parseErr == nil {
				if validateErr := localCfg.Validate(); validateErr == nil {
					addOrReplace(localCfg)
				}
			}
		}
	}

	return result
}
```

**Note:** This creates a circular import (`config` → `registry` → `config`). We need to break this by extracting `BuildRawURL` and `FetchMarketplaceConfig` usage. The simplest approach: pass a fetcher function instead of importing registry directly. Alternatively, move `BuildRawURL` to a shared location. Let's use a function parameter approach:

Actually, looking at the existing code, `registry` already imports `config`. So `config` cannot import `registry`. We'll put `LoadAllMarketplaces` in a new package or in `registry` instead. The cleanest approach: put the loader in `pkg/registry/` since it already imports `config`.

**Revised:** Create `pkg/registry/loader.go` instead of `internal/config/loader.go`.

```go
// pkg/registry/loader.go
package registry

import (
	"os"
	"path/filepath"

	"github.com/andrew-le-mfv/asds-marketplace-setup/internal/config"
)

// LoadAllMarketplaces loads marketplace configs using the 4-layer fallback chain:
//   1. Embedded default (always loaded as baseline)
//   2. User marketplaces from marketplacesConfigPath (~/.config/asds/marketplaces.yaml)
//   3. Remote fetch for each enabled marketplace URL
//   4. Project-local marketplace.yaml at projectRoot/marketplace.yaml
//
// Later layers override earlier ones by marketplace name.
func LoadAllMarketplaces(marketplacesConfigPath string, projectRoot string) []*config.MarketplaceConfig {
	seen := make(map[string]int)
	var result []*config.MarketplaceConfig

	addOrReplace := func(cfg *config.MarketplaceConfig) {
		if idx, ok := seen[cfg.Marketplace.Name]; ok {
			result[idx] = cfg
		} else {
			seen[cfg.Marketplace.Name] = len(result)
			result = append(result, cfg)
		}
	}

	// Layer 1: Embedded default
	if defaultCfg, err := config.DefaultMarketplaceConfig(); err == nil {
		addOrReplace(defaultCfg)
	}

	// Layer 2+3: User-configured marketplaces (read config, then fetch each)
	mktsCfg, err := config.ReadMarketplacesConfig(marketplacesConfigPath)
	if err == nil {
		for _, entry := range mktsCfg.EnabledMarketplaces() {
			rawURL := BuildRawURL(entry.URL)
			if fetched, fetchErr := FetchMarketplaceConfig(rawURL); fetchErr == nil {
				addOrReplace(fetched)
			}
		}
	}

	// Layer 4: Project-local marketplace.yaml
	if projectRoot != "" {
		localPath := filepath.Join(projectRoot, "marketplace.yaml")
		if data, readErr := os.ReadFile(localPath); readErr == nil {
			if localCfg, parseErr := config.ParseMarketplaceConfig(data); parseErr == nil {
				if validateErr := localCfg.Validate(); validateErr == nil {
					addOrReplace(localCfg)
				}
			}
		}
	}

	return result
}
```

Move the tests to `pkg/registry/loader_test.go` accordingly.

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./pkg/registry/ -v`
Expected: ALL PASS

- [ ] **Step 5: Commit**

```bash
git add pkg/registry/loader.go pkg/registry/loader_test.go
git commit -m "feat: add LoadAllMarketplaces with 4-layer fallback chain"
```

---

## Task 4: Update ASDSConfig to remove MarketplaceURL

**Files:**
- Modify: `internal/config/asdsconfig.go`
- Modify: `internal/config/asdsconfig_test.go`
- Modify: `internal/commands/install.go`
- Modify: `internal/commands/update.go`

- [ ] **Step 1: Update `internal/config/asdsconfig.go`**

Remove `MarketplaceURL` field and the `defaultMarketplaceURL` constant. The `ASDSConfig` struct is now a shell for future non-marketplace settings. Keep the read/write functions for forward compatibility:

```go
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ASDSConfig is the TUI's own configuration stored at ~/.config/asds/config.yaml.
type ASDSConfig struct {
	// Future non-marketplace settings go here.
}

// DefaultASDSConfig returns the default ASDS configuration.
func DefaultASDSConfig() ASDSConfig {
	return ASDSConfig{}
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
```

- [ ] **Step 2: Update `internal/config/asdsconfig_test.go`**

Remove tests referencing `MarketplaceURL`:

```go
package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/andrew-le-mfv/asds-marketplace-setup/internal/config"
)

func TestASDSConfig_Defaults(t *testing.T) {
	_ = config.DefaultASDSConfig()
	// Just verify it doesn't panic; struct is currently empty
}

func TestASDSConfig_WriteAndRead(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	cfg := config.ASDSConfig{}

	if err := config.WriteASDSConfig(path, &cfg); err != nil {
		t.Fatalf("WriteASDSConfig error: %v", err)
	}

	_, err := config.ReadASDSConfig(path)
	if err != nil {
		t.Fatalf("ReadASDSConfig error: %v", err)
	}
}

func TestReadASDSConfig_NotFound_ReturnsDefaults(t *testing.T) {
	_, err := config.ReadASDSConfig("/nonexistent/config.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
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
```

- [ ] **Step 3: Update `install.go` and `update.go` commands**

Replace `config.ReadASDSConfig` + `registry.FetchOrDefault` pattern with `registry.LoadAllMarketplaces`. For install, the `--role` flag needs to search across all loaded marketplaces.

In `install.go`, replace lines 53-61 with:
```go
// Load all marketplaces
mktsCfgPath := config.ResolveMarketplacesConfigPath()
allCfgs := registry.LoadAllMarketplaces(mktsCfgPath, projectRoot)

// Find role across all marketplaces
var mktCfg *config.MarketplaceConfig
var roleConfig config.Role
var found bool
for _, cfg := range allCfgs {
    if r, ok := cfg.Roles[role]; ok {
        mktCfg = cfg
        roleConfig = r
        found = true
        break
    }
}
if !found {
    return fmt.Errorf("unknown role %q across all marketplaces", role)
}
```

Remove the old `roleConfig, ok := mktCfg.Roles[role]` block and use the variables from above.

In `update.go`, replace lines 48-49 with:
```go
mktsCfgPath := config.ResolveMarketplacesConfigPath()
allCfgs := registry.LoadAllMarketplaces(mktsCfgPath, projectRoot)

// Find the marketplace that originally installed this role
var mktCfg *config.MarketplaceConfig
for _, cfg := range allCfgs {
    if _, ok := cfg.Roles[manifest.Role]; ok {
        mktCfg = cfg
        break
    }
}
if mktCfg == nil {
    return fmt.Errorf("role %q no longer exists in any marketplace", manifest.Role)
}
```

- [ ] **Step 4: Run all tests**

Run: `go test ./... -v`
Expected: ALL PASS

- [ ] **Step 5: Commit**

```bash
git add internal/config/asdsconfig.go internal/config/asdsconfig_test.go internal/commands/install.go internal/commands/update.go
git commit -m "refactor: remove MarketplaceURL from ASDSConfig, use multi-marketplace loader"
```

---

## Task 5: Update root command and TUI App to accept multiple marketplaces

**Files:**
- Modify: `internal/commands/root.go`
- Modify: `internal/tui/app.go`

- [ ] **Step 1: Update `internal/commands/root.go`**

Replace the `RunE` function to use `LoadAllMarketplaces` and pass `[]*config.MarketplaceConfig` to the TUI:

```go
RunE: func(cmd *cobra.Command, args []string) error {
    projectRoot, _ := claude.FindProjectRoot(".")

    mktsCfgPath := config.ResolveMarketplacesConfigPath()
    allCfgs := registry.LoadAllMarketplaces(mktsCfgPath, projectRoot)

    app := tui.NewApp(version, allCfgs, projectRoot)
    p := tea.NewProgram(app, tea.WithAltScreen())
    _, err := p.Run()
    return err
},
```

- [ ] **Step 2: Update `internal/tui/app.go`**

Change `NewApp` signature to accept `[]*config.MarketplaceConfig`:

```go
func NewApp(version string, cfgs []*config.MarketplaceConfig, projectRoot string) App {
    return App{
        activeTab:    TabSetup,
        tabs:         AllTabs(),
        keys:         DefaultKeyMap(),
        setupModel:   setup.New(cfgs, projectRoot),
        pluginsModel: plugins.New(cfgs, projectRoot),
        configModel:  tuiconfig.New(),
        statusModel:  status.New(projectRoot),
        aboutModel:   about.New(version),
    }
}
```

- [ ] **Step 3: Verify build compiles**

Run: `go build ./cmd/asds/`
Expected: SUCCESS (will need downstream tab updates in subsequent tasks)

- [ ] **Step 4: Commit**

```bash
git add internal/commands/root.go internal/tui/app.go
git commit -m "refactor: pass multi-marketplace configs through root command to TUI"
```

---

## Task 6: Update Setup tab for multi-marketplace

**Files:**
- Modify: `internal/tui/setup/model.go`
- Modify: `internal/tui/setup/update.go`
- Modify: `internal/tui/setup/view.go`

- [ ] **Step 1: Update `setup/model.go`**

Change `New` to accept `[]*config.MarketplaceConfig`. Aggregate roles from all marketplaces, adding a `MarketplaceName` field to `roleItem`:

```go
type roleItem struct {
	ID              string
	DisplayName     string
	Description     string
	PluginCount     int
	MarketplaceName string
}

func New(cfgs []*config.MarketplaceConfig, projectRoot string) Model {
	var roles []roleItem
	seen := make(map[string]bool) // roleID+marketplace -> dedup

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

	// ... rest same but store all cfgs
}
```

Store `marketplaceCfgs []*config.MarketplaceConfig` in Model instead of single `marketplaceCfg`. Update `doInstall` and `doUninstall` to find the correct marketplace config by the selected role's `MarketplaceName`.

- [ ] **Step 2: Update `setup/view.go`**

Show marketplace name next to each role in the selection list:
```go
line := fmt.Sprintf("%s%s — %s (%d plugins) [%s]%s", cursor, r.DisplayName, r.Description, r.PluginCount, r.MarketplaceName, badge)
```

- [ ] **Step 3: Update `setup/update.go`**

Update `doInstall` to look up the role from the correct marketplace config using the selected role's `MarketplaceName`.

- [ ] **Step 4: Verify build**

Run: `go build ./cmd/asds/`
Expected: SUCCESS

- [ ] **Step 5: Commit**

```bash
git add internal/tui/setup/
git commit -m "feat: setup tab supports roles from multiple marketplaces"
```

---

## Task 7: Update Plugins tab to aggregate from all marketplaces

**Files:**
- Modify: `internal/tui/plugins/model.go`
- Modify: `internal/tui/plugins/view.go`
- Modify: `internal/tui/plugins/update.go`

- [ ] **Step 1: Update `plugins/model.go`**

Change `New` to accept `[]*config.MarketplaceConfig`. Add `MarketplaceName` to `PluginItem`. Aggregate all unique plugins across all marketplaces:

```go
type PluginItem struct {
	Name            string
	Source          string
	Required        bool
	RoleName        string
	MarketplaceName string
}

func New(cfgs []*config.MarketplaceConfig, projectRoot string) Model {
	seen := make(map[string]bool) // source -> dedup
	var items []PluginItem

	for _, cfg := range cfgs {
		for _, roleName := range cfg.RoleNames() {
			role := cfg.Roles[roleName]
			for _, p := range role.Plugins {
				if seen[p.Source] {
					continue
				}
				seen[p.Source] = true
				items = append(items, PluginItem{
					Name:            p.Name,
					Source:          p.Source,
					Required:        p.Required,
					RoleName:        roleName,
					MarketplaceName: cfg.Marketplace.Name,
				})
			}
		}
	}

	// ... store all cfgs as marketplaceCfgs []*config.MarketplaceConfig
}
```

- [ ] **Step 2: Update `plugins/view.go`**

Show marketplace name in browse and detail views:
```go
lines = append(lines, styles.SubtitleStyle.Render(
    fmt.Sprintf("    Source: %s | Role: %s | Marketplace: %s", item.Source, item.RoleName, item.MarketplaceName)))
```

- [ ] **Step 3: Update `plugins/update.go`**

Update `doInstall` to find the correct marketplace config by `MarketplaceName` from the selected plugin:

```go
func (m Model) doInstall() tea.Cmd {
    return func() tea.Msg {
        plugin := m.SelectedPlugin()
        scope := m.SelectedScope()

        // Find the marketplace config for this plugin
        var mktCfg *config.MarketplaceConfig
        for _, cfg := range m.marketplaceCfgs {
            if cfg.Marketplace.Name == plugin.MarketplaceName {
                mktCfg = cfg
                break
            }
        }
        if mktCfg == nil {
            return InstallCompleteMsg{Error: fmt.Errorf("marketplace %q not found", plugin.MarketplaceName)}
        }

        inst := installer.NewInstaller(true)
        inst.RegisterMarketplace(mktCfg.Marketplace.Name, mktCfg.Marketplace.RegistryURL)
        // ... rest same
    }
}
```

- [ ] **Step 4: Verify build**

Run: `go build ./cmd/asds/`
Expected: SUCCESS

- [ ] **Step 5: Commit**

```bash
git add internal/tui/plugins/
git commit -m "feat: plugins tab aggregates from all marketplaces"
```

---

## Task 8: Rewrite Config tab as marketplace manager (TUI)

**Files:**
- Modify: `internal/tui/config/model.go`
- Modify: `internal/tui/config/update.go`
- Modify: `internal/tui/config/view.go`

- [ ] **Step 1: Rewrite `config/model.go`**

Replace the read-only config viewer with an interactive marketplace manager:

```go
package config

import (
	"github.com/charmbracelet/bubbles/textinput"

	appconfig "github.com/andrew-le-mfv/asds-marketplace-setup/internal/config"
)

type Step int

const (
	StepList Step = iota
	StepAdd
	StepEdit
	StepRemoveConfirm
	StepDiscovering
	StepDiscoverResult
	StepError
)

type addField int

const (
	fieldName addField = iota
	fieldURL
)

type Model struct {
	step       Step
	mktsCfg    *appconfig.MarketplacesConfig
	cfgPath    string
	cursor     int
	width      int
	height     int
	nameInput  textinput.Model
	urlInput   textinput.Model
	activeField addField
	errorMsg   string
	discoverOK bool
}

func New() Model {
	cfgPath := appconfig.ResolveMarketplacesConfigPath()
	cfg, _ := appconfig.ReadMarketplacesConfig(cfgPath)

	ni := textinput.New()
	ni.Placeholder = "marketplace-name"
	ni.CharLimit = 64

	ui := textinput.New()
	ui.Placeholder = "github.com/org/repo"
	ui.CharLimit = 256

	return Model{
		step:    StepList,
		mktsCfg: cfg,
		cfgPath: cfgPath,
		nameInput: ni,
		urlInput:  ui,
	}
}

func (m *Model) save() error {
	return appconfig.WriteMarketplacesConfig(m.cfgPath, m.mktsCfg)
}

func (m *Model) reload() {
	cfg, _ := appconfig.ReadMarketplacesConfig(m.cfgPath)
	m.mktsCfg = cfg
}
```

- [ ] **Step 2: Rewrite `config/update.go`**

Handle key events for list navigation, add/edit/remove flows:

```go
package config

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	appconfig "github.com/andrew-le-mfv/asds-marketplace-setup/internal/config"
	"github.com/andrew-le-mfv/asds-marketplace-setup/pkg/registry"
)

type DiscoverCompleteMsg struct {
	Config *appconfig.MarketplaceConfig
	Error  error
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case DiscoverCompleteMsg:
		if msg.Error != nil {
			m.step = StepError
			m.errorMsg = msg.Error.Error()
		} else {
			m.discoverOK = true
			m.step = StepDiscoverResult
		}
		return m, nil

	case tea.KeyMsg:
		switch m.step {
		case StepList:
			return m.updateList(msg)
		case StepAdd, StepEdit:
			return m.updateForm(msg)
		case StepRemoveConfirm:
			return m.updateRemoveConfirm(msg)
		case StepDiscovering:
			// Waiting, ignore keys
		case StepDiscoverResult:
			if msg.String() == "enter" || msg.String() == "esc" {
				m.reload()
				m.step = StepList
			}
		case StepError:
			if msg.String() == "enter" || msg.String() == "esc" {
				m.step = StepList
			}
		}
	}

	return m, nil
}

func (m Model) updateList(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.mktsCfg.Marketplaces)-1 {
			m.cursor++
		}
	case "a":
		m.nameInput.Reset()
		m.urlInput.Reset()
		m.nameInput.Focus()
		m.activeField = fieldName
		m.step = StepAdd
	case "e":
		if len(m.mktsCfg.Marketplaces) > 0 {
			entry := m.mktsCfg.Marketplaces[m.cursor]
			m.nameInput.SetValue(entry.Name)
			m.urlInput.SetValue(entry.URL)
			m.nameInput.Focus()
			m.activeField = fieldName
			m.step = StepEdit
		}
	case "d", "delete", "backspace":
		if len(m.mktsCfg.Marketplaces) > 0 {
			m.step = StepRemoveConfirm
		}
	case " ":
		// Toggle enabled
		if len(m.mktsCfg.Marketplaces) > 0 {
			m.mktsCfg.Marketplaces[m.cursor].Enabled = !m.mktsCfg.Marketplaces[m.cursor].Enabled
			m.save()
		}
	}
	return m, nil
}

func (m Model) updateForm(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "tab":
		if m.activeField == fieldName {
			m.nameInput.Blur()
			m.urlInput.Focus()
			m.activeField = fieldURL
		} else {
			m.urlInput.Blur()
			m.nameInput.Focus()
			m.activeField = fieldName
		}
		return m, nil
	case "enter":
		name := m.nameInput.Value()
		url := m.urlInput.Value()
		if name == "" || url == "" {
			m.errorMsg = "name and URL are required"
			m.step = StepError
			return m, nil
		}

		entry := appconfig.MarketplaceEntry{Name: name, URL: url, Enabled: true}

		if m.step == StepAdd {
			if err := m.mktsCfg.AddMarketplace(entry); err != nil {
				m.errorMsg = err.Error()
				m.step = StepError
				return m, nil
			}
		} else {
			oldName := m.mktsCfg.Marketplaces[m.cursor].Name
			m.mktsCfg.UpdateMarketplace(oldName, entry)
		}

		m.save()
		m.step = StepDiscovering
		return m, m.doDiscover(url)
	case "esc":
		m.step = StepList
		return m, nil
	}

	// Forward to active text input
	var cmd tea.Cmd
	if m.activeField == fieldName {
		m.nameInput, cmd = m.nameInput.Update(msg)
	} else {
		m.urlInput, cmd = m.urlInput.Update(msg)
	}
	return m, cmd
}

func (m Model) updateRemoveConfirm(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "y", "enter":
		name := m.mktsCfg.Marketplaces[m.cursor].Name
		m.mktsCfg.RemoveMarketplace(name)
		m.save()
		if m.cursor > 0 {
			m.cursor--
		}
		m.step = StepList
	case "n", "esc":
		m.step = StepList
	}
	return m, nil
}

func (m Model) doDiscover(url string) tea.Cmd {
	return func() tea.Msg {
		cfg, err := registry.DiscoverMarketplace(url)
		return DiscoverCompleteMsg{Config: cfg, Error: err}
	}
}
```

- [ ] **Step 3: Rewrite `config/view.go`**

```go
package config

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/andrew-le-mfv/asds-marketplace-setup/internal/tui/styles"
)

func (m Model) View() string {
	switch m.step {
	case StepList:
		return m.viewList()
	case StepAdd:
		return m.viewForm("Add Marketplace")
	case StepEdit:
		return m.viewForm("Edit Marketplace")
	case StepRemoveConfirm:
		return m.viewRemoveConfirm()
	case StepDiscovering:
		return m.viewDiscovering()
	case StepDiscoverResult:
		return m.viewDiscoverResult()
	case StepError:
		return m.viewError()
	default:
		return ""
	}
}

func (m Model) viewList() string {
	title := styles.TitleStyle.Render("⚙ Marketplaces")
	subtitle := styles.SubtitleStyle.Render("Manage your marketplace sources")

	var lines []string
	lines = append(lines, "", title, subtitle, "")

	if len(m.mktsCfg.Marketplaces) == 0 {
		lines = append(lines, styles.WarningStyle.Render("  No marketplaces configured"))
	} else {
		for i, mkt := range m.mktsCfg.Marketplaces {
			cursor := "  "
			style := styles.NormalStyle
			if i == m.cursor {
				cursor = "▸ "
				style = styles.SelectedStyle
			}

			enabledIcon := "✓"
			if !mkt.Enabled {
				enabledIcon = "✗"
			}

			line := fmt.Sprintf("%s[%s] %s", cursor, enabledIcon, mkt.Name)
			lines = append(lines, style.Render(line))
			if i == m.cursor {
				lines = append(lines, styles.SubtitleStyle.Render(fmt.Sprintf("    URL: %s", mkt.URL)))
			}
		}
	}

	lines = append(lines, "")
	lines = append(lines, styles.HelpStyle.Render("  Config: "+m.cfgPath))
	lines = append(lines, "")
	lines = append(lines, styles.HelpStyle.Render("↑↓ navigate  a add  e edit  d remove  space toggle"))

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m Model) viewForm(title string) string {
	t := styles.TitleStyle.Render(title)

	var lines []string
	lines = append(lines, "", t, "")

	nameLabel := "  Name: "
	urlLabel := "  URL:  "
	if m.activeField == fieldName {
		nameLabel = styles.SelectedStyle.Render(nameLabel)
	}
	if m.activeField == fieldURL {
		urlLabel = styles.SelectedStyle.Render(urlLabel)
	}

	lines = append(lines, nameLabel+m.nameInput.View())
	lines = append(lines, urlLabel+m.urlInput.View())
	lines = append(lines, "")
	lines = append(lines, styles.HelpStyle.Render("tab switch field  enter save  esc cancel"))

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m Model) viewRemoveConfirm() string {
	mkt := m.mktsCfg.Marketplaces[m.cursor]
	title := styles.WarningStyle.Render("⚠ Remove Marketplace")

	var lines []string
	lines = append(lines, "", title, "")
	lines = append(lines, styles.NormalStyle.Render(fmt.Sprintf("  Name: %s", mkt.Name)))
	lines = append(lines, styles.NormalStyle.Render(fmt.Sprintf("  URL:  %s", mkt.URL)))
	lines = append(lines, "")
	lines = append(lines, styles.HelpStyle.Render("y/enter confirm  n/esc cancel"))

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m Model) viewDiscovering() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		"",
		styles.TitleStyle.Render("Discovering marketplace..."),
		"",
		styles.SubtitleStyle.Render("Checking if URL is a valid marketplace"),
		"",
	)
}

func (m Model) viewDiscoverResult() string {
	var lines []string
	if m.discoverOK {
		lines = append(lines, "", styles.SuccessStyle.Render("✅ Marketplace validated and saved!"), "")
	} else {
		lines = append(lines, "", styles.WarningStyle.Render("⚠ Saved but could not validate marketplace"), "")
		lines = append(lines, styles.SubtitleStyle.Render("  The URL was saved. Plugins will load if URL becomes available."))
	}
	lines = append(lines, "", styles.HelpStyle.Render("enter to continue"))

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m Model) viewError() string {
	title := styles.ErrorStyle.Render("❌ Error")

	return lipgloss.JoinVertical(lipgloss.Left,
		"",
		title,
		"",
		styles.NormalStyle.Render("  "+m.errorMsg),
		"",
		styles.HelpStyle.Render("enter to go back"),
	)
}
```

- [ ] **Step 4: Verify build**

Run: `go build ./cmd/asds/`
Expected: SUCCESS

- [ ] **Step 5: Commit**

```bash
git add internal/tui/config/
git commit -m "feat: rewrite Config tab as interactive marketplace manager"
```

---

## Task 9: Add CLI marketplace subcommands

**Files:**
- Create: `internal/commands/marketplace.go`
- Modify: `internal/commands/root.go`

- [ ] **Step 1: Create `internal/commands/marketplace.go`**

```go
package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/andrew-le-mfv/asds-marketplace-setup/internal/config"
	"github.com/andrew-le-mfv/asds-marketplace-setup/pkg/registry"
)

func newMarketplaceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "marketplace",
		Short: "Manage marketplace sources",
		Long:  "List, add, edit, and remove marketplace sources.",
	}

	cmd.AddCommand(newMarketplaceListCmd())
	cmd.AddCommand(newMarketplaceAddCmd())
	cmd.AddCommand(newMarketplaceRemoveCmd())
	cmd.AddCommand(newMarketplaceUpdateCmd())

	return cmd
}

func newMarketplaceListCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List configured marketplaces",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgPath := config.ResolveMarketplacesConfigPath()
			cfg, err := config.ReadMarketplacesConfig(cfgPath)
			if err != nil {
				return fmt.Errorf("reading marketplaces config: %w", err)
			}

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(cfg.Marketplaces)
			}

			fmt.Println("📦 Configured Marketplaces")
			fmt.Println()
			for _, m := range cfg.Marketplaces {
				enabled := "✓"
				if !m.Enabled {
					enabled = "✗"
				}
				fmt.Printf("  [%s] %s — %s\n", enabled, m.Name, m.URL)
			}
			fmt.Printf("\nConfig: %s\n", cfgPath)
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	return cmd
}

func newMarketplaceAddCmd() *cobra.Command {
	var (
		name    string
		url     string
		noCheck bool
	)

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a marketplace source",
		RunE: func(cmd *cobra.Command, args []string) error {
			if url == "" {
				return fmt.Errorf("--url is required")
			}

			cfgPath := config.ResolveMarketplacesConfigPath()
			cfg, err := config.ReadMarketplacesConfig(cfgPath)
			if err != nil {
				return fmt.Errorf("reading marketplaces config: %w", err)
			}

			// If name not provided, try to discover it
			if name == "" && !noCheck {
				discovered, discErr := registry.DiscoverMarketplace(url)
				if discErr == nil {
					name = discovered.Marketplace.Name
					fmt.Printf("  ✓ Discovered marketplace: %s\n", name)
				}
			}
			if name == "" {
				return fmt.Errorf("--name is required (could not auto-discover)")
			}

			entry := config.MarketplaceEntry{Name: name, URL: url, Enabled: true}
			if err := cfg.AddMarketplace(entry); err != nil {
				return err
			}

			if err := config.WriteMarketplacesConfig(cfgPath, cfg); err != nil {
				return err
			}

			fmt.Printf("  ✓ Added marketplace %q (%s)\n", name, url)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Marketplace name (auto-discovered if omitted)")
	cmd.Flags().StringVar(&url, "url", "", "Marketplace URL")
	cmd.Flags().BoolVar(&noCheck, "no-check", false, "Skip marketplace validation")
	return cmd
}

func newMarketplaceRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove [name]",
		Short: "Remove a marketplace source",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			cfgPath := config.ResolveMarketplacesConfigPath()
			cfg, err := config.ReadMarketplacesConfig(cfgPath)
			if err != nil {
				return fmt.Errorf("reading marketplaces config: %w", err)
			}

			if err := cfg.RemoveMarketplace(name); err != nil {
				return err
			}

			if err := config.WriteMarketplacesConfig(cfgPath, cfg); err != nil {
				return err
			}

			fmt.Printf("  ✓ Removed marketplace %q\n", name)
			return nil
		},
	}

	return cmd
}

func newMarketplaceUpdateCmd() *cobra.Command {
	var (
		name string
		url  string
	)

	cmd := &cobra.Command{
		Use:   "update [name]",
		Short: "Update a marketplace source",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			targetName := args[0]

			cfgPath := config.ResolveMarketplacesConfigPath()
			cfg, err := config.ReadMarketplacesConfig(cfgPath)
			if err != nil {
				return fmt.Errorf("reading marketplaces config: %w", err)
			}

			existing := cfg.FindMarketplace(targetName)
			if existing == nil {
				return fmt.Errorf("marketplace %q not found", targetName)
			}

			updated := *existing
			if name != "" {
				updated.Name = name
			}
			if url != "" {
				updated.URL = url
			}

			if err := cfg.UpdateMarketplace(targetName, updated); err != nil {
				return err
			}

			if err := config.WriteMarketplacesConfig(cfgPath, cfg); err != nil {
				return err
			}

			fmt.Printf("  ✓ Updated marketplace %q\n", targetName)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "New marketplace name")
	cmd.Flags().StringVar(&url, "url", "", "New marketplace URL")
	return cmd
}
```

- [ ] **Step 2: Register in `root.go`**

Add `cmd.AddCommand(newMarketplaceCmd())` to `NewRootCmd()`.

- [ ] **Step 3: Verify build**

Run: `go build ./cmd/asds/`
Expected: SUCCESS

- [ ] **Step 4: Test CLI subcommands manually**

```bash
./bin/asds marketplace list
./bin/asds marketplace add --url github.com/anthropics/claude-plugins-official
./bin/asds marketplace list
./bin/asds marketplace remove test-name
```

- [ ] **Step 5: Commit**

```bash
git add internal/commands/marketplace.go internal/commands/root.go
git commit -m "feat: add marketplace CLI subcommands (list/add/remove/update)"
```

---

## Task 10: Final integration test and cleanup

**Files:**
- Modify: `internal/commands/commands_test.go` (if marketplace test needed)
- Verify all files compile

- [ ] **Step 1: Run full test suite**

Run: `go test ./... -v`
Expected: ALL PASS

- [ ] **Step 2: Run vet and fmt**

Run: `go vet ./... && gofmt -l .`
Expected: no issues

- [ ] **Step 3: Build and smoke test**

```bash
go build -o bin/asds ./cmd/asds/
./bin/asds marketplace list
./bin/asds --version
```

- [ ] **Step 4: Final commit**

```bash
git add -A
git commit -m "feat: complete multi-marketplace support with fallback chain, CLI, and TUI"
```
