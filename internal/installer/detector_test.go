package installer_test

import (
	"testing"

	"github.com/your-org/asds-marketplace-setup/internal/installer"
)

func TestDetectClaudeCode(t *testing.T) {
	result := installer.DetectClaudeCode()

	// We can't guarantee Claude Code is installed in test env,
	// but we can verify the function returns a valid result.
	t.Logf("Claude Code detected: %v, path: %q", result.Found, result.Path)

	if result.Found && result.Path == "" {
		t.Error("Found=true but Path is empty")
	}
}

func TestDetectClaudeCode_WithCustomPath(t *testing.T) {
	// Test with a known-nonexistent path
	result := installer.DetectClaudeCodeAt("/nonexistent/claude")
	if result.Found {
		t.Error("expected Found=false for nonexistent path")
	}
}
