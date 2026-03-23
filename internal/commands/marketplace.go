package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/andrew-le-mfv/asds-marketplace-setup/internal/config"
	"github.com/andrew-le-mfv/asds-marketplace-setup/pkg/registry"
)

func newMarketplaceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "marketplace",
		Short: "Manage marketplace sources",
		Long:  "List, add, edit, and remove marketplace sources.",
	}

	cmd.AddCommand(newMarketplaceListCmd())
	cmd.AddCommand(newMarketplaceAddCmd())
	cmd.AddCommand(newMarketplaceRemoveCmd())
	cmd.AddCommand(newMarketplaceUpdateCmd())

	return cmd
}

func newMarketplaceListCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List configured marketplaces",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgPath := config.ResolveMarketplacesConfigPath()
			cfg, err := config.ReadMarketplacesConfig(cfgPath)
			if err != nil {
				return fmt.Errorf("reading marketplaces config: %w", err)
			}

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(cfg.Marketplaces)
			}

			fmt.Println("📦 Configured Marketplaces")
			fmt.Println()
			for _, m := range cfg.Marketplaces {
				enabled := "✓"
				if !m.Enabled {
					enabled = "✗"
				}
				fmt.Printf("  [%s] %s — %s\n", enabled, m.Name, m.URL)
			}
			fmt.Printf("\nConfig: %s\n", cfgPath)
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	return cmd
}

func newMarketplaceAddCmd() *cobra.Command {
	var (
		name    string
		url     string
		noCheck bool
	)

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a marketplace source",
		RunE: func(cmd *cobra.Command, args []string) error {
			if url == "" {
				return fmt.Errorf("--url is required")
			}

			cfgPath := config.ResolveMarketplacesConfigPath()
			cfg, err := config.ReadMarketplacesConfig(cfgPath)
			if err != nil {
				return fmt.Errorf("reading marketplaces config: %w", err)
			}

			if name == "" && !noCheck {
				discovered, discErr := registry.DiscoverMarketplace(url, "")
				if discErr == nil {
					name = discovered.Marketplace.Name
					fmt.Printf("  ✓ Discovered marketplace: %s\n", name)
				}
			}
			if name == "" {
				return fmt.Errorf("--name is required (could not auto-discover)")
			}

			entry := config.MarketplaceEntry{Name: name, URL: url, Enabled: true}
			if err := cfg.AddMarketplace(entry); err != nil {
				return err
			}

			if err := config.WriteMarketplacesConfig(cfgPath, cfg); err != nil {
				return err
			}

			fmt.Printf("  ✓ Added marketplace %q (%s)\n", name, url)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Marketplace name (auto-discovered if omitted)")
	cmd.Flags().StringVar(&url, "url", "", "Marketplace URL")
	cmd.Flags().BoolVar(&noCheck, "no-check", false, "Skip marketplace validation")
	return cmd
}

func newMarketplaceRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove [name]",
		Short: "Remove a marketplace source",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			cfgPath := config.ResolveMarketplacesConfigPath()
			cfg, err := config.ReadMarketplacesConfig(cfgPath)
			if err != nil {
				return fmt.Errorf("reading marketplaces config: %w", err)
			}

			if err := cfg.RemoveMarketplace(name); err != nil {
				return err
			}

			if err := config.WriteMarketplacesConfig(cfgPath, cfg); err != nil {
				return err
			}

			fmt.Printf("  ✓ Removed marketplace %q\n", name)
			return nil
		},
	}

	return cmd
}

func newMarketplaceUpdateCmd() *cobra.Command {
	var (
		name string
		url  string
	)

	cmd := &cobra.Command{
		Use:   "update [name]",
		Short: "Update a marketplace source",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			targetName := args[0]

			cfgPath := config.ResolveMarketplacesConfigPath()
			cfg, err := config.ReadMarketplacesConfig(cfgPath)
			if err != nil {
				return fmt.Errorf("reading marketplaces config: %w", err)
			}

			existing := cfg.FindMarketplace(targetName)
			if existing == nil {
				return fmt.Errorf("marketplace %q not found", targetName)
			}

			updated := *existing
			if name != "" {
				updated.Name = name
			}
			if url != "" {
				updated.URL = url
			}

			if err := cfg.UpdateMarketplace(targetName, updated); err != nil {
				return err
			}

			if err := config.WriteMarketplacesConfig(cfgPath, cfg); err != nil {
				return err
			}

			fmt.Printf("  ✓ Updated marketplace %q\n", targetName)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "New marketplace name")
	cmd.Flags().StringVar(&url, "url", "", "New marketplace URL")
	return cmd
}
