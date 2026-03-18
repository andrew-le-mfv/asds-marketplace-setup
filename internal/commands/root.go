package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

const version = "0.1.0"

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "asds",
		Short: "ASDS — Agentic Software Development Suite",
		Long:  "A TUI for bootstrapping developers into curated Claude Code plugin sets organized by role.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Will launch dashboard TUI in Part 6
			fmt.Println("ASDS dashboard TUI — coming soon")
			return nil
		},
	}

	cmd.Version = version

	return cmd
}
