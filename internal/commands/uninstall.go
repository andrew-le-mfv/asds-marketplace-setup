package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/your-org/asds-marketplace-setup/internal/claude"
	"github.com/your-org/asds-marketplace-setup/internal/config"
	"github.com/your-org/asds-marketplace-setup/internal/installer"
)

func newUninstallCmd() *cobra.Command {
	var (
		scope       string
		projectRoot string
		yes         bool
	)

	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall ASDS plugins",
		Long:  "Remove all ASDS-installed plugins for the specified scope.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if projectRoot == "" {
				var err error
				projectRoot, err = claude.FindProjectRoot(".")
				if err != nil {
					return fmt.Errorf("finding project root: %w", err)
				}
			}

			if scope == "" {
				return fmt.Errorf("--scope is required (user, project, or local)")
			}

			s, err := config.ParseScope(scope)
			if err != nil {
				return err
			}

			// Read manifest
			manifestPath := claude.ManifestPath(s, projectRoot)
			manifest, err := config.ReadManifest(manifestPath)
			if err != nil {
				return fmt.Errorf("no ASDS installation found for scope %q: %v", scope, err)
			}

			if !yes {
				fmt.Printf("Will uninstall %d plugins for role %q (scope: %s)\n", len(manifest.Plugins), manifest.Role, scope)
			}

			// Uninstall plugins
			pluginRefs := make([]string, len(manifest.Plugins))
			for i, p := range manifest.Plugins {
				pluginRefs[i] = p.FullRef
			}

			inst := installer.NewInstaller(true)
			results, err := inst.Uninstall(pluginRefs, s, projectRoot)
			if err != nil {
				return fmt.Errorf("uninstalling plugins: %w", err)
			}

			for _, r := range results {
				if r.Success {
					fmt.Printf("  ✓ removed %s\n", r.PluginRef)
				} else {
					fmt.Printf("  ✗ %s: %v\n", r.PluginRef, r.Error)
				}
			}

			// Remove CLAUDE.md marker block
			if s != config.ScopeUser {
				claudeMDPath, err := claude.ClaudeMDPath(s, projectRoot)
				if err == nil {
					claude.RemoveMarkerBlock(claudeMDPath)
					fmt.Println("  ✓ CLAUDE.md cleaned")
				}
			}

			// Remove manifest
			if err := removeFile(manifestPath); err != nil {
				fmt.Printf("Warning: could not remove manifest: %v\n", err)
			}

			fmt.Println("\n✅ ASDS uninstall complete!")
			return nil
		},
	}

	cmd.Flags().StringVar(&scope, "scope", "", "Scope to uninstall from: user, project, or local")
	cmd.Flags().StringVar(&projectRoot, "project-root", "", "Override project root directory")
	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")

	return cmd
}
