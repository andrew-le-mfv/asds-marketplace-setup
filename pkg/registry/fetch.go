package registry

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/your-org/asds-marketplace-setup/internal/config"
)

const (
	defaultTimeout  = 15 * time.Second
	defaultYAMLFile = "asds-marketplace.yaml"
)

// BuildRawURL converts a registry URL to a raw content URL.
func BuildRawURL(registryURL string) string {
	if strings.HasPrefix(registryURL, "http://") || strings.HasPrefix(registryURL, "https://") {
		return registryURL
	}

	if strings.HasPrefix(registryURL, "github.com/") {
		path := strings.TrimPrefix(registryURL, "github.com/")
		return fmt.Sprintf("https://raw.githubusercontent.com/%s/main/%s", path, defaultYAMLFile)
	}

	return "https://" + registryURL
}

// FetchMarketplaceConfig fetches and parses a marketplace config from a URL.
func FetchMarketplaceConfig(url string) (*config.MarketplaceConfig, error) {
	client := &http.Client{Timeout: defaultTimeout}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetching marketplace config: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching marketplace config: HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	cfg, err := config.ParseMarketplaceConfig(data)
	if err != nil {
		return nil, fmt.Errorf("parsing remote marketplace config: %w", err)
	}

	return cfg, nil
}

// FetchOrDefault tries to fetch remote config, falls back to embedded default.
func FetchOrDefault(registryURL string) (*config.MarketplaceConfig, error) {
	rawURL := BuildRawURL(registryURL)
	cfg, err := FetchMarketplaceConfig(rawURL)
	if err != nil {
		return config.DefaultMarketplaceConfig()
	}
	return cfg, nil
}
