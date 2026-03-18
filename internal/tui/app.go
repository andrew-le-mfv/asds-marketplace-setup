package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/your-org/asds-marketplace-setup/internal/tui/about"
	"github.com/your-org/asds-marketplace-setup/internal/tui/styles"
)

// App is the root Bubble Tea model for the ASDS dashboard.
type App struct {
	activeTab TabID
	tabs      []TabInfo
	keys      KeyMap
	width     int
	height    int

	// Tab models
	aboutModel about.Model
}

// NewApp creates a new App model.
func NewApp(version string) App {
	return App{
		activeTab:  TabSetup,
		tabs:       AllTabs(),
		keys:       DefaultKeyMap(),
		aboutModel: about.New(version),
	}
}

// Init implements tea.Model.
func (a App) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, a.keys.Quit):
			return a, tea.Quit
		case key.Matches(msg, a.keys.NextTab):
			a.activeTab = TabID((int(a.activeTab) + 1) % TabCount())
			return a, nil
		case key.Matches(msg, a.keys.PrevTab):
			a.activeTab = TabID((int(a.activeTab) - 1 + TabCount()) % TabCount())
			return a, nil
		}

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
	}

	// Route to active tab
	var cmd tea.Cmd
	switch a.activeTab {
	case TabAbout:
		a.aboutModel, cmd = a.aboutModel.Update(msg)
	}

	return a, cmd
}

// View implements tea.Model.
func (a App) View() string {
	header := a.renderHeader()
	tabBar := a.renderTabBar()
	content := a.renderContent()
	footer := a.renderFooter()

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		tabBar,
		content,
		footer,
	)
}

func (a App) renderHeader() string {
	return styles.HeaderStyle.Render("🚀 ASDS — Agentic Software Development Suite")
}

func (a App) renderTabBar() string {
	var tabs []string
	for _, tab := range a.tabs {
		label := tab.Icon + " " + tab.Label
		if tab.ID == a.activeTab {
			tabs = append(tabs, styles.ActiveTabStyle.Render(label))
		} else {
			tabs = append(tabs, styles.InactiveTabStyle.Render(label))
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
}

func (a App) renderContent() string {
	switch a.activeTab {
	case TabSetup:
		return a.placeholderView("Setup", "Role selection wizard — coming in Part 7")
	case TabPlugins:
		return a.placeholderView("Plugins", "Plugin browser — coming in Part 7")
	case TabConfig:
		return a.placeholderView("Config", "Configuration viewer — coming in Part 7")
	case TabStatus:
		return a.placeholderView("Status", "Status dashboard — coming in Part 7")
	case TabAbout:
		return a.aboutModel.View()
	default:
		return ""
	}
}

func (a App) renderFooter() string {
	keys := []string{
		"↑↓ navigate",
		"tab/shift+tab switch",
		"enter select",
		"q quit",
	}
	return styles.FooterStyle.Render(strings.Join(keys, "  │  "))
}

func (a App) placeholderView(title, description string) string {
	content := lipgloss.JoinVertical(lipgloss.Left,
		"",
		styles.TitleStyle.Render(title),
		"",
		styles.SubtitleStyle.Render(description),
		"",
	)
	return styles.BoxStyle.Render(content)
}
