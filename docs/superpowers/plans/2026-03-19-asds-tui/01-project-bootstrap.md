# Part 1: Project Bootstrap + Core Types

**Dependencies:** None
**Estimated tasks:** 4

---

## Chunk 1: Project Bootstrap + Core Types

### Task 1: Initialize Go Module and Dependencies

**Files:**

- Create: `go.mod`
- Create: `go.sum` (auto-generated)

- [ ] **Step 1: Initialize Go module**

```bash
cd /Users/le.tuan.anh/Workspace/MFVWorkspace/asds-marketplace-setup
go mod init github.com/andrew-le-mfv/asds-marketplace-setup
```

- [ ] **Step 2: Add all dependencies**

```bash
go get github.com/charmbracelet/bubbletea@latest
go get github.com/charmbracelet/huh@latest
go get github.com/charmbracelet/lipgloss@latest
go get github.com/charmbracelet/bubbles@latest
go get github.com/spf13/cobra@latest
go get gopkg.in/yaml.v3@latest
```

- [ ] **Step 3: Verify go.mod looks correct**

Run: `cat go.mod`
Expected: module line + require block listing all 6 deps

- [ ] **Step 4: Commit**

```bash
git add go.mod go.sum
git commit -m "chore: initialize Go module with dependencies"
```

---

### Task 2: Create Entry Point and Minimal Root Command

**Files:**

- Create: `cmd/asds/main.go`
- Create: `internal/commands/root.go`

- [ ] **Step 1: Create root command**

Create `internal/commands/root.go`:

```go
package commands

import (
 "fmt"

 "github.com/spf13/cobra"
)

const version = "0.1.0"

func NewRootCmd() *cobra.Command {
 cmd := &cobra.Command{
  Use:   "asds",
  Short: "ASDS — Agentic Software Development Suite",
  Long:  "A TUI for bootstrapping developers into curated Claude Code plugin sets organized by role.",
  RunE: func(cmd *cobra.Command, args []string) error {
   // Will launch dashboard TUI in Part 6
   fmt.Println("ASDS dashboard TUI — coming soon")
   return nil
  },
 }

 cmd.Version = version

 return cmd
}
```

- [ ] **Step 2: Create main.go entry point**

Create `cmd/asds/main.go`:

```go
package main

import (
 "os"

 "github.com/andrew-le-mfv/asds-marketplace-setup/internal/commands"
)

func main() {
 cmd := commands.NewRootCmd()
 if err := cmd.Execute(); err != nil {
  os.Exit(1)
 }
}
```

- [ ] **Step 3: Verify it compiles and runs**

Run: `go run ./cmd/asds/`
Expected: `ASDS dashboard TUI — coming soon`

Run: `go run ./cmd/asds/ --version`
Expected: `asds version 0.1.0`

- [ ] **Step 4: Commit**

```bash
git add cmd/ internal/commands/
git commit -m "feat: add entry point and root cobra command"
```

---

### Task 3: Define Core Domain Types — Marketplace Config

**Files:**

- Create: `internal/config/marketplace.go`
- Create: `internal/config/marketplace_test.go`

- [ ] **Step 1: Write test for MarketplaceConfig parsing**

Create `internal/config/marketplace_test.go`:

```go
package config_test

import (
 "testing"

 "github.com/andrew-le-mfv/asds-marketplace-setup/internal/config"
)

func TestParseMarketplaceConfig(t *testing.T) {
 yamlData := []byte(`
schema_version: 1
marketplace:
  name: "test-marketplace"
  description: "Test"
  version: "1.0.0"
  registry_url: "github.com/test/marketplace"
roles:
  developer:
    display_name: "Software Developer"
    description: "Full-stack development"
    plugins:
      - name: "code-reviewer"
        source: "code-reviewer@test-marketplace"
        required: true
      - name: "commit-commands"
        source: "commit-commands@test-marketplace"
        required: false
    claude_md_snippets:
      - "Follow conventional commits"
      - "Always write tests"
defaults:
  scope: project
  auto_register_marketplace: true
`)

 cfg, err := config.ParseMarketplaceConfig(yamlData)
 if err != nil {
  t.Fatalf("unexpected error: %v", err)
 }

 if cfg.SchemaVersion != 1 {
  t.Errorf("schema_version = %d, want 1", cfg.SchemaVersion)
 }
 if cfg.Marketplace.Name != "test-marketplace" {
  t.Errorf("marketplace.name = %q, want %q", cfg.Marketplace.Name, "test-marketplace")
 }
 if len(cfg.Roles) != 1 {
  t.Fatalf("roles count = %d, want 1", len(cfg.Roles))
 }

 dev, ok := cfg.Roles["developer"]
 if !ok {
  t.Fatal("missing role 'developer'")
 }
 if dev.DisplayName != "Software Developer" {
  t.Errorf("display_name = %q, want %q", dev.DisplayName, "Software Developer")
 }
 if len(dev.Plugins) != 2 {
  t.Fatalf("plugins count = %d, want 2", len(dev.Plugins))
 }
 if dev.Plugins[0].Name != "code-reviewer" {
  t.Errorf("plugin[0].name = %q, want %q", dev.Plugins[0].Name, "code-reviewer")
 }
 if !dev.Plugins[0].Required {
  t.Error("plugin[0].required = false, want true")
 }
 if len(dev.ClaudeMDSnippets) != 2 {
  t.Errorf("claude_md_snippets count = %d, want 2", len(dev.ClaudeMDSnippets))
 }
 if cfg.Defaults.Scope != "project" {
  t.Errorf("defaults.scope = %q, want %q", cfg.Defaults.Scope, "project")
 }
}

func TestParseMarketplaceConfig_InvalidYAML(t *testing.T) {
 _, err := config.ParseMarketplaceConfig([]byte(`{{{invalid`))
 if err == nil {
  t.Error("expected error for invalid YAML, got nil")
 }
}

func TestMarketplaceConfig_RoleNames(t *testing.T) {
 yamlData := []byte(`
schema_version: 1
marketplace:
  name: "test"
  version: "1.0.0"
  registry_url: "github.com/test"
roles:
  developer:
    display_name: "Developer"
    description: "Dev"
  frontend:
    display_name: "Frontend"
    description: "FE"
defaults:
  scope: project
`)
 cfg, err := config.ParseMarketplaceConfig(yamlData)
 if err != nil {
  t.Fatalf("unexpected error: %v", err)
 }

 names := cfg.RoleNames()
 if len(names) != 2 {
  t.Fatalf("role names count = %d, want 2", len(names))
 }
 // RoleNames returns sorted keys
 if names[0] != "developer" || names[1] != "frontend" {
  t.Errorf("role names = %v, want [developer frontend]", names)
 }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/config/ -run TestParseMarketplace -v`
Expected: FAIL — `ParseMarketplaceConfig` not defined

- [ ] **Step 3: Implement marketplace types and parser**

Create `internal/config/marketplace.go`:

```go
package config

import (
 "fmt"
 "sort"

 "gopkg.in/yaml.v3"
)

// MarketplaceConfig represents the parsed asds-marketplace.yaml.
type MarketplaceConfig struct {
 SchemaVersion int                `yaml:"schema_version"`
 Marketplace   MarketplaceInfo    `yaml:"marketplace"`
 Roles         map[string]Role    `yaml:"roles"`
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
 DisplayName     string       `yaml:"display_name"`
 Description     string       `yaml:"description"`
 Plugins         []PluginRef  `yaml:"plugins"`
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
 Scope                    string `yaml:"scope"`
 AutoRegisterMarketplace  bool   `yaml:"auto_register_marketplace"`
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
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/config/ -run TestParseMarketplace -v`
Expected: all 3 tests PASS

Run: `go test ./internal/config/ -run TestMarketplaceConfig_RoleNames -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/config/marketplace.go internal/config/marketplace_test.go
git commit -m "feat: add MarketplaceConfig types and YAML parser"
```

---

### Task 4: Define Core Domain Types — Manifest and Scope

**Files:**

- Create: `internal/config/manifest.go`
- Create: `internal/config/manifest_test.go`

- [ ] **Step 1: Write test for Manifest serialization**

Create `internal/config/manifest_test.go`:

```go
package config_test

import (
 "encoding/json"
 "testing"
 "time"

 "github.com/andrew-le-mfv/asds-marketplace-setup/internal/config"
)

func TestManifest_JSON_Roundtrip(t *testing.T) {
 now := time.Date(2026, 3, 18, 10, 0, 0, 0, time.UTC)
 m := config.Manifest{
  SchemaVersion:     1,
  ASDSVersion:       "0.1.0",
  InstalledAt:       now,
  UpdatedAt:         now,
  Role:              "developer",
  Scope:             config.ScopeProject,
  MarketplaceSource: "github.com/test/marketplace",
  InstallMethod:     "direct",
  ClaudeCodeDetected: false,
  Plugins: []config.ManifestPlugin{
   {
    Name:        "code-reviewer",
    FullRef:     "code-reviewer@test-marketplace",
    Required:    true,
    InstalledAt: now,
   },
  },
  ClaudeMDModified: true,
  ScaffoldedFiles:  []string{".claude/settings.json", "CLAUDE.md"},
 }

 data, err := json.MarshalIndent(m, "", "  ")
 if err != nil {
  t.Fatalf("marshal error: %v", err)
 }

 var decoded config.Manifest
 if err := json.Unmarshal(data, &decoded); err != nil {
  t.Fatalf("unmarshal error: %v", err)
 }

 if decoded.Role != "developer" {
  t.Errorf("role = %q, want %q", decoded.Role, "developer")
 }
 if decoded.Scope != config.ScopeProject {
  t.Errorf("scope = %q, want %q", decoded.Scope, config.ScopeProject)
 }
 if len(decoded.Plugins) != 1 {
  t.Fatalf("plugins count = %d, want 1", len(decoded.Plugins))
 }
 if decoded.Plugins[0].Name != "code-reviewer" {
  t.Errorf("plugin name = %q, want %q", decoded.Plugins[0].Name, "code-reviewer")
 }
}

func TestScope_String(t *testing.T) {
 tests := []struct {
  scope config.Scope
  want  string
 }{
  {config.ScopeUser, "user"},
  {config.ScopeProject, "project"},
  {config.ScopeLocal, "local"},
 }

 for _, tt := range tests {
  if got := string(tt.scope); got != tt.want {
   t.Errorf("Scope(%q).String() = %q, want %q", tt.scope, got, tt.want)
  }
 }
}

func TestParseScope(t *testing.T) {
 tests := []struct {
  input string
  want  config.Scope
  err   bool
 }{
  {"user", config.ScopeUser, false},
  {"project", config.ScopeProject, false},
  {"local", config.ScopeLocal, false},
  {"invalid", "", true},
 }

 for _, tt := range tests {
  got, err := config.ParseScope(tt.input)
  if tt.err && err == nil {
   t.Errorf("ParseScope(%q): expected error, got nil", tt.input)
  }
  if !tt.err && err != nil {
   t.Errorf("ParseScope(%q): unexpected error: %v", tt.input, err)
  }
  if got != tt.want {
   t.Errorf("ParseScope(%q) = %q, want %q", tt.input, got, tt.want)
  }
 }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/config/ -run "TestManifest|TestScope|TestParseScope" -v`
Expected: FAIL — types not defined

- [ ] **Step 3: Implement Manifest types and Scope**

Create `internal/config/manifest.go`:

```go
package config

import (
 "fmt"
 "time"
)

// Scope represents a Claude Code plugin installation scope.
type Scope string

const (
 ScopeUser    Scope = "user"
 ScopeProject Scope = "project"
 ScopeLocal   Scope = "local"
)

// ParseScope parses a string into a Scope, returning an error for invalid values.
func ParseScope(s string) (Scope, error) {
 switch s {
 case "user":
  return ScopeUser, nil
 case "project":
  return ScopeProject, nil
 case "local":
  return ScopeLocal, nil
 default:
  return "", fmt.Errorf("invalid scope %q: must be one of user, project, local", s)
 }
}

// Manifest tracks what ASDS installed, enabling lifecycle operations.
type Manifest struct {
 SchemaVersion      int              `json:"schema_version"`
 ASDSVersion        string           `json:"asds_version"`
 InstalledAt        time.Time        `json:"installed_at"`
 UpdatedAt          time.Time        `json:"updated_at"`
 Role               string           `json:"role"`
 Scope              Scope            `json:"scope"`
 MarketplaceSource  string           `json:"marketplace_source"`
 InstallMethod      string           `json:"install_method"`
 ClaudeCodeDetected bool             `json:"claude_code_detected"`
 Plugins            []ManifestPlugin `json:"plugins"`
 ClaudeMDModified   bool             `json:"claude_md_modified"`
 ScaffoldedFiles    []string         `json:"scaffolded_files"`
}

// ManifestPlugin tracks a single installed plugin.
type ManifestPlugin struct {
 Name        string    `json:"name"`
 FullRef     string    `json:"full_ref"`
 Required    bool      `json:"required"`
 InstalledAt time.Time `json:"installed_at"`
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/config/ -run "TestManifest|TestScope|TestParseScope" -v`
Expected: all 3 tests PASS

- [ ] **Step 5: Run all config tests to confirm nothing is broken**

Run: `go test ./internal/config/ -v`
Expected: all tests PASS

- [ ] **Step 6: Commit**

```bash
git add internal/config/manifest.go internal/config/manifest_test.go
git commit -m "feat: add Manifest types and Scope enum"
```
