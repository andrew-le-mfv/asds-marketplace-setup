package configs

import (
	_ "embed"
)

//go:embed default-marketplace.yaml
var DefaultMarketplaceYAML []byte
