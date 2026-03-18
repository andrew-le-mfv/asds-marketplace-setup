package installer

import (
	"github.com/andrew-le-mfv/asds-marketplace-setup/internal/claude"
	"github.com/andrew-le-mfv/asds-marketplace-setup/internal/config"
)

// DirectInstaller writes enabledPlugins directly to Claude settings JSON files.
type DirectInstaller struct{}

// Install enables plugins by writing to the scope-appropriate settings file.
func (d *DirectInstaller) Install(plugins []config.PluginRef, scope config.Scope, projectRoot string) ([]InstallResult, error) {
	settingsPath := claude.SettingsPath(scope, projectRoot)

	settings, err := claude.ReadSettings(settingsPath)
	if err != nil {
		return nil, err
	}

	pluginMap := make(map[string]bool, len(plugins))
	for _, p := range plugins {
		pluginMap[p.Source] = true
	}

	claude.MergeEnabledPlugins(settings, pluginMap)

	if err := claude.WriteSettings(settingsPath, settings); err != nil {
		return nil, err
	}

	results := make([]InstallResult, len(plugins))
	for i, p := range plugins {
		results[i] = InstallResult{
			PluginRef: p.Source,
			Success:   true,
		}
	}
	return results, nil
}

// Uninstall removes plugins from the scope-appropriate settings file.
func (d *DirectInstaller) Uninstall(pluginRefs []string, scope config.Scope, projectRoot string) ([]InstallResult, error) {
	settingsPath := claude.SettingsPath(scope, projectRoot)

	settings, err := claude.ReadSettings(settingsPath)
	if err != nil {
		return nil, err
	}

	claude.DisablePlugins(settings, pluginRefs)

	if err := claude.WriteSettings(settingsPath, settings); err != nil {
		return nil, err
	}

	results := make([]InstallResult, len(pluginRefs))
	for i, ref := range pluginRefs {
		results[i] = InstallResult{
			PluginRef: ref,
			Success:   true,
		}
	}
	return results, nil
}

// RegisterMarketplace writes marketplace registration to user-level settings.
func (d *DirectInstaller) RegisterMarketplace(name string, registryURL string) error {
	settingsPath := claude.MarketplaceRegistrationPath()

	settings, err := claude.ReadSettings(settingsPath)
	if err != nil {
		return err
	}

	claude.MergeMarketplaceRegistration(settings, name, registryURL)

	return claude.WriteSettings(settingsPath, settings)
}

// Method returns "direct" for manifest tracking.
func (d *DirectInstaller) Method() string {
	return "direct"
}
