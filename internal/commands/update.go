package commands

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/andrew-le-mfv/asds-marketplace-setup/internal/claude"
	"github.com/andrew-le-mfv/asds-marketplace-setup/internal/config"
	"github.com/andrew-le-mfv/asds-marketplace-setup/internal/installer"
	"github.com/andrew-le-mfv/asds-marketplace-setup/pkg/registry"
)

func newUpdateCmd() *cobra.Command {
	var (
		scope       string
		projectRoot string
	)

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update ASDS plugins to latest marketplace config",
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

			manifestPath := claude.ManifestPath(s, projectRoot)
			manifest, err := config.ReadManifest(manifestPath)
			if err != nil {
				return fmt.Errorf("no ASDS installation found for scope %q", scope)
			}

			mktsCfgPath := config.ResolveMarketplacesConfigPath()
			allCfgs := registry.LoadAllMarketplaces(mktsCfgPath, projectRoot)

			var mktCfg *config.MarketplaceConfig
			var roleConfig config.Role
			for _, cfg := range allCfgs {
				if r, ok := cfg.Roles[manifest.Role]; ok {
					mktCfg = cfg
					roleConfig = r
					break
				}
			}
			if mktCfg == nil {
				return fmt.Errorf("role %q no longer exists in any marketplace", manifest.Role)
			}

			currentRefs := make(map[string]bool)
			for _, p := range manifest.Plugins {
				currentRefs[p.FullRef] = true
			}
			newRefs := make(map[string]bool)
			for _, p := range roleConfig.Plugins {
				newRefs[p.Source] = true
			}

			var toAdd []config.PluginRef
			for _, p := range roleConfig.Plugins {
				if !currentRefs[p.Source] {
					toAdd = append(toAdd, p)
				}
			}

			var toRemove []string
			for _, p := range manifest.Plugins {
				if !newRefs[p.FullRef] {
					toRemove = append(toRemove, p.FullRef)
				}
			}

			inst := installer.NewInstaller(true)

			if len(toAdd) > 0 {
				fmt.Println("Adding new plugins:")
				results, err := inst.Install(toAdd, s, projectRoot)
				if err != nil {
					return err
				}
				for _, r := range results {
					if r.Success {
						fmt.Printf("  ✓ %s\n", r.PluginRef)
					} else {
						fmt.Printf("  ✗ %s: %v\n", r.PluginRef, r.Error)
					}
				}
			}

			if len(toRemove) > 0 {
				fmt.Println("Removing old plugins:")
				results, err := inst.Uninstall(toRemove, s, projectRoot)
				if err != nil {
					return err
				}
				for _, r := range results {
					if r.Success {
						fmt.Printf("  ✓ removed %s\n", r.PluginRef)
					} else {
						fmt.Printf("  ✗ %s: %v\n", r.PluginRef, r.Error)
					}
				}
			}

			if len(toAdd) == 0 && len(toRemove) == 0 {
				fmt.Println("Everything is up to date!")
			}

			if s != config.ScopeUser && len(roleConfig.ClaudeMDSnippets) > 0 {
				claudeMDPath, _ := claude.ClaudeMDPath(s, projectRoot)
				claude.UpsertMarkerBlock(claudeMDPath, manifest.Role, roleConfig.ClaudeMDSnippets)
			}

			now := time.Now().UTC()
			manifestPlugins := make([]config.ManifestPlugin, len(roleConfig.Plugins))
			for i, p := range roleConfig.Plugins {
				manifestPlugins[i] = config.ManifestPlugin{
					Name:        p.Name,
					FullRef:     p.Source,
					Required:    p.Required,
					InstalledAt: now,
				}
			}
			manifest.UpdatedAt = now
			manifest.Plugins = manifestPlugins
			config.WriteManifest(manifestPath, manifest)

			fmt.Println("\n✅ ASDS update complete!")
			return nil
		},
	}

	cmd.Flags().StringVar(&scope, "scope", "", "Scope to update: user, project, or local")
	cmd.Flags().StringVar(&projectRoot, "project-root", "", "Override project root directory")

	return cmd
}
