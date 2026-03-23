package status

import (
	"github.com/andrew-le-mfv/asds-marketplace-setup/internal/claude"
	"github.com/andrew-le-mfv/asds-marketplace-setup/internal/config"
	"github.com/andrew-le-mfv/asds-marketplace-setup/internal/installer"
)

// Step tracks the status tab's current state.
type Step int

const (
	StepView Step = iota
	StepConfirm
	StepUninstalling
	StepResult
)

// ScopeInfo holds the status for one scope.
type ScopeInfo struct {
	Scope    config.Scope
	Found    bool
	Manifest *config.Manifest
}

// Model holds the status dashboard state.
type Model struct {
	step           Step
	claudeDetected bool
	claudePath     string
	projectRoot    string
	scopes         []ScopeInfo
	cursor         int
	uninstResults  []installer.InstallResult
	errorMsg       string
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
		step:           StepView,
		claudeDetected: detection.Found,
		claudePath:     detection.Path,
		projectRoot:    projectRoot,
		scopes:         scopes,
	}
}

// installedScopes returns the indices of scopes that have installations.
func (m Model) installedScopes() []int {
	var indices []int
	for i, si := range m.scopes {
		if si.Found {
			indices = append(indices, i)
		}
	}
	return indices
}

// selectedScopeIndex returns the actual scope index for the current cursor position.
func (m Model) selectedScopeIndex() int {
	installed := m.installedScopes()
	if m.cursor < len(installed) {
		return installed[m.cursor]
	}
	return -1
}

// refresh re-reads manifests from disk.
func (m *Model) refresh() {
	detection := installer.DetectClaudeCode()
	m.claudeDetected = detection.Found
	m.claudePath = detection.Path

	scopes := make([]ScopeInfo, 0, 3)
	for _, s := range []config.Scope{config.ScopeUser, config.ScopeProject, config.ScopeLocal} {
		mp := claude.ManifestPath(s, m.projectRoot)
		manifest, err := config.ReadManifest(mp)
		scopes = append(scopes, ScopeInfo{Scope: s, Found: err == nil, Manifest: manifest})
	}
	m.scopes = scopes
	m.cursor = 0
}
