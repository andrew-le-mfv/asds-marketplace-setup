package config

import (
	appconfig "github.com/andrew-le-mfv/asds-marketplace-setup/internal/config"
)

// Model holds the config viewer state.
type Model struct {
	asdsCfg *appconfig.ASDSConfig
	cfgPath string
	width   int
	height  int
}

// New creates a config viewer model.
func New() Model {
	cfgPath := appconfig.ResolveASDSConfigPath()
	cfg, _ := appconfig.ReadASDSConfig(cfgPath)

	return Model{
		asdsCfg: cfg,
		cfgPath: cfgPath,
	}
}
