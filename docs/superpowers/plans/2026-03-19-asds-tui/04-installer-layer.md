# Part 4: Installer Layer

**Dependencies:** Parts 2 and 3 (config types + claude settings operations)
**Estimated tasks:** 4

---

## Chunk 4: Installer Interface, Detector, DirectInstaller, CLIInstaller

### Task 13: Claude Code Detector

**Files:**

- Create: `internal/installer/detector.go`
- Create: `internal/installer/detector_test.go`

- [ ] **Step 1: Write tests for Claude Code detection**

Create `internal/installer/detector_test.go`:

```go
package installer_test

import (
 "testing"

 "github.com/your-org/asds-marketplace-setup/internal/installer"
)

func TestDetectClaudeCode(t *testing.T) {
 result := installer.DetectClaudeCode()

 // We can't guarantee Claude Code is installed in test env,
 // but we can verify the function returns a valid result.
 t.Logf("Claude Code detected: %v, path: %q", result.Found, result.Path)

 if result.Found && result.Path == "" {
  t.Error("Found=true but Path is empty")
 }
}

func TestDetectClaudeCode_WithCustomPath(t *testing.T) {
 // Test with a known-nonexistent path
 result := installer.DetectClaudeCodeAt("/nonexistent/claude")
 if result.Found {
  t.Error("expected Found=false for nonexistent path")
 }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/installer/ -run TestDetectClaudeCode -v`
Expected: FAIL — package doesn't exist

- [ ] **Step 3: Implement detector**

Create `internal/installer/detector.go`:

```go
package installer

import (
 "os/exec"
)

// DetectionResult holds the result of a Claude Code CLI detection.
type DetectionResult struct {
 Found bool
 Path  string
}

// DetectClaudeCode checks if the Claude Code CLI is available in PATH.
func DetectClaudeCode() DetectionResult {
 path, err := exec.LookPath("claude")
 if err != nil {
  return DetectionResult{Found: false}
 }
 return DetectionResult{Found: true, Path: path}
}

// DetectClaudeCodeAt checks if the Claude Code CLI exists at a specific path.
func DetectClaudeCodeAt(path string) DetectionResult {
 _, err := exec.LookPath(path)
 if err != nil {
  return DetectionResult{Found: false}
 }
 return DetectionResult{Found: true, Path: path}
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/installer/ -run TestDetectClaudeCode -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/installer/detector.go internal/installer/detector_test.go
git commit -m "feat: add Claude Code CLI detector"
```

---

### Task 14: Installer Interface (without factory — factory added after implementations)

**Files:**

- Create: `internal/installer/installer.go`

- [ ] **Step 1: Define the Installer interface and InstallResult (no factory yet)**

Create `internal/installer/installer.go`:

```go
package installer

import (
 "github.com/your-org/asds-marketplace-setup/internal/config"
)

// InstallResult holds the outcome of a plugin install/uninstall operation.
type InstallResult struct {
 PluginRef string
 Success   bool
 Error     error
}

// Installer abstracts over CLI and Direct installation methods.
type Installer interface {
 // Install enables the given plugins for the specified scope.
 Install(plugins []config.PluginRef, scope config.Scope, projectRoot string) ([]InstallResult, error)

 // Uninstall disables the given plugins for the specified scope.
 Uninstall(pluginRefs []string, scope config.Scope, projectRoot string) ([]InstallResult, error)

 // RegisterMarketplace registers a marketplace source.
 RegisterMarketplace(name string, registryURL string) error

 // Method returns "cli" or "direct" for manifest tracking.
 Method() string
}
```

- [ ] **Step 2: Verify it compiles**

Run: `go build ./internal/installer/`
Expected: compiles (interface + result type only, no references to unimplemented types)

- [ ] **Step 3: Commit**

```bash
git add internal/installer/installer.go
git commit -m "feat: add Installer interface"
```

---

### Task 15: DirectInstaller

**Files:**

- Create: `internal/installer/direct.go`
- Create: `internal/installer/direct_test.go`

- [ ] **Step 1: Write tests for DirectInstaller**

Create `internal/installer/direct_test.go`:

```go
package installer_test

import (
 "encoding/json"
 "os"
 "path/filepath"
 "testing"

 "github.com/your-org/asds-marketplace-setup/internal/config"
 "github.com/your-org/asds-marketplace-setup/internal/installer"
)

func TestDirectInstaller_Install(t *testing.T) {
 projectRoot := t.TempDir()
 claudeDir := filepath.Join(projectRoot, ".claude")
 os.MkdirAll(claudeDir, 0o755)

 inst := &installer.DirectInstaller{}

 plugins := []config.PluginRef{
  {Name: "code-reviewer", Source: "code-reviewer@asds", Required: true},
  {Name: "commit-commands", Source: "commit-commands@asds", Required: false},
 }

 results, err := inst.Install(plugins, config.ScopeProject, projectRoot)
 if err != nil {
  t.Fatalf("Install error: %v", err)
 }

 if len(results) != 2 {
  t.Fatalf("results count = %d, want 2", len(results))
 }

 for _, r := range results {
  if !r.Success {
   t.Errorf("plugin %q failed: %v", r.PluginRef, r.Error)
  }
 }

 // Verify settings file was written
 settingsPath := filepath.Join(claudeDir, "settings.json")
 data, err := os.ReadFile(settingsPath)
 if err != nil {
  t.Fatalf("settings file not created: %v", err)
 }

 var settings map[string]any
 json.Unmarshal(data, &settings)

 ep, ok := settings["enabledPlugins"].(map[string]any)
 if !ok {
  t.Fatal("enabledPlugins not found")
 }
 if ep["code-reviewer@asds"] != true {
  t.Error("code-reviewer not enabled")
 }
}

func TestDirectInstaller_Install_PreservesExisting(t *testing.T) {
 projectRoot := t.TempDir()
 claudeDir := filepath.Join(projectRoot, ".claude")
 os.MkdirAll(claudeDir, 0o755)

 // Write existing settings
 existing := map[string]any{
  "enabledPlugins": map[string]any{
   "other-plugin@somewhere": true,
  },
  "customSetting": "preserve",
 }
 data, _ := json.MarshalIndent(existing, "", "  ")
 os.WriteFile(filepath.Join(claudeDir, "settings.json"), data, 0o644)

 inst := &installer.DirectInstaller{}
 plugins := []config.PluginRef{
  {Name: "code-reviewer", Source: "code-reviewer@asds", Required: true},
 }

 _, err := inst.Install(plugins, config.ScopeProject, projectRoot)
 if err != nil {
  t.Fatalf("Install error: %v", err)
 }

 data, _ = os.ReadFile(filepath.Join(claudeDir, "settings.json"))
 var settings map[string]any
 json.Unmarshal(data, &settings)

 if settings["customSetting"] != "preserve" {
  t.Error("existing setting was overwritten")
 }

 ep := settings["enabledPlugins"].(map[string]any)
 if ep["other-plugin@somewhere"] != true {
  t.Error("existing plugin was removed")
 }
 if ep["code-reviewer@asds"] != true {
  t.Error("new plugin not added")
 }
}

func TestDirectInstaller_Uninstall(t *testing.T) {
 projectRoot := t.TempDir()
 claudeDir := filepath.Join(projectRoot, ".claude")
 os.MkdirAll(claudeDir, 0o755)

 // Setup: install first
 settings := map[string]any{
  "enabledPlugins": map[string]any{
   "code-reviewer@asds":    true,
   "commit-commands@asds":  true,
   "other-plugin@other":    true,
  },
 }
 data, _ := json.MarshalIndent(settings, "", "  ")
 os.WriteFile(filepath.Join(claudeDir, "settings.json"), data, 0o644)

 inst := &installer.DirectInstaller{}
 refs := []string{"code-reviewer@asds", "commit-commands@asds"}

 results, err := inst.Uninstall(refs, config.ScopeProject, projectRoot)
 if err != nil {
  t.Fatalf("Uninstall error: %v", err)
 }

 for _, r := range results {
  if !r.Success {
   t.Errorf("uninstall %q failed: %v", r.PluginRef, r.Error)
  }
 }

 data, _ = os.ReadFile(filepath.Join(claudeDir, "settings.json"))
 json.Unmarshal(data, &settings)
 ep := settings["enabledPlugins"].(map[string]any)

 if _, exists := ep["code-reviewer@asds"]; exists {
  t.Error("code-reviewer should be removed")
 }
 if ep["other-plugin@other"] != true {
  t.Error("unrelated plugin should be preserved")
 }
}

func TestDirectInstaller_RegisterMarketplace(t *testing.T) {
 // This writes to user-level settings, so we mock with temp HOME
 origHome := os.Getenv("HOME")
 tmpHome := t.TempDir()
 os.Setenv("HOME", tmpHome)
 defer os.Setenv("HOME", origHome)

 claudeDir := filepath.Join(tmpHome, ".claude")
 os.MkdirAll(claudeDir, 0o755)

 inst := &installer.DirectInstaller{}
 err := inst.RegisterMarketplace("asds-marketplace", "github.com/your-org/asds-marketplace")
 if err != nil {
  t.Fatalf("RegisterMarketplace error: %v", err)
 }

 data, _ := os.ReadFile(filepath.Join(claudeDir, "settings.json"))
 var settings map[string]any
 json.Unmarshal(data, &settings)

 ekm, ok := settings["extraKnownMarketplaces"].(map[string]any)
 if !ok {
  t.Fatal("extraKnownMarketplaces not found")
 }
 if _, ok := ekm["asds-marketplace"]; !ok {
  t.Error("marketplace not registered")
 }
}

func TestDirectInstaller_Method(t *testing.T) {
 inst := &installer.DirectInstaller{}
 if inst.Method() != "direct" {
  t.Errorf("Method() = %q, want %q", inst.Method(), "direct")
 }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/installer/ -run "TestDirectInstaller" -v`
Expected: FAIL — `DirectInstaller` type not fully defined

- [ ] **Step 3: Implement DirectInstaller**

Create `internal/installer/direct.go`:

```go
package installer

import (
 "github.com/your-org/asds-marketplace-setup/internal/claude"
 "github.com/your-org/asds-marketplace-setup/internal/config"
)

// DirectInstaller writes enabledPlugins directly to Claude settings JSON files.
// This is the config-only reconciliation mode — it declares intent; Claude Code
// will resolve and fetch actual plugin assets when it starts.
type DirectInstaller struct{}

// Install enables plugins by writing to the scope-appropriate settings file.
func (d *DirectInstaller) Install(plugins []config.PluginRef, scope config.Scope, projectRoot string) ([]InstallResult, error) {
 settingsPath := claude.SettingsPath(scope, projectRoot)

 settings, err := claude.ReadSettings(settingsPath)
 if err != nil {
  return nil, err
 }

 pluginMap := make(map[string]bool, len(plugins))
 for _, p := range plugins {
  pluginMap[p.Source] = true
 }

 claude.MergeEnabledPlugins(settings, pluginMap)

 if err := claude.WriteSettings(settingsPath, settings); err != nil {
  return nil, err
 }

 results := make([]InstallResult, len(plugins))
 for i, p := range plugins {
  results[i] = InstallResult{
   PluginRef: p.Source,
   Success:   true,
  }
 }
 return results, nil
}

// Uninstall removes plugins from the scope-appropriate settings file.
func (d *DirectInstaller) Uninstall(pluginRefs []string, scope config.Scope, projectRoot string) ([]InstallResult, error) {
 settingsPath := claude.SettingsPath(scope, projectRoot)

 settings, err := claude.ReadSettings(settingsPath)
 if err != nil {
  return nil, err
 }

 claude.DisablePlugins(settings, pluginRefs)

 if err := claude.WriteSettings(settingsPath, settings); err != nil {
  return nil, err
 }

 results := make([]InstallResult, len(pluginRefs))
 for i, ref := range pluginRefs {
  results[i] = InstallResult{
   PluginRef: ref,
   Success:   true,
  }
 }
 return results, nil
}

// RegisterMarketplace writes marketplace registration to user-level settings.
func (d *DirectInstaller) RegisterMarketplace(name string, registryURL string) error {
 settingsPath := claude.MarketplaceRegistrationPath()

 settings, err := claude.ReadSettings(settingsPath)
 if err != nil {
  return err
 }

 claude.MergeMarketplaceRegistration(settings, name, registryURL)

 return claude.WriteSettings(settingsPath, settings)
}

// Method returns "direct" for manifest tracking.
func (d *DirectInstaller) Method() string {
 return "direct"
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/installer/ -run "TestDirectInstaller" -v`
Expected: all tests PASS

- [ ] **Step 5: Commit**

```bash
git add internal/installer/direct.go internal/installer/direct_test.go
git commit -m "feat: add DirectInstaller (config-only JSON file manipulation)"
```

---

### Task 16: CLIInstaller

**Files:**

- Create: `internal/installer/cli.go`
- Create: `internal/installer/cli_test.go`

- [ ] **Step 1: Write tests for CLIInstaller**

Create `internal/installer/cli_test.go`:

```go
package installer_test

import (
 "testing"

 "github.com/your-org/asds-marketplace-setup/internal/installer"
)

func TestCLIInstaller_Method(t *testing.T) {
 inst := installer.NewCLIInstaller("/usr/bin/claude")
 if inst.Method() != "cli" {
  t.Errorf("Method() = %q, want %q", inst.Method(), "cli")
 }
}

func TestCLIInstaller_BuildArgs(t *testing.T) {
 inst := installer.NewCLIInstaller("/usr/bin/claude")

 tests := []struct {
  name string
  fn   func() []string
  want []string
 }{
  {
   "install args",
   func() []string { return inst.BuildInstallArgs("code-reviewer@asds-marketplace", "project") },
   []string{"plugin", "install", "code-reviewer@asds-marketplace", "--scope", "project"},
  },
  {
   "uninstall args",
   func() []string { return inst.BuildUninstallArgs("code-reviewer@asds-marketplace", "project") },
   []string{"plugin", "uninstall", "code-reviewer@asds-marketplace", "--scope", "project"},
  },
  {
   "marketplace add args",
   func() []string { return inst.BuildMarketplaceAddArgs("github.com/your-org/asds-marketplace") },
   []string{"plugin", "marketplace", "add", "github.com/your-org/asds-marketplace"},
  },
 }
 for _, tt := range tests {
  t.Run(tt.name, func(t *testing.T) {
   got := tt.fn()
   if len(got) != len(tt.want) {
    t.Fatalf("args length = %d, want %d", len(got), len(tt.want))
   }
   for i, a := range got {
    if a != tt.want[i] {
     t.Errorf("args[%d] = %q, want %q", i, a, tt.want[i])
    }
   }
  })
 }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/installer/ -run "TestCLIInstaller" -v`
Expected: FAIL — `CLIInstaller` / `NewCLIInstaller` not defined

- [ ] **Step 3: Implement CLIInstaller**

Create `internal/installer/cli.go`:

```go
package installer

import (
 "fmt"
 "os/exec"

 "github.com/your-org/asds-marketplace-setup/internal/config"
)

// CLIInstaller shells out to the Claude Code CLI for plugin management.
type CLIInstaller struct {
 claudePath string
}

// NewCLIInstaller creates a CLIInstaller using the given claude binary path.
func NewCLIInstaller(claudePath string) *CLIInstaller {
 return &CLIInstaller{claudePath: claudePath}
}

// BuildInstallArgs returns the claude CLI args for installing a plugin.
func (c *CLIInstaller) BuildInstallArgs(pluginRef string, scope string) []string {
 return []string{"plugin", "install", pluginRef, "--scope", scope}
}

// BuildUninstallArgs returns the claude CLI args for uninstalling a plugin.
func (c *CLIInstaller) BuildUninstallArgs(pluginRef string, scope string) []string {
 return []string{"plugin", "uninstall", pluginRef, "--scope", scope}
}

// BuildMarketplaceAddArgs returns the claude CLI args for adding a marketplace.
func (c *CLIInstaller) BuildMarketplaceAddArgs(source string) []string {
 return []string{"plugin", "marketplace", "add", source}
}

// Install enables plugins via the Claude Code CLI.
func (c *CLIInstaller) Install(plugins []config.PluginRef, scope config.Scope, projectRoot string) ([]InstallResult, error) {
 results := make([]InstallResult, 0, len(plugins))

 for _, p := range plugins {
  args := c.BuildInstallArgs(p.Source, string(scope))
  cmd := exec.Command(c.claudePath, args...)
  cmd.Dir = projectRoot

  if err := cmd.Run(); err != nil {
   results = append(results, InstallResult{
    PluginRef: p.Source,
    Success:   false,
    Error:     fmt.Errorf("claude plugin install %s: %w", p.Source, err),
   })
  } else {
   results = append(results, InstallResult{
    PluginRef: p.Source,
    Success:   true,
   })
  }
 }

 return results, nil
}

// Uninstall removes plugins via the Claude Code CLI.
func (c *CLIInstaller) Uninstall(pluginRefs []string, scope config.Scope, projectRoot string) ([]InstallResult, error) {
 results := make([]InstallResult, 0, len(pluginRefs))

 for _, ref := range pluginRefs {
  args := c.BuildUninstallArgs(ref, string(scope))
  cmd := exec.Command(c.claudePath, args...)
  cmd.Dir = projectRoot

  if err := cmd.Run(); err != nil {
   results = append(results, InstallResult{
    PluginRef: ref,
    Success:   false,
    Error:     fmt.Errorf("claude plugin uninstall %s: %w", ref, err),
   })
  } else {
   results = append(results, InstallResult{
    PluginRef: ref,
    Success:   true,
   })
  }
 }

 return results, nil
}

// RegisterMarketplace adds a marketplace via the Claude Code CLI.
func (c *CLIInstaller) RegisterMarketplace(name string, registryURL string) error {
 args := c.BuildMarketplaceAddArgs(registryURL)
 cmd := exec.Command(c.claudePath, args...)
 if err := cmd.Run(); err != nil {
  return fmt.Errorf("claude plugin marketplace add: %w", err)
 }
 return nil
}

// Method returns "cli" for manifest tracking.
func (c *CLIInstaller) Method() string {
 return "cli"
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/installer/ -run "TestCLIInstaller" -v`
Expected: all tests PASS

- [ ] **Step 5: Verify the whole installer package compiles and tests pass**

Run: `go test ./internal/installer/ -v`
Expected: all tests PASS

Run: `go build ./internal/installer/`
Expected: compiles without errors

- [ ] **Step 6: Commit**

```bash
git add internal/installer/cli.go internal/installer/cli_test.go
git commit -m "feat: add CLIInstaller (shells out to claude CLI)"
```

- [ ] **Step 7: Add NewInstaller factory to installer.go**

Now that both DirectInstaller and CLIInstaller exist, add the factory to `internal/installer/installer.go`:

```go
// NewInstaller creates the appropriate installer based on Claude Code availability.
// If preferCLI is true and Claude Code is detected, returns CLIInstaller.
// Otherwise returns DirectInstaller.
func NewInstaller(preferCLI bool) Installer {
 if preferCLI {
  detection := DetectClaudeCode()
  if detection.Found {
   return &CLIInstaller{claudePath: detection.Path}
  }
 }
 return &DirectInstaller{}
}
```

- [ ] **Step 8: Verify it compiles**

Run: `go build ./internal/installer/`
Expected: compiles without errors

- [ ] **Step 9: Commit**

```bash
git add internal/installer/installer.go
git commit -m "feat: add NewInstaller factory function"
```
