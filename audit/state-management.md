# Audit: State Management Hardening

## Files Examined
- internal/state/state.go
- internal/state/managed_manifest.go
- internal/app/app.go
- internal/cli/install.go
- internal/cli/run.go
- internal/cli/sync.go
- internal/components/uninstall/service.go

## Evidence
The state package (`internal/state/state.go`) provides `Read` and `Write` functions for persisting `InstallState` to `state.json`. These functions directly interact with the filesystem using `os.ReadFile` and `os.WriteFile` without any synchronization primitives (e.g., `sync.Mutex`, `sync.RWMutex`, or file locking).

The function `persistAssignments` in `internal/app/app.go` implements a read-modify-write pattern:
```go
current, err := state.Read(homeDir)
// ...
_ = state.Write(homeDir, current)
```
This is a classic race condition if multiple goroutines or parallel pipeline steps attempt to update the state simultaneously.

## Callers
- internal/cli/run.go
- internal/app/app.go
- internal/components/uninstall/service.go
- internal/cli/sync.go
- internal/cli/sync_test.go
- internal/app/app_test.go

## Vulnerability
- Pattern: unprotected read-modify-write
- Risk post-Phase 1: HIGH (concurrent pipeline steps)
- Fix: state.Manager with sync.Mutex and atomic file replacement or file locking.

## Verdict
MASTER-PLAN Phase 5 claim VALIDATED
