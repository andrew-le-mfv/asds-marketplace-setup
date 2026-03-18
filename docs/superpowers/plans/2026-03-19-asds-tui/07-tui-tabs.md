# Part 7: TUI Tabs

**Dependencies:** Parts 5 and 6 (CLI commands + TUI foundation)
**Estimated tasks:** 5

---

## Chunk 7: Setup Wizard, Plugin Browser, Config, Status, About

### Task 24: Setup Wizard Tab — Model and View

**Files:**

- Create: `internal/tui/setup/model.go`
- Create: `internal/tui/setup/update.go`
- Create: `internal/tui/setup/view.go`

- [ ] **Step 1: Define the setup wizard model**

Create `internal/tui/setup/model.go`:

```go
package setup

import (
 "github.com/andrew-le-mfv/asds-marketplace-setup/internal/config"
 "github.com/andrew-le-mfv/asds-marketplace-setup/internal/installer"
)

// Step tracks the wizard's current position.
type Step int

const (
 StepRoleSelect Step = iota
 StepScopeSelect
 StepConfirm
 StepInstalling
 StepComplete
 StepError
)

// Model holds the setup wizard state.
type Model struct {
 step           Step
 roles          []roleItem
 selectedRole   int
 scopes         []scopeItem
 selectedScope  int
 marketplaceCfg *config.MarketplaceConfig
 projectRoot    string
 installResults []installer.InstallResult
 errorMsg       string
 width          int
 height         int
}

type roleItem struct {
 ID          string
 DisplayName string
 Description string
 PluginCount int
}

type scopeItem struct {
 Scope       config.Scope
 Label       string
 Description string
}

// New creates a new setup wizard model.
func New(cfg *config.MarketplaceConfig, projectRoot string) Model {
 roles := make([]roleItem, 0, len(cfg.Roles))
 for _, name := range cfg.RoleNames() {
  r := cfg.Roles[name]
  roles = append(roles, roleItem{
   ID:          name,
   DisplayName: r.DisplayName,
   Description: r.Description,
   PluginCount: len(r.Plugins),
  })
 }

 scopes := []scopeItem{
  {Scope: config.ScopeUser, Label: "User (global)", Description: "Install for you — ~/.claude/settings.json"},
  {Scope: config.ScopeProject, Label: "Project (shared)", Description: "Install for this project — .claude/settings.json"},
  {Scope: config.ScopeLocal, Label: "Local (private)", Description: "Install locally — .claude/settings.local.json"},
 }

 return Model{
  step:           StepRoleSelect,
  roles:          roles,
  scopes:         scopes,
  marketplaceCfg: cfg,
  projectRoot:    projectRoot,
 }
}

// SelectedRoleID returns the currently selected role ID.
func (m Model) SelectedRoleID() string {
 if m.selectedRole < len(m.roles) {
  return m.roles[m.selectedRole].ID
 }
 return ""
}

// SelectedScope returns the currently selected scope.
func (m Model) SelectedScope() config.Scope {
 if m.selectedScope < len(m.scopes) {
  return m.scopes[m.selectedScope].Scope
 }
 return config.ScopeProject
}
```

- [ ] **Step 2: Implement update logic**

Create `internal/tui/setup/update.go`:

```go
package setup

import (
 "fmt"
 "time"

 tea "github.com/charmbracelet/bubbletea"

 "github.com/andrew-le-mfv/asds-marketplace-setup/internal/claude"
 "github.com/andrew-le-mfv/asds-marketplace-setup/internal/config"
 "github.com/andrew-le-mfv/asds-marketplace-setup/internal/installer"
)

// InstallCompleteMsg is sent when installation finishes.
type InstallCompleteMsg struct {
 Results []installer.InstallResult
 Error   error
}

// Update handles key events for the setup wizard.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
 switch msg := msg.(type) {
 case tea.WindowSizeMsg:
  m.width = msg.Width
  m.height = msg.Height

 case tea.KeyMsg:
  switch m.step {
  case StepRoleSelect:
   return m.updateRoleSelect(msg)
  case StepScopeSelect:
   return m.updateScopeSelect(msg)
  case StepConfirm:
   return m.updateConfirm(msg)
  case StepComplete, StepError:
   // Any key returns to role select
   if msg.String() == "enter" || msg.String() == "esc" {
    m.step = StepRoleSelect
   }
  }

 case InstallCompleteMsg:
  if msg.Error != nil {
   m.step = StepError
   m.errorMsg = msg.Error.Error()
  } else {
   m.step = StepComplete
   m.installResults = msg.Results
  }
 }

 return m, nil
}

func (m Model) updateRoleSelect(msg tea.KeyMsg) (Model, tea.Cmd) {
 switch msg.String() {
 case "up", "k":
  if m.selectedRole > 0 {
   m.selectedRole--
  }
 case "down", "j":
  if m.selectedRole < len(m.roles)-1 {
   m.selectedRole++
  }
 case "enter":
  m.step = StepScopeSelect
 }
 return m, nil
}

func (m Model) updateScopeSelect(msg tea.KeyMsg) (Model, tea.Cmd) {
 switch msg.String() {
 case "up", "k":
  if m.selectedScope > 0 {
   m.selectedScope--
  }
 case "down", "j":
  if m.selectedScope < len(m.scopes)-1 {
   m.selectedScope++
  }
 case "enter":
  m.step = StepConfirm
 case "esc":
  m.step = StepRoleSelect
 }
 return m, nil
}

func (m Model) updateConfirm(msg tea.KeyMsg) (Model, tea.Cmd) {
 switch msg.String() {
 case "enter", "y":
  m.step = StepInstalling
  return m, m.doInstall()
 case "esc", "n":
  m.step = StepScopeSelect
 }
 return m, nil
}

func (m Model) doInstall() tea.Cmd {
 return func() tea.Msg {
  roleID := m.SelectedRoleID()
  scope := m.SelectedScope()
  roleConfig := m.marketplaceCfg.Roles[roleID]

  inst := installer.NewInstaller(true)

  // Register marketplace (non-fatal if it fails)
  inst.RegisterMarketplace(
   m.marketplaceCfg.Marketplace.Name,
   m.marketplaceCfg.Marketplace.RegistryURL,
  )

  // Install plugins
  results, err := inst.Install(roleConfig.Plugins, scope, m.projectRoot)
  if err != nil {
   return InstallCompleteMsg{Error: err}
  }

  // Scaffold CLAUDE.md
  if scope != config.ScopeUser && len(roleConfig.ClaudeMDSnippets) > 0 {
   claudeMDPath, pathErr := claude.ClaudeMDPath(scope, m.projectRoot)
   if pathErr == nil {
    claude.UpsertMarkerBlock(claudeMDPath, roleID, roleConfig.ClaudeMDSnippets)
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
   ASDSVersion:        "0.1.0",
   InstalledAt:        now,
   UpdatedAt:          now,
   Role:               roleID,
   Scope:              scope,
   MarketplaceSource:  m.marketplaceCfg.Marketplace.RegistryURL,
   InstallMethod:      inst.Method(),
   ClaudeCodeDetected: installer.DetectClaudeCode().Found,
   Plugins:            manifestPlugins,
   ClaudeMDModified:   scope != config.ScopeUser,
  }

  manifestPath := claude.ManifestPath(scope, m.projectRoot)
  config.WriteManifest(manifestPath, manifest)

  // Gitignore for local scope
  if scope == config.ScopeLocal {
   claudeDir := claude.SettingsPath(scope, m.projectRoot)
   // Extract directory part
   for i := len(claudeDir) - 1; i >= 0; i-- {
    if claudeDir[i] == '/' {
     claude.EnsureGitignore(claudeDir[:i], ".asds-manifest.local.json")
     break
    }
   }
  }

  return InstallCompleteMsg{Results: results}
 }
}
```

- [ ] **Step 3: Implement the view**

Create `internal/tui/setup/view.go`:

```go
package setup

import (
 "fmt"

 "github.com/charmbracelet/lipgloss"

 "github.com/andrew-le-mfv/asds-marketplace-setup/internal/tui/styles"
)

// View renders the setup wizard.
func (m Model) View() string {
 switch m.step {
 case StepRoleSelect:
  return m.viewRoleSelect()
 case StepScopeSelect:
  return m.viewScopeSelect()
 case StepConfirm:
  return m.viewConfirm()
 case StepInstalling:
  return m.viewInstalling()
 case StepComplete:
  return m.viewComplete()
 case StepError:
  return m.viewError()
 default:
  return ""
 }
}

func (m Model) viewRoleSelect() string {
 title := styles.TitleStyle.Render("Select your role")
 subtitle := styles.SubtitleStyle.Render("Choose the role that best matches your work")

 var items []string
 for i, r := range m.roles {
  cursor := "  "
  style := styles.NormalStyle
  if i == m.selectedRole {
   cursor = "▸ "
   style = styles.SelectedStyle
  }
  line := fmt.Sprintf("%s%s — %s (%d plugins)", cursor, r.DisplayName, r.Description, r.PluginCount)
  items = append(items, style.Render(line))
 }

 help := styles.HelpStyle.Render("↑↓ navigate  enter select")

 content := lipgloss.JoinVertical(lipgloss.Left,
  append([]string{"", title, subtitle, ""}, append(items, "", help)...)...,
 )

 return styles.BoxStyle.Render(content)
}

func (m Model) viewScopeSelect() string {
 title := styles.TitleStyle.Render("Select installation scope")
 roleLine := styles.SubtitleStyle.Render(fmt.Sprintf("Role: %s", m.roles[m.selectedRole].DisplayName))

 var items []string
 for i, s := range m.scopes {
  cursor := "  "
  style := styles.NormalStyle
  if i == m.selectedScope {
   cursor = "▸ "
   style = styles.SelectedStyle
  }
  line := fmt.Sprintf("%s%s\n    %s", cursor, s.Label, styles.SubtitleStyle.Render(s.Description))
  items = append(items, style.Render(line))
 }

 help := styles.HelpStyle.Render("↑↓ navigate  enter select  esc back")

 content := lipgloss.JoinVertical(lipgloss.Left,
  append([]string{"", title, roleLine, ""}, append(items, "", help)...)...,
 )

 return styles.BoxStyle.Render(content)
}

func (m Model) viewConfirm() string {
 roleID := m.SelectedRoleID()
 role := m.marketplaceCfg.Roles[roleID]
 scope := m.SelectedScope()

 title := styles.TitleStyle.Render("Confirm installation")

 var lines []string
 lines = append(lines, "")
 lines = append(lines, title)
 lines = append(lines, "")
 lines = append(lines, styles.NormalStyle.Render(fmt.Sprintf("  Role:  %s (%s)", role.DisplayName, roleID)))
 lines = append(lines, styles.NormalStyle.Render(fmt.Sprintf("  Scope: %s", scope)))
 lines = append(lines, "")
 lines = append(lines, styles.NormalStyle.Render("  Plugins:"))

 for _, p := range role.Plugins {
  req := ""
  if p.Required {
   req = styles.WarningStyle.Render(" (required)")
  }
  lines = append(lines, styles.NormalStyle.Render(fmt.Sprintf("    • %s%s", p.Name, req)))
 }

 lines = append(lines, "")
 lines = append(lines, styles.HelpStyle.Render("enter/y confirm  esc/n go back"))

 return styles.BoxStyle.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

func (m Model) viewInstalling() string {
 content := lipgloss.JoinVertical(lipgloss.Left,
  "",
  styles.TitleStyle.Render("Installing..."),
  "",
  styles.SubtitleStyle.Render("Setting up ASDS plugins for "+m.roles[m.selectedRole].DisplayName),
  "",
 )
 return styles.BoxStyle.Render(content)
}

func (m Model) viewComplete() string {
 title := styles.SuccessStyle.Render("✅ Installation Complete!")

 var lines []string
 lines = append(lines, "", title, "")

 for _, r := range m.installResults {
  if r.Success {
   lines = append(lines, styles.SuccessStyle.Render(fmt.Sprintf("  ✓ %s", r.PluginRef)))
  } else {
   lines = append(lines, styles.ErrorStyle.Render(fmt.Sprintf("  ✗ %s: %v", r.PluginRef, r.Error)))
  }
 }

 lines = append(lines, "")
 lines = append(lines, styles.HelpStyle.Render("enter to continue"))

 return styles.BoxStyle.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

func (m Model) viewError() string {
 title := styles.ErrorStyle.Render("❌ Installation Failed")

 content := lipgloss.JoinVertical(lipgloss.Left,
  "",
  title,
  "",
  styles.NormalStyle.Render(m.errorMsg),
  "",
  styles.HelpStyle.Render("enter to go back"),
 )

 return styles.BoxStyle.Render(content)
}
```

- [ ] **Step 4: Verify it compiles**

Run: `go build ./internal/tui/setup/`
Expected: compiles without errors

- [ ] **Step 5: Commit**

```bash
git add internal/tui/setup/
git commit -m "feat: add setup wizard tab (role select, scope select, confirm, install)"
```

---

### Task 25: Status Tab

**Files:**

- Create: `internal/tui/status/model.go`
- Create: `internal/tui/status/update.go`
- Create: `internal/tui/status/view.go`

- [ ] **Step 1: Implement status model**

Create `internal/tui/status/model.go`:

```go
package status

import (
 "github.com/andrew-le-mfv/asds-marketplace-setup/internal/claude"
 "github.com/andrew-le-mfv/asds-marketplace-setup/internal/config"
 "github.com/andrew-le-mfv/asds-marketplace-setup/internal/installer"
)

// ScopeInfo holds the status for one scope.
type ScopeInfo struct {
 Scope    config.Scope
 Found    bool
 Manifest *config.Manifest
}

// Model holds the status dashboard state.
type Model struct {
 claudeDetected bool
 claudePath     string
 projectRoot    string
 scopes         []ScopeInfo
 width          int
 height         int
}

// New creates a status model by scanning all scopes.
func New(projectRoot string) Model {
 detection := installer.DetectClaudeCode()

 scopes := make([]ScopeInfo, 0, 3)
 for _, s := range []config.Scope{config.ScopeUser, config.ScopeProject, config.ScopeLocal} {
  mp := claude.ManifestPath(s, projectRoot)
  m, err := config.ReadManifest(mp)
  info := ScopeInfo{Scope: s, Found: err == nil, Manifest: m}
  scopes = append(scopes, info)
 }

 return Model{
  claudeDetected: detection.Found,
  claudePath:     detection.Path,
  projectRoot:    projectRoot,
  scopes:         scopes,
 }
}
```

- [ ] **Step 2: Implement update**

Create `internal/tui/status/update.go`:

```go
package status

import (
 tea "github.com/charmbracelet/bubbletea"
)

// Update handles messages for the status tab.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
 switch msg := msg.(type) {
 case tea.WindowSizeMsg:
  m.width = msg.Width
  m.height = msg.Height
 }
 return m, nil
}
```

- [ ] **Step 3: Implement view**

Create `internal/tui/status/view.go`:

```go
package status

import (
 "fmt"

 "github.com/charmbracelet/lipgloss"

 "github.com/andrew-le-mfv/asds-marketplace-setup/internal/tui/styles"
)

// View renders the status dashboard.
func (m Model) View() string {
 title := styles.TitleStyle.Render("📊 Status Dashboard")

 var lines []string
 lines = append(lines, "", title, "")

 // Claude Code detection
 if m.claudeDetected {
  lines = append(lines, styles.SuccessStyle.Render(fmt.Sprintf("  Claude Code: ✓ %s", m.claudePath)))
 } else {
  lines = append(lines, styles.WarningStyle.Render("  Claude Code: ✗ not detected"))
 }

 lines = append(lines, styles.NormalStyle.Render(fmt.Sprintf("  Project root: %s", m.projectRoot)))
 lines = append(lines, "")

 // Per-scope status
 for _, si := range m.scopes {
  if si.Found {
   lines = append(lines, styles.SuccessStyle.Render(
    fmt.Sprintf("  [%s] Role: %s | Plugins: %d | Method: %s | Installed: %s",
     si.Scope, si.Manifest.Role, len(si.Manifest.Plugins),
     si.Manifest.InstallMethod, si.Manifest.InstalledAt.Format("2006-01-02")),
   ))
  } else {
   lines = append(lines, styles.SubtitleStyle.Render(
    fmt.Sprintf("  [%s] Not installed", si.Scope),
   ))
  }
 }

 return styles.BoxStyle.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}
```

- [ ] **Step 4: Verify it compiles**

Run: `go build ./internal/tui/status/`
Expected: compiles without errors

- [ ] **Step 5: Commit**

```bash
git add internal/tui/status/
git commit -m "feat: add status dashboard tab"
```

---

### Task 26: Plugins Browser Tab

**Files:**

- Create: `internal/tui/plugins/model.go`
- Create: `internal/tui/plugins/update.go`
- Create: `internal/tui/plugins/view.go`

- [ ] **Step 1: Implement plugins model**

Create `internal/tui/plugins/model.go`:

```go
package plugins

import (
 "github.com/andrew-le-mfv/asds-marketplace-setup/internal/config"
)

// PluginItem represents a plugin in the browser list.
type PluginItem struct {
 Name     string
 Source   string
 Required bool
 RoleName string
}

// Model holds the plugin browser state.
type Model struct {
 items    []PluginItem
 cursor   int
 width    int
 height   int
}

// New creates a plugin browser from a marketplace config.
func New(cfg *config.MarketplaceConfig) Model {
 seen := make(map[string]bool)
 var items []PluginItem

 for _, roleName := range cfg.RoleNames() {
  role := cfg.Roles[roleName]
  for _, p := range role.Plugins {
   if seen[p.Source] {
    continue
   }
   seen[p.Source] = true
   items = append(items, PluginItem{
    Name:     p.Name,
    Source:   p.Source,
    Required: p.Required,
    RoleName: roleName,
   })
  }
 }

 return Model{items: items}
}
```

- [ ] **Step 2: Implement update**

Create `internal/tui/plugins/update.go`:

```go
package plugins

import (
 tea "github.com/charmbracelet/bubbletea"
)

// Update handles messages for the plugins tab.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
 switch msg := msg.(type) {
 case tea.WindowSizeMsg:
  m.width = msg.Width
  m.height = msg.Height

 case tea.KeyMsg:
  switch msg.String() {
  case "up", "k":
   if m.cursor > 0 {
    m.cursor--
   }
  case "down", "j":
   if m.cursor < len(m.items)-1 {
    m.cursor++
   }
  }
 }
 return m, nil
}
```

- [ ] **Step 3: Implement view**

Create `internal/tui/plugins/view.go`:

```go
package plugins

import (
 "fmt"

 "github.com/charmbracelet/lipgloss"

 "github.com/andrew-le-mfv/asds-marketplace-setup/internal/tui/styles"
)

// View renders the plugin browser.
func (m Model) View() string {
 title := styles.TitleStyle.Render("📦 Available Plugins")
 subtitle := styles.SubtitleStyle.Render("Browse all plugins from the marketplace")

 var lines []string
 lines = append(lines, "", title, subtitle, "")

 for i, item := range m.items {
  cursor := "  "
  style := styles.NormalStyle
  if i == m.cursor {
   cursor = "▸ "
   style = styles.SelectedStyle
  }

  req := ""
  if item.Required {
   req = styles.WarningStyle.Render(" [required]")
  }

  line := fmt.Sprintf("%s%s%s", cursor, item.Name, req)
  lines = append(lines, style.Render(line))
  if i == m.cursor {
   lines = append(lines, styles.SubtitleStyle.Render(fmt.Sprintf("    Source: %s | Role: %s", item.Source, item.RoleName)))
  }
 }

 lines = append(lines, "", styles.HelpStyle.Render("↑↓ navigate"))

 return styles.BoxStyle.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}
```

- [ ] **Step 4: Verify it compiles**

Run: `go build ./internal/tui/plugins/`
Expected: compiles without errors

- [ ] **Step 5: Commit**

```bash
git add internal/tui/plugins/
git commit -m "feat: add plugin browser tab"
```

---

### Task 27: Config Tab

**Files:**

- Create: `internal/tui/config/model.go`
- Create: `internal/tui/config/update.go`
- Create: `internal/tui/config/view.go`

- [ ] **Step 1: Implement config model**

Create `internal/tui/config/model.go`:

```go
package config

import (
 appconfig "github.com/andrew-le-mfv/asds-marketplace-setup/internal/config"
)

// Model holds the config viewer state.
type Model struct {
 asdsCfg  *appconfig.ASDSConfig
 cfgPath  string
 width    int
 height   int
}

// New creates a config viewer model.
func New() Model {
 cfgPath := appconfig.ResolveASDSConfigPath()
 cfg, _ := appconfig.ReadASDSConfig(cfgPath)

 return Model{
  asdsCfg: cfg,
  cfgPath: cfgPath,
 }
}
```

- [ ] **Step 2: Implement update**

Create `internal/tui/config/update.go`:

```go
package config

import (
 tea "github.com/charmbracelet/bubbletea"
)

// Update handles messages for the config tab.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
 switch msg := msg.(type) {
 case tea.WindowSizeMsg:
  m.width = msg.Width
  m.height = msg.Height
 }
 return m, nil
}
```

- [ ] **Step 3: Implement view**

Create `internal/tui/config/view.go`:

```go
package config

import (
 "fmt"

 "github.com/charmbracelet/lipgloss"

 "github.com/andrew-le-mfv/asds-marketplace-setup/internal/tui/styles"
)

// View renders the config viewer.
func (m Model) View() string {
 title := styles.TitleStyle.Render("⚙ Configuration")

 var lines []string
 lines = append(lines, "", title, "")

 lines = append(lines, styles.NormalStyle.Render(fmt.Sprintf("  Config file: %s", m.cfgPath)))
 lines = append(lines, "")

 if m.asdsCfg != nil {
  lines = append(lines, styles.NormalStyle.Render(fmt.Sprintf("  Marketplace URL: %s", m.asdsCfg.MarketplaceURL)))
 } else {
  lines = append(lines, styles.WarningStyle.Render("  No config file found (using defaults)"))
 }

 lines = append(lines, "")
 lines = append(lines, styles.HelpStyle.Render("Config location: ~/.config/asds/config.yaml"))
 lines = append(lines, styles.HelpStyle.Render("Precedence: CLI flags > env vars > config file > defaults"))

 return styles.BoxStyle.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}
```

- [ ] **Step 4: Verify it compiles**

Run: `go build ./internal/tui/config/`
Expected: compiles without errors

- [ ] **Step 5: Commit**

```bash
git add internal/tui/config/
git commit -m "feat: add config viewer tab"
```

---

### Task 28: Wire All Tabs into App Shell

**Files:**

- Modify: `internal/tui/app.go`

**Breaking change:** This task changes `NewApp` signature from `NewApp(version string)` to `NewApp(version string, cfg *config.MarketplaceConfig, projectRoot string)`. Step 4 updates the caller in `root.go` simultaneously.

- [ ] **Step 1: Add tab model imports and fields**

Update `internal/tui/app.go` to import all tab packages and add model fields:

```go
import (
 // ... existing imports ...
 tuiconfig "github.com/andrew-le-mfv/asds-marketplace-setup/internal/tui/config"
 "github.com/andrew-le-mfv/asds-marketplace-setup/internal/tui/plugins"
 "github.com/andrew-le-mfv/asds-marketplace-setup/internal/tui/setup"
 "github.com/andrew-le-mfv/asds-marketplace-setup/internal/tui/status"
 "github.com/andrew-le-mfv/asds-marketplace-setup/internal/config"
)
```

Update the `App` struct to include all tab models:

```go
type App struct {
 activeTab TabID
 tabs      []TabInfo
 keys      KeyMap
 width     int
 height    int

 setupModel   setup.Model
 pluginsModel plugins.Model
 configModel  tuiconfig.Model
 statusModel  status.Model
 aboutModel   about.Model
}
```

Update `NewApp` to accept marketplace config and project root:

```go
func NewApp(version string, cfg *config.MarketplaceConfig, projectRoot string) App {
 return App{
  activeTab:    TabSetup,
  tabs:         AllTabs(),
  keys:         DefaultKeyMap(),
  setupModel:   setup.New(cfg, projectRoot),
  pluginsModel: plugins.New(cfg),
  configModel:  tuiconfig.New(),
  statusModel:  status.New(projectRoot),
  aboutModel:   about.New(version),
 }
}
```

- [ ] **Step 2: Route Update to all tabs**

In the `Update` method, route messages to the active tab:

```go
 var cmd tea.Cmd
 switch a.activeTab {
 case TabSetup:
  a.setupModel, cmd = a.setupModel.Update(msg)
 case TabPlugins:
  a.pluginsModel, cmd = a.pluginsModel.Update(msg)
 case TabConfig:
  a.configModel, cmd = a.configModel.Update(msg)
 case TabStatus:
  a.statusModel, cmd = a.statusModel.Update(msg)
 case TabAbout:
  a.aboutModel, cmd = a.aboutModel.Update(msg)
 }
```

- [ ] **Step 3: Route View to all tabs**

In `renderContent`, replace placeholders with real tab views:

```go
func (a App) renderContent() string {
 switch a.activeTab {
 case TabSetup:
  return a.setupModel.View()
 case TabPlugins:
  return a.pluginsModel.View()
 case TabConfig:
  return a.configModel.View()
 case TabStatus:
  return a.statusModel.View()
 case TabAbout:
  return a.aboutModel.View()
 default:
  return ""
 }
}
```

- [ ] **Step 4: Update root command to pass config to NewApp**

Modify `internal/commands/root.go` — update the `RunE` to load config:

```go
  RunE: func(cmd *cobra.Command, args []string) error {
   projectRoot, _ := claude.FindProjectRoot(".")

   asdsCfg, _ := config.ReadASDSConfig(config.ResolveASDSConfigPath())
   mktCfg, err := registry.FetchOrDefault(asdsCfg.MarketplaceURL)
   if err != nil {
    return fmt.Errorf("loading marketplace config: %w", err)
   }

   app := tui.NewApp(version, mktCfg, projectRoot)
   p := tea.NewProgram(app, tea.WithAltScreen())
   _, err = p.Run()
   return err
  },
```

Add necessary imports: `config`, `claude`, `registry`.

- [ ] **Step 5: Verify everything compiles**

Run: `go build ./cmd/asds/`
Expected: compiles without errors

- [ ] **Step 6: Manual smoke test**

Run: `go run ./cmd/asds/`
Expected: full TUI with all 5 tabs working — setup wizard navigates roles/scopes, plugins browser scrolls, config shows settings, status shows detection, about shows info.

- [ ] **Step 7: Commit**

```bash
git add internal/tui/app.go internal/commands/root.go
git commit -m "feat: wire all TUI tabs into dashboard app shell"
```
