package skills_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rd-mg/architect-ai/internal/skills"
)

func TestWriteToAgentDirs(t *testing.T) {
	homeDir := t.TempDir()
	content := []byte("# test skill\n\nsome content")

	err := skills.WriteToAgentDirs(homeDir, "test-skill", content)
	if err != nil {
		t.Fatalf("WriteToAgentDirs: %v", err)
	}

	// Verify SKILL.md was written under some agent's skills directory.
	// We check a few common paths.
	opencodeDir := filepath.Join(homeDir, ".config", "opencode", "skills", "test-skill")
	data, err := os.ReadFile(filepath.Join(opencodeDir, "SKILL.md"))
	if err != nil {
		t.Fatalf("reading SKILL.md from opencode: %v", err)
	}
	if string(data) != string(content) {
		t.Errorf("content mismatch: got %q, want %q", string(data), string(content))
	}
}

func TestWriteToAgentDirs_EmptyContent(t *testing.T) {
	homeDir := t.TempDir()
	err := skills.WriteToAgentDirs(homeDir, "empty-skill", []byte{})
	if err != nil {
		t.Fatalf("WriteToAgentDirs with empty content: %v", err)
	}
}

func TestRemoveFromAgentDirs(t *testing.T) {
	homeDir := t.TempDir()
	content := []byte("# test")

	// Write first.
	if err := skills.WriteToAgentDirs(homeDir, "temp-skill", content); err != nil {
		t.Fatalf("WriteToAgentDirs: %v", err)
	}

	// Then remove.
	if err := skills.RemoveFromAgentDirs(homeDir, "temp-skill"); err != nil {
		t.Fatalf("RemoveFromAgentDirs: %v", err)
	}

	// Verify the skill directory is gone from at least one agent.
	opencodeDir := filepath.Join(homeDir, ".config", "opencode", "skills", "temp-skill")
	if _, err := os.Stat(opencodeDir); !os.IsNotExist(err) {
		t.Errorf("expected skill dir to be removed, stat err = %v", err)
	}
}

func TestRemoveFromAgentDirs_NonExistent(t *testing.T) {
	homeDir := t.TempDir()
	// Removing a non-existent skill should not error.
	err := skills.RemoveFromAgentDirs(homeDir, "does-not-exist")
	if err != nil {
		t.Errorf("RemoveFromAgentDirs on non-existent skill: %v", err)
	}
}
