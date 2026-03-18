# Part 5: Registry Fetch + CLI Commands

**Dependencies:** Part 4 (installer layer)
**Can run in parallel with:** Part 6 (TUI foundation)
**Estimated tasks:** 4

---

## Chunk 5: HTTP Registry Fetch and Cobra CLI Commands

### Task 17: HTTP Registry Fetch

**Files:**

- Create: `pkg/registry/fetch.go`
- Create: `pkg/registry/fetch_test.go`

- [ ] **Step 1: Write tests for registry fetch**

Create `pkg/registry/fetch_test.go`:

```go
package registry_test

import (
 "net/http"
 "net/http/httptest"
 "testing"

 "github.com/andrew-le-mfv/asds-marketplace-setup/pkg/registry"
)

func TestFetchMarketplaceConfig_Success(t *testing.T) {
 yamlContent := `
schema_version: 1
marketplace:
  name: "test-marketplace"
  description: "Test"
  version: "1.0.0"
  registry_url: "github.com/test"
roles:
  developer:
    display_name: "Developer"
    description: "Dev"
    plugins:
      - name: "test-plugin"
        source: "test-plugin@test"
        required: true
defaults:
  scope: project
`
 server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "text/yaml")
  w.Write([]byte(yamlContent))
 }))
 defer server.Close()

 cfg, err := registry.FetchMarketplaceConfig(server.URL)
 if err != nil {
  t.Fatalf("unexpected error: %v", err)
 }

 if cfg.Marketplace.Name != "test-marketplace" {
  t.Errorf("name = %q, want %q", cfg.Marketplace.Name, "test-marketplace")
 }
 if len(cfg.Roles) != 1 {
  t.Errorf("role count = %d, want 1", len(cfg.Roles))
 }
}

func TestFetchMarketplaceConfig_Errors(t *testing.T) {
 tests := []struct {
  name    string
  url     func() string
 }{
  {
   name: "HTTP 404",
   url: func() string {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
     w.WriteHeader(http.StatusNotFound)
    }))
    t.Cleanup(server.Close)
    return server.URL
   },
  },
  {
   name: "invalid YAML",
   url: func() string {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
     w.Write([]byte("{{{not yaml"))
    }))
    t.Cleanup(server.Close)
    return server.URL
   },
  },
  {
   name: "unreachable server",
   url: func() string {
    return "http://localhost:1"
   },
  },
 }

 for _, tt := range tests {
  t.Run(tt.name, func(t *testing.T) {
   _, err := registry.FetchMarketplaceConfig(tt.url())
   if err == nil {
    t.Error("expected error, got nil")
   }
  })
 }
}

func TestBuildRawURL(t *testing.T) {
 tests := []struct {
  input string
  want  string
 }{
  {
   "github.com/your-org/asds-marketplace",
   "https://raw.githubusercontent.com/your-org/asds-marketplace/main/asds-marketplace.yaml",
  },
  {
   "https://example.com/config.yaml",
   "https://example.com/config.yaml",
  },
 }

 for _, tt := range tests {
  got := registry.BuildRawURL(tt.input)
  if got != tt.want {
   t.Errorf("BuildRawURL(%q) = %q, want %q", tt.input, got, tt.want)
  }
 }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/registry/ -v`
Expected: FAIL — package doesn't exist

- [ ] **Step 3: Implement registry fetch**

Create `pkg/registry/fetch.go`:

```go
package registry

import (
 "fmt"
 "io"
 "net/http"
 "strings"
 "time"

 "github.com/andrew-le-mfv/asds-marketplace-setup/internal/config"
)

const (
 defaultTimeout    = 15 * time.Second
 defaultYAMLFile   = "asds-marketplace.yaml"
)

// BuildRawURL converts a registry URL to a raw content URL.
// If the URL already starts with "http", returns it as-is.
// If it looks like "github.com/org/repo", builds a raw.githubusercontent.com URL.
func BuildRawURL(registryURL string) string {
 if strings.HasPrefix(registryURL, "http://") || strings.HasPrefix(registryURL, "https://") {
  return registryURL
 }

 // Assume github.com/org/repo format
 if strings.HasPrefix(registryURL, "github.com/") {
  path := strings.TrimPrefix(registryURL, "github.com/")
  return fmt.Sprintf("https://raw.githubusercontent.com/%s/main/%s", path, defaultYAMLFile)
 }

 // Fallback: prepend https
 return "https://" + registryURL
}

// FetchMarketplaceConfig fetches and parses a marketplace config from a URL.
func FetchMarketplaceConfig(url string) (*config.MarketplaceConfig, error) {
 client := &http.Client{Timeout: defaultTimeout}

 resp, err := client.Get(url)
 if err != nil {
  return nil, fmt.Errorf("fetching marketplace config: %w", err)
 }
 defer resp.Body.Close()

 if resp.StatusCode != http.StatusOK {
  return nil, fmt.Errorf("fetching marketplace config: HTTP %d", resp.StatusCode)
 }

 data, err := io.ReadAll(resp.Body)
 if err != nil {
  return nil, fmt.Errorf("reading response body: %w", err)
 }

 cfg, err := config.ParseMarketplaceConfig(data)
 if err != nil {
  return nil, fmt.Errorf("parsing remote marketplace config: %w", err)
 }

 return cfg, nil
}

// FetchOrDefault tries to fetch remote config, falls back to embedded default.
func FetchOrDefault(registryURL string) (*config.MarketplaceConfig, error) {
 rawURL := BuildRawURL(registryURL)
 cfg, err := FetchMarketplaceConfig(rawURL)
 if err != nil {
  // Fall back to embedded default
  return config.DefaultMarketplaceConfig()
 }
 return cfg, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./pkg/registry/ -v`
Expected: all tests PASS

- [ ] **Step 5: Commit**

```bash
git add pkg/registry/fetch.go pkg/registry/fetch_test.go
git commit -m "feat: add HTTP registry fetch with fallback to embedded defaults"
```

---

### Task 18: Cobra Install Command

**Files:**

- Create: `internal/commands/install.go`

- [ ] **Step 1: Implement the install command**

Create `internal/commands/install.go`:

```go
package commands

import (
 "fmt"
 "time"

 "github.com/spf13/cobra"

 "github.com/andrew-le-mfv/asds-marketplace-setup/internal/claude"
 "github.com/andrew-le-mfv/asds-marketplace-setup/internal/config"
 "github.com/andrew-le-mfv/asds-marketplace-setup/internal/installer"
 "github.com/andrew-le-mfv/asds-marketplace-setup/pkg/registry"
)

func newInstallCmd() *cobra.Command {
 var (
  role        string
  scope       string
  projectRoot string
  yes         bool
 )

 cmd := &cobra.Command{
  Use:   "install",
  Short: "Install ASDS plugins for a role",
  Long:  "Install ASDS plugins for the selected role and scope. Prompts interactively if flags are missing.",
  RunE: func(cmd *cobra.Command, args []string) error {
   // Resolve project root
   if projectRoot == "" {
    cwd, _ := cmd.Flags().GetString("project-root")
    if cwd == "" {
     var err error
     projectRoot, err = claude.FindProjectRoot(".")
     if err != nil {
      return fmt.Errorf("finding project root: %w", err)
     }
    }
   }

   // If role or scope is missing, launch interactive mode
   // Interactive wizard is wired in Part 7 (TUI tabs)
   if role == "" || scope == "" {
    return fmt.Errorf("interactive mode not yet implemented; provide --role and --scope flags")
   }

   // Validate scope
   s, err := config.ParseScope(scope)
   if err != nil {
    return err
   }

   // Validate project root for project/local scope
   if s != config.ScopeUser && projectRoot == "" {
    return fmt.Errorf("no project root found; use --project-root or run from a git repository")
   }

   // Load ASDS config for marketplace URL
   asdsCfg, err := config.ReadASDSConfig(config.ResolveASDSConfigPath())
   if err != nil {
    return fmt.Errorf("reading ASDS config: %w", err)
   }

   // Fetch marketplace config
   mktCfg, err := registry.FetchOrDefault(asdsCfg.MarketplaceURL)
   if err != nil {
    return fmt.Errorf("loading marketplace config: %w", err)
   }

   // Validate role exists
   roleConfig, ok := mktCfg.Roles[role]
   if !ok {
    return fmt.Errorf("unknown role %q; available roles: %v", role, mktCfg.RoleNames())
   }

   // Show confirmation
   if !yes {
    fmt.Printf("Role: %s (%s)\n", roleConfig.DisplayName, roleConfig.Description)
    fmt.Printf("Scope: %s\n", scope)
    fmt.Printf("Plugins (%d):\n", len(roleConfig.Plugins))
    for _, p := range roleConfig.Plugins {
     req := ""
     if p.Required {
      req = " (required)"
     }
     fmt.Printf("  - %s%s\n", p.Name, req)
    }
    fmt.Println()
    // In non-interactive mode, --yes is required for no prompt
    // For now, just proceed (interactive confirmation in Part 7)
   }

   // Create installer
   inst := installer.NewInstaller(true)

   // Register marketplace
   if err := inst.RegisterMarketplace(mktCfg.Marketplace.Name, mktCfg.Marketplace.RegistryURL); err != nil {
    fmt.Printf("Warning: marketplace registration failed: %v\n", err)
   }

   // Install plugins
   results, err := inst.Install(roleConfig.Plugins, s, projectRoot)
   if err != nil {
    return fmt.Errorf("installing plugins: %w", err)
   }

   // Report results
   var failures int
   for _, r := range results {
    if r.Success {
     fmt.Printf("  ✓ %s\n", r.PluginRef)
    } else {
     fmt.Printf("  ✗ %s: %v\n", r.PluginRef, r.Error)
     failures++
    }
   }

   // Scaffold CLAUDE.md (project/local scope only)
   if s != config.ScopeUser && len(roleConfig.ClaudeMDSnippets) > 0 {
    claudeMDPath, err := claude.ClaudeMDPath(s, projectRoot)
    if err == nil {
     if err := claude.UpsertMarkerBlock(claudeMDPath, role, roleConfig.ClaudeMDSnippets); err != nil {
      fmt.Printf("Warning: CLAUDE.md update failed: %v\n", err)
     } else {
      fmt.Printf("  ✓ CLAUDE.md updated\n")
     }
    }
   }

   // Write manifest
   now := time.Now().UTC()
   manifestPlugins := make([]config.ManifestPlugin, len(roleConfig.Plugins))
   for i, p := range roleConfig.Plugins {
    manifestPlugins[i] = config.ManifestPlugin{
     Name:        p.Name,
     FullRef:     p.Source,
     Required:    p.Required,
     InstalledAt: now,
    }
   }

   manifest := &config.Manifest{
    SchemaVersion:      1,
    ASDSVersion:        version,
    InstalledAt:        now,
    UpdatedAt:          now,
    Role:               role,
    Scope:              s,
    MarketplaceSource:  mktCfg.Marketplace.RegistryURL,
    InstallMethod:      inst.Method(),
    ClaudeCodeDetected: installer.DetectClaudeCode().Found,
    Plugins:            manifestPlugins,
    ClaudeMDModified:   s != config.ScopeUser && len(roleConfig.ClaudeMDSnippets) > 0,
    ScaffoldedFiles:    []string{claude.SettingsPath(s, projectRoot)},
   }

   manifestPath := claude.ManifestPath(s, projectRoot)
   if err := config.WriteManifest(manifestPath, manifest); err != nil {
    fmt.Printf("Warning: manifest write failed: %v\n", err)
   }

   // Gitignore for local scope
   if s == config.ScopeLocal {
    claudeDir := claude.SettingsPath(s, projectRoot)
    claude.EnsureGitignore(
     claudeDir[:len(claudeDir)-len("settings.local.json")],
     ".asds-manifest.local.json",
    )
   }

   if failures > 0 {
    return fmt.Errorf("%d plugin(s) failed to install", failures)
   }

   fmt.Println("\n✅ ASDS setup complete!")
   return nil
  },
 }

 cmd.Flags().StringVar(&role, "role", "", "Role to install (e.g., developer, frontend, backend)")
 cmd.Flags().StringVar(&scope, "scope", "", "Installation scope: user, project, or local")
 cmd.Flags().StringVar(&projectRoot, "project-root", "", "Override project root directory")
 cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")

 return cmd
}
```

- [ ] **Step 2: Register install command in root**

Add to `internal/commands/root.go` inside `NewRootCmd()`, before `return cmd`:

```go
 cmd.AddCommand(newInstallCmd())
```

- [ ] **Step 3: Verify it compiles**

Run: `go build ./cmd/asds/`
Expected: compiles without errors

- [ ] **Step 4: Test help output**

Run: `go run ./cmd/asds/ install --help`
Expected: shows install usage with --role, --scope, --project-root, --yes flags

- [ ] **Step 5: Commit**

```bash
git add internal/commands/install.go internal/commands/root.go
git commit -m "feat: add install CLI command with full lifecycle"
```

---

### Task 19: Cobra Uninstall, Update, Status, Reset Commands

**Files:**

- Create: `internal/commands/uninstall.go`
- Create: `internal/commands/update.go`
- Create: `internal/commands/status.go`
- Create: `internal/commands/reset.go`

**Note:** All 4 command files (uninstall, status, update, reset) should be created before attempting to compile, as `removeFile` in `uninstall.go` is defined in `reset.go`. Alternatively, create the shared helper first.

- [ ] **Step 1: Implement uninstall command**

Create `internal/commands/uninstall.go`:

```go
package commands

import (
 "fmt"

 "github.com/spf13/cobra"

 "github.com/andrew-le-mfv/asds-marketplace-setup/internal/claude"
 "github.com/andrew-le-mfv/asds-marketplace-setup/internal/config"
 "github.com/andrew-le-mfv/asds-marketplace-setup/internal/installer"
)

func newUninstallCmd() *cobra.Command {
 var (
  scope       string
  projectRoot string
  yes         bool
 )

 cmd := &cobra.Command{
  Use:   "uninstall",
  Short: "Uninstall ASDS plugins",
  Long:  "Remove all ASDS-installed plugins for the specified scope.",
  RunE: func(cmd *cobra.Command, args []string) error {
   if projectRoot == "" {
    var err error
    projectRoot, err = claude.FindProjectRoot(".")
    if err != nil {
     return fmt.Errorf("finding project root: %w", err)
    }
   }

   if scope == "" {
    return fmt.Errorf("--scope is required (user, project, or local)")
   }

   s, err := config.ParseScope(scope)
   if err != nil {
    return err
   }

   // Read manifest
   manifestPath := claude.ManifestPath(s, projectRoot)
   manifest, err := config.ReadManifest(manifestPath)
   if err != nil {
    return fmt.Errorf("no ASDS installation found for scope %q: %v", scope, err)
   }

   if !yes {
    fmt.Printf("Will uninstall %d plugins for role %q (scope: %s)\n", len(manifest.Plugins), manifest.Role, scope)
    // Interactive confirmation would go here (Part 7)
   }

   // Uninstall plugins
   pluginRefs := make([]string, len(manifest.Plugins))
   for i, p := range manifest.Plugins {
    pluginRefs[i] = p.FullRef
   }

   inst := installer.NewInstaller(true)
   results, err := inst.Uninstall(pluginRefs, s, projectRoot)
   if err != nil {
    return fmt.Errorf("uninstalling plugins: %w", err)
   }

   for _, r := range results {
    if r.Success {
     fmt.Printf("  ✓ removed %s\n", r.PluginRef)
    } else {
     fmt.Printf("  ✗ %s: %v\n", r.PluginRef, r.Error)
    }
   }

   // Remove CLAUDE.md marker block
   if s != config.ScopeUser {
    claudeMDPath, err := claude.ClaudeMDPath(s, projectRoot)
    if err == nil {
     claude.RemoveMarkerBlock(claudeMDPath)
     fmt.Println("  ✓ CLAUDE.md cleaned")
    }
   }

   // Remove manifest
   if err := removeFile(manifestPath); err != nil {
    fmt.Printf("Warning: could not remove manifest: %v\n", err)
   }

   fmt.Println("\n✅ ASDS uninstall complete!")
   return nil
  },
 }

 cmd.Flags().StringVar(&scope, "scope", "", "Scope to uninstall from: user, project, or local")
 cmd.Flags().StringVar(&projectRoot, "project-root", "", "Override project root directory")
 cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")

 return cmd
}
```

- [ ] **Step 2: Implement status command**

Create `internal/commands/status.go`:

```go
package commands

import (
 "encoding/json"
 "fmt"
 "os"

 "github.com/spf13/cobra"

 "github.com/andrew-le-mfv/asds-marketplace-setup/internal/claude"
 "github.com/andrew-le-mfv/asds-marketplace-setup/internal/config"
 "github.com/andrew-le-mfv/asds-marketplace-setup/internal/installer"
)

func newStatusCmd() *cobra.Command {
 var (
  jsonOutput  bool
  projectRoot string
 )

 cmd := &cobra.Command{
  Use:   "status",
  Short: "Show current ASDS setup status",
  RunE: func(cmd *cobra.Command, args []string) error {
   if projectRoot == "" {
    var err error
    projectRoot, err = claude.FindProjectRoot(".")
    if err != nil {
     projectRoot = "."
    }
   }

   detection := installer.DetectClaudeCode()

   // Try to find manifests for all scopes
   type scopeStatus struct {
    Scope    string           `json:"scope"`
    Found    bool             `json:"found"`
    Manifest *config.Manifest `json:"manifest,omitempty"`
   }

   statuses := make([]scopeStatus, 0, 3)
   for _, s := range []config.Scope{config.ScopeUser, config.ScopeProject, config.ScopeLocal} {
    mp := claude.ManifestPath(s, projectRoot)
    m, err := config.ReadManifest(mp)
    if err != nil {
     statuses = append(statuses, scopeStatus{Scope: string(s), Found: false})
    } else {
     statuses = append(statuses, scopeStatus{Scope: string(s), Found: true, Manifest: m})
    }
   }

   if jsonOutput {
    output := map[string]any{
     "claude_code_detected": detection.Found,
     "claude_code_path":     detection.Path,
     "project_root":         projectRoot,
     "scopes":               statuses,
    }
    enc := json.NewEncoder(os.Stdout)
    enc.SetIndent("", "  ")
    return enc.Encode(output)
   }

   // Human-readable output
   fmt.Println("🔍 ASDS Status")
   fmt.Println()
   if detection.Found {
    fmt.Printf("  Claude Code: ✓ detected at %s\n", detection.Path)
   } else {
    fmt.Printf("  Claude Code: ✗ not detected\n")
   }
   fmt.Printf("  Project root: %s\n", projectRoot)
   fmt.Println()

   for _, ss := range statuses {
    if ss.Found {
     fmt.Printf("  [%s] Role: %s | Plugins: %d | Method: %s | Installed: %s\n",
      ss.Scope, ss.Manifest.Role, len(ss.Manifest.Plugins),
      ss.Manifest.InstallMethod, ss.Manifest.InstalledAt.Format("2006-01-02"))
    } else {
     fmt.Printf("  [%s] Not installed\n", ss.Scope)
    }
   }

   return nil
  },
 }

 cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
 cmd.Flags().StringVar(&projectRoot, "project-root", "", "Override project root directory")

 return cmd
}
```

- [ ] **Step 3: Implement update command**

Create `internal/commands/update.go`:

```go
package commands

import (
 "fmt"
 "time"

 "github.com/spf13/cobra"

 "github.com/andrew-le-mfv/asds-marketplace-setup/internal/claude"
 "github.com/andrew-le-mfv/asds-marketplace-setup/internal/config"
 "github.com/andrew-le-mfv/asds-marketplace-setup/internal/installer"
 "github.com/andrew-le-mfv/asds-marketplace-setup/pkg/registry"
)

func newUpdateCmd() *cobra.Command {
 var (
  scope       string
  projectRoot string
 )

 cmd := &cobra.Command{
  Use:   "update",
  Short: "Update ASDS plugins to latest marketplace config",
  RunE: func(cmd *cobra.Command, args []string) error {
   if projectRoot == "" {
    var err error
    projectRoot, err = claude.FindProjectRoot(".")
    if err != nil {
     return fmt.Errorf("finding project root: %w", err)
    }
   }

   if scope == "" {
    return fmt.Errorf("--scope is required (user, project, or local)")
   }

   s, err := config.ParseScope(scope)
   if err != nil {
    return err
   }

   // Read current manifest
   manifestPath := claude.ManifestPath(s, projectRoot)
   manifest, err := config.ReadManifest(manifestPath)
   if err != nil {
    return fmt.Errorf("no ASDS installation found for scope %q", scope)
   }

   // Fetch latest config
   asdsCfg, _ := config.ReadASDSConfig(config.ResolveASDSConfigPath())
   mktCfg, err := registry.FetchOrDefault(asdsCfg.MarketplaceURL)
   if err != nil {
    return fmt.Errorf("loading marketplace config: %w", err)
   }

   roleConfig, ok := mktCfg.Roles[manifest.Role]
   if !ok {
    return fmt.Errorf("role %q no longer exists in marketplace", manifest.Role)
   }

   // Compute diff
   currentRefs := make(map[string]bool)
   for _, p := range manifest.Plugins {
    currentRefs[p.FullRef] = true
   }
   newRefs := make(map[string]bool)
   for _, p := range roleConfig.Plugins {
    newRefs[p.Source] = true
   }

   // Plugins to add (in new but not current)
   var toAdd []config.PluginRef
   for _, p := range roleConfig.Plugins {
    if !currentRefs[p.Source] {
     toAdd = append(toAdd, p)
    }
   }

   // Plugins to remove (in current but not new)
   var toRemove []string
   for _, p := range manifest.Plugins {
    if !newRefs[p.FullRef] {
     toRemove = append(toRemove, p.FullRef)
    }
   }

   inst := installer.NewInstaller(true)

   if len(toAdd) > 0 {
    fmt.Println("Adding new plugins:")
    results, err := inst.Install(toAdd, s, projectRoot)
    if err != nil {
     return err
    }
    for _, r := range results {
     if r.Success {
      fmt.Printf("  ✓ %s\n", r.PluginRef)
     } else {
      fmt.Printf("  ✗ %s: %v\n", r.PluginRef, r.Error)
     }
    }
   }

   if len(toRemove) > 0 {
    fmt.Println("Removing old plugins:")
    results, err := inst.Uninstall(toRemove, s, projectRoot)
    if err != nil {
     return err
    }
    for _, r := range results {
     if r.Success {
      fmt.Printf("  ✓ removed %s\n", r.PluginRef)
     } else {
      fmt.Printf("  ✗ %s: %v\n", r.PluginRef, r.Error)
     }
    }
   }

   if len(toAdd) == 0 && len(toRemove) == 0 {
    fmt.Println("Everything is up to date!")
   }

   // Update CLAUDE.md
   if s != config.ScopeUser && len(roleConfig.ClaudeMDSnippets) > 0 {
    claudeMDPath, _ := claude.ClaudeMDPath(s, projectRoot)
    claude.UpsertMarkerBlock(claudeMDPath, manifest.Role, roleConfig.ClaudeMDSnippets)
   }

   // Update manifest
   now := time.Now().UTC()
   manifestPlugins := make([]config.ManifestPlugin, len(roleConfig.Plugins))
   for i, p := range roleConfig.Plugins {
    manifestPlugins[i] = config.ManifestPlugin{
     Name:        p.Name,
     FullRef:     p.Source,
     Required:    p.Required,
     InstalledAt: now,
    }
   }
   manifest.UpdatedAt = now
   manifest.Plugins = manifestPlugins
   config.WriteManifest(manifestPath, manifest)

   fmt.Println("\n✅ ASDS update complete!")
   return nil
  },
 }

 cmd.Flags().StringVar(&scope, "scope", "", "Scope to update: user, project, or local")
 cmd.Flags().StringVar(&projectRoot, "project-root", "", "Override project root directory")

 return cmd
}
```

- [ ] **Step 4: Implement reset command**

Create `internal/commands/reset.go`:

```go
package commands

import (
 "fmt"
 "os"

 "github.com/spf13/cobra"

 "github.com/andrew-le-mfv/asds-marketplace-setup/internal/claude"
 "github.com/andrew-le-mfv/asds-marketplace-setup/internal/config"
 "github.com/andrew-le-mfv/asds-marketplace-setup/internal/installer"
)

func newResetCmd() *cobra.Command {
 var (
  scope       string
  projectRoot string
  yes         bool
 )

 cmd := &cobra.Command{
  Use:   "reset",
  Short: "Remove all ASDS configuration for a scope",
  Long:  "Completely removes all ASDS traces for the specified scope. Marketplace registration is preserved.",
  RunE: func(cmd *cobra.Command, args []string) error {
   if projectRoot == "" {
    var err error
    projectRoot, err = claude.FindProjectRoot(".")
    if err != nil {
     return fmt.Errorf("finding project root: %w", err)
    }
   }

   if scope == "" {
    return fmt.Errorf("--scope is required (user, project, or local)")
   }

   s, err := config.ParseScope(scope)
   if err != nil {
    return err
   }

   if !yes {
    return fmt.Errorf("reset requires --yes flag for non-interactive mode")
   }

   // Read manifest to find what to clean up
   manifestPath := claude.ManifestPath(s, projectRoot)
   manifest, err := config.ReadManifest(manifestPath)
   if err != nil {
    fmt.Println("No ASDS installation found — nothing to reset.")
    return nil
   }

   // Remove plugins from settings
   pluginRefs := make([]string, len(manifest.Plugins))
   for i, p := range manifest.Plugins {
    pluginRefs[i] = p.FullRef
   }

   inst := installer.NewInstaller(true)
   inst.Uninstall(pluginRefs, s, projectRoot)

   // Remove CLAUDE.md marker
   if s != config.ScopeUser {
    claudeMDPath, err := claude.ClaudeMDPath(s, projectRoot)
    if err == nil {
     claude.RemoveMarkerBlock(claudeMDPath)
    }
   }

   // Delete manifest
   removeFile(manifestPath)

   fmt.Println("✅ ASDS reset complete for scope:", scope)
   return nil
  },
 }

 cmd.Flags().StringVar(&scope, "scope", "", "Scope to reset: user, project, or local")
 cmd.Flags().StringVar(&projectRoot, "project-root", "", "Override project root directory")
 cmd.Flags().BoolVar(&yes, "yes", false, "Confirm reset without prompting")

 return cmd
}

// removeFile removes a file, ignoring "not exist" errors.
func removeFile(path string) error {
 if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
  return err
 }
 return nil
}
```

- [ ] **Step 5: Register all new commands in root**

Modify `internal/commands/root.go` — add inside `NewRootCmd()` before `return cmd`:

```go
 cmd.AddCommand(newUninstallCmd())
 cmd.AddCommand(newUpdateCmd())
 cmd.AddCommand(newStatusCmd())
 cmd.AddCommand(newResetCmd())
```

- [ ] **Step 6: Verify everything compiles**

Run: `go build ./cmd/asds/`
Expected: compiles without errors

- [ ] **Step 7: Verify CLI help**

Run: `go run ./cmd/asds/ --help`
Expected: shows all 5 subcommands (install, uninstall, update, status, reset)

- [ ] **Step 8: Commit**

```bash
git add internal/commands/
git commit -m "feat: add uninstall, update, status, reset CLI commands"
```

---

### Task 20: CLI Integration Test (End-to-End)

**Files:**

- Create: `internal/commands/commands_test.go`

- [ ] **Step 1: Write an integration test**

Create `internal/commands/commands_test.go`:

```go
package commands_test

import (
 "bytes"
 "testing"

 "github.com/andrew-le-mfv/asds-marketplace-setup/internal/commands"
)

func TestRootCmd_Version(t *testing.T) {
 cmd := commands.NewRootCmd()
 buf := new(bytes.Buffer)
 cmd.SetOut(buf)
 cmd.SetArgs([]string{"--version"})

 if err := cmd.Execute(); err != nil {
  t.Fatalf("unexpected error: %v", err)
 }

 if !bytes.Contains(buf.Bytes(), []byte("0.1.0")) {
  t.Errorf("version output = %q, want to contain '0.1.0'", buf.String())
 }
}

func TestInstallCmd_MissingFlags(t *testing.T) {
 cmd := commands.NewRootCmd()
 cmd.SetArgs([]string{"install"})

 err := cmd.Execute()
 if err == nil {
  t.Error("expected error for missing flags, got nil")
 }
}

func TestStatusCmd_Runs(t *testing.T) {
 cmd := commands.NewRootCmd()
 buf := new(bytes.Buffer)
 cmd.SetOut(buf)
 cmd.SetArgs([]string{"status", "--project-root", t.TempDir()})

 err := cmd.Execute()
 if err != nil {
  t.Fatalf("status should not error: %v", err)
 }
}

func TestResetCmd_RequiresYes(t *testing.T) {
 cmd := commands.NewRootCmd()
 cmd.SetArgs([]string{"reset", "--scope", "project", "--project-root", t.TempDir()})

 err := cmd.Execute()
 if err == nil {
  t.Error("expected error without --yes flag")
 }
}
```

- [ ] **Step 2: Run integration tests**

Run: `go test ./internal/commands/ -v`
Expected: all tests PASS

- [ ] **Step 3: Run ALL project tests**

Run: `go test ./... -v`
Expected: all tests PASS across all packages

- [ ] **Step 4: Commit**

```bash
git add internal/commands/commands_test.go
git commit -m "test: add CLI integration tests"
```
