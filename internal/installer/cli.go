package installer

import (
	"fmt"
	"os/exec"

	"github.com/andrew-le-mfv/asds-marketplace-setup/internal/config"
)

// CLIInstaller shells out to the Claude Code CLI for plugin management.
type CLIInstaller struct {
	claudePath string
}

// NewCLIInstaller creates a CLIInstaller using the given claude binary path.
func NewCLIInstaller(claudePath string) *CLIInstaller {
	return &CLIInstaller{claudePath: claudePath}
}

// BuildInstallArgs returns the claude CLI args for installing a plugin.
func (c *CLIInstaller) BuildInstallArgs(pluginRef string, scope string) []string {
	return []string{"plugin", "install", pluginRef, "--scope", scope}
}

// BuildUninstallArgs returns the claude CLI args for uninstalling a plugin.
func (c *CLIInstaller) BuildUninstallArgs(pluginRef string, scope string) []string {
	return []string{"plugin", "uninstall", pluginRef, "--scope", scope}
}

// BuildMarketplaceAddArgs returns the claude CLI args for adding a marketplace.
func (c *CLIInstaller) BuildMarketplaceAddArgs(source string) []string {
	return []string{"plugin", "marketplace", "add", source}
}

// Install enables plugins via the Claude Code CLI.
func (c *CLIInstaller) Install(plugins []config.PluginRef, scope config.Scope, projectRoot string) ([]InstallResult, error) {
	results := make([]InstallResult, 0, len(plugins))

	for _, p := range plugins {
		args := c.BuildInstallArgs(p.Source, string(scope))
		cmd := exec.Command(c.claudePath, args...)
		cmd.Dir = projectRoot

		if err := cmd.Run(); err != nil {
			results = append(results, InstallResult{
				PluginRef: p.Source,
				Success:   false,
				Error:     fmt.Errorf("claude plugin install %s: %w", p.Source, err),
			})
		} else {
			results = append(results, InstallResult{
				PluginRef: p.Source,
				Success:   true,
			})
		}
	}

	return results, nil
}

// Uninstall removes plugins via the Claude Code CLI.
func (c *CLIInstaller) Uninstall(pluginRefs []string, scope config.Scope, projectRoot string) ([]InstallResult, error) {
	results := make([]InstallResult, 0, len(pluginRefs))

	for _, ref := range pluginRefs {
		args := c.BuildUninstallArgs(ref, string(scope))
		cmd := exec.Command(c.claudePath, args...)
		cmd.Dir = projectRoot

		if err := cmd.Run(); err != nil {
			results = append(results, InstallResult{
				PluginRef: ref,
				Success:   false,
				Error:     fmt.Errorf("claude plugin uninstall %s: %w", ref, err),
			})
		} else {
			results = append(results, InstallResult{
				PluginRef: ref,
				Success:   true,
			})
		}
	}

	return results, nil
}

// RegisterMarketplace adds a marketplace via the Claude Code CLI.
func (c *CLIInstaller) RegisterMarketplace(name string, registryURL string) error {
	args := c.BuildMarketplaceAddArgs(registryURL)
	cmd := exec.Command(c.claudePath, args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("claude plugin marketplace add: %w", err)
	}
	return nil
}

// Method returns "cli" for manifest tracking.
func (c *CLIInstaller) Method() string {
	return "cli"
}
