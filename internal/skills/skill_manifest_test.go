package skills_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rd-mg/architect-ai/internal/model"
	"github.com/rd-mg/architect-ai/internal/skills"
)

func TestLockfile_NotFound(t *testing.T) {
	m, err := skills.LoadLockfile(t.TempDir())
	if err != nil {
		t.Fatalf("LoadLockfile on empty dir: %v", err)
	}
	if m.Version == 0 {
		t.Error("Expected non-zero Version in new lockfile")
	}
}

func TestLockfile_SaveAndLoad(t *testing.T) {
	dir := t.TempDir()

	m := skills.SkillManifest{}
	m.Add(skills.SkillManifestEntry{ID: "ripgrep", Name: "Ripgrep", Source: "builtin", Kind: "System"})
	m.Add(skills.SkillManifestEntry{ID: "bash-expert", Name: "Bash Expert", Source: "builtin", Kind: "System"})

	if err := skills.SaveLockfile(dir, m); err != nil {
		t.Fatalf("SaveLockfile: %v", err)
	}

	loaded, err := skills.LoadLockfile(dir)
	if err != nil {
		t.Fatalf("LoadLockfile: %v", err)
	}
	if got := len(loaded.Skills); got != 2 {
		t.Fatalf("expected 2 skills, got %d", got)
	}
	if loaded.Skills[0].ID != "ripgrep" {
		t.Errorf("first skill ID = %q, want %q", loaded.Skills[0].ID, "ripgrep")
	}
}

func TestLockfile_AddUpdatesExisting(t *testing.T) {
	m := skills.SkillManifest{}
	m.Add(skills.SkillManifestEntry{ID: "rg", SHA: "abc"})
	m.Add(skills.SkillManifestEntry{ID: "rg", SHA: "def"})

	if len(m.Skills) != 1 {
		t.Fatalf("expected 1 entry after duplicate Add, got %d", len(m.Skills))
	}
	if m.Skills[0].SHA != "def" {
		t.Errorf("expected SHA=def after update, got %s", m.Skills[0].SHA)
	}
}

func TestLockfile_Remove(t *testing.T) {
	m := skills.SkillManifest{}
	m.Add(skills.SkillManifestEntry{ID: "a"})
	m.Add(skills.SkillManifestEntry{ID: "b"})

	if !m.Remove("a") {
		t.Error("expected Remove to return true for existing entry")
	}
	if m.FindByID("a") != nil {
		t.Error("entry a should be gone after Remove")
	}
	if m.FindByID("b") == nil {
		t.Error("entry b should remain after Remove(a)")
	}
	if m.Remove("nonexistent") {
		t.Error("expected Remove to return false for unknown ID")
	}
}

func TestLockfile_FindByID(t *testing.T) {
	m := skills.SkillManifest{}
	m.Add(skills.SkillManifestEntry{ID: "find-me"})

	entry := m.FindByID("find-me")
	if entry == nil {
		t.Fatal("FindByID returned nil for existing entry")
	}
	if entry.ID != "find-me" {
		t.Errorf("entry ID = %q, want %q", entry.ID, "find-me")
	}

	if m.FindByID("missing") != nil {
		t.Error("FindByID should return nil for missing ID")
	}
}

func TestLockfile_LegacyFormat(t *testing.T) {
	dir := t.TempDir()
	legacyJSON := `{
		"version": 1,
		"skills": {
			"ripgrep": {"source": "builtins/ripgrep", "sourceType": "builtin", "computedHash": "abc123"},
			"bash-expert": {"source": "builtins/bash-expert", "sourceType": "builtin", "computedHash": "def456"}
		}
	}`
	lockfilePath := filepath.Join(dir, "skills-lock.json")
	if err := os.WriteFile(lockfilePath, []byte(legacyJSON), 0o644); err != nil {
		t.Fatalf("writing legacy lockfile: %v", err)
	}

	loaded, err := skills.LoadLockfile(dir)
	if err != nil {
		t.Fatalf("LoadLockfile legacy: %v", err)
	}
	if got := len(loaded.Skills); got != 2 {
		t.Fatalf("expected 2 skills from legacy format, got %d", got)
	}
	found := map[string]bool{}
	for _, s := range loaded.Skills {
		found[s.ID] = true
	}
	if !found["ripgrep"] || !found["bash-expert"] {
		t.Errorf("legacy parse missing expected skills: %v", found)
	}
}

func TestLockfile_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	lockfilePath := filepath.Join(dir, "skills-lock.json")
	os.WriteFile(lockfilePath, []byte("not json"), 0o644)

	_, err := skills.LoadLockfile(dir)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestLoadCommunityManifestAsEntries(t *testing.T) {
	homeDir := t.TempDir()

	// First add a community skill.
	cm := skills.CommunityManifest{Version: skills.ManifestVersion}
	cm.Add(skills.CommunitySkillEntry{
		ID:              "cli-anything-gimp",
		Source:          "HKUDS/CLI-Anything",
		Path:            "skills/gimp",
		SHA:             "abc123",
		InstalledAt:     time.Now().UTC().Truncate(time.Second),
		InstalledAgents: []model.AgentID{model.AgentID("claude")},
	})
	if err := skills.SaveManifest(homeDir, cm); err != nil {
		t.Fatalf("SaveManifest: %v", err)
	}

	entries, err := skills.LoadCommunityManifestAsEntries(homeDir)
	if err != nil {
		t.Fatalf("LoadCommunityManifestAsEntries: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].ID != "cli-anything-gimp" {
		t.Errorf("entry ID = %q, want %q", entries[0].ID, "cli-anything-gimp")
	}
	if entries[0].Kind != "Community" {
		t.Errorf("entry Kind = %q, want %q", entries[0].Kind, "Community")
	}
	if len(entries[0].Agents) != 1 || entries[0].Agents[0] != model.AgentID("claude") {
		t.Errorf("entry Agents = %v, want [claude]", entries[0].Agents)
	}
}

func TestLoadCommunityManifestAsEntries_NoFile(t *testing.T) {
	entries, err := skills.LoadCommunityManifestAsEntries(t.TempDir())
	if err != nil {
		t.Fatalf("LoadCommunityManifestAsEntries on empty dir: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}
