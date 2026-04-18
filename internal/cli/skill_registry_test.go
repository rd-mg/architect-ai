package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLayeredSkillScanning(t *testing.T) {
	tmp := t.TempDir()

	// Set up mock .atl structure
	atlDir := filepath.Join(tmp, ".atl")
	os.MkdirAll(filepath.Join(atlDir, "overlays", "mock-overlay", "skills", "mock-skill"), 0755)
	os.WriteFile(filepath.Join(atlDir, "overlays", "mock-overlay", "skills", "mock-skill", "SKILL.md"), []byte("---\nname: mock-overlay-skill\n---"), 0644)
	os.WriteFile(filepath.Join(atlDir, "overlays", "mock-overlay", "manifest.json"), []byte(`{"name":"mock-overlay","activation_state":"active"}`), 0644)

	// Set up system skills in project
	os.MkdirAll(filepath.Join(tmp, ".agent", "skills", "sdd-init"), 0755)
	os.WriteFile(filepath.Join(tmp, ".agent", "skills", "sdd-init", "SKILL.md"), []byte("---\nname: sdd-init\n---"), 0644)

	// Set up shared rules in project
	os.MkdirAll(filepath.Join(tmp, ".agent", "skills", "_shared"), 0755)
	os.WriteFile(filepath.Join(tmp, ".agent", "skills", "_shared", "SKILL.md"), []byte("---\nname: _shared\n---"), 0644)

	// Set up project skills
	projectSkillDir := filepath.Join(tmp, ".agent", "skills", "mock-project-skill")
	os.MkdirAll(projectSkillDir, 0755)
	os.WriteFile(filepath.Join(projectSkillDir, "SKILL.md"), []byte("---\nname: mock-project-skill\n---"), 0644)

	// Set up user skills
	homeDir := t.TempDir()
	userSkillDir := filepath.Join(homeDir, ".gemini", "antigravity", "skills", "mock-user-skill")
	os.MkdirAll(userSkillDir, 0755)
	os.WriteFile(filepath.Join(userSkillDir, "SKILL.md"), []byte("---\nname: mock-user-skill\n---"), 0644)

	// We need to override the home directory for collectUserSkills
	// This is tricky as osUserHomeDir is a variable I added
	oldHomeDir := osUserHomeDir
	osUserHomeDir = func() (string, error) { return homeDir, nil }
	defer func() { osUserHomeDir = oldHomeDir }()
	// We'll just test that it writes the file and contains expected markers
	err := WriteLocalSkillRegistry(tmp)
	if err != nil {
		t.Fatalf("WriteLocalSkillRegistry failed: %v", err)
	}

	registryPath := filepath.Join(atlDir, "skill-registry.md")
	content, err := os.ReadFile(registryPath)
	if err != nil {
		t.Fatalf("failed to read registry file: %v", err)
	}

	markdown := string(content)

	// Verify markers are generated
	markers := []string{
		"<!-- architect-ai:registry:system -->",
		"<!-- architect-ai:registry:sharedrule -->",
		"<!-- architect-ai:registry:project -->",
		"<!-- architect-ai:registry:overlay -->",
		"<!-- architect-ai:registry:user -->",
		"<!-- architect-ai:registry:compact-rules -->",
		"<!-- architect-ai:registry:conventions -->",
	}

	for _, m := range markers {
		if !strings.Contains(markdown, m) {
			t.Errorf("expected marker %q not found in registry", m)
		}
	}
}
