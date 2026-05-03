package skills_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rd-mg/architect-ai/internal/components/skills"
)

func TestWriteAgentsIndex_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	entries := []skills.SkillEntry{
		{Name: "sdd-apply", Description: "Implement tasks", Path: ".agent/skills/sdd-apply/SKILL.md"},
	}
	if err := skills.WriteAgentsIndex(dir, entries); err != nil {
		t.Fatal(err)
	}
	content, _ := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if !strings.Contains(string(content), "sdd-apply") {
		t.Error("expected sdd-apply in AGENTS.md")
	}
}

func TestWriteAgentsIndex_RefreshesMarkers(t *testing.T) {
	dir := t.TempDir()
	initial := "# My Project\n\nUser content.\n\n" +
		"<!-- architect-ai:skills-index:start -->\n" +
		"| Old | Old | Old |\n|---|---|---|\n| old | old | old |\n" +
		"<!-- architect-ai:skills-index:end -->\n\nMore user content.\n"
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte(initial), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := skills.WriteAgentsIndex(dir, []skills.SkillEntry{{Name: "new-skill", Description: "New", Path: ".agent/skills/new/SKILL.md"}}); err != nil {
		t.Fatal(err)
	}
	content, _ := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	s := string(content)
	if strings.Contains(s, "Old") {
		t.Error("old content should be replaced")
	}
	if !strings.Contains(s, "new-skill") {
		t.Error("new content should appear")
	}
	if !strings.Contains(s, "# My Project") || !strings.Contains(s, "More user content.") {
		t.Error("surrounding content must be preserved")
	}
}

func TestParseFrontmatter(t *testing.T) {
	content := []byte("---\nname: sdd-apply\ndescription: >\n  Implement tasks. More detail.\n---\n# Content\n")
	name, desc := skills.ParseFrontmatter(content)
	if name != "sdd-apply" {
		t.Errorf("expected name=sdd-apply, got %q", name)
	}
	if desc == "" {
		t.Error("expected non-empty description")
	}
}

func TestParseFrontmatter_NoFrontmatter(t *testing.T) {
	name, desc := skills.ParseFrontmatter([]byte("# Just markdown\n"))
	if name != "" || desc != "" {
		t.Errorf("expected empty strings, got name=%q desc=%q", name, desc)
	}
}
