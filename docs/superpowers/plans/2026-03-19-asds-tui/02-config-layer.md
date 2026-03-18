# Part 2: Config Layer

**Dependencies:** Part 1 (core types must exist)
**Can run in parallel with:** Part 3
**Estimated tasks:** 4

---

## Chunk 2: Config Layer — Embedded Defaults, Manifest I/O, ASDS Config

### Task 5: Embedded Default Marketplace Config

**Files:**

- Create: `configs/default-marketplace.yaml`
- Create: `internal/config/defaults.go`
- Create: `internal/config/defaults_test.go`

- [ ] **Step 1: Create the default marketplace YAML**

Create `configs/default-marketplace.yaml`:

```yaml
schema_version: 1

marketplace:
  name: "asds-marketplace"
  description: "Agentic Software Development Suite"
  version: "1.0.0"
  registry_url: "github.com/your-org/asds-marketplace"

roles:
  developer:
    display_name: "Software Developer"
    description: "Full-stack development with code quality tools"
    plugins:
      - name: "code-reviewer"
        source: "code-reviewer@asds-marketplace"
        required: true
      - name: "commit-commands"
        source: "commit-commands@asds-marketplace"
        required: false
    claude_md_snippets:
      - "Follow conventional commits"
      - "Always write tests for new features"

  frontend:
    display_name: "Frontend Developer"
    description: "UI/UX focused development"
    plugins:
      - name: "frontend-design"
        source: "frontend-design@asds-marketplace"
        required: true
      - name: "playwright"
        source: "playwright@asds-marketplace"
        required: false

  backend:
    display_name: "Backend Developer"
    description: "API and server-side development"
    plugins:
      - name: "code-reviewer"
        source: "code-reviewer@asds-marketplace"
        required: true
      - name: "security-guidance"
        source: "security-guidance@asds-marketplace"
        required: false

  devops:
    display_name: "DevOps Engineer"
    description: "CI/CD, infrastructure, and deployment"
    plugins:
      - name: "security-guidance"
        source: "security-guidance@asds-marketplace"
        required: true

  tester:
    display_name: "QA / Tester"
    description: "Testing and quality assurance"
    plugins:
      - name: "playwright"
        source: "playwright@asds-marketplace"
        required: true

  security:
    display_name: "Security Engineer"
    description: "Security auditing and compliance"
    plugins:
      - name: "security-guidance"
        source: "security-guidance@asds-marketplace"
        required: true

  techlead:
    display_name: "Tech Lead"
    description: "Architecture, code review, and team standards"
    plugins:
      - name: "code-reviewer"
        source: "code-reviewer@asds-marketplace"
        required: true
      - name: "code-simplifier"
        source: "code-simplifier@asds-marketplace"
        required: false

  data-engineer:
    display_name: "Data Engineer"
    description: "Data pipelines and analytics"
    plugins:
      - name: "code-simplifier"
        source: "code-simplifier@asds-marketplace"
        required: false

  pm:
    display_name: "Product Manager"
    description: "Product planning and requirements"
    plugins:
      - name: "feature-dev"
        source: "feature-dev@asds-marketplace"
        required: true

defaults:
  scope: project
  auto_register_marketplace: true
```

- [ ] **Step 2: Write test for embedded defaults**

Create `internal/config/defaults_test.go`:

```go
package config_test

import (
 "testing"

 "github.com/andrew-le-mfv/asds-marketplace-setup/internal/config"
)

func TestDefaultMarketplaceConfig(t *testing.T) {
 cfg, err := config.DefaultMarketplaceConfig()
 if err != nil {
  t.Fatalf("unexpected error loading default config: %v", err)
 }

 if cfg.SchemaVersion != 1 {
  t.Errorf("schema_version = %d, want 1", cfg.SchemaVersion)
 }
 if cfg.Marketplace.Name != "asds-marketplace" {
  t.Errorf("marketplace.name = %q, want %q", cfg.Marketplace.Name, "asds-marketplace")
 }

 expectedRoles := []string{
  "backend", "data-engineer", "developer", "devops",
  "frontend", "pm", "security", "techlead", "tester",
 }
 names := cfg.RoleNames()
 if len(names) != len(expectedRoles) {
  t.Fatalf("role count = %d, want %d", len(names), len(expectedRoles))
 }
 for i, name := range names {
  if name != expectedRoles[i] {
   t.Errorf("role[%d] = %q, want %q", i, name, expectedRoles[i])
  }
 }
}
```

- [ ] **Step 3: Run test to verify it fails**

Run: `go test ./internal/config/ -run TestDefaultMarketplaceConfig -v`
Expected: FAIL — `DefaultMarketplaceConfig` not defined

- [ ] **Step 4: Implement embedded defaults**

Since `go:embed` cannot reference files outside the package directory using `..`, we use a `configs` package to embed the YAML and reference it from `internal/config`.

Create `configs/embed.go`:

```go
package configs

import (
 _ "embed"
)

//go:embed default-marketplace.yaml
var DefaultMarketplaceYAML []byte
```

Then create `internal/config/defaults.go`:

```go
package config

import (
 "github.com/andrew-le-mfv/asds-marketplace-setup/configs"
)

// DefaultMarketplaceConfig returns the embedded fallback marketplace configuration.
func DefaultMarketplaceConfig() (*MarketplaceConfig, error) {
 return ParseMarketplaceConfig(configs.DefaultMarketplaceYAML)
}
```

- [ ] **Step 5: Run test to verify it passes**

Run: `go test ./internal/config/ -run TestDefaultMarketplaceConfig -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add configs/ internal/config/defaults.go internal/config/defaults_test.go
git commit -m "feat: add embedded default marketplace config with go:embed"
```

---

### Task 6: Manifest File I/O (Read/Write to Disk)

**Files:**

- Modify: `internal/config/manifest.go` (add ReadManifest, WriteManifest)
- Modify: `internal/config/manifest_test.go` (add file I/O tests)

- [ ] **Step 1: Write test for manifest file I/O**

Append to `internal/config/manifest_test.go`:

```go
func TestManifest_WriteAndRead(t *testing.T) {
 dir := t.TempDir()
 path := filepath.Join(dir, ".asds-manifest.json")

 now := time.Date(2026, 3, 18, 10, 0, 0, 0, time.UTC)
 m := config.Manifest{
  SchemaVersion:      1,
  ASDSVersion:        "0.1.0",
  InstalledAt:        now,
  UpdatedAt:          now,
  Role:               "developer",
  Scope:              config.ScopeProject,
  MarketplaceSource:  "github.com/test/marketplace",
  InstallMethod:      "direct",
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
  ScaffoldedFiles:  []string{".claude/settings.json"},
 }

 if err := config.WriteManifest(path, &m); err != nil {
  t.Fatalf("WriteManifest error: %v", err)
 }

 loaded, err := config.ReadManifest(path)
 if err != nil {
  t.Fatalf("ReadManifest error: %v", err)
 }

 if loaded.Role != "developer" {
  t.Errorf("role = %q, want %q", loaded.Role, "developer")
 }
 if loaded.Scope != config.ScopeProject {
  t.Errorf("scope = %q, want %q", loaded.Scope, config.ScopeProject)
 }
 if len(loaded.Plugins) != 1 {
  t.Fatalf("plugins count = %d, want 1", len(loaded.Plugins))
 }
}

func TestReadManifest_NotFound(t *testing.T) {
 _, err := config.ReadManifest("/nonexistent/path/manifest.json")
 if err == nil {
  t.Error("expected error for missing file, got nil")
 }
}
```

Add import `"path/filepath"` to the test file's imports.

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/config/ -run "TestManifest_WriteAndRead|TestReadManifest_NotFound" -v`
Expected: FAIL — functions not defined

- [ ] **Step 3: Implement ReadManifest and WriteManifest**

Add to `internal/config/manifest.go`:

```go
import (
 "encoding/json"
 "fmt"
 "os"
 "path/filepath"
 "time"
)

// WriteManifest writes the manifest to disk as formatted JSON.
func WriteManifest(path string, m *Manifest) error {
 if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
  return fmt.Errorf("creating manifest directory: %w", err)
 }

 data, err := json.MarshalIndent(m, "", "  ")
 if err != nil {
  return fmt.Errorf("marshaling manifest: %w", err)
 }

 if err := os.WriteFile(path, data, 0o644); err != nil {
  return fmt.Errorf("writing manifest: %w", err)
 }
 return nil
}

// ReadManifest reads a manifest from disk.
func ReadManifest(path string) (*Manifest, error) {
 data, err := os.ReadFile(path)
 if err != nil {
  return nil, fmt.Errorf("reading manifest: %w", err)
 }

 var m Manifest
 if err := json.Unmarshal(data, &m); err != nil {
  return nil, fmt.Errorf("parsing manifest: %w", err)
 }
 return &m, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/config/ -run "TestManifest_WriteAndRead|TestReadManifest_NotFound" -v`
Expected: both PASS

- [ ] **Step 5: Commit**

```bash
git add internal/config/manifest.go internal/config/manifest_test.go
git commit -m "feat: add manifest file read/write operations"
```

---

### Task 7: ASDS Own Config (~/.config/asds/config.yaml)

**Files:**

- Create: `internal/config/asdsconfig.go`
- Create: `internal/config/asdsconfig_test.go`

- [ ] **Step 1: Write test for ASDS config**

Create `internal/config/asdsconfig_test.go`:

```go
package config_test

import (
 "os"
 "path/filepath"
 "testing"

 "github.com/andrew-le-mfv/asds-marketplace-setup/internal/config"
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/config/ -run "TestASDSConfig|TestResolveASDSConfig" -v`
Expected: FAIL — types not defined

- [ ] **Step 3: Implement ASDSConfig**

Create `internal/config/asdsconfig.go`:

```go
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
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/config/ -run "TestASDSConfig|TestResolveASDSConfig" -v`
Expected: all PASS

- [ ] **Step 5: Commit**

```bash
git add internal/config/asdsconfig.go internal/config/asdsconfig_test.go
git commit -m "feat: add ASDS own config read/write with defaults"
```

---

### Task 8: MarketplaceConfig Validation

**Files:**

- Modify: `internal/config/marketplace.go` (add Validate method)
- Modify: `internal/config/marketplace_test.go` (add validation tests)

- [ ] **Step 1: Write test for validation**

Append to `internal/config/marketplace_test.go`:

```go
func TestMarketplaceConfig_Validate(t *testing.T) {
 tests := []struct {
  name    string
  yaml    string
  wantErr bool
 }{
  {
   name: "valid config",
   yaml: `
schema_version: 1
marketplace:
  name: "test"
  version: "1.0.0"
  registry_url: "github.com/test"
roles:
  developer:
    display_name: "Developer"
    description: "Dev"
    plugins:
      - name: "plugin-a"
        source: "plugin-a@test"
defaults:
  scope: project
`,
   wantErr: false,
  },
  {
   name: "missing schema_version",
   yaml: `
marketplace:
  name: "test"
  version: "1.0.0"
  registry_url: "github.com/test"
roles:
  developer:
    display_name: "Developer"
    description: "Dev"
defaults:
  scope: project
`,
   wantErr: true,
  },
  {
   name: "no roles",
   yaml: `
schema_version: 1
marketplace:
  name: "test"
  version: "1.0.0"
  registry_url: "github.com/test"
defaults:
  scope: project
`,
   wantErr: true,
  },
  {
   name: "missing marketplace name",
   yaml: `
schema_version: 1
marketplace:
  version: "1.0.0"
  registry_url: "github.com/test"
roles:
  developer:
    display_name: "Developer"
    description: "Dev"
defaults:
  scope: project
`,
   wantErr: true,
  },
 }

 for _, tt := range tests {
  t.Run(tt.name, func(t *testing.T) {
   cfg, err := config.ParseMarketplaceConfig([]byte(tt.yaml))
   if err != nil {
    if !tt.wantErr {
     t.Fatalf("unexpected parse error: %v", err)
    }
    return
   }
   err = cfg.Validate()
   if tt.wantErr && err == nil {
    t.Error("expected validation error, got nil")
   }
   if !tt.wantErr && err != nil {
    t.Errorf("unexpected validation error: %v", err)
   }
  })
 }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/config/ -run TestMarketplaceConfig_Validate -v`
Expected: FAIL — `Validate` not defined

- [ ] **Step 3: Implement Validate**

Add to `internal/config/marketplace.go`:

```go
// Validate checks the MarketplaceConfig for required fields.
func (c *MarketplaceConfig) Validate() error {
 if c.SchemaVersion < 1 {
  return fmt.Errorf("schema_version must be >= 1, got %d", c.SchemaVersion)
 }
 if c.Marketplace.Name == "" {
  return fmt.Errorf("marketplace.name is required")
 }
 if len(c.Roles) == 0 {
  return fmt.Errorf("at least one role must be defined")
 }
 return nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/config/ -run TestMarketplaceConfig_Validate -v`
Expected: all sub-tests PASS

- [ ] **Step 5: Run all config tests**

Run: `go test ./internal/config/ -v`
Expected: all tests PASS

- [ ] **Step 6: Commit**

```bash
git add internal/config/marketplace.go internal/config/marketplace_test.go
git commit -m "feat: add marketplace config validation"
```
