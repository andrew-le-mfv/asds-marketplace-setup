package config

import (
	"github.com/andrew-le-mfv/asds-marketplace-setup/configs"
)

// DefaultMarketplaceConfig returns the embedded fallback marketplace configuration.
func DefaultMarketplaceConfig() (*MarketplaceConfig, error) {
	return ParseMarketplaceConfig(configs.DefaultMarketplaceYAML)
}
