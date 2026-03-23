# AGENTS.md — ASDS Marketplace Setup

> Guidance for AI coding agents when working with this repository.

## Project Overview

Go TUI application (Bubble Tea + Cobra) that bootstraps developers into ASDS (Agentic Software Development Suite).

## Build & Test

- Build: `go build -o bin/asds ./cmd/asds/` or `make build`
- Run: `go run ./cmd/asds/` or `make run`
- Test: `go test ./...` or `make test`
- Test verbose: `go test ./... -v`
- Test single package: `go test ./internal/config/ -v` or `make test-package PKG=./internal/config/`
- Format: `gofmt -s -w .` or `make fmt`
- Vet: `go vet ./...` or `make vet`
- Lint: `make lint` (requires golangci-lint)

## Code Style

- Follow standard Go conventions (gofmt, go vet)
- Use `internal/` for private packages, `pkg/` for public
- Error messages: lowercase, no trailing punctuation
- Error wrapping: use `fmt.Errorf("...: %w", err)`
- Test files: `*_test.go` in the same package (or `_test` suffix for black-box)

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
- Three scopes for Claude settings: user, project, local (maps to different file paths)
