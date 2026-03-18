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

func newInstallCmd() *cobra.Command {
	var (
		role        string
		scope       string
		projectRoot string
		yes         bool
	)

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install ASDS plugins for a role",
		Long:  "Install ASDS plugins for the selected role and scope. Prompts interactively if flags are missing.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Resolve project root
			if projectRoot == "" {
				var err error
				projectRoot, err = claude.FindProjectRoot(".")
				if err != nil {
					return fmt.Errorf("finding project root: %w", err)
				}
			}

			// If role or scope is missing, launch interactive mode
			if role == "" || scope == "" {
				return fmt.Errorf("interactive mode not yet implemented; provide --role and --scope flags")
			}

			// Validate scope
			s, err := config.ParseScope(scope)
			if err != nil {
				return err
			}

			// Validate project root for project/local scope
			if s != config.ScopeUser && projectRoot == "" {
				return fmt.Errorf("no project root found; use --project-root or run from a git repository")
			}

			// Load ASDS config for marketplace URL
			asdsCfg, err := config.ReadASDSConfig(config.ResolveASDSConfigPath())
			if err != nil {
				return fmt.Errorf("reading ASDS config: %w", err)
			}

			// Fetch marketplace config
			mktCfg, err := registry.FetchOrDefault(asdsCfg.MarketplaceURL)
			if err != nil {
				return fmt.Errorf("loading marketplace config: %w", err)
			}

			// Validate role exists
			roleConfig, ok := mktCfg.Roles[role]
			if !ok {
				return fmt.Errorf("unknown role %q; available roles: %v", role, mktCfg.RoleNames())
			}

			// Show confirmation
			if !yes {
				fmt.Printf("Role: %s (%s)\n", roleConfig.DisplayName, roleConfig.Description)
				fmt.Printf("Scope: %s\n", scope)
				fmt.Printf("Plugins (%d):\n", len(roleConfig.Plugins))
				for _, p := range roleConfig.Plugins {
					req := ""
					if p.Required {
						req = " (required)"
					}
					fmt.Printf("  - %s%s\n", p.Name, req)
				}
				fmt.Println()
			}

			// Create installer
			inst := installer.NewInstaller(true)

			// Register marketplace
			if err := inst.RegisterMarketplace(mktCfg.Marketplace.Name, mktCfg.Marketplace.RegistryURL); err != nil {
				fmt.Printf("Warning: marketplace registration failed: %v\n", err)
			}

			// Install plugins
			results, err := inst.Install(roleConfig.Plugins, s, projectRoot)
			if err != nil {
				return fmt.Errorf("installing plugins: %w", err)
			}

			// Report results
			var failures int
			for _, r := range results {
				if r.Success {
					fmt.Printf("  ✓ %s\n", r.PluginRef)
				} else {
					fmt.Printf("  ✗ %s: %v\n", r.PluginRef, r.Error)
					failures++
				}
			}

			// Scaffold CLAUDE.md (project/local scope only)
			if s != config.ScopeUser && len(roleConfig.ClaudeMDSnippets) > 0 {
				claudeMDPath, err := claude.ClaudeMDPath(s, projectRoot)
				if err == nil {
					if err := claude.UpsertMarkerBlock(claudeMDPath, role, roleConfig.ClaudeMDSnippets); err != nil {
						fmt.Printf("Warning: CLAUDE.md update failed: %v\n", err)
					} else {
						fmt.Printf("  ✓ CLAUDE.md updated\n")
					}
				}
			}

			// Write manifest
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

			manifest := &config.Manifest{
				SchemaVersion:      1,
				ASDSVersion:        version,
				InstalledAt:        now,
				UpdatedAt:          now,
				Role:               role,
				Scope:              s,
				MarketplaceSource:  mktCfg.Marketplace.RegistryURL,
				InstallMethod:      inst.Method(),
				ClaudeCodeDetected: installer.DetectClaudeCode().Found,
				Plugins:            manifestPlugins,
				ClaudeMDModified:   s != config.ScopeUser && len(roleConfig.ClaudeMDSnippets) > 0,
				ScaffoldedFiles:    []string{claude.SettingsPath(s, projectRoot)},
			}

			manifestPath := claude.ManifestPath(s, projectRoot)
			if err := config.WriteManifest(manifestPath, manifest); err != nil {
				fmt.Printf("Warning: manifest write failed: %v\n", err)
			}

			// Gitignore for local scope
			if s == config.ScopeLocal {
				claudeDir := claude.SettingsPath(s, projectRoot)
				claude.EnsureGitignore(
					claudeDir[:len(claudeDir)-len("settings.local.json")],
					".asds-manifest.local.json",
				)
			}

			if failures > 0 {
				return fmt.Errorf("%d plugin(s) failed to install", failures)
			}

			fmt.Println("\n✅ ASDS setup complete!")
			return nil
		},
	}

	cmd.Flags().StringVar(&role, "role", "", "Role to install (e.g., developer, frontend, backend)")
	cmd.Flags().StringVar(&scope, "scope", "", "Installation scope: user, project, or local")
	cmd.Flags().StringVar(&projectRoot, "project-root", "", "Override project root directory")
	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")

	return cmd
}
