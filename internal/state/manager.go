package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// Manager provides thread-safe access to the install state file.
// All state operations for concurrent callers MUST go through Manager;
// direct state.Read/state.Write are deprecated for concurrent use.
//
// The Manager serializes read-modify-write cycles via a mutex, ensuring
// that Merge operations are atomic and no data is lost under concurrent access.
type Manager struct {
	homeDir string
	mu      sync.Mutex
}

// NewManager creates a Manager for the given home directory.
// The homeDir is used to resolve the path to .architect-ai/state.json.
func NewManager(homeDir string) *Manager {
	return &Manager{homeDir: homeDir}
}

// Read reads the state file while holding the mutex.
// This is safe for concurrent use.
func (m *Manager) Read() (InstallState, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return Read(m.homeDir)
}

// Write persists the full install state to disk while holding the mutex.
// This is safe for concurrent use.
func (m *Manager) Write(s InstallState) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return Write(m.homeDir, s)
}

// Merge reads the current state, applies fn, and writes the result atomically.
// The entire read-modify-write cycle is mutex-protected, so concurrent Merge
// calls will not cause data loss or corruption.
//
// fn receives a copy of the current state and returns the modified copy.
// If the state file does not exist, fn receives an empty InstallState.
func (m *Manager) Merge(fn func(InstallState) InstallState) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	current, err := Read(m.homeDir)
	if err != nil {
		// Missing state file is treated as empty state — not an error.
		current = InstallState{}
	}
	updated := fn(current)
	return Write(m.homeDir, updated)
}

// StateFilePath returns the absolute path to state.json for this Manager's homeDir.
// This is useful for tests and diagnostics.
func (m *Manager) StateFilePath() string {
	return Path(m.homeDir)
}

// atomicWrite writes data to a file using the write-then-rename pattern to
// avoid partial writes that could corrupt state.json during crashes.
func atomicWrite(targetPath string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(targetPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	tmpPath := targetPath + ".tmp"
	if err := os.WriteFile(tmpPath, data, perm); err != nil {
		return err
	}

	// os.Rename on most Unix systems is atomic within the same filesystem.
	// On Windows, os.Rename requires the target to not exist, so we use the
	// portable approach: remove target first, then rename.
	_ = os.Remove(targetPath)
	return os.Rename(tmpPath, targetPath)
}

// WriteAtomic persists state using atomic write (write-then-rename).
// This is the same as Write but uses atomic file replacement to prevent
// corruption from partial writes during crashes.
func WriteAtomic(homeDir string, s InstallState) error {
	dir := filepath.Join(homeDir, stateDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return atomicWrite(Path(homeDir), append(data, '\n'), 0o644)
}