package skills

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/rd-mg/architect-ai/internal/model"
)

const LockfileVersion = 1

type SkillManifest struct {
	Version int                   `json:"version"`
	Skills  []SkillManifestEntry  `json:"skills"`
}

type SkillManifestEntry struct {
	ID         string          `json:"id"`
	Name       string          `json:"name"`
	Source     string          `json:"source"` // "project", "community", "overlay/{name}", "builtin"
	Path       string          `json:"path"`
	SHA        string          `json:"sha"`
	Trigger    string          `json:"trigger,omitempty"`
	Kind       string          `json:"kind"` // "System", "Project", "Overlay", "User", "Community"
	Deprecated bool            `json:"deprecated,omitempty"`
	InstalledAt time.Time      `json:"installed_at,omitempty"`
	Agents     []model.AgentID `json:"agents,omitempty"`
}

func lockfilePath(projectRoot string) string {
	return filepath.Join(projectRoot, "skills-lock.json")
}

func LoadLockfile(projectRoot string) (SkillManifest, error) {
	data, err := os.ReadFile(lockfilePath(projectRoot))
	if errors.Is(err, os.ErrNotExist) {
		return SkillManifest{Version: LockfileVersion}, nil
	}
	if err != nil {
		return SkillManifest{}, err
	}

	// Try new list-based format first.
	var m SkillManifest
	if err := json.Unmarshal(data, &m); err == nil && len(m.Skills) > 0 {
		return m, nil
	}

	// Fall back to legacy map-based format.
	var legacy struct {
		Version int                          `json:"version"`
		Skills  map[string]legacyLockfileEntry `json:"skills"`
	}
	if err := json.Unmarshal(data, &legacy); err != nil {
		return SkillManifest{}, err
	}
	m = SkillManifest{Version: LockfileVersion}
	for id, e := range legacy.Skills {
		source := "builtin"
		if e.SourceType == "community" {
			source = "community"
		}
		kind := "System"
		if source == "community" {
			kind = "Community"
		}
		m.Skills = append(m.Skills, SkillManifestEntry{
			ID:    id,
			Source: source,
			SHA:   e.ComputedHash,
			Kind:  kind,
		})
	}
	return m, nil
}

type legacyLockfileEntry struct {
	Source       string `json:"source"`
	SourceType   string `json:"sourceType"`
	ComputedHash string `json:"computedHash"`
}

func SaveLockfile(projectRoot string, m SkillManifest) error {
	m.Version = LockfileVersion
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(lockfilePath(projectRoot), append(data, '\n'), 0o644)
}

func (m *SkillManifest) Add(e SkillManifestEntry) {
	for i, s := range m.Skills {
		if s.ID == e.ID {
			m.Skills[i] = e
			return
		}
	}
	m.Skills = append(m.Skills, e)
}

func (m *SkillManifest) Remove(id string) bool {
	for i, s := range m.Skills {
		if s.ID == id {
			m.Skills = append(m.Skills[:i], m.Skills[i+1:]...)
			return true
		}
	}
	return false
}

func (m *SkillManifest) FindByID(id string) *SkillManifestEntry {
	for i := range m.Skills {
		if m.Skills[i].ID == id {
			return &m.Skills[i]
		}
	}
	return nil
}

// LoadCommunityManifestAsEntries reads community-skills.json and converts to SkillManifestEntry slice.
func LoadCommunityManifestAsEntries(homeDir string) ([]SkillManifestEntry, error) {
	cm, err := LoadManifest(homeDir)
	if err != nil {
		return nil, err
	}
	var entries []SkillManifestEntry
	for _, e := range cm.Skills {
		entries = append(entries, SkillManifestEntry{
			ID:         e.ID,
			Source:     "community",
			Path:       e.Path,
			SHA:        e.SHA,
			Kind:       "Community",
			InstalledAt: e.InstalledAt,
			Agents:     e.InstalledAgents,
		})
	}
	return entries, nil
}
