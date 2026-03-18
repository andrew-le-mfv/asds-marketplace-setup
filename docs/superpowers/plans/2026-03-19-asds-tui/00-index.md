# ASDS TUI Implementation Plan — Index

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a full-screen Golang TUI that bootstraps developers into the ASDS (Agentic Software Development Suite) — installing curated Claude Code plugins by role, with full lifecycle management.

**Architecture:** Cobra CLI dispatches commands. Each command resolves whether to launch interactive TUI (Bubble Tea) or execute non-interactively. Core logic lives in `internal/` packages: config parsing, Claude settings manipulation, installer abstraction, and TUI models. A `pkg/registry` package handles remote marketplace fetching.

**Tech Stack:** Go 1.23+, charmbracelet/bubbletea, charmbracelet/huh, charmbracelet/lipgloss, charmbracelet/bubbles, spf13/cobra, gopkg.in/yaml.v3, goreleaser

**Spec:** `docs/specs/2026-03-18-asds-tui-design.md`

---

## Parts

Execute parts **in order**. Parts 1–4 are foundational and must be sequential. Parts 5–7 have limited parallelism noted below.

| Part | File | Description | Dependencies |
|------|------|-------------|--------------|
| 1 | [01-project-bootstrap.md](./01-project-bootstrap.md) | Go module, deps, project skeleton, core domain types | None |
| 2 | [02-config-layer.md](./02-config-layer.md) | Marketplace YAML parser, manifest JSON read/write, ASDS config, embedded defaults | Part 1 |
| 3 | [03-claude-integration.md](./03-claude-integration.md) | Claude settings JSON read/write/merge, path resolution, CLAUDE.md marker blocks | Part 1 |
| 4 | [04-installer-layer.md](./04-installer-layer.md) | Installer interface, Claude Code detector, DirectInstaller, CLIInstaller | Parts 2, 3 |
| 5 | [05-registry-and-cli.md](./05-registry-and-cli.md) | HTTP registry fetch, Cobra CLI commands (install, uninstall, update, status, reset) | Part 4 |
| 6 | [06-tui-foundation.md](./06-tui-foundation.md) | Bubble Tea app shell, tab navigation, theme/styles | Part 1 |
| 7 | [07-tui-tabs.md](./07-tui-tabs.md) | Setup wizard, Plugins browser, Config viewer, Status dashboard, About tab | Parts 5, 6 |
| 8 | [08-distribution.md](./08-distribution.md) | GoReleaser config, install script, README | Part 7 |

### Parallelism opportunities

- **Parts 2 & 3** can run in parallel (both depend only on Part 1).
- **Parts 5 & 6** can run in parallel (5 depends on Part 4; 6 depends on Part 1 only).
- All other parts are sequential.

---

## File Structure Overview

```
asds-marketplace-setup/
├── cmd/
│   └── asds/
│       └── main.go
├── internal/
│   ├── config/
│   │   ├── marketplace.go        # MarketplaceConfig type + parser
│   │   ├── marketplace_test.go
│   │   ├── manifest.go           # Manifest type + read/write
│   │   ├── manifest_test.go
│   │   ├── asdsconfig.go         # ASDS own config (~/.config/asds/config.yaml)
│   │   ├── asdsconfig_test.go
│   │   └── defaults.go           # go:embed fallback marketplace YAML
│   ├── installer/
│   │   ├── installer.go          # Installer interface + factory
│   │   ├── detector.go           # Claude Code CLI detection
│   │   ├── detector_test.go
│   │   ├── direct.go             # DirectInstaller (JSON file manipulation)
│   │   ├── direct_test.go
│   │   ├── cli.go                # CLIInstaller (shells out to claude CLI)
│   │   └── cli_test.go
│   ├── claude/
│   │   ├── settings.go           # Read/write/merge Claude settings JSON
│   │   ├── settings_test.go
│   │   ├── paths.go              # Scope path resolution
│   │   ├── paths_test.go
│   │   ├── claudemd.go           # CLAUDE.md marker block management
│   │   └── claudemd_test.go
│   ├── tui/
│   │   ├── app.go                # Root Bubble Tea model
│   │   ├── tabs.go               # Tab navigation component
│   │   ├── keymap.go             # Shared key bindings
│   │   ├── styles/
│   │   │   └── theme.go          # Lipgloss palette + styles
│   │   ├── setup/
│   │   │   ├── model.go
│   │   │   ├── update.go
│   │   │   └── view.go
│   │   ├── plugins/
│   │   │   ├── model.go
│   │   │   ├── update.go
│   │   │   └── view.go
│   │   ├── config/
│   │   │   ├── model.go
│   │   │   ├── update.go
│   │   │   └── view.go
│   │   ├── status/
│   │   │   ├── model.go
│   │   │   ├── update.go
│   │   │   └── view.go
│   │   └── about/
│   │       └── view.go
│   └── commands/
│       ├── root.go               # Root cobra command (launches TUI)
│       ├── install.go            # install subcommand
│       ├── uninstall.go          # uninstall subcommand
│       ├── update.go             # update subcommand
│       ├── status.go             # status subcommand
│       └── reset.go              # reset subcommand
├── pkg/
│   └── registry/
│       ├── fetch.go              # HTTP fetch marketplace config
│       └── fetch_test.go
├── configs/
│   └── default-marketplace.yaml  # Embedded fallback
├── scripts/
│   └── install.sh
├── .goreleaser.yaml
├── go.mod
├── go.sum
├── CLAUDE.md
└── README.md
```
