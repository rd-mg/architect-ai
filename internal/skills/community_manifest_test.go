package skills_test

import (
	"testing"
	"time"

	"github.com/rd-mg/architect-ai/internal/skills"
)

func TestManifest_RoundTrip(t *testing.T) {
	homeDir := t.TempDir()

	m, err := skills.LoadManifest(homeDir)
	if err != nil {
		t.Fatal(err)
	}
	m.Add(skills.CommunitySkillEntry{
		ID:          "cli-anything-gimp",
		Source:      "HKUDS/CLI-Anything",
		Path:        "skills/gimp",
		SHA:         "abc123",
		InstalledAt: time.Now().UTC().Truncate(time.Second),
	})
	if err := skills.SaveManifest(homeDir, m); err != nil {
		t.Fatal(err)
	}

	loaded, err := skills.LoadManifest(homeDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Skills) != 1 || loaded.Skills[0].ID != "cli-anything-gimp" {
		t.Errorf("round-trip failed: %+v", loaded.Skills)
	}
}

func TestManifest_AddUpdatesExisting(t *testing.T) {
	m := skills.CommunityManifest{Version: skills.ManifestVersion}
	m.Add(skills.CommunitySkillEntry{ID: "a", SHA: "old"})
	m.Add(skills.CommunitySkillEntry{ID: "a", SHA: "new"})
	if len(m.Skills) != 1 {
		t.Errorf("expected 1 entry, got %d", len(m.Skills))
	}
	if m.Skills[0].SHA != "new" {
		t.Errorf("expected SHA=new, got %s", m.Skills[0].SHA)
	}
}

func TestManifest_Remove(t *testing.T) {
	m := skills.CommunityManifest{Version: skills.ManifestVersion}
	m.Add(skills.CommunitySkillEntry{ID: "a"})
	m.Add(skills.CommunitySkillEntry{ID: "b"})
	if !m.Remove("a") {
		t.Error("expected Remove to return true")
	}
	if m.FindByID("a") != nil {
		t.Error("a should be gone")
	}
	if m.FindByID("b") == nil {
		t.Error("b should remain")
	}
}

func TestManifest_EmptyOnFirstLoad(t *testing.T) {
	m, err := skills.LoadManifest(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Skills) != 0 {
		t.Errorf("expected 0 skills on first load, got %d", len(m.Skills))
	}
}
