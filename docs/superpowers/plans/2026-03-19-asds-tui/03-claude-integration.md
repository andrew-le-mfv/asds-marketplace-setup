# Part 3: Claude Integration

**Dependencies:** Part 1 (core types must exist)
**Can run in parallel with:** Part 2
**Estimated tasks:** 4

---

## Chunk 3: Claude Settings, Path Resolution, CLAUDE.md Marker Blocks

### Task 9: Scope Path Resolution

**Files:**

- Create: `internal/claude/paths.go`
- Create: `internal/claude/paths_test.go`

- [ ] **Step 1: Write tests for path resolution**

Create `internal/claude/paths_test.go`:

```go
package claude_test

import (
 "os"
 "path/filepath"
 "testing"

 "github.com/your-org/asds-marketplace-setup/internal/claude"
 "github.com/your-org/asds-marketplace-setup/internal/config"
)

func TestSettingsPath(t *testing.T) {
 projectRoot := "/tmp/test-project"
 home, _ := os.UserHomeDir()

 tests := []struct {
  scope config.Scope
  want  string
 }{
  {config.ScopeUser, filepath.Join(home, ".claude", "settings.json")},
  {config.ScopeProject, filepath.Join(projectRoot, ".claude", "settings.json")},
  {config.ScopeLocal, filepath.Join(projectRoot, ".claude", "settings.local.json")},
 }

 for _, tt := range tests {
  t.Run(string(tt.scope), func(t *testing.T) {
   got := claude.SettingsPath(tt.scope, projectRoot)
   if got != tt.want {
    t.Errorf("SettingsPath(%q) = %q, want %q", tt.scope, got, tt.want)
   }
  })
 }
}

func TestManifestPath(t *testing.T) {
 projectRoot := "/tmp/test-project"
 home, _ := os.UserHomeDir()

 tests := []struct {
  scope config.Scope
  want  string
 }{
  {config.ScopeUser, filepath.Join(home, ".claude", ".asds-manifest.json")},
  {config.ScopeProject, filepath.Join(projectRoot, ".claude", ".asds-manifest.json")},
  {config.ScopeLocal, filepath.Join(projectRoot, ".claude", ".asds-manifest.local.json")},
 }

 for _, tt := range tests {
  t.Run(string(tt.scope), func(t *testing.T) {
   got := claude.ManifestPath(tt.scope, projectRoot)
   if got != tt.want {
    t.Errorf("ManifestPath(%q) = %q, want %q", tt.scope, got, tt.want)
   }
  })
 }
}

func TestClaudeMDPath(t *testing.T) {
 projectRoot := "/tmp/test-project"

 tests := []struct {
  scope   config.Scope
  want    string
  wantErr bool
 }{
  {config.ScopeProject, filepath.Join(projectRoot, "CLAUDE.md"), false},
  {config.ScopeLocal, filepath.Join(projectRoot, "CLAUDE.md"), false},
  {config.ScopeUser, "", true},
 }

 for _, tt := range tests {
  t.Run(string(tt.scope), func(t *testing.T) {
   got, err := claude.ClaudeMDPath(tt.scope, projectRoot)
   if tt.wantErr {
    if err == nil {
     t.Error("expected error, got nil")
    }
    return
   }
   if err != nil {
    t.Fatalf("unexpected error: %v", err)
   }
   if got != tt.want {
    t.Errorf("ClaudeMDPath(%q) = %q, want %q", tt.scope, got, tt.want)
   }
  })
 }
}

func TestMarketplaceRegistrationPath(t *testing.T) {
 home, _ := os.UserHomeDir()
 want := filepath.Join(home, ".claude", "settings.json")
 got := claude.MarketplaceRegistrationPath()
 if got != want {
  t.Errorf("MarketplaceRegistrationPath() = %q, want %q", got, want)
 }
}

func TestFindProjectRoot(t *testing.T) {
 // Create a temp dir with .git
 dir := t.TempDir()
 gitDir := filepath.Join(dir, ".git")
 if err := os.Mkdir(gitDir, 0o755); err != nil {
  t.Fatal(err)
 }

 subDir := filepath.Join(dir, "src", "pkg")
 if err := os.MkdirAll(subDir, 0o755); err != nil {
  t.Fatal(err)
 }

 got, err := claude.FindProjectRoot(subDir)
 if err != nil {
  t.Fatalf("unexpected error: %v", err)
 }
 if got != dir {
  t.Errorf("FindProjectRoot() = %q, want %q", got, dir)
 }
}

func TestFindProjectRoot_NoGit(t *testing.T) {
 dir := t.TempDir()
 got, err := claude.FindProjectRoot(dir)
 if err != nil {
  t.Fatalf("unexpected error: %v", err)
 }
 // Falls back to cwd (the dir itself)
 if got != dir {
  t.Errorf("FindProjectRoot() = %q, want %q (fallback)", got, dir)
 }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/claude/ -v`
Expected: FAIL — package doesn't exist

- [ ] **Step 3: Implement path resolution**

Create `internal/claude/paths.go`:

```go
package claude

import (
 "fmt"
 "os"
 "path/filepath"

 "github.com/your-org/asds-marketplace-setup/internal/config"
)

// SettingsPath returns the Claude settings file path for the given scope.
func SettingsPath(scope config.Scope, projectRoot string) string {
 switch scope {
 case config.ScopeUser:
  return filepath.Join(userClaudeDir(), "settings.json")
 case config.ScopeProject:
  return filepath.Join(projectRoot, ".claude", "settings.json")
 case config.ScopeLocal:
  return filepath.Join(projectRoot, ".claude", "settings.local.json")
 default:
  return ""
 }
}

// ManifestPath returns the ASDS manifest file path for the given scope.
func ManifestPath(scope config.Scope, projectRoot string) string {
 switch scope {
 case config.ScopeUser:
  return filepath.Join(userClaudeDir(), ".asds-manifest.json")
 case config.ScopeProject:
  return filepath.Join(projectRoot, ".claude", ".asds-manifest.json")
 case config.ScopeLocal:
  return filepath.Join(projectRoot, ".claude", ".asds-manifest.local.json")
 default:
  return ""
 }
}

// ClaudeMDPath returns the CLAUDE.md path for project/local scopes.
// Returns an error for user scope (CLAUDE.md is project-only).
func ClaudeMDPath(scope config.Scope, projectRoot string) (string, error) {
 if scope == config.ScopeUser {
  return "", fmt.Errorf("CLAUDE.md is not applicable for user scope")
 }
 return filepath.Join(projectRoot, "CLAUDE.md"), nil
}

// MarketplaceRegistrationPath returns the path where marketplace registration
// is always written — user-level settings regardless of plugin scope.
func MarketplaceRegistrationPath() string {
 return filepath.Join(userClaudeDir(), "settings.json")
}

// FindProjectRoot walks up from startDir looking for .git/.
// Falls back to startDir if no git root is found.
func FindProjectRoot(startDir string) (string, error) {
 dir, err := filepath.Abs(startDir)
 if err != nil {
  return "", fmt.Errorf("resolving absolute path: %w", err)
 }

 for {
  gitPath := filepath.Join(dir, ".git")
  if info, err := os.Stat(gitPath); err == nil && info.IsDir() {
   return dir, nil
  }

  parent := filepath.Dir(dir)
  if parent == dir {
   // Reached filesystem root, fall back to startDir
   return filepath.Abs(startDir)
  }
  dir = parent
 }
}

func userClaudeDir() string {
 home, err := os.UserHomeDir()
 if err != nil {
  home = "~"
 }
 return filepath.Join(home, ".claude")
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/claude/ -v`
Expected: all tests PASS

- [ ] **Step 5: Commit**

```bash
git add internal/claude/paths.go internal/claude/paths_test.go
git commit -m "feat: add Claude scope path resolution and project root finder"
```

---

### Task 10: Claude Settings JSON Read/Write/Merge

**Files:**

- Create: `internal/claude/settings.go`
- Create: `internal/claude/settings_test.go`

- [ ] **Step 1: Write tests for settings merge**

Create `internal/claude/settings_test.go`:

```go
package claude_test

import (
 "encoding/json"
 "os"
 "path/filepath"
 "testing"

 "github.com/your-org/asds-marketplace-setup/internal/claude"
)

func TestReadSettings_NewFile(t *testing.T) {
 path := filepath.Join(t.TempDir(), "settings.json")
 settings, err := claude.ReadSettings(path)
 if err != nil {
  t.Fatalf("unexpected error: %v", err)
 }
 if settings == nil {
  t.Fatal("expected empty settings map, got nil")
 }
}

func TestReadSettings_ExistingFile(t *testing.T) {
 dir := t.TempDir()
 path := filepath.Join(dir, "settings.json")

 existing := map[string]any{
  "enabledPlugins": map[string]any{
   "existing-plugin@other": true,
  },
  "customKey": "preserve-me",
 }
 data, _ := json.MarshalIndent(existing, "", "  ")
 os.WriteFile(path, data, 0o644)

 settings, err := claude.ReadSettings(path)
 if err != nil {
  t.Fatalf("unexpected error: %v", err)
 }

 // Verify existing keys are preserved
 if settings["customKey"] != "preserve-me" {
  t.Errorf("customKey not preserved")
 }
}

func TestMergeEnabledPlugins(t *testing.T) {
 settings := map[string]any{
  "enabledPlugins": map[string]any{
   "existing@other": true,
  },
  "customKey": "keep",
 }

 plugins := map[string]bool{
  "code-reviewer@asds":    true,
  "commit-commands@asds":  true,
 }

 claude.MergeEnabledPlugins(settings, plugins)

 ep, ok := settings["enabledPlugins"].(map[string]any)
 if !ok {
  t.Fatal("enabledPlugins not a map")
 }

 // Existing plugin preserved
 if ep["existing@other"] != true {
  t.Error("existing plugin was removed")
 }
 // New plugins added
 if ep["code-reviewer@asds"] != true {
  t.Error("code-reviewer not added")
 }
 if ep["commit-commands@asds"] != true {
  t.Error("commit-commands not added")
 }
 // Unrelated key preserved
 if settings["customKey"] != "keep" {
  t.Error("customKey was removed")
 }
}

func TestMergeMarketplaceRegistration(t *testing.T) {
 settings := map[string]any{}

 claude.MergeMarketplaceRegistration(settings, "asds-marketplace", "github.com/your-org/asds-marketplace")

 ekm, ok := settings["extraKnownMarketplaces"].(map[string]any)
 if !ok {
  t.Fatal("extraKnownMarketplaces not a map")
 }

 entry, ok := ekm["asds-marketplace"].(map[string]any)
 if !ok {
  t.Fatal("marketplace entry not a map")
 }

 source, ok := entry["source"].(map[string]any)
 if !ok {
  t.Fatal("source not a map")
 }

 if source["source"] != "github" {
  t.Errorf("source.source = %v, want 'github'", source["source"])
 }
}

func TestWriteSettings(t *testing.T) {
 dir := t.TempDir()
 path := filepath.Join(dir, ".claude", "settings.json")

 settings := map[string]any{
  "enabledPlugins": map[string]any{
   "test@asds": true,
  },
 }

 if err := claude.WriteSettings(path, settings); err != nil {
  t.Fatalf("WriteSettings error: %v", err)
 }

 // Verify file exists and is valid JSON
 data, err := os.ReadFile(path)
 if err != nil {
  t.Fatalf("file not created: %v", err)
 }

 var loaded map[string]any
 if err := json.Unmarshal(data, &loaded); err != nil {
  t.Fatalf("invalid JSON: %v", err)
 }
}

func TestDisablePlugins(t *testing.T) {
 settings := map[string]any{
  "enabledPlugins": map[string]any{
   "code-reviewer@asds":   true,
   "commit-commands@asds": true,
   "existing@other":       true,
  },
 }

 pluginsToDisable := []string{"code-reviewer@asds", "commit-commands@asds"}
 claude.DisablePlugins(settings, pluginsToDisable)

 ep := settings["enabledPlugins"].(map[string]any)

 // ASDS plugins removed
 if _, exists := ep["code-reviewer@asds"]; exists {
  t.Error("code-reviewer should be removed")
 }
 if _, exists := ep["commit-commands@asds"]; exists {
  t.Error("commit-commands should be removed")
 }
 // Non-ASDS plugin preserved
 if ep["existing@other"] != true {
  t.Error("existing plugin should be preserved")
 }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/claude/ -run "TestReadSettings|TestMerge|TestWriteSettings|TestDisable" -v`
Expected: FAIL — functions not defined

- [ ] **Step 3: Implement settings operations**

Create `internal/claude/settings.go`:

```go
package claude

import (
 "encoding/json"
 "fmt"
 "os"
 "path/filepath"
)

// ReadSettings reads a Claude settings JSON file.
// Returns an empty map if the file doesn't exist.
func ReadSettings(path string) (map[string]any, error) {
 data, err := os.ReadFile(path)
 if err != nil {
  if os.IsNotExist(err) {
   return make(map[string]any), nil
  }
  return nil, fmt.Errorf("reading settings: %w", err)
 }

 var settings map[string]any
 if err := json.Unmarshal(data, &settings); err != nil {
  return nil, fmt.Errorf("parsing settings: %w", err)
 }
 return settings, nil
}

// WriteSettings writes a Claude settings map to disk as formatted JSON.
func WriteSettings(path string, settings map[string]any) error {
 if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
  return fmt.Errorf("creating settings directory: %w", err)
 }

 data, err := json.MarshalIndent(settings, "", "  ")
 if err != nil {
  return fmt.Errorf("marshaling settings: %w", err)
 }

 if err := os.WriteFile(path, data, 0o644); err != nil {
  return fmt.Errorf("writing settings: %w", err)
 }
 return nil
}

// MergeEnabledPlugins merges the given plugins into the enabledPlugins map.
// Existing plugins are preserved; new plugins are added.
func MergeEnabledPlugins(settings map[string]any, plugins map[string]bool) {
 ep, ok := settings["enabledPlugins"].(map[string]any)
 if !ok {
  ep = make(map[string]any)
 }

 for name, enabled := range plugins {
  ep[name] = enabled
 }
 settings["enabledPlugins"] = ep
}

// DisablePlugins removes the specified plugin keys from enabledPlugins.
func DisablePlugins(settings map[string]any, pluginRefs []string) {
 ep, ok := settings["enabledPlugins"].(map[string]any)
 if !ok {
  return
 }

 for _, ref := range pluginRefs {
  delete(ep, ref)
 }
}

// MergeMarketplaceRegistration adds a marketplace entry to extraKnownMarketplaces.
func MergeMarketplaceRegistration(settings map[string]any, marketplaceName, registryURL string) {
 ekm, ok := settings["extraKnownMarketplaces"].(map[string]any)
 if !ok {
  ekm = make(map[string]any)
 }

 ekm[marketplaceName] = map[string]any{
  "source": map[string]any{
   "source": "github",
   "repo":   registryURL,
  },
 }
 settings["extraKnownMarketplaces"] = ekm
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/claude/ -run "TestReadSettings|TestMerge|TestWriteSettings|TestDisable" -v`
Expected: all tests PASS

- [ ] **Step 5: Commit**

```bash
git add internal/claude/settings.go internal/claude/settings_test.go
git commit -m "feat: add Claude settings JSON read/write/merge operations"
```

---

### Task 11: CLAUDE.md Marker Block Management

**Files:**

- Create: `internal/claude/claudemd.go`
- Create: `internal/claude/claudemd_test.go`

- [ ] **Step 1: Write tests for marker block operations**

Create `internal/claude/claudemd_test.go`:

```go
package claude_test

import (
 "os"
 "path/filepath"
 "strings"
 "testing"

 "github.com/your-org/asds-marketplace-setup/internal/claude"
)

func TestUpsertMarkerBlock_NewFile(t *testing.T) {
 dir := t.TempDir()
 path := filepath.Join(dir, "CLAUDE.md")

 snippets := []string{"Follow conventional commits", "Always write tests"}
 err := claude.UpsertMarkerBlock(path, "developer", snippets)
 if err != nil {
  t.Fatalf("unexpected error: %v", err)
 }

 data, _ := os.ReadFile(path)
 content := string(data)

 if !strings.Contains(content, "<!-- ASDS:BEGIN role=developer -->") {
  t.Error("missing begin marker")
 }
 if !strings.Contains(content, "<!-- ASDS:END -->") {
  t.Error("missing end marker")
 }
 if !strings.Contains(content, "- Follow conventional commits") {
  t.Error("missing snippet")
 }
}

func TestUpsertMarkerBlock_ExistingContent(t *testing.T) {
 dir := t.TempDir()
 path := filepath.Join(dir, "CLAUDE.md")

 // Write existing content
 os.WriteFile(path, []byte("# Project\n\nExisting instructions.\n"), 0o644)

 snippets := []string{"New snippet"}
 err := claude.UpsertMarkerBlock(path, "developer", snippets)
 if err != nil {
  t.Fatalf("unexpected error: %v", err)
 }

 data, _ := os.ReadFile(path)
 content := string(data)

 if !strings.Contains(content, "# Project") {
  t.Error("existing content was removed")
 }
 if !strings.Contains(content, "Existing instructions.") {
  t.Error("existing content was removed")
 }
 if !strings.Contains(content, "<!-- ASDS:BEGIN role=developer -->") {
  t.Error("marker block not added")
 }
}

func TestUpsertMarkerBlock_ReplaceExisting(t *testing.T) {
 dir := t.TempDir()
 path := filepath.Join(dir, "CLAUDE.md")

 // Write with existing ASDS block
 existing := `# Project

<!-- ASDS:BEGIN role=frontend -->
- Old snippet
<!-- ASDS:END -->

Other content.
`
 os.WriteFile(path, []byte(existing), 0o644)

 snippets := []string{"New developer snippet"}
 err := claude.UpsertMarkerBlock(path, "developer", snippets)
 if err != nil {
  t.Fatalf("unexpected error: %v", err)
 }

 data, _ := os.ReadFile(path)
 content := string(data)

 if strings.Contains(content, "Old snippet") {
  t.Error("old snippet should be replaced")
 }
 if strings.Contains(content, "role=frontend") {
  t.Error("old role marker should be replaced")
 }
 if !strings.Contains(content, "role=developer") {
  t.Error("new role marker not present")
 }
 if !strings.Contains(content, "New developer snippet") {
  t.Error("new snippet not present")
 }
 if !strings.Contains(content, "Other content.") {
  t.Error("surrounding content was removed")
 }
}

func TestRemoveMarkerBlock(t *testing.T) {
 dir := t.TempDir()
 path := filepath.Join(dir, "CLAUDE.md")

 content := `# Project

Some text.

<!-- ASDS:BEGIN role=developer -->
- Follow conventional commits
<!-- ASDS:END -->

More text.
`
 os.WriteFile(path, []byte(content), 0o644)

 err := claude.RemoveMarkerBlock(path)
 if err != nil {
  t.Fatalf("unexpected error: %v", err)
 }

 data, _ := os.ReadFile(path)
 result := string(data)

 if strings.Contains(result, "ASDS:BEGIN") {
  t.Error("marker block not removed")
 }
 if strings.Contains(result, "conventional commits") {
  t.Error("snippet not removed")
 }
 if !strings.Contains(result, "Some text.") {
  t.Error("surrounding content was removed")
 }
 if !strings.Contains(result, "More text.") {
  t.Error("surrounding content was removed")
 }
}

func TestRemoveMarkerBlock_NoBlock(t *testing.T) {
 dir := t.TempDir()
 path := filepath.Join(dir, "CLAUDE.md")

 os.WriteFile(path, []byte("# Project\nNo ASDS block here.\n"), 0o644)

 err := claude.RemoveMarkerBlock(path)
 if err != nil {
  t.Fatalf("unexpected error: %v", err)
 }

 data, _ := os.ReadFile(path)
 if string(data) != "# Project\nNo ASDS block here.\n" {
  t.Error("content was modified despite no marker block")
 }
}

func TestHasMarkerBlock(t *testing.T) {
 tests := []struct {
  name    string
  content string
  want    bool
 }{
  {
   "with markers",
   "# Project\n\n<!-- ASDS:BEGIN role=developer -->\n- snippet\n<!-- ASDS:END -->\n",
   true,
  },
  {
   "without markers",
   "# No markers here",
   false,
  },
 }
 for _, tt := range tests {
  t.Run(tt.name, func(t *testing.T) {
   got := claude.HasMarkerBlock(tt.content)
   if got != tt.want {
    t.Errorf("HasMarkerBlock() = %v, want %v", got, tt.want)
   }
  })
 }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/claude/ -run "TestUpsert|TestRemove|TestHas" -v`
Expected: FAIL — functions not defined

- [ ] **Step 3: Implement CLAUDE.md marker block operations**

Create `internal/claude/claudemd.go`:

```go
package claude

import (
 "fmt"
 "os"
 "path/filepath"
 "regexp"
 "strings"
)

const (
 markerBeginPrefix = "<!-- ASDS:BEGIN"
 markerEnd         = "<!-- ASDS:END -->"
)

var markerBlockRegex = regexp.MustCompile(`(?s)<!-- ASDS:BEGIN[^\n]*-->\n.*?<!-- ASDS:END -->\n?`)

// HasMarkerBlock returns true if the content contains an ASDS marker block.
func HasMarkerBlock(content string) bool {
 return markerBlockRegex.MatchString(content)
}

// buildMarkerBlock constructs the marker block text for a role and snippets.
func buildMarkerBlock(role string, snippets []string) string {
 var b strings.Builder
 fmt.Fprintf(&b, "<!-- ASDS:BEGIN role=%s -->\n", role)
 for _, s := range snippets {
  fmt.Fprintf(&b, "- %s\n", s)
 }
 b.WriteString(markerEnd + "\n")
 return b.String()
}

// UpsertMarkerBlock adds or replaces the ASDS marker block in a CLAUDE.md file.
// Creates the file if it doesn't exist.
func UpsertMarkerBlock(path string, role string, snippets []string) error {
 block := buildMarkerBlock(role, snippets)

 data, err := os.ReadFile(path)
 if err != nil {
  if os.IsNotExist(err) {
   if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
    return fmt.Errorf("creating directory: %w", err)
   }
   return os.WriteFile(path, []byte(block), 0o644)
  }
  return fmt.Errorf("reading CLAUDE.md: %w", err)
 }

 content := string(data)
 if HasMarkerBlock(content) {
  content = markerBlockRegex.ReplaceAllString(content, block)
 } else {
  if !strings.HasSuffix(content, "\n") {
   content += "\n"
  }
  content += "\n" + block
 }

 return os.WriteFile(path, []byte(content), 0o644)
}

// RemoveMarkerBlock removes the ASDS marker block from a CLAUDE.md file.
// No-op if the file doesn't exist or has no marker block.
func RemoveMarkerBlock(path string) error {
 data, err := os.ReadFile(path)
 if err != nil {
  if os.IsNotExist(err) {
   return nil
  }
  return fmt.Errorf("reading CLAUDE.md: %w", err)
 }

 content := string(data)
 if !HasMarkerBlock(content) {
  return nil
 }

 content = markerBlockRegex.ReplaceAllString(content, "")
 // Clean up any double blank lines left behind
 for strings.Contains(content, "\n\n\n") {
  content = strings.ReplaceAll(content, "\n\n\n", "\n\n")
 }

 return os.WriteFile(path, []byte(content), 0o644)
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/claude/ -run "TestUpsert|TestRemove|TestHas" -v`
Expected: all tests PASS

- [ ] **Step 5: Run all claude package tests**

Run: `go test ./internal/claude/ -v`
Expected: all tests PASS

- [ ] **Step 6: Commit**

```bash
git add internal/claude/claudemd.go internal/claude/claudemd_test.go
git commit -m "feat: add CLAUDE.md marker block management (upsert, remove)"
```

---

### Task 12: Gitignore Helper for Local Scope

**Files:**

- Modify: `internal/claude/paths.go` (add EnsureGitignore)
- Modify: `internal/claude/paths_test.go` (add gitignore test)

- [ ] **Step 1: Write test for gitignore management**

Append to `internal/claude/paths_test.go`:

```go
func TestEnsureGitignore(t *testing.T) {
 dir := t.TempDir()
 claudeDir := filepath.Join(dir, ".claude")
 os.MkdirAll(claudeDir, 0o755)

 err := claude.EnsureGitignore(claudeDir, ".asds-manifest.local.json")
 if err != nil {
  t.Fatalf("unexpected error: %v", err)
 }

 data, _ := os.ReadFile(filepath.Join(claudeDir, ".gitignore"))
 if !strings.Contains(string(data), ".asds-manifest.local.json") {
  t.Error("entry not added to .gitignore")
 }

 // Idempotent: calling again should not duplicate
 claude.EnsureGitignore(claudeDir, ".asds-manifest.local.json")
 data, _ = os.ReadFile(filepath.Join(claudeDir, ".gitignore"))
 count := strings.Count(string(data), ".asds-manifest.local.json")
 if count != 1 {
  t.Errorf("entry duplicated: found %d times", count)
 }
}
```

Add `"strings"` to the test file imports if not already present.

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/claude/ -run TestEnsureGitignore -v`
Expected: FAIL — `EnsureGitignore` not defined

- [ ] **Step 3: Implement EnsureGitignore**

Add to `internal/claude/paths.go`:

```go
// EnsureGitignore ensures the given entry is present in dir/.gitignore.
// Creates the .gitignore file if it doesn't exist.
func EnsureGitignore(dir string, entry string) error {
 gitignorePath := filepath.Join(dir, ".gitignore")

 data, err := os.ReadFile(gitignorePath)
 if err != nil && !os.IsNotExist(err) {
  return fmt.Errorf("reading .gitignore: %w", err)
 }

 content := string(data)
 // Check if entry already exists
 for _, line := range strings.Split(content, "\n") {
  if strings.TrimSpace(line) == entry {
   return nil // Already present
  }
 }

 // Append entry
 if len(content) > 0 && !strings.HasSuffix(content, "\n") {
  content += "\n"
 }
 content += entry + "\n"

 if err := os.MkdirAll(dir, 0o755); err != nil {
  return fmt.Errorf("creating directory: %w", err)
 }
 return os.WriteFile(gitignorePath, []byte(content), 0o644)
}
```

Add `"strings"` to `paths.go` imports.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/claude/ -run TestEnsureGitignore -v`
Expected: PASS

- [ ] **Step 5: Run all tests**

Run: `go test ./internal/claude/ -v`
Expected: all tests PASS

- [ ] **Step 6: Commit**

```bash
git add internal/claude/paths.go internal/claude/paths_test.go
git commit -m "feat: add .gitignore helper for local scope manifest"
```
