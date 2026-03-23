#!/bin/sh
set -e

# ASDS Marketplace Setup — Install Script
# Usage: curl -fsSL https://raw.githubusercontent.com/andrew-le-mfv/asds-marketplace-setup/master/scripts/install.sh | sh

REPO="andrew-le-mfv/asds-marketplace-setup"
MODULE="github.com/${REPO}/cmd/asds"
BINARY_NAME="asds"
INSTALL_DIR="${ASDS_INSTALL_DIR:-${HOME}/.local/bin}"

TMPDIR=""
TMPGOBIN=""
cleanup() {
	[ -n "$TMPDIR" ] && rm -rf "$TMPDIR"
	[ -n "$TMPGOBIN" ] && rm -rf "$TMPGOBIN"
}
trap cleanup EXIT

# Extract first "tag_name" string from GitHub API JSON (no jq required).
extract_tag_name() {
	printf '%s' "$1" | grep -o '"tag_name"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | sed 's/.*"\([^"]*\)"$/\1/'
}

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

ensure_install_dir() {
	if [ ! -d "$INSTALL_DIR" ]; then
		mkdir -p "$INSTALL_DIR" || {
			echo "Error: could not create ${INSTALL_DIR}"
			exit 1
		}
	fi
	if [ ! -w "$INSTALL_DIR" ]; then
		echo "Error: ${INSTALL_DIR} is not writable"
		echo "Set ASDS_INSTALL_DIR to a directory you own (default is \$HOME/.local/bin)."
		exit 1
	fi
}

echo "Fetching latest release..."
JSON=$(curl -sS "https://api.github.com/repos/${REPO}/releases/latest")
LATEST=$(extract_tag_name "$JSON")

if [ -z "$LATEST" ]; then
	JSON=$(curl -sS "https://api.github.com/repos/${REPO}/releases?per_page=1")
	LATEST=$(extract_tag_name "$JSON")
fi

install_from_release() {
	VERSION="${LATEST#v}"
	FILENAME="${BINARY_NAME}_${VERSION}_${OS}_${ARCH}.tar.gz"
	URL="https://github.com/${REPO}/releases/download/${LATEST}/${FILENAME}"

	echo "Downloading ${BINARY_NAME} ${LATEST} for ${OS}/${ARCH}..."
	TMPDIR=$(mktemp -d)

	curl -fsSL "$URL" -o "${TMPDIR}/${FILENAME}"

	echo "Extracting..."
	tar -xzf "${TMPDIR}/${FILENAME}" -C "$TMPDIR"

	ensure_install_dir
	echo "Installing to ${INSTALL_DIR}..."
	mv "${TMPDIR}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
	chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
}

install_with_go() {
	if ! command -v go >/dev/null 2>&1; then
		return 1
	fi
	echo "No GitHub release with binaries found; installing with go install..."
	TMPGOBIN=$(mktemp -d)
	GOBIN="$TMPGOBIN" go install "${MODULE}@latest"

	ensure_install_dir
	echo "Installing to ${INSTALL_DIR}..."
	mv "${TMPGOBIN}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
	chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
	LATEST="(go install @latest)"
}

if [ -n "$LATEST" ]; then
	install_from_release
else
	if ! install_with_go; then
		echo "Error: no GitHub releases published for ${REPO} and Go is not available."
		echo "Either publish a release (see .goreleaser.yaml in the repo), install Go and re-run, or build from a clone:"
		echo "  git clone https://github.com/${REPO}.git && cd asds-marketplace-setup && go build -o asds ./cmd/asds/"
		exit 1
	fi
fi

echo ""
echo "✅ ${BINARY_NAME} ${LATEST} installed to ${INSTALL_DIR}/${BINARY_NAME}"
case ":${PATH}:" in
	*":${INSTALL_DIR}:"*) ;;
	*)
		echo ""
		echo "Add this directory to your PATH if needed, e.g.:"
		echo "  export PATH=\"${INSTALL_DIR}:\$PATH\""
		;;
esac
echo ""
echo "Get started:"
echo "  ${BINARY_NAME}              # Launch dashboard TUI"
echo "  ${BINARY_NAME} install      # Install plugins for a role"
echo "  ${BINARY_NAME} --help       # Show all commands"
