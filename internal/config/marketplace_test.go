package config_test

import (
	"testing"

	"github.com/andrew-le-mfv/asds-marketplace-setup/internal/config"
)

func TestParseMarketplaceConfig(t *testing.T) {
	yamlData := []byte(`
schema_version: 1
marketplace:
  name: "test-marketplace"
  description: "Test"
  version: "1.0.0"
  registry_url: "github.com/test/marketplace"
roles:
  developer:
    display_name: "Software Developer"
    description: "Full-stack development"
    plugins:
      - name: "code-reviewer"
        source: "code-reviewer@test-marketplace"
        required: true
      - name: "commit-commands"
        source: "commit-commands@test-marketplace"
        required: false
    claude_md_snippets:
      - "Follow conventional commits"
      - "Always write tests"
defaults:
  scope: project
  auto_register_marketplace: true
`)

	cfg, err := config.ParseMarketplaceConfig(yamlData)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.SchemaVersion != 1 {
		t.Errorf("schema_version = %d, want 1", cfg.SchemaVersion)
	}
	if cfg.Marketplace.Name != "test-marketplace" {
		t.Errorf("marketplace.name = %q, want %q", cfg.Marketplace.Name, "test-marketplace")
	}
	if len(cfg.Roles) != 1 {
		t.Fatalf("roles count = %d, want 1", len(cfg.Roles))
	}

	dev, ok := cfg.Roles["developer"]
	if !ok {
		t.Fatal("missing role 'developer'")
	}
	if dev.DisplayName != "Software Developer" {
		t.Errorf("display_name = %q, want %q", dev.DisplayName, "Software Developer")
	}
	if len(dev.Plugins) != 2 {
		t.Fatalf("plugins count = %d, want 2", len(dev.Plugins))
	}
	if dev.Plugins[0].Name != "code-reviewer" {
		t.Errorf("plugin[0].name = %q, want %q", dev.Plugins[0].Name, "code-reviewer")
	}
	if !dev.Plugins[0].Required {
		t.Error("plugin[0].required = false, want true")
	}
	if len(dev.ClaudeMDSnippets) != 2 {
		t.Errorf("claude_md_snippets count = %d, want 2", len(dev.ClaudeMDSnippets))
	}
	if cfg.Defaults.Scope != "project" {
		t.Errorf("defaults.scope = %q, want %q", cfg.Defaults.Scope, "project")
	}
}

func TestParseMarketplaceConfig_InvalidYAML(t *testing.T) {
	_, err := config.ParseMarketplaceConfig([]byte(`{{{invalid`))
	if err == nil {
		t.Error("expected error for invalid YAML, got nil")
	}
}

func TestMarketplaceConfig_RoleNames(t *testing.T) {
	yamlData := []byte(`
schema_version: 1
marketplace:
  name: "test"
  version: "1.0.0"
  registry_url: "github.com/test"
roles:
  developer:
    display_name: "Developer"
    description: "Dev"
  frontend:
    display_name: "Frontend"
    description: "FE"
defaults:
  scope: project
`)
	cfg, err := config.ParseMarketplaceConfig(yamlData)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	names := cfg.RoleNames()
	if len(names) != 2 {
		t.Fatalf("role names count = %d, want 2", len(names))
	}
	// RoleNames returns sorted keys
	if names[0] != "developer" || names[1] != "frontend" {
		t.Errorf("role names = %v, want [developer frontend]", names)
	}
}

func TestMarketplaceConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr bool
	}{
		{
			name: "valid config",
			yaml: `
schema_version: 1
marketplace:
  name: "test"
  version: "1.0.0"
  registry_url: "github.com/test"
roles:
  developer:
    display_name: "Developer"
    description: "Dev"
    plugins:
      - name: "plugin-a"
        source: "plugin-a@test"
defaults:
  scope: project
`,
			wantErr: false,
		},
		{
			name: "missing schema_version",
			yaml: `
marketplace:
  name: "test"
  version: "1.0.0"
  registry_url: "github.com/test"
roles:
  developer:
    display_name: "Developer"
    description: "Dev"
defaults:
  scope: project
`,
			wantErr: true,
		},
		{
			name: "no roles",
			yaml: `
schema_version: 1
marketplace:
  name: "test"
  version: "1.0.0"
  registry_url: "github.com/test"
defaults:
  scope: project
`,
			wantErr: true,
		},
		{
			name: "missing marketplace name",
			yaml: `
schema_version: 1
marketplace:
  version: "1.0.0"
  registry_url: "github.com/test"
roles:
  developer:
    display_name: "Developer"
    description: "Dev"
defaults:
  scope: project
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := config.ParseMarketplaceConfig([]byte(tt.yaml))
			if err != nil {
				if !tt.wantErr {
					t.Fatalf("unexpected parse error: %v", err)
				}
				return
			}
			err = cfg.Validate()
			if tt.wantErr && err == nil {
				t.Error("expected validation error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected validation error: %v", err)
			}
		})
	}
}
