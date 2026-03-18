# ASDS Marketplace Setup TUI — Design Specification

**Date:** 2026-03-18
**Status:** Draft
**Schema version:** 1
**Author:** Brainstorm session

---

## 1. Overview

A Golang TUI application that bootstraps developers into the **Agentic Software Development Suite (ASDS)** — a curated collection of Claude Code plugins, settings, and project configurations organized by role.

The tool provides a full-screen interactive dashboard for browsing, installing, uninstalling, updating, and reconfiguring ASDS plugins. Every operation is also available via CLI flags for non-interactive use (CI/CD, scripting).

### Goals

- **Zero-to-productive onboarding**: a developer selects their role and scope, and the TUI installs all relevant Claude Code plugins and scaffolds project configuration.
- **Full lifecycle management**: install, uninstall, update, re-configure (switch roles), and reset.
- **Hybrid registry model**: supports a default public ASDS marketplace AND custom registries (authenticated/private registry support is limited to existing ambient credentials in v0.1).
- **Two interaction modes**: rich full-screen TUI (default) and non-interactive CLI flags.

### Non-Goals

- Replacing the Claude Code CLI plugin manager.
- Building or authoring plugins (use Claude Code's own tooling for that).
- Managing non-ASDS plugins.

---

## 2. User Experience

### 2.1 Two Modes

| Mode | Trigger | Description |
|------|---------|-------------|
| **Interactive** | `asds` or subcommand without all required flags | Full-screen Bubble Tea dashboard / wizard prompts |
| **Non-interactive** | All required flags provided, or `--yes` flag | Direct execution, no TUI, suitable for CI/scripts |

### 2.2 CLI Commands

```
asds                                              → launches dashboard TUI
asds install                                      → interactive wizard for install
asds install --role developer --scope project      → non-interactive install
asds install --role developer                      → prompts for missing --scope only
asds uninstall                                     → interactive selective uninstall
asds uninstall --role developer --scope project    → non-interactive uninstall
asds update                                        → update all installed ASDS plugins
asds status                                        → print current ASDS setup (human-readable)
asds status --json                                 → machine-readable JSON output
asds reset --scope project --yes                   → remove all ASDS configs (no prompt)
```

**Mode resolution per command:**
- `asds` (no subcommand) → always dashboard TUI
- `asds install` with all required flags (`--role` + `--scope`) → non-interactive
- `asds install` with partial or no flags → interactive wizard (prompts for missing values)
- `asds status`, `asds update` → non-interactive by nature (no TUI needed)
- `asds reset` → requires `--yes` for non-interactive; otherwise prompts for confirmation

### 2.3 Dashboard TUI Layout

```
┌──────────────────────────────────────────────────────────────┐
│  🚀 ASDS — Agentic Software Development Suite          v0.1 │
├───────────┬───────────┬───────────┬───────────┬──────────────┤
│  ⬡ Setup  │ 📦 Plugins │ ⚙ Config  │ 📊 Status │   ℹ About   │
├───────────┴───────────┴───────────┴───────────┴──────────────┤
│                                                              │
│  (Active tab content renders here)                           │
│                                                              │
│                                                              │
│                                                              │
├──────────────────────────────────────────────────────────────┤
│  ↑↓ navigate  tab/shift+tab switch  enter select  q quit     │
└──────────────────────────────────────────────────────────────┘
```

**Tabs:**

| Tab | Purpose |
|-----|---------|
| **Setup** | The onboarding wizard: role selection → scope selection → confirmation → install with progress |
| **Plugins** | Browse all plugins from the registry. Enable/disable/install/uninstall individual plugins. |
| **Config** | View/edit marketplace source URL and registry settings. |
| **Status** | Current role, scope, installed plugins, Claude Code detection status. |
| **About** | Version info, links, help text. |

---

## 3. Behavior Contract

### 3.1 Scope Model (v0.1: Single Scope)

In v0.1, the user selects **one scope** and **all plugins install to that scope**. The `recommended_scope` field in the marketplace YAML is informational only — displayed as a hint in the TUI but not enforced.

This keeps install, uninstall, update, and reset logic simple. Mixed-scope installs are a future consideration.

### 3.2 Source of Truth

**Live settings files are the source of truth.** The `.asds-manifest.json` is advisory — it tracks what ASDS installed to enable smart lifecycle operations, but the TUI always reads live Claude settings to determine actual state.

When showing `asds status`, the TUI:
1. Reads the manifest to know what ASDS *intended* to install.
2. Reads live settings to see what *actually* exists.
3. Reports any drift (e.g., "plugin X was installed by ASDS but is now disabled").

### 3.3 DirectInstaller Contract

The DirectInstaller performs **config-only reconciliation**. It writes `enabledPlugins` and `extraKnownMarketplaces` to the appropriate settings JSON file. This declares intent — Claude Code will resolve and fetch the actual plugin assets the next time it starts.

If Claude Code is not yet installed, the user gets a clear message:
> "Plugins have been configured. Install Claude Code and run `claude` to activate them."

### 3.4 File Ownership and Idempotency

- **Settings JSON**: ASDS only touches `enabledPlugins` and `extraKnownMarketplaces` keys. All other keys are preserved during merge. Running install twice is idempotent.
- **CLAUDE.md**: ASDS uses marker blocks for all edits:
  ```md
  <!-- ASDS:BEGIN role=developer -->
  - Follow conventional commits
  - Always write tests for new features
  <!-- ASDS:END -->
  ```
  On re-install or role switch, the existing ASDS block is replaced (not duplicated). On reset, the block is removed.
- **`.gitignore`**: if the local scope creates `.claude/.asds-manifest.local.json`, ASDS ensures it's listed in `.claude/.gitignore`.

### 3.5 Project Root Resolution

Project root is determined by:
1. Nearest parent directory containing `.git/` (git root).
2. If no git root found, use the current working directory.
3. Can be overridden with `--project-root <path>`.

User scope ignores project root entirely (writes to `~/.claude/`).

### 3.6 Marketplace Registration Location

Marketplace registration (`extraKnownMarketplaces`) is always written to `~/.claude/settings.json` (user-level) regardless of plugin installation scope. This ensures Claude Code can resolve `plugin@asds-marketplace` references regardless of which project the user opens.

### 3.7 Trust and Security (v0.1)

- Before installing from any marketplace, the TUI displays the full list of plugins and their source repo, requiring explicit confirmation.
- If installing from a non-default (custom) registry, an additional warning is shown: "This is a third-party marketplace. Only install from sources you trust."
- The TUI does not inspect plugin internals (hooks, MCP servers, etc.) — trust is delegated to the marketplace maintainer and Claude Code's own trust mechanisms.

---

## 4. Core Installation Flow

```
1. Detect Claude Code installation
   → If found: prompt user to use native CLI installer
   → If not found: use direct file manipulation
2. Fetch asds-marketplace.yaml
   → Try remote marketplace URL first
   → Fall back to embedded default config
3. User selects Role (huh.Select or --role flag)
4. User selects Scope: User / Project / Local (huh.Select or --scope flag)
5. Display plugin list for the selected role → user confirms
6. Install plugins (CLI or direct method, with progress bar)
7. Scaffold .claude/ directory (CLAUDE.md snippets, settings adjustments)
8. Register ASDS marketplace in extraKnownMarketplaces
9. Write .asds-manifest.json to track what was installed
10. Display summary screen with next steps
```

---

## 5. Lifecycle Semantics

### 5.1 `asds install`

1. Fetch marketplace config.
2. User selects role and scope.
3. Display plugin list with confirmation.
4. Register marketplace in `~/.claude/settings.json` (`extraKnownMarketplaces`).
5. Enable all role plugins in the scope-appropriate settings file (`enabledPlugins`).
6. Scaffold `CLAUDE.md` with role snippets (project/local scope only) using marker blocks.
7. Write `.asds-manifest.json`.

If ASDS is already installed for this scope, the wizard shows the current state and asks to overwrite.

### 5.2 `asds uninstall`

1. Read manifest to find ASDS-installed plugins.
2. Set all ASDS plugins to `false` in `enabledPlugins` (or remove the keys).
3. Remove ASDS marker block from `CLAUDE.md`.
4. Remove `.asds-manifest.json`.
5. Marketplace registration is preserved (user may have other plugins from it).

### 5.3 `asds update`

1. Re-fetch latest marketplace config.
2. Compare current role's plugin list (from manifest) to latest registry.
3. **Added plugins**: enable newly added plugins for the current role.
4. **Removed plugins**: disable plugins no longer in the role.
5. **CLAUDE.md**: replace marker block with latest snippets.
6. Update manifest timestamps and plugin list.

### 5.4 `asds reset`

Removes all ASDS traces for the specified scope:
1. Remove all ASDS `enabledPlugins` entries from settings file.
2. Remove ASDS marker block from `CLAUDE.md`.
3. Delete `.asds-manifest.json`.
4. Marketplace registration is preserved.

Requires `--yes` for non-interactive, otherwise prompts for confirmation.

### 5.5 Role Switching (`asds install` with a different role)

1. Detect existing manifest for the scope.
2. Compute diff: plugins to add (new role) and plugins to remove (old role only).
3. Disable old-role-only plugins; enable new-role plugins.
4. Replace CLAUDE.md marker block with new role's snippets.
5. Update manifest with new role.

---

## 6. Plugin Installation — 3 Scopes

All three Claude Code plugin scopes are supported. The TUI always asks the user to choose.

| Scope | Label in TUI | Settings file written | Shared via git? | Use case |
|-------|-------------|----------------------|-----------------|----------|
| **User** | "Install for you" | `~/.claude/settings.json` | N/A (global) | Personal tools across all projects |
| **Project** | "Install for this project" | `.claude/settings.json` | ✅ Yes | Team standards, shared with collaborators |
| **Local** | "Install locally" | `.claude/settings.local.json` | ❌ No (gitignored) | Personal preference within this repo only |

### Installation Method Selection

```
┌─ Claude Code detected? ─┐
│                          │
├── Yes ──→ Prompt: "Claude Code is installed. Use native plugin installer? (recommended)"
│           ├── Yes ──→ CLIInstaller: shells out to `claude plugin install <name> --scope <scope>`
│           └── No  ──→ DirectInstaller: writes enabledPlugins to settings JSON
│
└── No  ──→ DirectInstaller: writes enabledPlugins to settings JSON
            + Info: "Install Claude Code later to manage plugins natively"
```

### 6.2 CLIInstaller

Executes Claude Code CLI commands:
```sh
claude plugin marketplace add <marketplace-source>
claude plugin install <plugin-name>@<marketplace-name> --scope <scope>
claude plugin uninstall <plugin-name>@<marketplace-name> --scope <scope>
claude plugin enable <plugin-name>@<marketplace-name> --scope <scope>
claude plugin disable <plugin-name>@<marketplace-name> --scope <scope>
```

### 6.3 DirectInstaller (Config-Only Mode)

Writes JSON directly to the appropriate settings file:
```json
{
  "enabledPlugins": {
    "code-reviewer@asds-marketplace": true,
    "commit-commands@asds-marketplace": true
  },
  "extraKnownMarketplaces": {
    "asds-marketplace": {
      "source": {
        "source": "github",
        "repo": "your-org/asds-marketplace"
      }
    }
  }
}
```

The DirectInstaller merges with existing settings (never overwrites unrelated keys).

---

## 7. Marketplace Configuration Format

### 7.1 Remote Config (`asds-marketplace.yaml`)

Lives in the ASDS marketplace repository. The TUI fetches this file at startup to discover available roles and plugins.

**Role IDs** are the canonical identifiers used in CLI flags (`--role developer`), manifest files, and YAML keys. They must be lowercase kebab-case.

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

### 7.2 Fallback

If the remote `asds-marketplace.yaml` cannot be fetched (network error, missing file), the TUI uses an embedded default compiled into the binary via Go's `embed` package. This default contains the same role/plugin structure shown above.

---

## 8. State Tracking

The TUI writes an `.asds-manifest.json` file to track what it installed, enabling uninstall, update, and role switching.

**Location:** determined by scope:
- **User scope:** `~/.claude/.asds-manifest.json`
- **Project scope:** `.claude/.asds-manifest.json`
- **Local scope:** `.claude/.asds-manifest.local.json` (gitignored)

```json
{
  "schema_version": 1,
  "asds_version": "0.1.0",
  "installed_at": "2026-03-18T10:00:00Z",
  "updated_at": "2026-03-18T10:00:00Z",
  "role": "developer",
  "scope": "project",
  "marketplace_source": "github.com/your-org/asds-marketplace",
  "install_method": "cli",
  "claude_code_detected": true,
  "plugins": [
    {
      "name": "code-reviewer",
      "full_ref": "code-reviewer@asds-marketplace",
      "required": true,
      "installed_at": "2026-03-18T10:00:00Z"
    },
    {
      "name": "commit-commands",
      "full_ref": "commit-commands@asds-marketplace",
      "required": false,
      "installed_at": "2026-03-18T10:00:00Z"
    }
  ],
  "claude_md_modified": true,
  "scaffolded_files": [
    ".claude/settings.json",
    "CLAUDE.md"
  ]
}
```

---

## 9. Project Structure

```
asds-marketplace-setup/
├── cmd/
│   └── asds/
│       └── main.go                  # Entry point, cobra root command
├── internal/
│   ├── config/
│   │   ├── marketplace.go           # Parse asds-marketplace.yaml
│   │   ├── manifest.go              # Read/write .asds-manifest.json
│   │   ├── asdsconfig.go            # Read/write ~/.config/asds/config.yaml
│   │   └── defaults.go              # Embedded fallback config (go:embed)
│   ├── installer/
│   │   ├── installer.go             # Installer interface + factory
│   │   ├── cli_installer.go         # Shells out to `claude plugin install`
│   │   ├── direct_installer.go      # Writes settings JSON files directly
│   │   └── detector.go              # Detect Claude Code CLI in PATH
│   ├── tui/
│   │   ├── app.go                   # Root Bubble Tea model, tab routing
│   │   ├── tabs.go                  # Tab navigation component
│   │   ├── setup/                   # Setup wizard tab
│   │   │   ├── model.go             # State: step, selectedRole, selectedScope
│   │   │   ├── update.go            # Handle key events, form transitions
│   │   │   └── view.go              # Render wizard steps
│   │   ├── plugins/                 # Plugin browser tab
│   │   │   ├── model.go
│   │   │   ├── update.go
│   │   │   └── view.go
│   │   ├── config/                  # Config viewer/editor tab
│   │   │   ├── model.go
│   │   │   ├── update.go
│   │   │   └── view.go
│   │   ├── status/                  # Status dashboard tab
│   │   │   ├── model.go
│   │   │   ├── update.go
│   │   │   └── view.go
│   │   ├── about/                   # About/help tab
│   │   │   └── view.go
│   │   └── styles/
│   │       └── theme.go             # Lipgloss color palette + styles
│   └── claude/
│       ├── settings.go              # Read/write/merge Claude settings JSON
│       └── paths.go                 # Resolve ~/.claude, ./.claude, scope paths
├── pkg/
│   └── registry/
│       └── fetch.go                 # HTTP fetch marketplace config from remote
├── configs/
│   └── default-marketplace.yaml     # Embedded fallback config
├── scripts/
│   └── install.sh                   # curl | sh installer script
├── docs/
│   ├── researches/
│   │   └── quick-research.md
│   └── specs/
│       └── 2026-03-18-asds-tui-design.md
├── .goreleaser.yaml                 # Cross-platform build + GitHub Releases
├── go.mod
├── go.sum
└── README.md
```

---

## 10. Dependencies

| Library | Version | Purpose |
|---------|---------|---------|
| `charmbracelet/bubbletea` | latest | Core TUI framework (Elm architecture) |
| `charmbracelet/huh` | latest | Interactive form components (Select, Confirm) |
| `charmbracelet/lipgloss` | latest | Terminal styling (colors, borders, layout) |
| `charmbracelet/bubbles` | latest | Reusable components (spinner, progress, table) |
| `spf13/cobra` | latest | CLI command structure and flag parsing |
| `gopkg.in/yaml.v3` | latest | Parse marketplace YAML configuration |
| `goreleaser/goreleaser` | latest | Cross-compilation and GitHub Release automation |

---

## 11. Claude Code Integration Details

### 11.1 What the TUI Scaffolds

**For project/local scope**, the TUI may scaffold or modify:

| File | Action | Content |
|------|--------|---------|
| `.claude/settings.json` | Merge `enabledPlugins` + `extraKnownMarketplaces` | Plugin enable flags, marketplace registration |
| `.claude/settings.local.json` | Merge `enabledPlugins` (local scope only) | Plugin enable flags for local-only plugins |
| `CLAUDE.md` | Append snippets | Role-specific instructions from `claude_md_snippets` in config |

**For user scope**, the TUI only modifies `~/.claude/settings.json` and the user-level manifest. It does **not** touch project files (`CLAUDE.md`, `.claude/settings.json`).

### 11.2 Supported Plugin Component Types

The ASDS marketplace can distribute plugins containing any of these 7 component types:

| # | Component | Plugin Location | Description |
|---|-----------|----------------|-------------|
| 1 | Skills | `skills/*/SKILL.md` | Reusable prompts Claude auto-invokes based on context |
| 2 | Commands | `commands/*.md` | Slash commands users invoke manually |
| 3 | Agents | `agents/*.md` | Custom AI personas with system prompts and tool restrictions |
| 4 | Hooks | `hooks/hooks.json` | Shell commands triggered on lifecycle events |
| 5 | MCP Servers | `.mcp.json` | External tool integrations (APIs, databases, services) |
| 6 | LSP Servers | `.lsp.json` | Language server protocol for code intelligence |
| 7 | Settings | `settings.json` | Default configuration applied when plugin is enabled |

The TUI does not need to understand plugin internals — it only needs to enable/disable them. Claude Code handles loading components from installed plugins.

---

## 12. Distribution

- **Primary:** GitHub Releases with pre-compiled binaries for macOS (amd64, arm64), Linux (amd64, arm64), and Windows (amd64).
- **Install script (Unix):** `curl -fsSL https://raw.githubusercontent.com/your-org/asds-marketplace-setup/main/scripts/install.sh | sh`
- **Windows:** download binary from GitHub Releases manually (no PowerShell install script in v0.1).
- **Build tool:** GoReleaser for cross-compilation, checksums, and release automation.

---

## 13. ASDS Config File

The TUI stores its own preferences in `~/.config/asds/config.yaml`:

```yaml
marketplace_url: "github.com/your-org/asds-marketplace"
```

**Precedence:** CLI flags > environment variables > config file > embedded defaults.

This is separate from Claude Code settings and only controls the TUI's own behavior.

---

## 14. Error Handling

| Scenario | Behavior |
|----------|----------|
| No network access | Fall back to embedded default `asds-marketplace.yaml` |
| Claude Code CLI not found | Switch to DirectInstaller, inform user |
| `claude plugin install` fails | Log error, continue with remaining plugins, show summary of failures |
| Target settings file doesn't exist | Create it with proper structure |
| Settings file has conflicting keys | Merge (ASDS keys only), never overwrite unrelated settings |
| Invalid marketplace YAML | Show error with line number, abort gracefully |
| User runs `asds` outside a git repo | User scope works fine; Project/Local scopes error: "No project root found. Use --project-root or run from a git repository." |

---

## 15. Future Considerations (Out of Scope for v0.1)

- **Mixed-scope installs**: allow per-plugin scope selection based on `recommended_scope`.
- **Plugin dependency resolution**: automatically install plugin dependencies.
- **Role composition**: select multiple roles and merge their plugin sets.
- **Authenticated private registries**: token-based auth for private marketplace repos.
- **Plugin health checks**: verify installed plugins are functional after install.
- **Auto-update the TUI itself**: self-update mechanism for the binary.
- **PowerShell install script**: for Windows users.
