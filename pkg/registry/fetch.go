package registry

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/andrew-le-mfv/asds-marketplace-setup/internal/config"
	"gopkg.in/yaml.v3"
)

const (
	defaultTimeout  = 15 * time.Second
	defaultYAMLFile = "asds-marketplace.yaml"
	pluginYAMLFile  = ".claude-plugin/marketplace.yaml"
	pluginJSONFile  = ".claude-plugin/marketplace.json"
)

// githubAPIBase is the GitHub API base URL. Override in tests via SetGitHubAPIBase.
var githubAPIBase = "https://api.github.com"

// GitHubAPIBase returns the current GitHub API base URL.
func GitHubAPIBase() string { return githubAPIBase }

// SetGitHubAPIBase overrides the GitHub API base URL (for testing).
func SetGitHubAPIBase(base string) { githubAPIBase = base }

var (
	ghTokenOnce  sync.Once
	ghTokenValue string
)

// githubToken returns a GitHub token from environment or gh CLI, cached after first call.
func githubToken() string {
	ghTokenOnce.Do(func() {
		if t := os.Getenv("GH_TOKEN"); t != "" {
			ghTokenValue = t
			return
		}
		if t := os.Getenv("GITHUB_TOKEN"); t != "" {
			ghTokenValue = t
			return
		}
		if out, err := exec.Command("gh", "auth", "token").Output(); err == nil {
			ghTokenValue = strings.TrimSpace(string(out))
		}
	})
	return ghTokenValue
}

// githubGet performs an HTTP GET with GitHub token auth if available.
func githubGet(client *http.Client, url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if token := githubToken(); token != "" {
		req.Header.Set("Authorization", "token "+token)
	}
	return client.Do(req)
}

// BuildRawURL converts a registry URL to a raw content URL.
func BuildRawURL(registryURL string) string {
	// Strip scheme prefix to normalise, but remember if it was http://.
	stripped := registryURL
	scheme := "https://"
	if strings.HasPrefix(stripped, "https://") {
		stripped = strings.TrimPrefix(stripped, "https://")
	} else if strings.HasPrefix(stripped, "http://") {
		stripped = strings.TrimPrefix(stripped, "http://")
		scheme = "http://"
	}

	if strings.HasPrefix(stripped, "github.com/") {
		path := strings.TrimPrefix(stripped, "github.com/")
		return fmt.Sprintf("https://raw.githubusercontent.com/%s/main/%s", path, defaultYAMLFile)
	}

	return scheme + stripped
}

// FetchMarketplaceConfig fetches and parses a marketplace config from a URL.
func FetchMarketplaceConfig(url string) (*config.MarketplaceConfig, error) {
	client := &http.Client{Timeout: defaultTimeout}

	resp, err := githubGet(client, url)
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

	// Try YAML first, then JSON.
	cfg, err := config.ParseMarketplaceConfig(data)
	if err != nil {
		cfg, err = parseJSONMarketplaceConfig(data)
		if err != nil {
			return nil, fmt.Errorf("parsing remote marketplace config: %w", err)
		}
	}

	return cfg, nil
}

// buildGitHubRawURL builds a raw.githubusercontent.com URL for a given repo path, branch, and file.
func buildGitHubRawURL(repoPath, branch, file string) string {
	return fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s", repoPath, branch, file)
}

// githubRepoPath extracts the org/repo path from a URL, stripping scheme and github.com/.
func githubRepoPath(url string) string {
	stripped := strings.TrimPrefix(strings.TrimPrefix(url, "https://"), "http://")
	stripped = strings.TrimPrefix(stripped, "github.com/")
	return strings.TrimSuffix(stripped, "/")
}

// DiscoverMarketplace attempts to fetch a marketplace config from a URL.
// For GitHub URLs it tries multiple file locations (asds-marketplace.yaml,
// .claude-plugin/marketplace.yaml) and branches (main, master).
// cacheName overrides the name used for caching; if empty, the name is derived
// from the URL or the discovered config.
// Returns the parsed config after validation.
func DiscoverMarketplace(url string, cacheName string) (*config.MarketplaceConfig, error) {
	if isGitHubURL(url) {
		repoPath := githubRepoPath(url)
		// Try ASDS marketplace config files first, then .claude-plugin/marketplace.yaml.
		candidates := []string{
			buildGitHubRawURL(repoPath, "main", defaultYAMLFile),
			buildGitHubRawURL(repoPath, "master", defaultYAMLFile),
			buildGitHubRawURL(repoPath, "main", pluginYAMLFile),
			buildGitHubRawURL(repoPath, "master", pluginYAMLFile),
		}

		for _, candidate := range candidates {
			cfg, err := FetchMarketplaceConfig(candidate)
			if err != nil {
				continue
			}
			if err := cfg.Validate(); err != nil {
				continue
			}
			saveName := cfg.Marketplace.Name
			if cacheName != "" {
				saveName = cacheName
			}
			_ = SaveCachedMarketplaceConfig(saveName, cfg)
			return cfg, nil
		}

		// Fallback: discover plugins from the repo's plugins/ directory.
		name := marketplaceNameFromURL(url)
		if cacheName != "" {
			name = cacheName
		}
		cfg, fallbackErr := discoverPluginsFromGitHub(repoPath, name)
		if fallbackErr != nil {
			return nil, fmt.Errorf("discovering marketplace at %q: %w", url, fallbackErr)
		}
		_ = SaveCachedMarketplaceConfig(name, cfg)
		return cfg, nil
	}

	rawURL := BuildRawURL(url)
	cfg, err := FetchMarketplaceConfig(rawURL)
	if err != nil {
		return nil, fmt.Errorf("discovering marketplace at %q: %w", url, err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid marketplace at %q: %w", url, err)
	}

	return cfg, nil
}

// claudePluginJSON represents the .claude-plugin/marketplace.json format.
type claudePluginJSON struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Metadata struct {
		Description string `json:"description"`
	} `json:"metadata"`
	Plugins []struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Source      string `json:"source"`
		Category    string `json:"category"`
	} `json:"plugins"`
}

// parseJSONMarketplaceConfig converts a .claude-plugin/marketplace.json into MarketplaceConfig.
func parseJSONMarketplaceConfig(data []byte) (*config.MarketplaceConfig, error) {
	var j claudePluginJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return nil, fmt.Errorf("parsing JSON marketplace config: %w", err)
	}
	if j.Name == "" || len(j.Plugins) == 0 {
		return nil, fmt.Errorf("invalid JSON marketplace: missing name or plugins")
	}

	// Build a single "default" role containing all plugins.
	var plugins []config.PluginRef
	for _, p := range j.Plugins {
		plugins = append(plugins, config.PluginRef{
			Name:   p.Name,
			Source: p.Source,
		})
	}

	return &config.MarketplaceConfig{
		SchemaVersion: 1,
		Marketplace: config.MarketplaceInfo{
			Name:        j.Name,
			Description: j.Metadata.Description,
			Version:     j.Version,
		},
		Roles: map[string]config.Role{
			"default": {
				DisplayName: "Default",
				Description: "All plugins from this marketplace",
				Plugins:     plugins,
			},
		},
	}, nil
}

// githubContentEntry represents a single item from the GitHub Contents API.
type githubContentEntry struct {
	Name string `json:"name"`
	Type string `json:"type"` // "file" or "dir"
}

// fetchPluginRegistryName fetches .claude-plugin/marketplace.json from a GitHub repo
// and returns the "name" field, which is used as the @<name> suffix for plugin install.
func fetchPluginRegistryName(repoPath string) string {
	client := &http.Client{Timeout: defaultTimeout}
	for _, branch := range []string{"main", "master"} {
		rawURL := buildGitHubRawURL(repoPath, branch, pluginJSONFile)
		resp, err := githubGet(client, rawURL)
		if err != nil {
			continue
		}
		data, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK || readErr != nil {
			continue
		}
		var j struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal(data, &j); err == nil && j.Name != "" {
			return j.Name
		}
	}
	return ""
}

// discoverPluginsFromGitHub scans a GitHub repo for plugin directories.
// It checks each of pluginScanDirs (plugins/, external_plugins/) up to 3 levels
// deep via the GitHub Contents API. A directory is a valid plugin if it contains
// a .claude-plugin/ subdirectory.
func discoverPluginsFromGitHub(repoPath, marketplaceName string) (*config.MarketplaceConfig, error) {
	client := &http.Client{Timeout: defaultTimeout}

	// Get the plugin registry name from .claude-plugin/marketplace.json.
	// Falls back to the repo name (last path segment) if not available.
	registryName := fetchPluginRegistryName(repoPath)
	if registryName == "" {
		parts := strings.Split(repoPath, "/")
		registryName = parts[len(parts)-1]
	}

	seen := make(map[string]bool)
	var plugins []config.PluginRef

	for _, scanDir := range pluginScanDirs {
		found := discoverPluginsInGitHub(client, repoPath, scanDir, registryName, 3, seen)
		plugins = append(plugins, found...)
	}

	if len(plugins) == 0 {
		return nil, fmt.Errorf("no plugin directories found in %s", repoPath)
	}

	return &config.MarketplaceConfig{
		SchemaVersion: 1,
		Marketplace: config.MarketplaceInfo{
			Name:        marketplaceName,
			Description: "Auto-discovered from " + repoPath,
		},
		Roles: map[string]config.Role{
			"default": {
				DisplayName: "Default",
				Description: "All plugins from this marketplace",
				Plugins:     plugins,
			},
		},
	}, nil
}

// discoverPluginsInGitHub recursively lists directories via the GitHub Contents
// API. A directory is a valid plugin if its children include a .claude-plugin/ entry.
func discoverPluginsInGitHub(client *http.Client, repoPath, dirPath, registryName string, maxDepth int, seen map[string]bool) []config.PluginRef {
	if maxDepth <= 0 {
		return nil
	}

	apiURL := fmt.Sprintf("%s/repos/%s/contents/%s", githubAPIBase, repoPath, dirPath)
	resp, err := githubGet(client, apiURL)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	var entries []githubContentEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil
	}

	// Collect subdirectories and check for .claude-plugin marker.
	hasClaudePlugin := false
	var subdirs []githubContentEntry
	for _, e := range entries {
		if e.Type == "dir" {
			if e.Name == ".claude-plugin" {
				hasClaudePlugin = true
			} else {
				subdirs = append(subdirs, e)
			}
		}
	}

	// If this directory has .claude-plugin, it is a plugin itself.
	if hasClaudePlugin {
		name := filepath.Base(dirPath)
		if !seen[name] {
			seen[name] = true
			return []config.PluginRef{{
				Name:   name,
				Source: name + "@" + registryName,
			}}
		}
		return nil
	}

	// Otherwise recurse into subdirectories.
	var plugins []config.PluginRef
	for _, sub := range subdirs {
		childPath := dirPath + "/" + sub.Name
		plugins = append(plugins, discoverPluginsInGitHub(client, repoPath, childPath, registryName, maxDepth-1, seen)...)
	}
	return plugins
}

// pluginScanDirs are the top-level directories scanned for plugins.
var pluginScanDirs = []string{"plugins", "external_plugins"}

// DiscoverLocalPlugins scans a local project for plugin directories.
// It looks in each of pluginScanDirs (plugins/, external_plugins/) up to 3 levels
// deep. Any directory containing a .claude-plugin/ subdirectory is treated as a
// valid plugin. The result is not persisted — discovery runs fresh each time so
// it always reflects the current directory layout.
func DiscoverLocalPlugins(projectRoot string) (*config.MarketplaceConfig, error) {
	// Derive marketplace name from the directory basename.
	name := filepath.Base(projectRoot)

	seen := make(map[string]bool)
	var plugins []config.PluginRef

	for _, scanDir := range pluginScanDirs {
		root := filepath.Join(projectRoot, scanDir)
		found := discoverPluginsIn(root, name, 3, seen)
		plugins = append(plugins, found...)
	}

	if len(plugins) == 0 {
		return nil, fmt.Errorf("no plugin directories found in %s", projectRoot)
	}

	return &config.MarketplaceConfig{
		SchemaVersion: 1,
		Marketplace: config.MarketplaceInfo{
			Name:        name,
			Description: "Auto-discovered from " + projectRoot,
			Version:     "0.0.0",
		},
		Roles: map[string]config.Role{
			"default": {
				DisplayName: "Default",
				Description: "All plugins from this marketplace",
				Plugins:     plugins,
			},
		},
	}, nil
}

// discoverPluginsIn walks dir up to maxDepth levels, collecting directories
// that contain a .claude-plugin/ subdirectory. seen deduplicates by plugin name.
func discoverPluginsIn(dir string, marketplaceName string, maxDepth int, seen map[string]bool) []config.PluginRef {
	if maxDepth <= 0 {
		return nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var plugins []config.PluginRef
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		child := filepath.Join(dir, e.Name())

		// Check if this directory has a .claude-plugin/ subdirectory.
		if info, statErr := os.Stat(filepath.Join(child, ".claude-plugin")); statErr == nil && info.IsDir() {
			if !seen[e.Name()] {
				seen[e.Name()] = true
				plugins = append(plugins, config.PluginRef{
					Name:   e.Name(),
					Source: e.Name() + "@" + marketplaceName,
				})
			}
			continue // don't recurse into a discovered plugin
		}

		// Recurse one level deeper.
		plugins = append(plugins, discoverPluginsIn(child, marketplaceName, maxDepth-1, seen)...)
	}

	return plugins
}

// SaveCachedMarketplaceConfig saves a marketplace config to ~/.config/asds/<name>.yaml.
func SaveCachedMarketplaceConfig(name string, cfg *config.MarketplaceConfig) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("resolving home directory: %w", err)
	}

	dir := filepath.Join(home, ".config", "asds")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling marketplace config: %w", err)
	}

	path := filepath.Join(dir, name+".yaml")
	return os.WriteFile(path, data, 0o644)
}

// CachedMarketplaceConfigPath returns the path to ~/.config/asds/<name>.yaml.
func CachedMarketplaceConfigPath(name string) string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "asds", name+".yaml")
}

// LoadCachedMarketplaceConfig loads a cached marketplace config from ~/.config/asds/<name>.yaml.
func LoadCachedMarketplaceConfig(name string) (*config.MarketplaceConfig, error) {
	data, err := os.ReadFile(CachedMarketplaceConfigPath(name))
	if err != nil {
		return nil, err
	}
	return config.ParseMarketplaceConfig(data)
}

func isGitHubURL(url string) bool {
	stripped := strings.TrimPrefix(strings.TrimPrefix(url, "https://"), "http://")
	return strings.HasPrefix(stripped, "github.com/")
}

// marketplaceNameFromURL derives a marketplace name from a URL (last path segment).
func marketplaceNameFromURL(url string) string {
	stripped := strings.TrimPrefix(strings.TrimPrefix(url, "https://"), "http://")
	stripped = strings.TrimSuffix(stripped, "/")
	parts := strings.Split(stripped, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return "unknown"
}

// FetchOrDefault tries to fetch remote config, falls back to embedded default.
func FetchOrDefault(registryURL string) (*config.MarketplaceConfig, error) {
	rawURL := BuildRawURL(registryURL)
	cfg, err := FetchMarketplaceConfig(rawURL)

	if err != nil && isGitHubURL(registryURL) {
		masterURL := strings.Replace(rawURL, "/main/", "/master/", 1)
		cfg, err = FetchMarketplaceConfig(masterURL)
	}

	if err != nil {
		return config.DefaultMarketplaceConfig()
	}
	return cfg, nil
}
