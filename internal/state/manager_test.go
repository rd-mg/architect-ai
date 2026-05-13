package state

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// TestManagerReadAfterWrite verifies Manager.Read returns state written via Manager.Write.
func TestManagerReadAfterWrite(t *testing.T) {
	home := t.TempDir()
	mgr := NewManager(home)

	want := InstallState{
		InstalledAgents: []string{"claude-code", "opencode"},
		ClaudeModelAssignments: map[string]string{
			"orchestrator": "opus",
		},
	}

	if err := mgr.Write(want); err != nil {
		t.Fatalf("Manager.Write() error = %v", err)
	}

	got, err := mgr.Read()
	if err != nil {
		t.Fatalf("Manager.Read() error = %v", err)
	}

	if len(got.InstalledAgents) != len(want.InstalledAgents) {
		t.Errorf("InstalledAgents length = %d, want %d", len(got.InstalledAgents), len(want.InstalledAgents))
	}
	if got.ClaudeModelAssignments["orchestrator"] != "opus" {
		t.Errorf("ClaudeModelAssignments[orchestrator] = %q, want %q", got.ClaudeModelAssignments["orchestrator"], "opus")
	}
}

// TestManagerReadMissing verifies Manager.Read returns error for missing state file.
func TestManagerReadMissing(t *testing.T) {
	home := t.TempDir()
	mgr := NewManager(home)

	_, err := mgr.Read()
	if err == nil {
		t.Fatal("Manager.Read() expected error for missing file, got nil")
	}
}

// TestManagerMergeAtomic verifies Merge reads-modify-writes atomically.
func TestManagerMergeAtomic(t *testing.T) {
	home := t.TempDir()
	mgr := NewManager(home)

	// Write initial state
	initial := InstallState{
		InstalledAgents: []string{"claude-code"},
	}
	if err := mgr.Write(initial); err != nil {
		t.Fatalf("Manager.Write() error = %v", err)
	}

	// Merge should read-merge-write atomically
	err := mgr.Merge(func(s InstallState) InstallState {
		s.InstalledAgents = append(s.InstalledAgents, "opencode")
		s.ClaudeModelAssignments = map[string]string{"sdd-explore": "sonnet"}
		return s
	})
	if err != nil {
		t.Fatalf("Manager.Merge() error = %v", err)
	}

	got, err := mgr.Read()
	if err != nil {
		t.Fatalf("Manager.Read() error = %v", err)
	}

	// Verify merged state has both agents and the model assignment
	if len(got.InstalledAgents) != 2 {
		t.Errorf("InstalledAgents length = %d, want 2", len(got.InstalledAgents))
	}
	if got.ClaudeModelAssignments["sdd-explore"] != "sonnet" {
		t.Errorf("ClaudeModelAssignments[sdd-explore] = %q, want %q", got.ClaudeModelAssignments["sdd-explore"], "sonnet")
	}
}

// TestManagerMergeFromEmpty verifies Merge works when state file doesn't exist yet.
func TestManagerMergeFromEmpty(t *testing.T) {
	home := t.TempDir()
	mgr := NewManager(home)

	// No state file written yet — Merge should treat as empty state
	err := mgr.Merge(func(s InstallState) InstallState {
		s.InstalledAgents = []string{"kiro"}
		return s
	})
	if err != nil {
		t.Fatalf("Manager.Merge() error = %v", err)
	}

	got, err := mgr.Read()
	if err != nil {
		t.Fatalf("Manager.Read() error = %v", err)
	}

	if len(got.InstalledAgents) != 1 || got.InstalledAgents[0] != "kiro" {
		t.Errorf("InstalledAgents = %v, want [kiro]", got.InstalledAgents)
	}
}

// TestManagerConcurrentMerge verifies that 10 goroutines calling Merge concurrently
// produce a consistent final state (no data loss, no corruption).
func TestManagerConcurrentMerge(t *testing.T) {
	home := t.TempDir()
	mgr := NewManager(home)

	// Write initial state
	if err := mgr.Write(InstallState{}); err != nil {
		t.Fatalf("Manager.Write() error = %v", err)
	}

	const goroutines = 10
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			err := mgr.Merge(func(s InstallState) InstallState {
				s.InstalledAgents = append(s.InstalledAgents, "agent-"+string(rune('A'+idx)))
				return s
			})
			if err != nil {
				t.Errorf("Merge goroutine %d error: %v", idx, err)
			}
		}(i)
	}

	wg.Wait()

	got, err := mgr.Read()
	if err != nil {
		t.Fatalf("Manager.Read() after concurrent merges error = %v", err)
	}

	// All 10 agents must be present — no data loss from races
	if len(got.InstalledAgents) != goroutines {
		t.Errorf("InstalledAgents length = %d, want %d (data lost in race)", len(got.InstalledAgents), goroutines)
	}
}

// TestManagerConcurrentReadWrite verifies that concurrent reads and writes
// do not trigger the race detector.
func TestManagerConcurrentReadWrite(t *testing.T) {
	home := t.TempDir()
	mgr := NewManager(home)

	if err := mgr.Write(InstallState{
		InstalledAgents: []string{"initial"},
	}); err != nil {
		t.Fatalf("Manager.Write() error = %v", err)
	}

	const iterations = 50
	var wg sync.WaitGroup
	wg.Add(iterations * 3) // 3x: readers, writers, mergers

	// Concurrent readers
	for i := 0; i < iterations; i++ {
		go func() {
			defer wg.Done()
			_, _ = mgr.Read()
		}()
	}

	// Concurrent writers
	for i := 0; i < iterations; i++ {
		go func() {
			defer wg.Done()
			_ = mgr.Write(InstallState{
				InstalledAgents: []string{"writer"},
			})
		}()
	}

	// Concurrent mergers
	for i := 0; i < iterations; i++ {
		go func() {
			defer wg.Done()
			_ = mgr.Merge(func(s InstallState) InstallState {
				s.InstalledAgents = append(s.InstalledAgents, "merged")
				return s
			})
		}()
	}

	wg.Wait()
	// If we get here without race detector firing, the test passes.
}

// TestManagerStateFilePath verifies managerStateFilePath returns expected path.
func TestManagerStateFilePath(t *testing.T) {
	home := t.TempDir()
	mgr := NewManager(home)

	got := mgr.StateFilePath()
	want := filepath.Join(home, ".architect-ai", "state.json")
	if got != want {
		t.Errorf("StateFilePath() = %q, want %q", got, want)
	}
}

// TestManagerMergePreservesExisting verifies that Merge only modifies what
// the callback changes — existing fields not touched by the callback survive.
func TestManagerMergePreservesExisting(t *testing.T) {
	home := t.TempDir()
	mgr := NewManager(home)

	existing := InstallState{
		InstalledAgents: []string{"claude-code", "opencode"},
		ClaudeModelAssignments: map[string]string{
			"orchestrator": "opus",
		},
		KiroModelAssignments: map[string]string{
			"sdd-design": "sonnet",
		},
	}
	if err := mgr.Write(existing); err != nil {
		t.Fatalf("Manager.Write() error = %v", err)
	}

	// Merge that only adds model assignments — should NOT wipe InstalledAgents
	// or KiroModelAssignments
	err := mgr.Merge(func(s InstallState) InstallState {
		s.ClaudeModelAssignments = map[string]string{
			"orchestrator": "opus",
			"sdd-explore":  "sonnet",
		}
		return s
	})
	if err != nil {
		t.Fatalf("Manager.Merge() error = %v", err)
	}

	got, err := mgr.Read()
	if err != nil {
		t.Fatalf("Manager.Read() error = %v", err)
	}

	// InstalledAgents must survive
	if len(got.InstalledAgents) != 2 {
		t.Errorf("InstalledAgents length = %d, want 2 (preserved by merge)", len(got.InstalledAgents))
	}

	// KiroModelAssignments must survive
	if got.KiroModelAssignments["sdd-design"] != "sonnet" {
		t.Errorf("KiroModelAssignments[sdd-design] = %q, want %q (preserved by merge)", got.KiroModelAssignments["sdd-design"], "sonnet")
	}

	// New assignment added
	if got.ClaudeModelAssignments["sdd-explore"] != "sonnet" {
		t.Errorf("ClaudeModelAssignments[sdd-explore] = %q, want %q", got.ClaudeModelAssignments["sdd-explore"], "sonnet")
	}
}

// TestManagerWriteOverwrite verifies a second Write replaces previous state.
func TestManagerWriteOverwrite(t *testing.T) {
	home := t.TempDir()
	mgr := NewManager(home)

	if err := mgr.Write(InstallState{InstalledAgents: []string{"old"}}); err != nil {
		t.Fatalf("Manager.Write() first error = %v", err)
	}

	if err := mgr.Write(InstallState{InstalledAgents: []string{"new"}}); err != nil {
		t.Fatalf("Manager.Write() second error = %v", err)
	}

	got, err := mgr.Read()
	if err != nil {
		t.Fatalf("Manager.Read() error = %v", err)
	}

	if len(got.InstalledAgents) != 1 || got.InstalledAgents[0] != "new" {
		t.Errorf("InstalledAgents = %v, want [new]", got.InstalledAgents)
	}
}

// TestManagerCreatesStateDir verifies Manager creates the .architect-ai directory.
func TestManagerCreatesStateDir(t *testing.T) {
	home := t.TempDir()
	mgr := NewManager(home)

	if err := mgr.Write(InstallState{InstalledAgents: []string{"opencode"}}); err != nil {
		t.Fatalf("Manager.Write() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(home, stateDir)); err != nil {
		t.Errorf("Directory %s not created: %v", stateDir, err)
	}
}