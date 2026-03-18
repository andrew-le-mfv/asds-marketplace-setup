package claude_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andrew-le-mfv/asds-marketplace-setup/internal/claude"
	"github.com/andrew-le-mfv/asds-marketplace-setup/internal/config"
)

func TestSettingsPath(t *testing.T) {
	projectRoot := "/tmp/test-project"
	home, _ := os.UserHomeDir()

	tests := []struct {
		scope config.Scope
		want  string
	}{
		{config.ScopeUser, filepath.Join(home, ".claude", "settings.json")},
		{config.ScopeProject, filepath.Join(projectRoot, ".claude", "settings.json")},
		{config.ScopeLocal, filepath.Join(projectRoot, ".claude", "settings.local.json")},
	}

	for _, tt := range tests {
		t.Run(string(tt.scope), func(t *testing.T) {
			got := claude.SettingsPath(tt.scope, projectRoot)
			if got != tt.want {
				t.Errorf("SettingsPath(%q) = %q, want %q", tt.scope, got, tt.want)
			}
		})
	}
}

func TestManifestPath(t *testing.T) {
	projectRoot := "/tmp/test-project"
	home, _ := os.UserHomeDir()

	tests := []struct {
		scope config.Scope
		want  string
	}{
		{config.ScopeUser, filepath.Join(home, ".claude", ".asds-manifest.json")},
		{config.ScopeProject, filepath.Join(projectRoot, ".claude", ".asds-manifest.json")},
		{config.ScopeLocal, filepath.Join(projectRoot, ".claude", ".asds-manifest.local.json")},
	}

	for _, tt := range tests {
		t.Run(string(tt.scope), func(t *testing.T) {
			got := claude.ManifestPath(tt.scope, projectRoot)
			if got != tt.want {
				t.Errorf("ManifestPath(%q) = %q, want %q", tt.scope, got, tt.want)
			}
		})
	}
}

func TestClaudeMDPath(t *testing.T) {
	projectRoot := "/tmp/test-project"

	tests := []struct {
		scope   config.Scope
		want    string
		wantErr bool
	}{
		{config.ScopeProject, filepath.Join(projectRoot, "CLAUDE.md"), false},
		{config.ScopeLocal, filepath.Join(projectRoot, "CLAUDE.md"), false},
		{config.ScopeUser, "", true},
	}

	for _, tt := range tests {
		t.Run(string(tt.scope), func(t *testing.T) {
			got, err := claude.ClaudeMDPath(tt.scope, projectRoot)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("ClaudeMDPath(%q) = %q, want %q", tt.scope, got, tt.want)
			}
		})
	}
}

func TestMarketplaceRegistrationPath(t *testing.T) {
	home, _ := os.UserHomeDir()
	want := filepath.Join(home, ".claude", "settings.json")
	got := claude.MarketplaceRegistrationPath()
	if got != want {
		t.Errorf("MarketplaceRegistrationPath() = %q, want %q", got, want)
	}
}

func TestFindProjectRoot(t *testing.T) {
	// Create a temp dir with .git
	dir := t.TempDir()
	gitDir := filepath.Join(dir, ".git")
	if err := os.Mkdir(gitDir, 0o755); err != nil {
		t.Fatal(err)
	}

	subDir := filepath.Join(dir, "src", "pkg")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := claude.FindProjectRoot(subDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != dir {
		t.Errorf("FindProjectRoot() = %q, want %q", got, dir)
	}
}

func TestFindProjectRoot_NoGit(t *testing.T) {
	dir := t.TempDir()
	got, err := claude.FindProjectRoot(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Falls back to cwd (the dir itself)
	if got != dir {
		t.Errorf("FindProjectRoot() = %q, want %q (fallback)", got, dir)
	}
}

func TestEnsureGitignore(t *testing.T) {
	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	os.MkdirAll(claudeDir, 0o755)

	err := claude.EnsureGitignore(claudeDir, ".asds-manifest.local.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(claudeDir, ".gitignore"))
	if !strings.Contains(string(data), ".asds-manifest.local.json") {
		t.Error("entry not added to .gitignore")
	}

	// Idempotent: calling again should not duplicate
	claude.EnsureGitignore(claudeDir, ".asds-manifest.local.json")
	data, _ = os.ReadFile(filepath.Join(claudeDir, ".gitignore"))
	count := strings.Count(string(data), ".asds-manifest.local.json")
	if count != 1 {
		t.Errorf("entry duplicated: found %d times", count)
	}
}
