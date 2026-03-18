package claude_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andrew-le-mfv/asds-marketplace-setup/internal/claude"
)

func TestUpsertMarkerBlock_NewFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "CLAUDE.md")

	snippets := []string{"Follow conventional commits", "Always write tests"}
	err := claude.UpsertMarkerBlock(path, "developer", snippets)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(path)
	content := string(data)

	if !strings.Contains(content, "<!-- ASDS:BEGIN role=developer -->") {
		t.Error("missing begin marker")
	}
	if !strings.Contains(content, "<!-- ASDS:END -->") {
		t.Error("missing end marker")
	}
	if !strings.Contains(content, "- Follow conventional commits") {
		t.Error("missing snippet")
	}
}

func TestUpsertMarkerBlock_ExistingContent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "CLAUDE.md")

	// Write existing content
	os.WriteFile(path, []byte("# Project\n\nExisting instructions.\n"), 0o644)

	snippets := []string{"New snippet"}
	err := claude.UpsertMarkerBlock(path, "developer", snippets)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(path)
	content := string(data)

	if !strings.Contains(content, "# Project") {
		t.Error("existing content was removed")
	}
	if !strings.Contains(content, "Existing instructions.") {
		t.Error("existing content was removed")
	}
	if !strings.Contains(content, "<!-- ASDS:BEGIN role=developer -->") {
		t.Error("marker block not added")
	}
}

func TestUpsertMarkerBlock_ReplaceExisting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "CLAUDE.md")

	// Write with existing ASDS block
	existing := `# Project

<!-- ASDS:BEGIN role=frontend -->
- Old snippet
<!-- ASDS:END -->

Other content.
`
	os.WriteFile(path, []byte(existing), 0o644)

	snippets := []string{"New developer snippet"}
	err := claude.UpsertMarkerBlock(path, "developer", snippets)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(path)
	content := string(data)

	if strings.Contains(content, "Old snippet") {
		t.Error("old snippet should be replaced")
	}
	if strings.Contains(content, "role=frontend") {
		t.Error("old role marker should be replaced")
	}
	if !strings.Contains(content, "role=developer") {
		t.Error("new role marker not present")
	}
	if !strings.Contains(content, "New developer snippet") {
		t.Error("new snippet not present")
	}
	if !strings.Contains(content, "Other content.") {
		t.Error("surrounding content was removed")
	}
}

func TestRemoveMarkerBlock(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "CLAUDE.md")

	content := `# Project

Some text.

<!-- ASDS:BEGIN role=developer -->
- Follow conventional commits
<!-- ASDS:END -->

More text.
`
	os.WriteFile(path, []byte(content), 0o644)

	err := claude.RemoveMarkerBlock(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(path)
	result := string(data)

	if strings.Contains(result, "ASDS:BEGIN") {
		t.Error("marker block not removed")
	}
	if strings.Contains(result, "conventional commits") {
		t.Error("snippet not removed")
	}
	if !strings.Contains(result, "Some text.") {
		t.Error("surrounding content was removed")
	}
	if !strings.Contains(result, "More text.") {
		t.Error("surrounding content was removed")
	}
}

func TestRemoveMarkerBlock_NoBlock(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "CLAUDE.md")

	os.WriteFile(path, []byte("# Project\nNo ASDS block here.\n"), 0o644)

	err := claude.RemoveMarkerBlock(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(path)
	if string(data) != "# Project\nNo ASDS block here.\n" {
		t.Error("content was modified despite no marker block")
	}
}

func TestHasMarkerBlock(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    bool
	}{
		{
			"with markers",
			"# Project\n\n<!-- ASDS:BEGIN role=developer -->\n- snippet\n<!-- ASDS:END -->\n",
			true,
		},
		{
			"without markers",
			"# No markers here",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := claude.HasMarkerBlock(tt.content)
			if got != tt.want {
				t.Errorf("HasMarkerBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}
