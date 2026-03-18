package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Scope represents a Claude Code plugin installation scope.
type Scope string

const (
	ScopeUser    Scope = "user"
	ScopeProject Scope = "project"
	ScopeLocal   Scope = "local"
)

// ParseScope parses a string into a Scope, returning an error for invalid values.
func ParseScope(s string) (Scope, error) {
	switch s {
	case "user":
		return ScopeUser, nil
	case "project":
		return ScopeProject, nil
	case "local":
		return ScopeLocal, nil
	default:
		return "", fmt.Errorf("invalid scope %q: must be one of user, project, local", s)
	}
}

// Manifest tracks what ASDS installed, enabling lifecycle operations.
type Manifest struct {
	SchemaVersion      int              `json:"schema_version"`
	ASDSVersion        string           `json:"asds_version"`
	InstalledAt        time.Time        `json:"installed_at"`
	UpdatedAt          time.Time        `json:"updated_at"`
	Role               string           `json:"role"`
	Scope              Scope            `json:"scope"`
	MarketplaceSource  string           `json:"marketplace_source"`
	InstallMethod      string           `json:"install_method"`
	ClaudeCodeDetected bool             `json:"claude_code_detected"`
	Plugins            []ManifestPlugin `json:"plugins"`
	ClaudeMDModified   bool             `json:"claude_md_modified"`
	ScaffoldedFiles    []string         `json:"scaffolded_files"`
}

// ManifestPlugin tracks a single installed plugin.
type ManifestPlugin struct {
	Name        string    `json:"name"`
	FullRef     string    `json:"full_ref"`
	Required    bool      `json:"required"`
	InstalledAt time.Time `json:"installed_at"`
}

// WriteManifest writes the manifest to disk as formatted JSON.
func WriteManifest(path string, m *Manifest) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating manifest directory: %w", err)
	}

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling manifest: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing manifest: %w", err)
	}
	return nil
}

// ReadManifest reads a manifest from disk.
func ReadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading manifest: %w", err)
	}

	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parsing manifest: %w", err)
	}
	return &m, nil
}
