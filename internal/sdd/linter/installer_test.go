// internal/sdd/linter/installer_test.go
package linter

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInstallScripts_CreatesFiles(t *testing.T) {
	dir := t.TempDir()
	if err := InstallScripts(dir); err != nil {
		t.Fatalf("InstallScripts: %v", err)
	}
	expected := []string{"scripts/assumption-linter.sh", "scripts/validate-result-contract.sh"}
	for _, f := range expected {
		path := filepath.Join(dir, f)
		info, err := os.Stat(path)
		if os.IsNotExist(err) {
			t.Errorf("script not created: %s", f)
			continue
		}
		if info.Mode()&0111 == 0 {
			t.Errorf("script not executable: %s", f)
		}
	}
}

func TestInstallScripts_Idempotent(t *testing.T) {
	dir := t.TempDir()
	if err := InstallScripts(dir); err != nil { t.Fatal(err) }
	if err := InstallScripts(dir); err != nil { t.Fatal("second install should succeed") }
}
