# Phase 5 — State Management Hardening & Concurrent-Safe Patterns

> **Cognitive Mode**: +++Systemic +++Adversarial +++Empirical  
> **CCLD Tag**: `[PHASE-5][STATE][CONCURRENT][MUTEX][GO]`  
> **Status**: BLOCKED until Phase 1 complete (parallel pipeline introduces race window)  
> **Estimated Duration**: 1 session  
> **Depends On**: Phase 1 complete, `audit/state-management.md`

---

## 5.1 Objective

Harden the `internal/state` package and all call sites to be safe under concurrent access introduced by Phase 1's parallel pipeline. Add mutex protection to state file R/W, make backup operations context-aware, and enforce a read-merge-write pattern with conflict detection.

**Target Outcome**: Zero data races on state.json under concurrent pipeline execution. Verified by `-race` detector.

---

## 5.2 Root Cause (from Phase 0 Audit)

### 5.2.1 No Mutex on state.json

```go
// internal/app/app.go — current
func tuiExecute(...) pipeline.ExecutionResult {
    execResult := orchestrator.Execute(stagePlan) // ← parallel pipeline after Phase 1
    if execResult.Err == nil {
        _ = state.Write(homeDir, state.InstallState{...}) // ← no mutex, no CAS
    }
    return execResult
}

func loadPersistedAssignments(homeDir string, selection *model.Selection) {
    s, err := state.Read(homeDir) // ← concurrent read possible
    ...
}
```

**Problem**: After Phase 1, `tuiExecute` runs multiple pipeline steps concurrently. If two steps both modify `state.json` (unlikely in current codebase but possible after extension), they race.

More concretely: the TUI can call `tuiSync` and `tuiExecute` concurrently (user fires two actions before first completes). Both call `state.Write`. Race.

### 5.2.2 Backup Operations Not Context-Aware

```go
// internal/cli/install.go (inferred from backup_metadata_test.go)
step := prepareBackupStep{
    snapshotter: backup.NewSnapshotter(),
    targets:     []string{configPath},
    // ... no context.Context parameter
}
```

**Problem**: `backup.NewSnapshotter().Snapshot()` performs I/O without a `context.Context`. Cannot be cancelled. Cannot timeout. With parallel pipeline (Phase 1), a hung backup step blocks the entire group.

### 5.2.3 `persistAssignments` Race Pattern

```go
func persistAssignments(homeDir string, selection model.Selection) {
    current, err := state.Read(homeDir)  // read
    // ... merge selection into current  // modify
    _ = state.Write(homeDir, current)    // write
}
```

**Problem**: Classic read-modify-write race if two callers run concurrently. No file locking. No optimistic concurrency control.

---

## 5.3 Refactoring Plan

### 5.3.1 `StateManager` — Mutex-Protected R/W

```go
// internal/state/manager.go — NEW FILE

package state

import (
    "encoding/json"
    "os"
    "path/filepath"
    "sync"
)

// Manager provides thread-safe access to the install state file.
// All state operations MUST go through Manager; direct state.Read/state.Write
// are deprecated for concurrent callers.
type Manager struct {
    homeDir string
    mu      sync.Mutex
}

func NewManager(homeDir string) *Manager {
    return &Manager{homeDir: homeDir}
}

func (m *Manager) Read() (InstallState, error) {
    m.mu.Lock()
    defer m.mu.Unlock()
    return Read(m.homeDir)
}

func (m *Manager) Write(s InstallState) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    return Write(m.homeDir, s)
}

// Merge reads current state, applies fn, and writes the result atomically.
// fn receives a copy of current state and returns the modified copy.
// The entire read-modify-write cycle is mutex-protected.
func (m *Manager) Merge(fn func(InstallState) InstallState) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    current, err := Read(m.homeDir)
    if err != nil {
        current = InstallState{} // treat missing state as empty
    }
    updated := fn(current)
    return Write(m.homeDir, updated)
}

// stateFilePath returns the path to state.json.
func (m *Manager) stateFilePath() string {
    return filepath.Join(m.homeDir, ".architect-ai", "state.json")
}
```

**Usage in `app.go`**:
```go
// internal/app/app.go — AFTER
// Declare package-level manager (initialized in RunArgs)
var stateManager *state.Manager

func RunArgs(args []string, stdout io.Writer) error {
    homeDir, err := os.UserHomeDir()
    if err != nil { return err }
    stateManager = state.NewManager(homeDir)
    ...
}

func tuiExecute(...) pipeline.ExecutionResult {
    execResult := orchestrator.Execute(stagePlan)
    if execResult.Err == nil {
        _ = stateManager.Merge(func(s state.InstallState) state.InstallState {
            agentIDs := make([]string, 0, len(selection.Agents))
            for _, a := range selection.Agents {
                agentIDs = append(agentIDs, string(a))
            }
            s.InstalledAgents = agentIDs
            s.ClaudeModelAssignments = claudeAliasesToStrings(selection.ClaudeModelAssignments)
            s.ModelAssignments = modelAssignmentsToState(selection.ModelAssignments)
            return s
        })
    }
    return execResult
}
```

---

### 5.3.2 Context-Aware Backup Steps

```go
// internal/cli/install.go — MODIFY prepareBackupStep

type prepareBackupStep struct {
    id          string
    snapshotter backup.Snapshotter
    snapshotDir string
    targets     []string
    state       *runtimeState
    source      backup.BackupSource
    description string
    appVersion  string
    ctx         context.Context // NEW: cancellation support
}

func (s prepareBackupStep) Run() error {
    ctx := s.ctx
    if ctx == nil {
        ctx = context.Background()
    }
    
    // Wrap snapshot in a channel for context cancellation
    type result struct {
        manifest backup.Manifest
        err      error
    }
    ch := make(chan result, 1)
    go func() {
        m, err := s.snapshotter.Snapshot(s.snapshotDir, s.targets, backup.SnapshotOptions{
            Source:      s.source,
            Description: s.description,
            AppVersion:  s.appVersion,
        })
        ch <- result{m, err}
    }()
    
    select {
    case r := <-ch:
        if r.err != nil {
            return fmt.Errorf("backup snapshot: %w", r.err)
        }
        s.state.setManifest(r.manifest)
        return nil
    case <-ctx.Done():
        return fmt.Errorf("backup snapshot cancelled: %w", ctx.Err())
    }
}
```

---

### 5.3.3 `persistAssignments` — Use Manager.Merge

```go
// internal/app/app.go — AFTER

func persistAssignments(selection model.Selection) {
    if len(selection.ClaudeModelAssignments) == 0 && 
       len(selection.KiroModelAssignments) == 0 && 
       len(selection.ModelAssignments) == 0 {
        return
    }
    
    _ = stateManager.Merge(func(current state.InstallState) state.InstallState {
        if len(selection.ClaudeModelAssignments) > 0 {
            current.ClaudeModelAssignments = claudeAliasesToStrings(selection.ClaudeModelAssignments)
        }
        if len(selection.KiroModelAssignments) > 0 {
            current.KiroModelAssignments = claudeAliasesToStrings(selection.KiroModelAssignments)
        }
        if len(selection.ModelAssignments) > 0 {
            current.ModelAssignments = modelAssignmentsToState(selection.ModelAssignments)
        }
        return current
    })
}
```

Remove the `homeDir` parameter — `stateManager` is package-level, initialized once.

---

### 5.3.4 `loadPersistedAssignments` — Use Manager.Read

```go
// internal/app/app.go — AFTER

func loadPersistedAssignments(selection *model.Selection) {
    s, err := stateManager.Read()
    if err != nil {
        return
    }
    // ... rest unchanged
}
```

---

### 5.3.5 TUI Event Serialization

The TUI uses `tea.Program` which processes messages on a single goroutine. However, the `ExecuteFn`, `SyncFn`, `UpgradeFn` callbacks are invoked from goroutines spawned by `tea.Program`. Add explicit notes to TUI callbacks:

```go
// internal/app/app.go — ADD COMMENTS

// tuiExecute: called from tea.Program worker goroutine. 
// State writes MUST go through stateManager.Merge (mutex-protected).
// DO NOT call state.Write directly from this function.
func tuiExecute(...) pipeline.ExecutionResult {
    ...
}

// tuiSync: called from tea.Program worker goroutine.
// Same constraint as tuiExecute.
func tuiSync(homeDir string) tui.SyncFunc {
    return func(overrides *model.SyncOverrides) (int, error) {
        ...
        persistAssignments(selection) // now uses stateManager.Merge
        ...
    }
}
```

---

### 5.3.6 `runtimeState` Concurrent-Safety (Pipeline Internal)

The `runtimeState` struct is shared between pipeline steps and the orchestrator to track backup manifest and other transient state:

```go
// Inferred from backup_metadata_test.go
type runtimeState struct {
    manifest backup.Manifest
    // ... other fields
}
```

After Phase 1 parallelization, if two prepare steps run concurrently and both call `state.setManifest(...)`, they race. Fix:

```go
// internal/cli/install.go — ADD

type runtimeState struct {
    mu       sync.Mutex
    manifest backup.Manifest
}

func (s *runtimeState) setManifest(m backup.Manifest) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.manifest = m
}

func (s *runtimeState) getManifest() backup.Manifest {
    s.mu.Lock()
    defer s.mu.Unlock()
    return s.manifest
}
```

---

## 5.4 Files to Create / Modify

| File | Action | Notes |
|---|---|---|
| `internal/state/manager.go` | CREATE | `StateManager` with mutex, `Read`, `Write`, `Merge` |
| `internal/state/manager_test.go` | CREATE | Concurrent R/W tests, race detector |
| `internal/app/app.go` | MODIFY | Use `stateManager`; remove `homeDir` param from persist functions |
| `internal/cli/install.go` | MODIFY | `prepareBackupStep` context-aware; `runtimeState` mutex |
| `internal/cli/install_test.go` | MODIFY | Add concurrent backup step test |

---

## 5.5 Testing Requirements

| Test | Location | Verifies |
|---|---|---|
| `TestStateManagerConcurrentMerge` | `state/manager_test.go` | 10 goroutines calling `Merge` concurrently; final state consistent |
| `TestStateManagerReadWriteRace` | `state/manager_test.go` | `-race` clean under concurrent R/W |
| `TestPrepareBackupStepContextCancellation` | `cli/install_test.go` | Step respects context cancellation |
| `TestRuntimeStateConcurrentSetManifest` | `cli/install_test.go` | No race on concurrent setManifest |
| `TestTuiExecuteAndSyncConcurrent` | `app/app_test.go` | Concurrent tuiExecute + tuiSync don't corrupt state |

### Race Detector Full Sweep
```bash
go test -race ./internal/state/...
go test -race ./internal/app/...
go test -race ./internal/cli/...
go test -race ./internal/pipeline/...
```

All must be clean before Phase 5 is closed.

---

## 5.6 Acceptance Criteria

- [ ] `state.Manager` type exists with `Read`, `Write`, `Merge` methods
- [ ] All `state.Read` / `state.Write` direct calls in `app.go` replaced with `stateManager.*`
- [ ] `runtimeState` has mutex protection on `setManifest` / `getManifest`
- [ ] `prepareBackupStep.Run()` respects `context.Context` cancellation
- [ ] `-race` clean on all packages listed above
- [ ] `TestStateManagerConcurrentMerge`: 10 concurrent goroutines produce consistent final state
- [ ] All existing tests pass without modification

---

## 5.7 Sub-Agent Delegation

```
[PHASE-5 ORCHESTRATOR]
    │
    ├── [5A] go-writer-agent     → state/manager.go StateManager
    ├── [5B] go-writer-agent     → cli/install.go context-aware backup + runtimeState mutex
    ├── [5C] go-writer-agent     → app.go stateManager integration (depends 5A)
    ├── [5D] go-tester-agent     → state/manager_test.go + cli/install_test.go (depends 5A+5B)
    └── [5E] go-tester-agent     → app/app_test.go concurrent TUI test (depends 5C)
```

5A, 5B launch in parallel.  
5C launches after 5A.  
5D launches after 5A + 5B.  
5E launches after 5C.
