package installer_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/andrew-le-mfv/asds-marketplace-setup/internal/config"
	"github.com/andrew-le-mfv/asds-marketplace-setup/internal/installer"
)

func TestDirectInstaller_Install(t *testing.T) {
	projectRoot := t.TempDir()
	claudeDir := filepath.Join(projectRoot, ".claude")
	os.MkdirAll(claudeDir, 0o755)

	inst := &installer.DirectInstaller{}

	plugins := []config.PluginRef{
		{Name: "code-reviewer", Source: "code-reviewer@asds", Required: true},
		{Name: "commit-commands", Source: "commit-commands@asds", Required: false},
	}

	results, err := inst.Install(plugins, config.ScopeProject, projectRoot)
	if err != nil {
		t.Fatalf("Install error: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("results count = %d, want 2", len(results))
	}

	for _, r := range results {
		if !r.Success {
			t.Errorf("plugin %q failed: %v", r.PluginRef, r.Error)
		}
	}

	// Verify settings file was written
	settingsPath := filepath.Join(claudeDir, "settings.json")
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("settings file not created: %v", err)
	}

	var settings map[string]any
	json.Unmarshal(data, &settings)

	ep, ok := settings["enabledPlugins"].(map[string]any)
	if !ok {
		t.Fatal("enabledPlugins not found")
	}
	if ep["code-reviewer@asds"] != true {
		t.Error("code-reviewer not enabled")
	}
}

func TestDirectInstaller_Install_PreservesExisting(t *testing.T) {
	projectRoot := t.TempDir()
	claudeDir := filepath.Join(projectRoot, ".claude")
	os.MkdirAll(claudeDir, 0o755)

	// Write existing settings
	existing := map[string]any{
		"enabledPlugins": map[string]any{
			"other-plugin@somewhere": true,
		},
		"customSetting": "preserve",
	}
	data, _ := json.MarshalIndent(existing, "", "  ")
	os.WriteFile(filepath.Join(claudeDir, "settings.json"), data, 0o644)

	inst := &installer.DirectInstaller{}
	plugins := []config.PluginRef{
		{Name: "code-reviewer", Source: "code-reviewer@asds", Required: true},
	}

	_, err := inst.Install(plugins, config.ScopeProject, projectRoot)
	if err != nil {
		t.Fatalf("Install error: %v", err)
	}

	data, _ = os.ReadFile(filepath.Join(claudeDir, "settings.json"))
	var settings map[string]any
	json.Unmarshal(data, &settings)

	if settings["customSetting"] != "preserve" {
		t.Error("existing setting was overwritten")
	}

	ep := settings["enabledPlugins"].(map[string]any)
	if ep["other-plugin@somewhere"] != true {
		t.Error("existing plugin was removed")
	}
	if ep["code-reviewer@asds"] != true {
		t.Error("new plugin not added")
	}
}

func TestDirectInstaller_Uninstall(t *testing.T) {
	projectRoot := t.TempDir()
	claudeDir := filepath.Join(projectRoot, ".claude")
	os.MkdirAll(claudeDir, 0o755)

	// Setup: install first
	settings := map[string]any{
		"enabledPlugins": map[string]any{
			"code-reviewer@asds":   true,
			"commit-commands@asds": true,
			"other-plugin@other":   true,
		},
	}
	data, _ := json.MarshalIndent(settings, "", "  ")
	os.WriteFile(filepath.Join(claudeDir, "settings.json"), data, 0o644)

	inst := &installer.DirectInstaller{}
	refs := []string{"code-reviewer@asds", "commit-commands@asds"}

	results, err := inst.Uninstall(refs, config.ScopeProject, projectRoot)
	if err != nil {
		t.Fatalf("Uninstall error: %v", err)
	}

	for _, r := range results {
		if !r.Success {
			t.Errorf("uninstall %q failed: %v", r.PluginRef, r.Error)
		}
	}

	data, _ = os.ReadFile(filepath.Join(claudeDir, "settings.json"))
	json.Unmarshal(data, &settings)
	ep := settings["enabledPlugins"].(map[string]any)

	if _, exists := ep["code-reviewer@asds"]; exists {
		t.Error("code-reviewer should be removed")
	}
	if ep["other-plugin@other"] != true {
		t.Error("unrelated plugin should be preserved")
	}
}

func TestDirectInstaller_RegisterMarketplace(t *testing.T) {
	// This writes to user-level settings, so we mock with temp HOME
	origHome := os.Getenv("HOME")
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	claudeDir := filepath.Join(tmpHome, ".claude")
	os.MkdirAll(claudeDir, 0o755)

	inst := &installer.DirectInstaller{}
	err := inst.RegisterMarketplace("asds-marketplace", "github.com/your-org/asds-marketplace")
	if err != nil {
		t.Fatalf("RegisterMarketplace error: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(claudeDir, "settings.json"))
	var settings map[string]any
	json.Unmarshal(data, &settings)

	ekm, ok := settings["extraKnownMarketplaces"].(map[string]any)
	if !ok {
		t.Fatal("extraKnownMarketplaces not found")
	}
	if _, ok := ekm["asds-marketplace"]; !ok {
		t.Error("marketplace not registered")
	}
}

func TestDirectInstaller_Method(t *testing.T) {
	inst := &installer.DirectInstaller{}
	if inst.Method() != "direct" {
		t.Errorf("Method() = %q, want %q", inst.Method(), "direct")
	}
}
