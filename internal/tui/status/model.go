package status

import (
	"github.com/your-org/asds-marketplace-setup/internal/claude"
	"github.com/your-org/asds-marketplace-setup/internal/config"
	"github.com/your-org/asds-marketplace-setup/internal/installer"
)

// ScopeInfo holds the status for one scope.
type ScopeInfo struct {
	Scope    config.Scope
	Found    bool
	Manifest *config.Manifest
}

// Model holds the status dashboard state.
type Model struct {
	claudeDetected bool
	claudePath     string
	projectRoot    string
	scopes         []ScopeInfo
	width          int
	height         int
}

// New creates a status model by scanning all scopes.
func New(projectRoot string) Model {
	detection := installer.DetectClaudeCode()

	scopes := make([]ScopeInfo, 0, 3)
	for _, s := range []config.Scope{config.ScopeUser, config.ScopeProject, config.ScopeLocal} {
		mp := claude.ManifestPath(s, projectRoot)
		m, err := config.ReadManifest(mp)
		info := ScopeInfo{Scope: s, Found: err == nil, Manifest: m}
		scopes = append(scopes, info)
	}

	return Model{
		claudeDetected: detection.Found,
		claudePath:     detection.Path,
		projectRoot:    projectRoot,
		scopes:         scopes,
	}
}
