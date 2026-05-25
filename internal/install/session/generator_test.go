package session

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitialSession_ContainsRequiredFields(t *testing.T) {
	content := InitialSession("my-project")
	required := []string{
		"version:", "project:", "execution_mode:", "delivery_strategy:",
		"artifact_store:", "tdd_mode:", "active_change:", "mcp:", "history:",
	}
	for _, field := range required {
		if !strings.Contains(content, field) {
			t.Errorf("missing field: %s", field)
		}
	}
}

func TestInitialSession_DefaultValues(t *testing.T) {
	content := InitialSession("test")
	if !strings.Contains(content, `execution_mode: "interactive"`) {
		t.Error("default execution_mode should be interactive")
	}
	if !strings.Contains(content, `delivery_strategy: "ask-on-risk"`) {
		t.Error("default delivery_strategy should be ask-on-risk")
	}
	if !strings.Contains(content, `artifact_store: "hybrid"`) {
		t.Error("default artifact_store should be hybrid")
	}
}

func TestWriteSession_Atomic(t *testing.T) {
	dir := t.TempDir()
	content := InitialSession("test")
	if err := WriteSession(dir, content); err != nil {
		t.Fatalf("WriteSession: %v", err)
	}
	path := filepath.Join(dir, "session.yaml")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("session.yaml not created")
	}
	if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
		t.Error("tmp file should be removed after atomic write")
	}
}

func TestInstallRollbackScripts_CreatesAllScripts(t *testing.T) {
	dir := t.TempDir()
	if err := InstallRollbackScripts(dir); err != nil {
		t.Fatalf("InstallRollbackScripts: %v", err)
	}
	expected := []string{
		"scripts/rollback-apply.sh",
		"scripts/resolve-task-order.py",
		"scripts/backup-before-mutate.sh",
	}
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
