package installer

import (
	"github.com/your-org/asds-marketplace-setup/internal/config"
)

// InstallResult holds the outcome of a plugin install/uninstall operation.
type InstallResult struct {
	PluginRef string
	Success   bool
	Error     error
}

// Installer abstracts over CLI and Direct installation methods.
type Installer interface {
	// Install enables the given plugins for the specified scope.
	Install(plugins []config.PluginRef, scope config.Scope, projectRoot string) ([]InstallResult, error)

	// Uninstall disables the given plugins for the specified scope.
	Uninstall(pluginRefs []string, scope config.Scope, projectRoot string) ([]InstallResult, error)

	// RegisterMarketplace registers a marketplace source.
	RegisterMarketplace(name string, registryURL string) error

	// Method returns "cli" or "direct" for manifest tracking.
	Method() string
}
