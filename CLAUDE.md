# ASDS Marketplace Setup

## Project Overview
This is a Go TUI application that bootstraps developers into the ASDS (Agentic Software Development Suite).

## Build & Test
- Build: `go build -o bin/asds ./cmd/asds/`
- Run: `go run ./cmd/asds/`
- Test: `go test ./...`
- Test verbose: `go test ./... -v`
- Test single package: `go test ./internal/config/ -v`

## Code Style
- Follow standard Go conventions (gofmt, go vet)
- Use `internal/` for private packages, `pkg/` for public
- Error messages: lowercase, no trailing punctuation
- Test files: `*_test.go` in the same package (or `_test` suffix package for black-box)

## Architecture
- `cmd/asds/main.go` — entry point
- `internal/commands/` — Cobra CLI commands
- `internal/config/` — marketplace config, manifest, ASDS config
- `internal/claude/` — Claude settings, paths, CLAUDE.md management
- `internal/installer/` — installer interface, direct + CLI implementations
- `internal/tui/` — Bubble Tea TUI (app shell, tabs, styles)
- `pkg/registry/` — HTTP marketplace config fetch
- `configs/` — embedded default marketplace YAML

## Key Patterns
- Installer interface with Direct and CLI implementations
- CLAUDE.md marker blocks: `<!-- ASDS:BEGIN role=X -->` / `<!-- ASDS:END -->`
- Settings JSON merge (never overwrite unrelated keys)
- go:embed for fallback marketplace config
