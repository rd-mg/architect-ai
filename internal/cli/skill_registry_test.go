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

func TestWriteLocalSkillRegistry_VersionGating(t *testing.T) {
	tmp := t.TempDir()
	atlDir := filepath.Join(tmp, ".atl")

	// Create Odoo 19 project
	os.WriteFile(filepath.Join(tmp, "__manifest__.py"), []byte("'version': '19.0'"), 0644)

	// Create Overlay with v18 and v19 skills
	overlayRoot := filepath.Join(atlDir, "overlays", "odoo-development-skill")
	os.MkdirAll(filepath.Join(overlayRoot, "skills", "odoo-18"), 0755)
	os.WriteFile(filepath.Join(overlayRoot, "skills", "odoo-18", "SKILL.md"), []byte("---\nname: odoo-18\n---\nCompact Rules 18"), 0644)

	os.MkdirAll(filepath.Join(overlayRoot, "skills", "odoo-19"), 0755)
	os.WriteFile(filepath.Join(overlayRoot, "skills", "odoo-19", "SKILL.md"), []byte("---\nname: odoo-19\n---\nCompact Rules 19"), 0644)

	os.MkdirAll(filepath.Join(overlayRoot, "skills", "odoo-agnostic"), 0755)
	os.WriteFile(filepath.Join(overlayRoot, "skills", "odoo-agnostic", "SKILL.md"), []byte("---\nname: odoo-agnostic\n---\nCompact Rules Agnostic"), 0644)

	// Create manifest
	os.WriteFile(filepath.Join(overlayRoot, "manifest.json"), []byte(`{"name":"odoo-development-skill","activation_state":"active"}`), 0644)

	err := WriteLocalSkillRegistry(tmp)
	if err != nil {
		t.Fatalf("WriteLocalSkillRegistry failed: %v", err)
	}

	content, _ := os.ReadFile(filepath.Join(atlDir, "skill-registry.md"))
	markdown := string(content)

	if !strings.Contains(markdown, "odoo-19") {
		t.Errorf("expected odoo-19 in registry")
	}
	if !strings.Contains(markdown, "odoo-agnostic") {
		t.Errorf("expected odoo-agnostic in registry")
	}
	// This is the RED part: currently it WILL contain odoo-18
	if strings.Contains(markdown, "odoo-18") {
		t.Errorf("did NOT expect odoo-18 in registry for v19 project")
	}
}
