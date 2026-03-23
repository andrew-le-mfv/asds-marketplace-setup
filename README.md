# ASDS — Agentic Software Development Suite

A TUI for bootstrapping developers into curated Claude Code plugin sets organized by role.

## Quick Start

### Install

**One-liner** (requires a [public](https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/managing-repository-settings/setting-repository-visibility) repo so `raw.githubusercontent.com` can serve the script):

```sh
curl -fsSL https://raw.githubusercontent.com/andrew-le-mfv/asds-marketplace-setup/master/scripts/install.sh | sh
```

If that command fails with **404**, GitHub is not serving that URL (common causes: repo is private, not pushed yet, or lives under a different `org/repo`). Fix it by publishing `scripts/install.sh` on the default branch and making the repo public, **or** install from a checkout:

```sh
git clone https://github.com/andrew-le-mfv/asds-marketplace-setup.git
cd asds-marketplace-setup
sh scripts/install.sh
```

The script downloads a release asset when [GitHub Releases](https://github.com/andrew-le-mfv/asds-marketplace-setup/releases) exist; if there are none yet, it falls back to `go install` when Go is on your `PATH`.

To ship binaries for everyone without requiring Go, tag a version and run [GoReleaser](https://goreleaser.com/) (see `.goreleaser.yaml` in this repo).

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
