package backup

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBackupBefore_CreatesBackup(t *testing.T) {
	dir := t.TempDir()
	mgr := New(dir)

	target := filepath.Join(dir, "opencode.json")
	os.WriteFile(target, []byte(`{"test": "value"}`), 0644)

	backupPath, err := mgr.BackupBefore(target)
	if err != nil { t.Fatalf("BackupBefore: %v", err) }
	if backupPath == "" { t.Error("backup path should not be empty") }
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Error("backup file should exist")
	}
}

func TestBackupBefore_NoFileYet(t *testing.T) {
	dir := t.TempDir()
	mgr := New(dir)

	backupPath, err := mgr.BackupBefore(filepath.Join(dir, "nonexistent.json"))
	if err != nil { t.Fatal(err) }
	if backupPath != "" { t.Error("no backup should be created for nonexistent file") }
}

func TestBackupBefore_PurgesOldBackups(t *testing.T) {
	dir := t.TempDir()
	mgr := New(dir)

	target := filepath.Join(dir, "test.json")
	for i := 0; i < 7; i++ {
		os.WriteFile(target, []byte(`{}`), 0644)
		mgr.BackupBefore(target)
	}

	backups := mgr.ListBackups("test.json")
	if len(backups) > maxBackupsPerFile {
		t.Errorf("should keep max %d backups, got %d", maxBackupsPerFile, len(backups))
	}
}

func TestRestore_RestoresMostRecent(t *testing.T) {
	dir := t.TempDir()
	mgr := New(dir)

	target := filepath.Join(dir, "config.json")
	os.WriteFile(target, []byte(`{"version": "1"}`), 0644)
	mgr.BackupBefore(target)

	// Corrupt the file
	os.WriteFile(target, []byte(`{CORRUPT`), 0644)

	if err := mgr.Restore(target); err != nil { t.Fatalf("Restore: %v", err) }

	data, _ := os.ReadFile(target)
	if string(data) != `{"version": "1"}` {
		t.Errorf("restored content incorrect: %s", string(data))
	}
}
