package skills

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/rd-mg/architect-ai/internal/model"
)

const ManifestVersion = 1

type CommunitySkillEntry struct {
	ID              string          `json:"id"`
	Source          string          `json:"source"`           // "HKUDS/CLI-Anything"
	Path            string          `json:"path"`             // "skills/gimp"
	SHA             string          `json:"sha"`              // for update detection
	InstalledAt     time.Time       `json:"installed_at"`
	InstalledAgents []model.AgentID `json:"installed_agents"`
}

type CommunityManifest struct {
	Version int                   `json:"version"`
	Skills  []CommunitySkillEntry `json:"skills"`
}

func manifestPath(homeDir string) string {
	return filepath.Join(homeDir, ".config", "architect-ai", "community-skills.json")
}

// LoadManifest reads the manifest. Returns empty manifest if the file does not exist.
func LoadManifest(homeDir string) (CommunityManifest, error) {
	data, err := os.ReadFile(manifestPath(homeDir))
	if errors.Is(err, os.ErrNotExist) {
		return CommunityManifest{Version: ManifestVersion}, nil
	}
	if err != nil {
		return CommunityManifest{}, err
	}
	var m CommunityManifest
	return m, json.Unmarshal(data, &m)
}

// SaveManifest writes the manifest, creating the directory if needed.
func SaveManifest(homeDir string, m CommunityManifest) error {
	path := manifestPath(homeDir)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

func (m *CommunityManifest) Add(e CommunitySkillEntry) {
	for i, s := range m.Skills {
		if s.ID == e.ID {
			m.Skills[i] = e
			return
		}
	}
	m.Skills = append(m.Skills, e)
}

func (m *CommunityManifest) Remove(id string) bool {
	for i, s := range m.Skills {
		if s.ID == id {
			m.Skills = append(m.Skills[:i], m.Skills[i+1:]...)
			return true
		}
	}
	return false
}

func (m *CommunityManifest) FindByID(id string) *CommunitySkillEntry {
	for i := range m.Skills {
		if m.Skills[i].ID == id {
			return &m.Skills[i]
		}
	}
	return nil
}
