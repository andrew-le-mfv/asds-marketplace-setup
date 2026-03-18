package config_test

import (
	"testing"

	"github.com/andrew-le-mfv/asds-marketplace-setup/internal/config"
)

func TestDefaultMarketplaceConfig(t *testing.T) {
	cfg, err := config.DefaultMarketplaceConfig()
	if err != nil {
		t.Fatalf("unexpected error loading default config: %v", err)
	}

	if cfg.SchemaVersion != 1 {
		t.Errorf("schema_version = %d, want 1", cfg.SchemaVersion)
	}
	if cfg.Marketplace.Name != "asds-marketplace" {
		t.Errorf("marketplace.name = %q, want %q", cfg.Marketplace.Name, "asds-marketplace")
	}

	expectedRoles := []string{
		"backend", "data-engineer", "developer", "devops",
		"frontend", "pm", "security", "techlead", "tester",
	}
	names := cfg.RoleNames()
	if len(names) != len(expectedRoles) {
		t.Fatalf("role count = %d, want %d", len(names), len(expectedRoles))
	}
	for i, name := range names {
		if name != expectedRoles[i] {
			t.Errorf("role[%d] = %q, want %q", i, name, expectedRoles[i])
		}
	}
}
