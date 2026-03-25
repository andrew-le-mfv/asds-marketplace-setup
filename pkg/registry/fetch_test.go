package registry_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/andrew-le-mfv/asds-marketplace-setup/pkg/registry"
)

func TestFetchMarketplaceConfig_Success(t *testing.T) {
	yamlContent := `
schema_version: 1
marketplace:
  name: "test-marketplace"
  description: "Test"
  version: "1.0.0"
  registry_url: "github.com/test"
roles:
  developer:
    display_name: "Developer"
    description: "Dev"
    plugins:
      - name: "test-plugin"
        source: "test-plugin@test"
        required: true
defaults:
  scope: project
`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/yaml")
		w.Write([]byte(yamlContent))
	}))
	defer server.Close()

	cfg, err := registry.FetchMarketplaceConfig(server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Marketplace.Name != "test-marketplace" {
		t.Errorf("name = %q, want %q", cfg.Marketplace.Name, "test-marketplace")
	}
	if len(cfg.Roles) != 1 {
		t.Errorf("role count = %d, want 1", len(cfg.Roles))
	}
}

func TestFetchMarketplaceConfig_Errors(t *testing.T) {
	tests := []struct {
		name string
		url  func() string
	}{
		{
			name: "HTTP 404",
			url: func() string {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotFound)
				}))
				t.Cleanup(server.Close)
				return server.URL
			},
		},
		{
			name: "invalid YAML",
			url: func() string {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Write([]byte("{{{not yaml"))
				}))
				t.Cleanup(server.Close)
				return server.URL
			},
		},
		{
			name: "unreachable server",
			url: func() string {
				return "http://localhost:1"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := registry.FetchMarketplaceConfig(tt.url())
			if err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestDiscoverMarketplace(t *testing.T) {
	yamlContent := `
schema_version: 1
marketplace:
  name: "discovered"
  description: "Test"
  version: "1.0.0"
  registry_url: "github.com/test"
roles:
  dev:
    display_name: "Dev"
    description: "Dev"
    plugins:
      - name: "p"
        source: "p@test"
        required: true
defaults:
  scope: project
`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/yaml")
		w.Write([]byte(yamlContent))
	}))
	defer server.Close()

	cfg, err := registry.DiscoverMarketplace(server.URL, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Marketplace.Name != "discovered" {
		t.Errorf("name = %q, want %q", cfg.Marketplace.Name, "discovered")
	}
}

func TestDiscoverMarketplace_InvalidURL(t *testing.T) {
	_, err := registry.DiscoverMarketplace("http://localhost:1", "")
	if err == nil {
		t.Error("expected error for unreachable URL")
	}
}

func TestDiscoverMarketplace_InvalidResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not yaml at all {{{"))
	}))
	defer server.Close()

	_, err := registry.DiscoverMarketplace(server.URL, "")
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestDiscoverPluginsFromGitHub_NestedWithMarker(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/repos/test-org/test-repo/contents/plugins":
			w.Write([]byte(`[
				{"name": "sample-plugin", "type": "dir"},
				{"name": "asds", "type": "dir"},
				{"name": "README.md", "type": "file"}
			]`))
		case "/repos/test-org/test-repo/contents/plugins/sample-plugin":
			w.Write([]byte(`[
				{"name": ".claude-plugin", "type": "dir"},
				{"name": "skills", "type": "dir"}
			]`))
		case "/repos/test-org/test-repo/contents/plugins/asds":
			// Group dir without .claude-plugin — should recurse.
			w.Write([]byte(`[
				{"name": "asds-core", "type": "dir"},
				{"name": "asds-deploy", "type": "dir"},
				{"name": "_source", "type": "dir"}
			]`))
		case "/repos/test-org/test-repo/contents/plugins/asds/asds-core":
			w.Write([]byte(`[
				{"name": ".claude-plugin", "type": "dir"},
				{"name": "agents", "type": "dir"}
			]`))
		case "/repos/test-org/test-repo/contents/plugins/asds/asds-deploy":
			w.Write([]byte(`[
				{"name": ".claude-plugin", "type": "dir"},
				{"name": "skills", "type": "dir"}
			]`))
		case "/repos/test-org/test-repo/contents/plugins/asds/_source":
			// No .claude-plugin — should NOT be discovered.
			w.Write([]byte(`[{"name": "hooks", "type": "dir"}]`))
		case "/repos/test-org/test-repo/contents/external_plugins":
			w.Write([]byte(`[{"name": "ext-plugin", "type": "dir"}]`))
		case "/repos/test-org/test-repo/contents/external_plugins/ext-plugin":
			w.Write([]byte(`[
				{"name": ".claude-plugin", "type": "dir"},
				{"name": "skills", "type": "dir"}
			]`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Override GitHub API base for this test.
	oldBase := registry.GitHubAPIBase()
	registry.SetGitHubAPIBase(server.URL)
	defer registry.SetGitHubAPIBase(oldBase)

	cfg, err := registry.DiscoverMarketplace("github.com/test-org/test-repo", "test-mkt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	role, ok := cfg.Roles["default"]
	if !ok {
		t.Fatal("expected 'default' role")
	}

	names := map[string]bool{}
	for _, p := range role.Plugins {
		names[p.Name] = true
	}

	// Should discover: sample-plugin, asds-core, asds-deploy, ext-plugin
	// Should NOT discover: asds (no .claude-plugin), _source (no .claude-plugin)
	expected := []string{"sample-plugin", "asds-core", "asds-deploy", "ext-plugin"}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("expected plugin %q to be discovered, got %v", name, names)
		}
	}
	if len(role.Plugins) != len(expected) {
		t.Errorf("expected %d plugins, got %d: %v", len(expected), len(role.Plugins), role.Plugins)
	}

	// Verify _source and asds were NOT discovered.
	for _, bad := range []string{"asds", "_source"} {
		if names[bad] {
			t.Errorf("did not expect %q to be discovered", bad)
		}
	}
}

func TestBuildRawURL(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{
			"github.com/your-org/asds-marketplace",
			"https://raw.githubusercontent.com/your-org/asds-marketplace/main/asds-marketplace.yaml",
		},
		{
			"https://github.com/your-org/asds-marketplace",
			"https://raw.githubusercontent.com/your-org/asds-marketplace/main/asds-marketplace.yaml",
		},
		{
			"https://example.com/config.yaml",
			"https://example.com/config.yaml",
		},
	}

	for _, tt := range tests {
		got := registry.BuildRawURL(tt.input)
		if got != tt.want {
			t.Errorf("BuildRawURL(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
