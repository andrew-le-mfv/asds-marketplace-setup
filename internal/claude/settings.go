package claude

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ReadSettings reads a Claude settings JSON file.
// Returns an empty map if the file doesn't exist.
func ReadSettings(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]any), nil
		}
		return nil, fmt.Errorf("reading settings: %w", err)
	}

	var settings map[string]any
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, fmt.Errorf("parsing settings: %w", err)
	}
	return settings, nil
}

// WriteSettings writes a Claude settings map to disk as formatted JSON.
func WriteSettings(path string, settings map[string]any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating settings directory: %w", err)
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling settings: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing settings: %w", err)
	}
	return nil
}

// MergeEnabledPlugins merges the given plugins into the enabledPlugins map.
// Existing plugins are preserved; new plugins are added.
func MergeEnabledPlugins(settings map[string]any, plugins map[string]bool) {
	ep, ok := settings["enabledPlugins"].(map[string]any)
	if !ok {
		ep = make(map[string]any)
	}

	for name, enabled := range plugins {
		ep[name] = enabled
	}
	settings["enabledPlugins"] = ep
}

// DisablePlugins removes the specified plugin keys from enabledPlugins.
func DisablePlugins(settings map[string]any, pluginRefs []string) {
	ep, ok := settings["enabledPlugins"].(map[string]any)
	if !ok {
		return
	}

	for _, ref := range pluginRefs {
		delete(ep, ref)
	}
}

// MergeMarketplaceRegistration adds a marketplace entry to extraKnownMarketplaces.
func MergeMarketplaceRegistration(settings map[string]any, marketplaceName, registryURL string) {
	ekm, ok := settings["extraKnownMarketplaces"].(map[string]any)
	if !ok {
		ekm = make(map[string]any)
	}

	ekm[marketplaceName] = map[string]any{
		"source": map[string]any{
			"source": "github",
			"repo":   registryURL,
		},
	}
	settings["extraKnownMarketplaces"] = ekm
}
