package openspec

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func minimalState() *State {
	t0 := time.Date(2026, 4, 17, 14, 22, 0, 0, time.UTC)
	t1 := t0.Add(time.Hour)
	return &State{
		SchemaVersion: 1,
		ChangeName:    "add-user-export",
		CreatedAt:     t0,
		UpdatedAt:     t1,
		ArtifactStore: "openspec",
		Phases: map[string]*Phase{
			"sdd-propose": {Status: "completed", CompletedAt: &t0, Artifact: "proposal.md"},
			"sdd-spec":    {Status: "pending", DependsOn: []string{"sdd-propose"}},
			"sdd-design":  {Status: "pending", DependsOn: []string{"sdd-spec"}},
			"sdd-tasks":   {Status: "pending", DependsOn: []string{"sdd-design"}},
			"sdd-apply":   {Status: "pending", DependsOn: []string{"sdd-tasks"}},
			"sdd-verify":  {Status: "pending", DependsOn: []string{"sdd-apply"}},
			"sdd-archive": {Status: "pending", DependsOn: []string{"sdd-verify"}},
		},
	}
}

func TestValidate_Happy(t *testing.T) {
	s := minimalState()
	if err := Validate(s, "add-user-export"); err != nil {
		t.Fatalf("want nil, got %v", err)
	}
}

func TestValidate_SchemaVersion(t *testing.T) {
	s := minimalState()
	s.SchemaVersion = 2
	err := Validate(s, "add-user-export")
	if !errors.Is(err, ErrSchemaVersion) {
		t.Fatalf("want ErrSchemaVersion, got %v", err)
	}
}

func TestValidate_MissingChangeName(t *testing.T) {
	s := minimalState()
	s.ChangeName = ""
	err := Validate(s, "")
	if !errors.Is(err, ErrMissingField) {
		t.Fatalf("want ErrMissingField, got %v", err)
	}
}

func TestValidate_BadKebab(t *testing.T) {
	s := minimalState()
	s.ChangeName = "AddUserExport"
	if err := Validate(s, "AddUserExport"); err == nil {
		t.Fatal("want error for non-kebab change_name")
	}
}

func TestValidate_FolderMismatch(t *testing.T) {
	s := minimalState()
	err := Validate(s, "different-name")
	if !errors.Is(err, ErrChangeNameMismatch) {
		t.Fatalf("want ErrChangeNameMismatch, got %v", err)
	}
}

func TestValidate_TimestampOrder(t *testing.T) {
	s := minimalState()
	s.UpdatedAt = s.CreatedAt.Add(-time.Minute)
	err := Validate(s, "add-user-export")
	if !errors.Is(err, ErrTimestampOrder) {
		t.Fatalf("want ErrTimestampOrder, got %v", err)
	}
}

func TestValidate_UnknownPhase(t *testing.T) {
	s := minimalState()
	s.Phases["sdd-mystery"] = &Phase{Status: "pending"}
	err := Validate(s, "add-user-export")
	if !errors.Is(err, ErrUnknownPhase) {
		t.Fatalf("want ErrUnknownPhase, got %v", err)
	}
}

func TestValidate_DanglingDepends(t *testing.T) {
	s := minimalState()
	// remove sdd-propose but keep sdd-spec's reference
	delete(s.Phases, "sdd-propose")
	err := Validate(s, "add-user-export")
	if !errors.Is(err, ErrDanglingDepends) {
		t.Fatalf("want ErrDanglingDepends, got %v", err)
	}
}

func TestValidate_CycleRejected(t *testing.T) {
	// Force a cycle by cross-linking two phases.
	s := minimalState()
	s.Phases["sdd-propose"].DependsOn = []string{"sdd-spec"} // spec -> propose already
	err := Validate(s, "add-user-export")
	if !errors.Is(err, ErrCycle) {
		t.Fatalf("want ErrCycle, got %v", err)
	}
}

func TestValidate_CompletedRequiresTimestamp(t *testing.T) {
	s := minimalState()
	s.Phases["sdd-propose"].CompletedAt = nil
	err := Validate(s, "add-user-export")
	if !errors.Is(err, ErrMissingField) {
		t.Fatalf("want ErrMissingField, got %v", err)
	}
}

func TestSaveThenLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	changeDir := filepath.Join(dir, "add-user-export")
	if err := os.MkdirAll(changeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(changeDir, "state.yaml")
	s := minimalState()
	if err := Save(path, s); err != nil {
		t.Fatalf("Save: %v", err)
	}
	// tmp file must not remain
	if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
		t.Fatalf("state.yaml.tmp should not exist after Save: %v", err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.ChangeName != s.ChangeName {
		t.Fatalf("round-trip lost change_name")
	}
	// Load refreshes UpdatedAt via Save; it must be >= CreatedAt
	if loaded.UpdatedAt.Before(loaded.CreatedAt) {
		t.Fatal("UpdatedAt before CreatedAt after round-trip")
	}
}

func TestLoad_CorruptYAML(t *testing.T) {
	dir := t.TempDir()
	changeDir := filepath.Join(dir, "add-user-export")
	if err := os.MkdirAll(changeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(changeDir, "state.yaml")
	if err := os.WriteFile(path, []byte("schema_version: 1\n:::this is not yaml"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Load(path)
	if err == nil {
		t.Fatal("want error on corrupt YAML")
	}
	if !strings.Contains(err.Error(), "parse") {
		t.Fatalf("want parse error, got %v", err)
	}
}
