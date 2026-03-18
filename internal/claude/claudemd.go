package claude

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	markerBeginPrefix = "<!-- ASDS:BEGIN"
	markerEnd         = "<!-- ASDS:END -->"
)

var markerBlockRegex = regexp.MustCompile(`(?s)<!-- ASDS:BEGIN[^\n]*-->\n.*?<!-- ASDS:END -->\n?`)

// HasMarkerBlock returns true if the content contains an ASDS marker block.
func HasMarkerBlock(content string) bool {
	return markerBlockRegex.MatchString(content)
}

// buildMarkerBlock constructs the marker block text for a role and snippets.
func buildMarkerBlock(role string, snippets []string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "<!-- ASDS:BEGIN role=%s -->\n", role)
	for _, s := range snippets {
		fmt.Fprintf(&b, "- %s\n", s)
	}
	b.WriteString(markerEnd + "\n")
	return b.String()
}

// UpsertMarkerBlock adds or replaces the ASDS marker block in a CLAUDE.md file.
// Creates the file if it doesn't exist.
func UpsertMarkerBlock(path string, role string, snippets []string) error {
	block := buildMarkerBlock(role, snippets)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				return fmt.Errorf("creating directory: %w", err)
			}
			return os.WriteFile(path, []byte(block), 0o644)
		}
		return fmt.Errorf("reading CLAUDE.md: %w", err)
	}

	content := string(data)
	if HasMarkerBlock(content) {
		content = markerBlockRegex.ReplaceAllString(content, block)
	} else {
		if !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
		content += "\n" + block
	}

	return os.WriteFile(path, []byte(content), 0o644)
}

// RemoveMarkerBlock removes the ASDS marker block from a CLAUDE.md file.
// No-op if the file doesn't exist or has no marker block.
func RemoveMarkerBlock(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("reading CLAUDE.md: %w", err)
	}

	content := string(data)
	if !HasMarkerBlock(content) {
		return nil
	}

	content = markerBlockRegex.ReplaceAllString(content, "")
	// Clean up any double blank lines left behind
	for strings.Contains(content, "\n\n\n") {
		content = strings.ReplaceAll(content, "\n\n\n", "\n\n")
	}

	return os.WriteFile(path, []byte(content), 0o644)
}
