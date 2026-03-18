package claude

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/your-org/asds-marketplace-setup/internal/config"
)

// SettingsPath returns the Claude settings file path for the given scope.
func SettingsPath(scope config.Scope, projectRoot string) string {
	switch scope {
	case config.ScopeUser:
		return filepath.Join(userClaudeDir(), "settings.json")
	case config.ScopeProject:
		return filepath.Join(projectRoot, ".claude", "settings.json")
	case config.ScopeLocal:
		return filepath.Join(projectRoot, ".claude", "settings.local.json")
	default:
		return ""
	}
}

// ManifestPath returns the ASDS manifest file path for the given scope.
func ManifestPath(scope config.Scope, projectRoot string) string {
	switch scope {
	case config.ScopeUser:
		return filepath.Join(userClaudeDir(), ".asds-manifest.json")
	case config.ScopeProject:
		return filepath.Join(projectRoot, ".claude", ".asds-manifest.json")
	case config.ScopeLocal:
		return filepath.Join(projectRoot, ".claude", ".asds-manifest.local.json")
	default:
		return ""
	}
}

// ClaudeMDPath returns the CLAUDE.md path for project/local scopes.
// Returns an error for user scope (CLAUDE.md is project-only).
func ClaudeMDPath(scope config.Scope, projectRoot string) (string, error) {
	if scope == config.ScopeUser {
		return "", fmt.Errorf("CLAUDE.md is not applicable for user scope")
	}
	return filepath.Join(projectRoot, "CLAUDE.md"), nil
}

// MarketplaceRegistrationPath returns the path where marketplace registration
// is always written — user-level settings regardless of plugin scope.
func MarketplaceRegistrationPath() string {
	return filepath.Join(userClaudeDir(), "settings.json")
}

// FindProjectRoot walks up from startDir looking for .git/.
// Falls back to startDir if no git root is found.
func FindProjectRoot(startDir string) (string, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", fmt.Errorf("resolving absolute path: %w", err)
	}

	for {
		gitPath := filepath.Join(dir, ".git")
		if info, err := os.Stat(gitPath); err == nil && info.IsDir() {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root, fall back to startDir
			return filepath.Abs(startDir)
		}
		dir = parent
	}
}

func userClaudeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "~"
	}
	return filepath.Join(home, ".claude")
}
