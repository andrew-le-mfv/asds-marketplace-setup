# Part 6: TUI Foundation

**Dependencies:** Part 1 (core types)
**Can run in parallel with:** Part 5 (registry + CLI)
**Estimated tasks:** 3

---

## Chunk 6: Bubble Tea App Shell, Tab Navigation, Theme

### Task 21: Theme and Styles

**Files:**

- Create: `internal/tui/styles/theme.go`

- [ ] **Step 1: Implement the theme**

Create `internal/tui/styles/theme.go`:

```go
package styles

import (
 "github.com/charmbracelet/lipgloss"
)

// Colors — a cohesive palette for the ASDS TUI.
var (
 Primary     = lipgloss.Color("#7C3AED") // Purple
 Secondary   = lipgloss.Color("#06B6D4") // Cyan
 Success     = lipgloss.Color("#22C55E") // Green
 Warning     = lipgloss.Color("#F59E0B") // Amber
 Danger      = lipgloss.Color("#EF4444") // Red
 Muted       = lipgloss.Color("#6B7280") // Gray
 Text        = lipgloss.Color("#F9FAFB") // Almost white
 TextDim     = lipgloss.Color("#9CA3AF") // Light gray
 Background  = lipgloss.Color("#111827") // Dark
 Surface     = lipgloss.Color("#1F2937") // Slightly lighter
 Border      = lipgloss.Color("#374151") // Gray border
)

// Layout styles
var (
 AppStyle = lipgloss.NewStyle().
  Padding(0, 1)

 HeaderStyle = lipgloss.NewStyle().
  Bold(true).
  Foreground(Primary).
  Padding(0, 1)

 TitleStyle = lipgloss.NewStyle().
  Bold(true).
  Foreground(Text)

 SubtitleStyle = lipgloss.NewStyle().
  Foreground(TextDim)

 FooterStyle = lipgloss.NewStyle().
  Foreground(Muted).
  Padding(0, 1)
)

// Tab styles
var (
 ActiveTabStyle = lipgloss.NewStyle().
  Bold(true).
  Foreground(Primary).
  Border(lipgloss.NormalBorder(), false, false, true, false).
  BorderForeground(Primary).
  Padding(0, 2)

 InactiveTabStyle = lipgloss.NewStyle().
  Foreground(TextDim).
  Padding(0, 2)

 TabGapStyle = lipgloss.NewStyle().
  Border(lipgloss.NormalBorder(), false, false, true, false).
  BorderForeground(Border)
)

// Content styles
var (
 SelectedStyle = lipgloss.NewStyle().
  Foreground(Primary).
  Bold(true)

 NormalStyle = lipgloss.NewStyle().
  Foreground(Text)

 SuccessStyle = lipgloss.NewStyle().
  Foreground(Success)

 ErrorStyle = lipgloss.NewStyle().
  Foreground(Danger)

 WarningStyle = lipgloss.NewStyle().
  Foreground(Warning)

 BoxStyle = lipgloss.NewStyle().
  Border(lipgloss.RoundedBorder()).
  BorderForeground(Border).
  Padding(1, 2)

 HelpStyle = lipgloss.NewStyle().
  Foreground(Muted)
)
```

- [ ] **Step 2: Verify it compiles**

Run: `go build ./internal/tui/styles/`
Expected: compiles without errors

- [ ] **Step 3: Commit**

```bash
git add internal/tui/styles/
git commit -m "feat: add TUI theme and lipgloss styles"
```

---

### Task 22: Tab Navigation Component

**Files:**

- Create: `internal/tui/keymap.go`
- Create: `internal/tui/tabs.go`

- [ ] **Step 1: Define shared key bindings**

Create `internal/tui/keymap.go`:

```go
package tui

import (
 "github.com/charmbracelet/bubbles/key"
)

// KeyMap defines the global key bindings for the TUI.
type KeyMap struct {
 NextTab  key.Binding
 PrevTab  key.Binding
 Quit     key.Binding
 Help     key.Binding
 Select   key.Binding
 Back     key.Binding
 Up       key.Binding
 Down     key.Binding
}

// DefaultKeyMap returns the default key bindings.
func DefaultKeyMap() KeyMap {
 return KeyMap{
  NextTab: key.NewBinding(
   key.WithKeys("tab"),
   key.WithHelp("tab", "next tab"),
  ),
  PrevTab: key.NewBinding(
   key.WithKeys("shift+tab"),
   key.WithHelp("shift+tab", "prev tab"),
  ),
  Quit: key.NewBinding(
   key.WithKeys("q", "ctrl+c"),
   key.WithHelp("q", "quit"),
  ),
  Help: key.NewBinding(
   key.WithKeys("?"),
   key.WithHelp("?", "help"),
  ),
  Select: key.NewBinding(
   key.WithKeys("enter"),
   key.WithHelp("enter", "select"),
  ),
  Back: key.NewBinding(
   key.WithKeys("esc"),
   key.WithHelp("esc", "back"),
  ),
  Up: key.NewBinding(
   key.WithKeys("up", "k"),
   key.WithHelp("↑/k", "up"),
  ),
  Down: key.NewBinding(
   key.WithKeys("down", "j"),
   key.WithHelp("↓/j", "down"),
  ),
 }
}
```

- [ ] **Step 2: Implement tab navigation**

Create `internal/tui/tabs.go`:

```go
package tui

// TabID identifies each tab in the dashboard.
type TabID int

const (
 TabSetup TabID = iota
 TabPlugins
 TabConfig
 TabStatus
 TabAbout
)

// TabInfo holds metadata for a tab.
type TabInfo struct {
 ID    TabID
 Label string
 Icon  string
}

// AllTabs returns the ordered list of dashboard tabs.
func AllTabs() []TabInfo {
 return []TabInfo{
  {ID: TabSetup, Label: "Setup", Icon: "⬡"},
  {ID: TabPlugins, Label: "Plugins", Icon: "📦"},
  {ID: TabConfig, Label: "Config", Icon: "⚙"},
  {ID: TabStatus, Label: "Status", Icon: "📊"},
  {ID: TabAbout, Label: "About", Icon: "ℹ"},
 }
}

// TabCount returns the total number of tabs.
func TabCount() int {
 return len(AllTabs())
}
```

- [ ] **Step 3: Verify it compiles**

Run: `go build ./internal/tui/`
Expected: compiles without errors

- [ ] **Step 4: Commit**

```bash
git add internal/tui/keymap.go internal/tui/tabs.go
git commit -m "feat: add TUI tab navigation and keymap"
```

---

### Task 23: Root Bubble Tea Model (App Shell)

**Files:**

- Create: `internal/tui/app.go`
- Create: `internal/tui/about/view.go` (simplest tab, to prove the shell works)

- [ ] **Step 1: Create the About tab (simplest tab)**

Create `internal/tui/about/view.go`:

```go
package about

import (
 tea "github.com/charmbracelet/bubbletea"
 "github.com/charmbracelet/lipgloss"

 "github.com/your-org/asds-marketplace-setup/internal/tui/styles"
)

// Model is the About tab model.
type Model struct {
 version string
 width   int
 height  int
}

// New creates a new About tab model.
func New(version string) Model {
 return Model{version: version}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
 return nil
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
 switch msg := msg.(type) {
 case tea.WindowSizeMsg:
  m.width = msg.Width
  m.height = msg.Height
 }
 return m, nil
}

// View implements tea.Model.
func (m Model) View() string {
 title := styles.TitleStyle.Render("🚀 ASDS — Agentic Software Development Suite")

 info := lipgloss.JoinVertical(lipgloss.Left,
  "",
  title,
  "",
  styles.NormalStyle.Render("Version: "+m.version),
  "",
  styles.SubtitleStyle.Render("A TUI for bootstrapping developers into curated"),
  styles.SubtitleStyle.Render("Claude Code plugin sets organized by role."),
  "",
  styles.HelpStyle.Render("GitHub: github.com/your-org/asds-marketplace-setup"),
  styles.HelpStyle.Render("Docs:   https://your-org.github.io/asds-marketplace-setup"),
 )

 return styles.BoxStyle.Render(info)
}
```

- [ ] **Step 2: Create the root app model**

Create `internal/tui/app.go`:

```go
package tui

import (
 "strings"

 "github.com/charmbracelet/bubbles/key"
 tea "github.com/charmbracelet/bubbletea"
 "github.com/charmbracelet/lipgloss"

 "github.com/your-org/asds-marketplace-setup/internal/tui/about"
 "github.com/your-org/asds-marketplace-setup/internal/tui/styles"
)

// App is the root Bubble Tea model for the ASDS dashboard.
type App struct {
 activeTab TabID
 tabs      []TabInfo
 keys      KeyMap
 width     int
 height    int

 // Tab models
 aboutModel about.Model

 // Placeholder views for tabs not yet implemented
}

// NewApp creates a new App model.
func NewApp(version string) App {
 return App{
  activeTab:  TabSetup,
  tabs:       AllTabs(),
  keys:       DefaultKeyMap(),
  aboutModel: about.New(version),
 }
}

// Init implements tea.Model.
func (a App) Init() tea.Cmd {
 return nil
}

// Update implements tea.Model.
func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
 switch msg := msg.(type) {
 case tea.KeyMsg:
  switch {
  case key.Matches(msg, a.keys.Quit):
   return a, tea.Quit
  case key.Matches(msg, a.keys.NextTab):
   a.activeTab = TabID((int(a.activeTab) + 1) % TabCount())
   return a, nil
  case key.Matches(msg, a.keys.PrevTab):
   a.activeTab = TabID((int(a.activeTab) - 1 + TabCount()) % TabCount())
   return a, nil
  }

 case tea.WindowSizeMsg:
  a.width = msg.Width
  a.height = msg.Height
 }

 // Route to active tab
 var cmd tea.Cmd
 switch a.activeTab {
 case TabAbout:
  a.aboutModel, cmd = a.aboutModel.Update(msg)
 }

 return a, cmd
}

// View implements tea.Model.
func (a App) View() string {
 header := a.renderHeader()
 tabBar := a.renderTabBar()
 content := a.renderContent()
 footer := a.renderFooter()

 return lipgloss.JoinVertical(lipgloss.Left,
  header,
  tabBar,
  content,
  footer,
 )
}

func (a App) renderHeader() string {
 return styles.HeaderStyle.Render("🚀 ASDS — Agentic Software Development Suite")
}

func (a App) renderTabBar() string {
 var tabs []string
 for _, tab := range a.tabs {
  label := tab.Icon + " " + tab.Label
  if tab.ID == a.activeTab {
   tabs = append(tabs, styles.ActiveTabStyle.Render(label))
  } else {
   tabs = append(tabs, styles.InactiveTabStyle.Render(label))
  }
 }
 return lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
}

func (a App) renderContent() string {
 switch a.activeTab {
 case TabSetup:
  return a.placeholderView("Setup", "Role selection wizard — coming in Part 7")
 case TabPlugins:
  return a.placeholderView("Plugins", "Plugin browser — coming in Part 7")
 case TabConfig:
  return a.placeholderView("Config", "Configuration viewer — coming in Part 7")
 case TabStatus:
  return a.placeholderView("Status", "Status dashboard — coming in Part 7")
 case TabAbout:
  return a.aboutModel.View()
 default:
  return ""
 }
}

func (a App) renderFooter() string {
 keys := []string{
  "↑↓ navigate",
  "tab/shift+tab switch",
  "enter select",
  "q quit",
 }
 return styles.FooterStyle.Render(strings.Join(keys, "  │  "))
}

func (a App) placeholderView(title, description string) string {
 content := lipgloss.JoinVertical(lipgloss.Left,
  "",
  styles.TitleStyle.Render(title),
  "",
  styles.SubtitleStyle.Render(description),
  "",
 )
 return styles.BoxStyle.Render(content)
}
```

- [ ] **Step 3: Wire the TUI into the root cobra command**

Modify `internal/commands/root.go` — replace the `RunE` function body:

```go
  RunE: func(cmd *cobra.Command, args []string) error {
   app := tui.NewApp(version)
   p := tea.NewProgram(app, tea.WithAltScreen())
   _, err := p.Run()
   return err
  },
```

Add import `"github.com/your-org/asds-marketplace-setup/internal/tui"` and
`tea "github.com/charmbracelet/bubbletea"` to root.go.

- [ ] **Step 4: Verify it compiles**

Run: `go build ./cmd/asds/`
Expected: compiles without errors

- [ ] **Step 5: Manual smoke test**

Run: `go run ./cmd/asds/`
Expected: full-screen TUI launches with tab bar, placeholder content, can switch tabs with tab/shift+tab, quit with q

- [ ] **Step 6: Commit**

```bash
git add internal/tui/ internal/commands/root.go
git commit -m "feat: add TUI app shell with tab navigation and about tab"
```
