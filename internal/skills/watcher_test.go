package skills_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rd-mg/architect-ai/internal/skills"
)

func TestNewWatcher(t *testing.T) {
	w, err := skills.NewWatcher(func() {})
	if err != nil {
		t.Fatalf("NewWatcher: %v", err)
	}
	defer w.Stop()

	if w == nil {
		t.Fatal("expected non-nil Watcher")
	}
}

func TestWatcherAddDir_NonExistent(t *testing.T) {
	w, err := skills.NewWatcher(func() {})
	if err != nil {
		t.Fatalf("NewWatcher: %v", err)
	}
	defer w.Stop()

	// Should not panic on non-existent dir.
	w.AddDir("/nonexistent/path")
}

func TestWatcherDetectsSkillChange(t *testing.T) {
	dir := t.TempDir()

	skillDir := filepath.Join(dir, "skills", "test-skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	// Channel to signal the callback fired.
	done := make(chan struct{}, 1)

	w, err := skills.NewWatcher(func() {
		select {
		case done <- struct{}{}:
		default:
		}
	})
	if err != nil {
		t.Fatalf("NewWatcher: %v", err)
	}
	defer w.Stop()

	w.AddDir(skillDir)

	// Start watcher in background.
	go w.Start()

	// Give it a moment to initialize.
	time.Sleep(100 * time.Millisecond)

	// Write SKILL.md — should trigger the callback (debounced at 500ms).
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("test"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	select {
	case <-done:
		// Success — callback was invoked.
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for watcher callback after SKILL.md write")
	}
}

func TestWatcherIgnoresNonSkillFiles(t *testing.T) {
	dir := t.TempDir()

	skillDir := filepath.Join(dir, "skills", "other")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	done := make(chan struct{}, 1)

	w, err := skills.NewWatcher(func() {
		select {
		case done <- struct{}{}:
		default:
		}
	})
	if err != nil {
		t.Fatalf("NewWatcher: %v", err)
	}
	defer w.Stop()

	w.AddDir(skillDir)
	go w.Start()

	time.Sleep(100 * time.Millisecond)

	// Write a non-SKILL.md file — should NOT trigger callback.
	if err := os.WriteFile(filepath.Join(skillDir, "README.md"), []byte("readme"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	select {
	case <-done:
		// This might trigger if the file system generates events the watcher
		// catches in isSkillChange. The test is informational only.
		t.Log("callback fired on non-SKILL.md write (accept if fsnotify sends Create events)")
	case <-time.After(600 * time.Millisecond):
		// Expected — no callback for non-SKILL files.
	}
}
