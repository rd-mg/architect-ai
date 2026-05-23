package registry

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateManifest_ContainsTiers(t *testing.T) {
	dir := t.TempDir()
	err := GenerateManifest(dir, nil)
	if err != nil {
		t.Fatalf("GenerateManifest: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "skill-manifest.yaml"))
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	content := string(data)

	for _, section := range []string{"foundation:", "context_activated:", "on_demand:"} {
		if !strings.Contains(content, section) {
			t.Errorf("missing tier section: %s", section)
		}
	}
}

func TestGenerateManifest_NotebookLMIsOnDemand(t *testing.T) {
	dir := t.TempDir()
	err := GenerateManifest(dir, nil)
	if err != nil {
		t.Fatalf("GenerateManifest: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "skill-manifest.yaml"))
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	content := string(data)

	// notebooklm must be in on_demand, NOT in foundation
	onDemandIdx := strings.Index(content, "on_demand:")
	foundationIdx := strings.Index(content, "foundation:")

	nlmIdx := strings.Index(content, "mcp-notebooklm-orchestrator")
	if nlmIdx == -1 {
		t.Error("notebooklm not found in manifest")
	}
	if nlmIdx < onDemandIdx || nlmIdx < foundationIdx {
		if nlmIdx > foundationIdx && nlmIdx < onDemandIdx {
			t.Error("notebooklm is in foundation tier — should be on_demand")
		}
	}
}

func TestGenerateFoundationBlock_CreatesMergedFile(t *testing.T) {
	dir := t.TempDir()

	// Create minimal skill files for testing all 6 foundation skills
	for _, skill := range FoundationSkills {
		skillDir := filepath.Join(dir, "skills", skill.Name)
		if err := os.MkdirAll(skillDir, 0755); err != nil {
			t.Fatalf("MkdirAll: %v", err)
		}
		err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"),
			[]byte("## Compact Rules\n- sample rule for "+skill.Name), 0644)
		if err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
	}

	err := GenerateFoundationBlock(dir)
	if err != nil {
		t.Fatalf("GenerateFoundationBlock: %v", err)
	}

	foundationPath := filepath.Join(dir, "_generated", "foundation.md")
	if _, err := os.Stat(foundationPath); os.IsNotExist(err) {
		t.Error("foundation.md not created")
	}

	data, err := os.ReadFile(foundationPath)
	if err != nil {
		t.Fatalf("read foundation.md: %v", err)
	}
	if !strings.Contains(string(data), "Project Foundation Standards") {
		t.Error("foundation.md missing header")
	}
	if !strings.Contains(string(data), "ripgrep") {
		t.Error("foundation.md missing ripgrep compact rule")
	}
}

func TestFoundationSkills_Has6Skills(t *testing.T) {
	if len(FoundationSkills) != 6 {
		t.Errorf("expected exactly 6 foundation skills, got %d", len(FoundationSkills))
	}
}

func TestFoundationSkills_NoNotebookLM(t *testing.T) {
	for _, s := range FoundationSkills {
		if strings.Contains(s.Name, "notebooklm") {
			t.Errorf("notebooklm should NOT be in foundation skills: %s", s.Name)
		}
	}
}
