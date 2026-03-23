package commands

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/andrew-le-mfv/asds-marketplace-setup/internal/claude"
	"github.com/andrew-le-mfv/asds-marketplace-setup/internal/config"
	"github.com/andrew-le-mfv/asds-marketplace-setup/internal/tui"
	"github.com/andrew-le-mfv/asds-marketplace-setup/pkg/registry"
)

const version = "0.1.0"

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "asds",
		Short: "ASDS — Agentic Software Development Suite",
		Long:  "A TUI for bootstrapping developers into curated Claude Code plugin sets organized by role.",
		RunE: func(cmd *cobra.Command, args []string) error {
			projectRoot, _ := claude.FindProjectRoot(".")

			mktsCfgPath := config.ResolveMarketplacesConfigPath()
			allCfgs := registry.LoadAllMarketplaces(mktsCfgPath, projectRoot)

			app := tui.NewApp(version, allCfgs, projectRoot)
			p := tea.NewProgram(app, tea.WithAltScreen())
			_, err := p.Run()
			return err
		},
	}

	cmd.Version = version

	cmd.AddCommand(newInstallCmd())
	cmd.AddCommand(newUninstallCmd())
	cmd.AddCommand(newUpdateCmd())
	cmd.AddCommand(newStatusCmd())
	cmd.AddCommand(newResetCmd())
	cmd.AddCommand(newMarketplaceCmd())

	return cmd
}
