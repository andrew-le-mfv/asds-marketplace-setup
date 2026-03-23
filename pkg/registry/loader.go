package registry

import (
	"os"
	"path/filepath"

	"github.com/andrew-le-mfv/asds-marketplace-setup/internal/config"
)

// LoadAllMarketplaces loads marketplace configs using the 4-layer fallback chain:
//  1. Embedded default (always loaded as baseline)
//  2. User marketplaces from marketplacesConfigPath (~/.config/asds/marketplaces.yaml)
//  3. Remote fetch for each enabled marketplace URL
//  4. Project-local marketplace.yaml at projectRoot/marketplace.yaml
//
// Later layers override earlier ones by marketplace name.
func LoadAllMarketplaces(marketplacesConfigPath string, projectRoot string) []*config.MarketplaceConfig {
	seen := make(map[string]int)
	var result []*config.MarketplaceConfig

	addOrReplace := func(cfg *config.MarketplaceConfig) {
		if idx, ok := seen[cfg.Marketplace.Name]; ok {
			result[idx] = cfg
		} else {
			seen[cfg.Marketplace.Name] = len(result)
			result = append(result, cfg)
		}
	}

	// Layer 1: Embedded default
	if defaultCfg, err := config.DefaultMarketplaceConfig(); err == nil {
		addOrReplace(defaultCfg)
	}

	// Layer 2+3: User-configured marketplaces (fetch remote, fallback to cached)
	mktsCfg, err := config.ReadMarketplacesConfig(marketplacesConfigPath)
	if err == nil {
		for _, entry := range mktsCfg.EnabledMarketplaces() {
			// Try simple remote fetch first (for repos with a config file).
			rawURL := BuildRawURL(entry.URL)
			if fetched, fetchErr := FetchMarketplaceConfig(rawURL); fetchErr == nil {
				addOrReplace(fetched)
				continue
			}
			// Fallback: load cached config saved during discovery.
			if cached, cacheErr := LoadCachedMarketplaceConfig(entry.Name); cacheErr == nil {
				addOrReplace(cached)
			}
		}
	}

	// Layer 4: Project-local marketplace.yaml
	if projectRoot != "" {
		localPath := filepath.Join(projectRoot, "marketplace.yaml")
		if data, readErr := os.ReadFile(localPath); readErr == nil {
			if localCfg, parseErr := config.ParseMarketplaceConfig(data); parseErr == nil {
				if validateErr := localCfg.Validate(); validateErr == nil {
					addOrReplace(localCfg)
				}
			}
		}
	}

	return result
}
