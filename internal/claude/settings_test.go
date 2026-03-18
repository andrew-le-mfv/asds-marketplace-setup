package claude_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/andrew-le-mfv/asds-marketplace-setup/internal/claude"
)

func TestReadSettings_NewFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	settings, err := claude.ReadSettings(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if settings == nil {
		t.Fatal("expected empty settings map, got nil")
	}
}

func TestReadSettings_ExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")

	existing := map[string]any{
		"enabledPlugins": map[string]any{
			"existing-plugin@other": true,
		},
		"customKey": "preserve-me",
	}
	data, _ := json.MarshalIndent(existing, "", "  ")
	os.WriteFile(path, data, 0o644)

	settings, err := claude.ReadSettings(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify existing keys are preserved
	if settings["customKey"] != "preserve-me" {
		t.Errorf("customKey not preserved")
	}
}

func TestMergeEnabledPlugins(t *testing.T) {
	settings := map[string]any{
		"enabledPlugins": map[string]any{
			"existing@other": true,
		},
		"customKey": "keep",
	}

	plugins := map[string]bool{
		"code-reviewer@asds":   true,
		"commit-commands@asds": true,
	}

	claude.MergeEnabledPlugins(settings, plugins)

	ep, ok := settings["enabledPlugins"].(map[string]any)
	if !ok {
		t.Fatal("enabledPlugins not a map")
	}

	// Existing plugin preserved
	if ep["existing@other"] != true {
		t.Error("existing plugin was removed")
	}
	// New plugins added
	if ep["code-reviewer@asds"] != true {
		t.Error("code-reviewer not added")
	}
	if ep["commit-commands@asds"] != true {
		t.Error("commit-commands not added")
	}
	// Unrelated key preserved
	if settings["customKey"] != "keep" {
		t.Error("customKey was removed")
	}
}

func TestMergeMarketplaceRegistration(t *testing.T) {
	settings := map[string]any{}

	claude.MergeMarketplaceRegistration(settings, "asds-marketplace", "github.com/your-org/asds-marketplace")

	ekm, ok := settings["extraKnownMarketplaces"].(map[string]any)
	if !ok {
		t.Fatal("extraKnownMarketplaces not a map")
	}

	entry, ok := ekm["asds-marketplace"].(map[string]any)
	if !ok {
		t.Fatal("marketplace entry not a map")
	}

	source, ok := entry["source"].(map[string]any)
	if !ok {
		t.Fatal("source not a map")
	}

	if source["source"] != "github" {
		t.Errorf("source.source = %v, want 'github'", source["source"])
	}
}

func TestWriteSettings(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".claude", "settings.json")

	settings := map[string]any{
		"enabledPlugins": map[string]any{
			"test@asds": true,
		},
	}

	if err := claude.WriteSettings(path, settings); err != nil {
		t.Fatalf("WriteSettings error: %v", err)
	}

	// Verify file exists and is valid JSON
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("file not created: %v", err)
	}

	var loaded map[string]any
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
}

func TestDisablePlugins(t *testing.T) {
	settings := map[string]any{
		"enabledPlugins": map[string]any{
			"code-reviewer@asds":   true,
			"commit-commands@asds": true,
			"existing@other":       true,
		},
	}

	pluginsToDisable := []string{"code-reviewer@asds", "commit-commands@asds"}
	claude.DisablePlugins(settings, pluginsToDisable)

	ep := settings["enabledPlugins"].(map[string]any)

	// ASDS plugins removed
	if _, exists := ep["code-reviewer@asds"]; exists {
		t.Error("code-reviewer should be removed")
	}
	if _, exists := ep["commit-commands@asds"]; exists {
		t.Error("commit-commands should be removed")
	}
	// Non-ASDS plugin preserved
	if ep["existing@other"] != true {
		t.Error("existing plugin should be preserved")
	}
}
