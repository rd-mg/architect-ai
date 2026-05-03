package skills

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rd-mg/architect-ai/internal/agents"
)

// WriteToAgentDirs writes the skill content (SKILL.md) to all active agent skill directories.
func WriteToAgentDirs(homeDir, skillID string, content []byte) error {
	reg, err := agents.NewDefaultRegistry()
	if err != nil {
		return err
	}

	written := 0
	for _, id := range reg.SupportedAgents() {
		adapter, ok := reg.Get(id)
		if !ok || !adapter.SupportsSkills() {
			continue
		}

		dir := adapter.SkillsDir(homeDir)
		if dir == "" {
			continue
		}

		skillDir := filepath.Join(dir, skillID)
		if err := os.MkdirAll(skillDir, 0o755); err != nil {
			return fmt.Errorf("create skill directory %s: %w", skillDir, err)
		}

		if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), content, 0o644); err != nil {
			return fmt.Errorf("write SKILL.md to %s: %w", skillDir, err)
		}
		written++
	}

	if written == 0 {
		return fmt.Errorf("no agent skill directories found to write to")
	}
	return nil
}

// RemoveFromAgentDirs removes the skill directory from all active agent skill directories.
func RemoveFromAgentDirs(homeDir, skillID string) error {
	reg, err := agents.NewDefaultRegistry()
	if err != nil {
		return err
	}

	for _, id := range reg.SupportedAgents() {
		adapter, ok := reg.Get(id)
		if !ok || !adapter.SupportsSkills() {
			continue
		}

		dir := adapter.SkillsDir(homeDir)
		if dir == "" {
			continue
		}

		skillDir := filepath.Join(dir, skillID)
		if _, err := os.Stat(skillDir); err == nil {
			if err := os.RemoveAll(skillDir); err != nil {
				return fmt.Errorf("remove skill directory %s: %w", skillDir, err)
			}
		}
	}
	return nil
}
