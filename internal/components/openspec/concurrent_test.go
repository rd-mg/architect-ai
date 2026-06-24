package openspec

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestSave_ConcurrentWriters_NoCorruption(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	changeName := "test-change"
	stateDir := filepath.Join(dir, changeName)
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	path := filepath.Join(stateDir, "state.yaml")

	makeState := func(n int) *State {
		now := time.Now().UTC()
		return &State{
			SchemaVersion: SchemaVersion,
			ChangeName:    changeName,
			CreatedAt:     now,
			UpdatedAt:     now,
			ArtifactStore: "openspec",
			Phases: map[string]*Phase{
				"sdd-explore": {
					Status: "in_progress",
					StartedAt: func() *time.Time {
						t := time.Now().UTC()
						return &t
					}(),
					Meta: map[string]any{"writer": n},
				},
			},
		}
	}

	const writers = 20
	var wg sync.WaitGroup
	errs := make([]error, writers)

	for i := 0; i < writers; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			errs[n] = Save(path, makeState(n))
		}(i)
	}
	wg.Wait()

	successes := 0
	for _, err := range errs {
		if err == nil {
			successes++
		} else if err.Error() == "" {
			t.Errorf("unexpected empty error")
		}
	}
	if successes == 0 {
		t.Fatal("at least one writer must succeed")
	}

	final, err := Load(path)
	if err != nil {
		t.Fatalf("final state.yaml corrupted after concurrent writes: %v", err)
	}
	if final.ChangeName != changeName {
		t.Errorf("change_name corrupted: got %q, want %q", final.ChangeName, changeName)
	}
}

func TestSave_EnumInProgress_Accepted(t *testing.T) {
	dir := t.TempDir()
	changeName := "enum-test"
	stateDir := filepath.Join(dir, changeName)
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	path := filepath.Join(stateDir, "state.yaml")

	now := time.Now().UTC()
	s := &State{
		SchemaVersion: SchemaVersion,
		ChangeName:    changeName,
		CreatedAt:     now,
		UpdatedAt:     now,
		ArtifactStore: "openspec",
		Phases: map[string]*Phase{
			"sdd-explore": {
				Status: "in_progress",
				StartedAt: func() *time.Time {
					t := time.Now().UTC()
					return &t
				}(),
			},
		},
	}
	if err := Save(path, s); err != nil {
		t.Fatalf("Save with in_progress should succeed: %v", err)
	}
}

func TestSave_EnumRunning_Rejected(t *testing.T) {
	dir := t.TempDir()
	changeName := "enum-running"
	stateDir := filepath.Join(dir, changeName)
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	path := filepath.Join(stateDir, "state.yaml")

	now := time.Now().UTC()
	s := &State{
		SchemaVersion: SchemaVersion,
		ChangeName:    changeName,
		CreatedAt:     now,
		UpdatedAt:     now,
		ArtifactStore: "openspec",
		Phases: map[string]*Phase{
			"sdd-explore": {Status: "running"},
		},
	}
	if err := Save(path, s); err == nil {
		t.Fatal(`Save with status "running" must fail — "running" is not a valid status`)
	}
}

func TestSave_NewStatuses_Accepted(t *testing.T) {
	newStatuses := []string{
		"abandoned", "infrastructure_blocked",
		"partially_completed", "rollback_in_progress",
	}
	for _, status := range newStatuses {
		t.Run(fmt.Sprintf("status_%s", status), func(t *testing.T) {
			dir := t.TempDir()
			changeName := "status-test"
			stateDir := filepath.Join(dir, changeName)
			if err := os.MkdirAll(stateDir, 0o755); err != nil {
				t.Fatalf("mkdir: %v", err)
			}
			path := filepath.Join(stateDir, "state.yaml")

			now := time.Now().UTC()
			ph := &Phase{Status: status}
			if status == "in_progress" {
				ph.StartedAt = func() *time.Time {
					t := time.Now().UTC()
					return &t
				}()
			}
			if status == "completed" {
				ph.CompletedAt = func() *time.Time {
					t := time.Now().UTC()
					return &t
				}()
			}
			if status == "failed" {
				ph.Error = "test error"
			}
			s := &State{
				SchemaVersion: SchemaVersion,
				ChangeName:    changeName,
				CreatedAt:     now,
				UpdatedAt:     now,
				ArtifactStore: "openspec",
				Phases:        map[string]*Phase{"sdd-explore": ph},
			}
			if err := Save(path, s); err != nil {
				t.Errorf("Save with status %q should succeed: %v", status, err)
			}
		})
	}
}
