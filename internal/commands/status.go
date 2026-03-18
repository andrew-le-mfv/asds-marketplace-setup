package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/your-org/asds-marketplace-setup/internal/claude"
	"github.com/your-org/asds-marketplace-setup/internal/config"
	"github.com/your-org/asds-marketplace-setup/internal/installer"
)

func newStatusCmd() *cobra.Command {
	var (
		jsonOutput  bool
		projectRoot string
	)

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show current ASDS setup status",
		RunE: func(cmd *cobra.Command, args []string) error {
			if projectRoot == "" {
				var err error
				projectRoot, err = claude.FindProjectRoot(".")
				if err != nil {
					projectRoot = "."
				}
			}

			detection := installer.DetectClaudeCode()

			type scopeStatus struct {
				Scope    string           `json:"scope"`
				Found    bool             `json:"found"`
				Manifest *config.Manifest `json:"manifest,omitempty"`
			}

			statuses := make([]scopeStatus, 0, 3)
			for _, s := range []config.Scope{config.ScopeUser, config.ScopeProject, config.ScopeLocal} {
				mp := claude.ManifestPath(s, projectRoot)
				m, err := config.ReadManifest(mp)
				if err != nil {
					statuses = append(statuses, scopeStatus{Scope: string(s), Found: false})
				} else {
					statuses = append(statuses, scopeStatus{Scope: string(s), Found: true, Manifest: m})
				}
			}

			if jsonOutput {
				output := map[string]any{
					"claude_code_detected": detection.Found,
					"claude_code_path":     detection.Path,
					"project_root":         projectRoot,
					"scopes":               statuses,
				}
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(output)
			}

			fmt.Println("🔍 ASDS Status")
			fmt.Println()
			if detection.Found {
				fmt.Printf("  Claude Code: ✓ detected at %s\n", detection.Path)
			} else {
				fmt.Printf("  Claude Code: ✗ not detected\n")
			}
			fmt.Printf("  Project root: %s\n", projectRoot)
			fmt.Println()

			for _, ss := range statuses {
				if ss.Found {
					fmt.Printf("  [%s] Role: %s | Plugins: %d | Method: %s | Installed: %s\n",
						ss.Scope, ss.Manifest.Role, len(ss.Manifest.Plugins),
						ss.Manifest.InstallMethod, ss.Manifest.InstalledAt.Format("2006-01-02"))
				} else {
					fmt.Printf("  [%s] Not installed\n", ss.Scope)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	cmd.Flags().StringVar(&projectRoot, "project-root", "", "Override project root directory")

	return cmd
}
