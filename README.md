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
