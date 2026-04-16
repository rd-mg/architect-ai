package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLayeredSkillScanning(t *testing.T) {
	tmp := t.TempDir()

	// Set up mock `.atl` structure
	atlDir := filepath.Join(tmp, ".atl")
	os.MkdirAll(atlDir, 0755)

	// Set up user skills
	userSkillDir := filepath.Join(tmp, ".gemini", "antigravity", "skills")
	os.MkdirAll(filepath.Join(userSkillDir, "my-custom-skill"), 0755)
	os.WriteFile(filepath.Join(userSkillDir, "my-custom-skill", "SKILL.md"), []byte("---\nname: my-custom-skill\n---"), 0644)

	skills := []skillEntry{
		{Name: "sdd-init", Kind: "System", Origin: "overlay", Trigger: "on project creation"},
		{Name: "clean-arch", Kind: "SharedRule", Origin: "overlay", Trigger: "always"},
		{Name: "my-custom-skill", Kind: "User", Origin: "user", Trigger: "manual"},
		{Name: "project-specific", Kind: "Project", Origin: "project", Trigger: "file save"},
	}

	markdown := buildRegistryMarkdown(tmp, skills, nil, nil)

	// Verify sections are generated
	if !strings.Contains(markdown, "## System Skills") {
		t.Error("expected '## System Skills' section in markdown")
	}
	if !strings.Contains(markdown, "## SharedRule Skills") {
		t.Error("expected '## SharedRule Skills' section in markdown")
	}
	if !strings.Contains(markdown, "## User Skills") {
		t.Error("expected '## User Skills' section in markdown")
	}
	if !strings.Contains(markdown, "## Project Skills") {
		t.Error("expected '## Project Skills' section in markdown")
	}
}
