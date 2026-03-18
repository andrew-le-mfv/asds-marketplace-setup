package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/your-org/asds-marketplace-setup/internal/claude"
	"github.com/your-org/asds-marketplace-setup/internal/config"
	"github.com/your-org/asds-marketplace-setup/internal/installer"
)

func newResetCmd() *cobra.Command {
	var (
		scope       string
		projectRoot string
		yes         bool
	)

	cmd := &cobra.Command{
		Use:   "reset",
		Short: "Remove all ASDS configuration for a scope",
		Long:  "Completely removes all ASDS traces for the specified scope. Marketplace registration is preserved.",
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

			if !yes {
				return fmt.Errorf("reset requires --yes flag for non-interactive mode")
			}

			manifestPath := claude.ManifestPath(s, projectRoot)
			manifest, err := config.ReadManifest(manifestPath)
			if err != nil {
				fmt.Println("No ASDS installation found — nothing to reset.")
				return nil
			}

			pluginRefs := make([]string, len(manifest.Plugins))
			for i, p := range manifest.Plugins {
				pluginRefs[i] = p.FullRef
			}

			inst := installer.NewInstaller(true)
			inst.Uninstall(pluginRefs, s, projectRoot)

			if s != config.ScopeUser {
				claudeMDPath, err := claude.ClaudeMDPath(s, projectRoot)
				if err == nil {
					claude.RemoveMarkerBlock(claudeMDPath)
				}
			}

			removeFile(manifestPath)

			fmt.Println("✅ ASDS reset complete for scope:", scope)
			return nil
		},
	}

	cmd.Flags().StringVar(&scope, "scope", "", "Scope to reset: user, project, or local")
	cmd.Flags().StringVar(&projectRoot, "project-root", "", "Override project root directory")
	cmd.Flags().BoolVar(&yes, "yes", false, "Confirm reset without prompting")

	return cmd
}

// removeFile removes a file, ignoring "not exist" errors.
func removeFile(path string) error {
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
