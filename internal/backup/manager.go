package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const maxBackupsPerFile = 5

// Manager handles backup and restore of critical config files
type Manager struct {
	BackupDir string
}

// New creates a Manager with the given backup directory
func New(atDir string) *Manager {
	return &Manager{BackupDir: filepath.Join(atDir, "backups")}
}

// BackupBefore creates a timestamped backup of target before mutation
// Returns the backup path, or "" if target doesn't exist yet
func (m *Manager) BackupBefore(targetPath string) (string, error) {
	if err := os.MkdirAll(m.BackupDir, 0755); err != nil {
		return "", fmt.Errorf("create backup dir: %w", err)
	}

	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		return "", nil // nothing to backup — file doesn't exist yet
	}

	ts := time.Now().Format("20060102-150405")
	basename := filepath.Base(targetPath)
	backupPath := filepath.Join(m.BackupDir, fmt.Sprintf("%s.%s.bak", basename, ts))

	src, err := os.ReadFile(targetPath)
	if err != nil {
		return "", fmt.Errorf("read source: %w", err)
	}
	if err := os.WriteFile(backupPath, src, 0644); err != nil {
		return "", fmt.Errorf("write backup: %w", err)
	}

	// Purge old backups (keep last maxBackupsPerFile)
	m.purgeOld(basename)

	return backupPath, nil
}

// Restore restores the most recent backup of a file
func (m *Manager) Restore(targetPath string) error {
	basename := filepath.Base(targetPath)
	backups := m.listBackups(basename)
	if len(backups) == 0 {
		return fmt.Errorf("no backup found for %s", basename)
	}

	// Most recent backup is last after sort
	latest := backups[len(backups)-1]
	src, err := os.ReadFile(filepath.Join(m.BackupDir, latest))
	if err != nil {
		return fmt.Errorf("read backup: %w", err)
	}

	tmp := targetPath + ".restore.tmp"
	if err := os.WriteFile(tmp, src, 0644); err != nil {
		return fmt.Errorf("write restore tmp: %w", err)
	}
	return os.Rename(tmp, targetPath)
}

// ListBackups returns all backups for a file, sorted oldest to newest
func (m *Manager) ListBackups(basename string) []string {
	return m.listBackups(basename)
}

func (m *Manager) listBackups(basename string) []string {
	entries, err := os.ReadDir(m.BackupDir)
	if err != nil {
		return nil
	}
	var backups []string
	prefix := basename + "."
	for _, e := range entries {
		if !e.IsDir() && strings.HasPrefix(e.Name(), prefix) && strings.HasSuffix(e.Name(), ".bak") {
			backups = append(backups, e.Name())
		}
	}
	sort.Strings(backups) // lexicographic = chronological for timestamp format
	return backups
}

func (m *Manager) purgeOld(basename string) {
	backups := m.listBackups(basename)
	for len(backups) > maxBackupsPerFile {
		oldest := filepath.Join(m.BackupDir, backups[0])
		os.Remove(oldest) //nolint
		backups = backups[1:]
	}
}
