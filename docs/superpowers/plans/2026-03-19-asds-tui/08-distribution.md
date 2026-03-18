# Part 8: Distribution

**Dependencies:** Part 7 (all features implemented)
**Estimated tasks:** 3

---

## Chunk 8: GoReleaser, Install Script, README

### Task 29: GoReleaser Configuration

**Files:**

- Create: `.goreleaser.yaml`

- [ ] **Step 1: Create GoReleaser config**

Create `.goreleaser.yaml`:

```yaml
version: 2

project_name: asds

before:
  hooks:
    - go mod tidy
    - go test ./...

builds:
  - id: asds
    main: ./cmd/asds/
    binary: asds
    env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w

archives:
  - id: default
    format: tar.gz
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"
    format_overrides:
      - goos: windows
        format: zip

checksum:
  name_template: "checksums.txt"

changelog:
  sort: asc
  filters:
    exclude:
      - "^docs:"
      - "^test:"
      - "^chore:"

release:
  github:
    owner: your-org
    name: asds-marketplace-setup
  draft: false
  prerelease: auto
```

- [ ] **Step 2: Verify GoReleaser config (if goreleaser is installed)**

Run: `goreleaser check` (skip if not installed)
Expected: no config errors

- [ ] **Step 3: Test local build**

Run: `go build -o bin/asds ./cmd/asds/`
Expected: binary created at `bin/asds`

Run: `./bin/asds --version`
Expected: `asds version 0.1.0`

- [ ] **Step 4: Commit**

```bash
git add .goreleaser.yaml
echo "bin/" >> .gitignore
git add .gitignore
git commit -m "chore: add GoReleaser config for cross-platform builds"
```

---

### Task 30: Install Script

**Files:**

- Create: `scripts/install.sh`

- [ ] **Step 1: Create the install script**

Create `scripts/install.sh`:

```bash
#!/bin/sh
set -e

# ASDS Marketplace Setup — Install Script
# Usage: curl -fsSL https://raw.githubusercontent.com/your-org/asds-marketplace-setup/main/scripts/install.sh | sh

REPO="your-org/asds-marketplace-setup"
BINARY_NAME="asds"
INSTALL_DIR="${ASDS_INSTALL_DIR:-/usr/local/bin}"

# Detect OS and architecture
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
    x86_64)  ARCH="amd64" ;;
    aarch64) ARCH="arm64" ;;
    arm64)   ARCH="arm64" ;;
    *)
        echo "Error: unsupported architecture: $ARCH"
        exit 1
        ;;
esac

case "$OS" in
    linux)  OS="linux" ;;
    darwin) OS="darwin" ;;
    *)
        echo "Error: unsupported OS: $OS"
        exit 1
        ;;
esac

# Get latest release tag
echo "Fetching latest release..."
LATEST=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST" ]; then
    echo "Error: could not determine latest release"
    exit 1
fi

VERSION="${LATEST#v}"
FILENAME="${BINARY_NAME}_${VERSION}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${LATEST}/${FILENAME}"

echo "Downloading ${BINARY_NAME} ${LATEST} for ${OS}/${ARCH}..."
TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

curl -fsSL "$URL" -o "${TMPDIR}/${FILENAME}"

echo "Extracting..."
tar -xzf "${TMPDIR}/${FILENAME}" -C "$TMPDIR"

echo "Installing to ${INSTALL_DIR}..."
if [ -w "$INSTALL_DIR" ]; then
    mv "${TMPDIR}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
else
    sudo mv "${TMPDIR}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
fi

chmod +x "${INSTALL_DIR}/${BINARY_NAME}"

echo ""
echo "✅ ${BINARY_NAME} ${LATEST} installed to ${INSTALL_DIR}/${BINARY_NAME}"
echo ""
echo "Get started:"
echo "  ${BINARY_NAME}              # Launch dashboard TUI"
echo "  ${BINARY_NAME} install      # Install plugins for a role"
echo "  ${BINARY_NAME} --help       # Show all commands"
```

- [ ] **Step 2: Make it executable**

Run: `chmod +x scripts/install.sh`

- [ ] **Step 3: Verify script syntax**

Run: `bash -n scripts/install.sh`
Expected: no syntax errors

- [ ] **Step 4: Commit**

```bash
git add scripts/install.sh
git commit -m "chore: add curl install script for Unix"
```

---

### Task 31: README and CLAUDE.md

**Files:**

- Create: `README.md`
- Create: `CLAUDE.md`

- [ ] **Step 1: Create README**

Create `README.md`:

```markdown
# ASDS — Agentic Software Development Suite

A TUI for bootstrapping developers into curated Claude Code plugin sets organized by role.

## Quick Start

### Install

```sh
curl -fsSL https://raw.githubusercontent.com/your-org/asds-marketplace-setup/main/scripts/install.sh | sh
```

Or download a binary from [GitHub Releases](https://github.com/andrew-le-mfv/asds-marketplace-setup/releases).

### Usage

```sh
# Launch the interactive dashboard
asds

# Install plugins for a role (non-interactive)
asds install --role developer --scope project

# Check current setup
asds status

# Update plugins to latest marketplace config
asds update --scope project

# Uninstall ASDS plugins
asds uninstall --scope project --yes

# Reset everything
asds reset --scope project --yes
```

## Roles

| Role | Description |
|------|-------------|
| `developer` | Full-stack development with code quality tools |
| `frontend` | UI/UX focused development |
| `backend` | API and server-side development |
| `devops` | CI/CD, infrastructure, and deployment |
| `tester` | Testing and quality assurance |
| `security` | Security auditing and compliance |
| `techlead` | Architecture, code review, and team standards |
| `data-engineer` | Data pipelines and analytics |
| `pm` | Product planning and requirements |

## Scopes

| Scope | Settings File | Shared via Git? |
|-------|--------------|----------------|
| `user` | `~/.claude/settings.json` | N/A |
| `project` | `.claude/settings.json` | ✅ Yes |
| `local` | `.claude/settings.local.json` | ❌ No |

## Development

```sh
# Run locally
go run ./cmd/asds/

# Run tests
go test ./...

# Build binary
go build -o bin/asds ./cmd/asds/
```

## License

MIT

```

- [ ] **Step 2: Create CLAUDE.md**

Create `CLAUDE.md`:

```markdown
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
```

- [ ] **Step 3: Run final full test suite**

Run: `go test ./... -v`
Expected: all tests PASS

- [ ] **Step 4: Final build verification**

Run: `go build -o bin/asds ./cmd/asds/`
Run: `./bin/asds --version`
Expected: `asds version 0.1.0`

Run: `./bin/asds status --project-root .`
Expected: shows status output

- [ ] **Step 5: Commit**

```bash
git add README.md CLAUDE.md
git commit -m "docs: add README and CLAUDE.md"
```
