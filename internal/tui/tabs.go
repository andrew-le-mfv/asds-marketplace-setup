package tui

// TabID identifies each tab in the dashboard.
type TabID int

const (
	TabSetup TabID = iota
	TabPlugins
	TabConfig
	TabStatus
	TabAbout
)

// TabInfo holds metadata for a tab.
type TabInfo struct {
	ID    TabID
	Label string
	Icon  string
}

// AllTabs returns the ordered list of dashboard tabs.
func AllTabs() []TabInfo {
	return []TabInfo{
		{ID: TabSetup, Label: "Setup", Icon: "⬡"},
		{ID: TabPlugins, Label: "Plugins", Icon: "📦"},
		{ID: TabConfig, Label: "Config", Icon: "⚙"},
		{ID: TabStatus, Label: "Status", Icon: "📊"},
		{ID: TabAbout, Label: "About", Icon: "ℹ"},
	}
}

// TabCount returns the total number of tabs.
func TabCount() int {
	return len(AllTabs())
}
