package registry_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/your-org/asds-marketplace-setup/pkg/registry"
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
