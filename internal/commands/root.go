package commands

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/your-org/asds-marketplace-setup/internal/tui"
)

const version = "0.1.0"

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "asds",
		Short: "ASDS — Agentic Software Development Suite",
		Long:  "A TUI for bootstrapping developers into curated Claude Code plugin sets organized by role.",
		RunE: func(cmd *cobra.Command, args []string) error {
			app := tui.NewApp(version)
			p := tea.NewProgram(app, tea.WithAltScreen())
			_, err := p.Run()
			return err
		},
	}

	cmd.Version = version

	cmd.AddCommand(newInstallCmd())

	return cmd
}
