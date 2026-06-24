package openspec

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// MigrateV03 scans all state.yaml files under atDir and replaces the obsolete
// "running" status with "in_progress". Writes atomically via Save().
// Returns the list of migrated file paths and any first error encountered.
func MigrateV03(atDir string) (migrated []string, err error) {
	changesDir := filepath.Join(atDir, "openspec", "changes")
	entries, readErr := os.ReadDir(changesDir)
	if readErr != nil {
		if os.IsNotExist(readErr) {
			return nil, nil
		}
		return nil, fmt.Errorf("read changes dir: %w", readErr)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		statePath := filepath.Join(changesDir, entry.Name(), "state.yaml")
		data, readErr := os.ReadFile(statePath)
		if readErr != nil {
			if os.IsNotExist(readErr) {
				continue
			}
			return migrated, fmt.Errorf("read %s: %w", statePath, readErr)
		}

		content := string(data)
		if !strings.Contains(content, `status: "running"`) &&
			!strings.Contains(content, "status: running\n") {
			continue
		}

		fixed := strings.ReplaceAll(content, `status: "running"`, `status: "in_progress"`)
		fixed = strings.ReplaceAll(fixed, "status: running\n", "status: in_progress\n")

		var s State
		if parseErr := yaml.Unmarshal([]byte(fixed), &s); parseErr != nil {
			return migrated, fmt.Errorf("parse migrated %s: %w", statePath, parseErr)
		}

		if saveErr := Save(statePath, &s); saveErr != nil {
			return migrated, fmt.Errorf("save migrated %s: %w", statePath, saveErr)
		}
		migrated = append(migrated, statePath)
	}
	return migrated, nil
}
