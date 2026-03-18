package config_test

import (
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/andrew-le-mfv/asds-marketplace-setup/internal/config"
)

func TestManifest_JSON_Roundtrip(t *testing.T) {
	now := time.Date(2026, 3, 18, 10, 0, 0, 0, time.UTC)
	m := config.Manifest{
		SchemaVersion:      1,
		ASDSVersion:        "0.1.0",
		InstalledAt:        now,
		UpdatedAt:          now,
		Role:               "developer",
		Scope:              config.ScopeProject,
		MarketplaceSource:  "github.com/test/marketplace",
		InstallMethod:      "direct",
		ClaudeCodeDetected: false,
		Plugins: []config.ManifestPlugin{
			{
				Name:        "code-reviewer",
				FullRef:     "code-reviewer@test-marketplace",
				Required:    true,
				InstalledAt: now,
			},
		},
		ClaudeMDModified: true,
		ScaffoldedFiles:  []string{".claude/settings.json", "CLAUDE.md"},
	}

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var decoded config.Manifest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if decoded.Role != "developer" {
		t.Errorf("role = %q, want %q", decoded.Role, "developer")
	}
	if decoded.Scope != config.ScopeProject {
		t.Errorf("scope = %q, want %q", decoded.Scope, config.ScopeProject)
	}
	if len(decoded.Plugins) != 1 {
		t.Fatalf("plugins count = %d, want 1", len(decoded.Plugins))
	}
	if decoded.Plugins[0].Name != "code-reviewer" {
		t.Errorf("plugin name = %q, want %q", decoded.Plugins[0].Name, "code-reviewer")
	}
}

func TestScope_String(t *testing.T) {
	tests := []struct {
		scope config.Scope
		want  string
	}{
		{config.ScopeUser, "user"},
		{config.ScopeProject, "project"},
		{config.ScopeLocal, "local"},
	}

	for _, tt := range tests {
		if got := string(tt.scope); got != tt.want {
			t.Errorf("Scope(%q).String() = %q, want %q", tt.scope, got, tt.want)
		}
	}
}

func TestParseScope(t *testing.T) {
	tests := []struct {
		input string
		want  config.Scope
		err   bool
	}{
		{"user", config.ScopeUser, false},
		{"project", config.ScopeProject, false},
		{"local", config.ScopeLocal, false},
		{"invalid", "", true},
	}

	for _, tt := range tests {
		got, err := config.ParseScope(tt.input)
		if tt.err && err == nil {
			t.Errorf("ParseScope(%q): expected error, got nil", tt.input)
		}
		if !tt.err && err != nil {
			t.Errorf("ParseScope(%q): unexpected error: %v", tt.input, err)
		}
		if got != tt.want {
			t.Errorf("ParseScope(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestManifest_WriteAndRead(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".asds-manifest.json")

	now := time.Date(2026, 3, 18, 10, 0, 0, 0, time.UTC)
	m := config.Manifest{
		SchemaVersion:      1,
		ASDSVersion:        "0.1.0",
		InstalledAt:        now,
		UpdatedAt:          now,
		Role:               "developer",
		Scope:              config.ScopeProject,
		MarketplaceSource:  "github.com/test/marketplace",
		InstallMethod:      "direct",
		ClaudeCodeDetected: false,
		Plugins: []config.ManifestPlugin{
			{
				Name:        "code-reviewer",
				FullRef:     "code-reviewer@test-marketplace",
				Required:    true,
				InstalledAt: now,
			},
		},
		ClaudeMDModified: true,
		ScaffoldedFiles:  []string{".claude/settings.json"},
	}

	if err := config.WriteManifest(path, &m); err != nil {
		t.Fatalf("WriteManifest error: %v", err)
	}

	loaded, err := config.ReadManifest(path)
	if err != nil {
		t.Fatalf("ReadManifest error: %v", err)
	}

	if loaded.Role != "developer" {
		t.Errorf("role = %q, want %q", loaded.Role, "developer")
	}
	if loaded.Scope != config.ScopeProject {
		t.Errorf("scope = %q, want %q", loaded.Scope, config.ScopeProject)
	}
	if len(loaded.Plugins) != 1 {
		t.Fatalf("plugins count = %d, want 1", len(loaded.Plugins))
	}
}

func TestReadManifest_NotFound(t *testing.T) {
	_, err := config.ReadManifest("/nonexistent/path/manifest.json")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}
