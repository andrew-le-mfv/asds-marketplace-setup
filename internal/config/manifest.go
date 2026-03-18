package config

import (
	"fmt"
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
