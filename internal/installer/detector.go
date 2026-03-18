package installer

import (
	"os/exec"
)

// DetectionResult holds the result of a Claude Code CLI detection.
type DetectionResult struct {
	Found bool
	Path  string
}

// DetectClaudeCode checks if the Claude Code CLI is available in PATH.
func DetectClaudeCode() DetectionResult {
	path, err := exec.LookPath("claude")
	if err != nil {
		return DetectionResult{Found: false}
	}
	return DetectionResult{Found: true, Path: path}
}

// DetectClaudeCodeAt checks if the Claude Code CLI exists at a specific path.
func DetectClaudeCodeAt(path string) DetectionResult {
	_, err := exec.LookPath(path)
	if err != nil {
		return DetectionResult{Found: false}
	}
	return DetectionResult{Found: true, Path: path}
}
