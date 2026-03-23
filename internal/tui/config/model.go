package config

import (
	"github.com/charmbracelet/bubbles/textinput"

	appconfig "github.com/andrew-le-mfv/asds-marketplace-setup/internal/config"
)

// Step tracks the config tab's current state.
type Step int

const (
	StepList Step = iota
	StepAdd
	StepEdit
	StepRemoveConfirm
	StepDiscovering
	StepDiscoverResult
	StepError
)

type addField int

const (
	fieldName addField = iota
	fieldURL
)

// Model holds the marketplace manager state.
type Model struct {
	step        Step
	mktsCfg     *appconfig.MarketplacesConfig
	cfgPath     string
	cursor      int
	width       int
	height      int
	nameInput   textinput.Model
	urlInput    textinput.Model
	activeField addField
	errorMsg     string
	discoverOK   bool
	discoverErr  string
}

// New creates a marketplace manager model.
func New() Model {
	cfgPath := appconfig.ResolveMarketplacesConfigPath()
	cfg, _ := appconfig.ReadMarketplacesConfig(cfgPath)

	ni := textinput.New()
	ni.Placeholder = "marketplace-name"
	ni.CharLimit = 64

	ui := textinput.New()
	ui.Placeholder = "github.com/org/repo"
	ui.CharLimit = 256

	return Model{
		step:      StepList,
		mktsCfg:   cfg,
		cfgPath:   cfgPath,
		nameInput: ni,
		urlInput:  ui,
	}
}

// InForm returns true when a text input is active (add/edit mode).
func (m Model) InForm() bool {
	return m.step == StepAdd || m.step == StepEdit
}

func (m *Model) save() error {
	return appconfig.WriteMarketplacesConfig(m.cfgPath, m.mktsCfg)
}

func (m *Model) reload() {
	cfg, _ := appconfig.ReadMarketplacesConfig(m.cfgPath)
	m.mktsCfg = cfg
}
