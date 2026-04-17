package uninstall

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/rd-mg/architect-ai/internal/backup"
	"github.com/rd-mg/architect-ai/internal/model"
)

// PurgeScope describes which categories of artifacts a DeepPurge removes.
// Any combination is legal. All-zero means "nothing extra beyond managed config".
type PurgeScope struct {
	// ManagedConfig removes whatever CompleteUninstall already removes.
	// Always set to true in practice; users who don't want this should use
	// the existing CompleteUninstall entry point instead.
	ManagedConfig bool

	// EngramProject deletes the Engram project memories for the workspace.
	// Requires Engram MCP to be reachable; failure is logged but does not
	// abort other purge steps.
	EngramProject bool

	// WorkspaceATL removes the `.atl/` directory in the workspace (skill
	// registry, overlay manifests, context packs, local state).
	WorkspaceATL bool

	// GlobalArchitectAI removes `~/.architect-ai/` (backups, state, cache).
	// HIGH-IMPACT — backups go away. Only set after the user has confirmed
	// and a pre-purge snapshot has been captured.
	GlobalArchitectAI bool

	// Binary removes the architect-ai binary via the platform's package
	// manager (brew/apt/pacman). Best-effort; silently skipped if no
	// matching package manager is found.
	Binary bool
}

// AllButGlobal returns a conservative default — everything except the global
// state directory and the binary. Good default for a "project cleanup".
func AllButGlobal() PurgeScope {
	return PurgeScope{
		ManagedConfig: true,
		EngramProject: true,
		WorkspaceATL:  true,
	}
}

// Nuclear returns the maximal purge scope. Use with explicit user confirmation.
func Nuclear() PurgeScope {
	return PurgeScope{
		ManagedConfig:     true,
		EngramProject:     true,
		WorkspaceATL:      true,
		GlobalArchitectAI: true,
		Binary:            true,
	}
}

// EngramDeleter is a minimal interface implemented by whatever component
// speaks to the Engram MCP. The Service holds an implementation; tests
// can pass a stub.
type EngramDeleter interface {
	DeleteProject(project string) error
	IsReachable() bool
}

// PurgeResult is the post-purge report shown to the user.
type PurgeResult struct {
	Result                   // embed the standard uninstall result
	SnapshotPath      string // absolute path of the pre-purge snapshot
	EngramRemoved     bool
	EngramError       string
	ATLRemoved        bool
	GlobalRemoved     bool
	BinaryRemoved     bool
	BinaryCommandUsed string
	BinaryError       string
	PurgeDurationMs   int64
	ScopeRequested    PurgeScope
}

// DeepPurge runs a deep uninstall across the requested scope.
//
// Order (each step is best-effort; failures in one step do NOT abort the others):
//  1. Capture pre-purge snapshot (always)
//  2. CompleteUninstall — managed config (delegated to existing Service method)
//  3. Engram project memories (mem_delete_project via EngramDeleter)
//  4. Workspace .atl/ directory
//  5. Global ~/.architect-ai/ directory
//  6. Binary removal via package manager
//
// A DeepPurge ALWAYS captures a snapshot first, regardless of scope. The
// snapshot allows restore even of a Nuclear purge.
func (s *Service) DeepPurge(scope PurgeScope, project string, engram EngramDeleter) (PurgeResult, error) {
	start := time.Now()
	res := PurgeResult{ScopeRequested: scope}

	// 1. Pre-purge snapshot — ALWAYS, even if scope is minimal.
	snapshotDir := filepath.Join(s.backupRoot, fmt.Sprintf("pre-purge-%s", start.Format("20060102-150405")))
	snapshot, err := s.capturePrePurgeSnapshot(snapshotDir)
	if err != nil {
		return res, fmt.Errorf("capture pre-purge snapshot: %w", err)
	}
	res.SnapshotPath = snapshotDir
	res.Manifest = snapshot

	// 2. Managed config — delegate to existing CompleteUninstall.
	if scope.ManagedConfig {
		completeRes, err := s.CompleteUninstall()
		if err != nil {
			return res, fmt.Errorf("complete uninstall: %w", err)
		}
		res.Result = completeRes
		res.Result.Manifest = snapshot // override with our snapshot
	}

	// 3. Engram project memories.
	if scope.EngramProject {
		if engram == nil || !engram.IsReachable() {
			res.EngramError = "Engram MCP unreachable — skipped"
		} else if err := engram.DeleteProject(project); err != nil {
			res.EngramError = err.Error()
		} else {
			res.EngramRemoved = true
		}
	}

	// 4. Workspace .atl/ directory.
	if scope.WorkspaceATL {
		atlPath := filepath.Join(s.workspaceDir, ".atl")
		if err := os.RemoveAll(atlPath); err == nil {
			res.ATLRemoved = true
			res.Result.RemovedDirectories = append(res.Result.RemovedDirectories, atlPath)
		} else if !os.IsNotExist(err) {
			res.Result.ManualActions = append(res.Result.ManualActions,
				fmt.Sprintf("Remove %s manually: %v", atlPath, err))
		}
	}

	// 5. Global ~/.architect-ai/ directory.
	if scope.GlobalArchitectAI {
		globalPath := filepath.Join(s.homeDir, ".architect-ai")
		if err := os.RemoveAll(globalPath); err == nil {
			res.GlobalRemoved = true
			res.Result.RemovedDirectories = append(res.Result.RemovedDirectories, globalPath)
		} else if !os.IsNotExist(err) {
			res.Result.ManualActions = append(res.Result.ManualActions,
				fmt.Sprintf("Remove %s manually: %v", globalPath, err))
		}
	}

	// 6. Binary via package manager.
	if scope.Binary {
		cmd, err := removeBinary()
		res.BinaryCommandUsed = cmd
		if err != nil {
			res.BinaryError = err.Error()
			res.Result.ManualActions = append(res.Result.ManualActions,
				fmt.Sprintf("Remove binary manually (tried: %s): %v", cmd, err))
		} else {
			res.BinaryRemoved = true
		}
	}

	res.PurgeDurationMs = time.Since(start).Milliseconds()
	return res, nil
}

// capturePrePurgeSnapshot records the paths that a full restore would need.
// Uses the Service.registry.SupportedAgents() to enumerate agent IDs, then
// builds per-agent candidate paths conservatively. Non-existent paths are
// filtered out by the snapshotter itself.
func (s *Service) capturePrePurgeSnapshot(dest string) (backup.Manifest, error) {
	paths := []string{
		filepath.Join(s.workspaceDir, ".atl"),
		filepath.Join(s.homeDir, ".architect-ai"),
	}

	// Enumerate supported agent IDs from the registry and add their
	// conventional config paths. The Registry exposes SupportedAgents()
	// returning []model.AgentID; no .All() or .ID() methods exist.
	if s.registry != nil {
		for _, agentID := range s.registry.SupportedAgents() {
			paths = append(paths, candidatePathsForAgent(s.homeDir, s.workspaceDir, agentID)...)
		}
	}

	paths = dedupe(paths)

	if s.snapshotter == nil {
		return backup.Manifest{}, fmt.Errorf("no snapshotter configured")
	}
	return s.snapshotter.Create(dest, paths)
}

// candidatePathsForAgent returns the conventional config locations for a given
// agent. The snapshotter filters out non-existent paths.
func candidatePathsForAgent(homeDir, workspaceDir string, id model.AgentID) []string {
	agent := string(id)
	return []string{
		filepath.Join(homeDir, "."+agent),
		filepath.Join(workspaceDir, "."+agent),
		filepath.Join(workspaceDir, "."+agent+"-settings.json"),
	}
}

// removeBinary tries the platform-typical package manager to uninstall
// architect-ai. Returns the command line attempted (for diagnostics) and
// an error if none worked.
func removeBinary() (string, error) {
	candidates := []struct {
		cmd  string
		args []string
	}{
		{"brew", []string{"uninstall", "architect-ai"}},
		{"apt-get", []string{"remove", "-y", "architect-ai"}},
		{"pacman", []string{"-R", "--noconfirm", "architect-ai"}},
		{"snap", []string{"remove", "architect-ai"}},
	}

	for _, c := range candidates {
		if _, err := exec.LookPath(c.cmd); err != nil {
			continue
		}
		full := c.cmd + " " + strings.Join(c.args, " ")
		cmd := exec.Command(c.cmd, c.args...)
		if _, err := cmd.CombinedOutput(); err != nil {
			// Try next candidate
			continue
		}
		return full, nil
	}

	if runtime.GOOS == "windows" {
		return "", fmt.Errorf("binary removal on Windows: use winget/choco manually")
	}
	return "", fmt.Errorf("no supported package manager found (tried brew, apt, pacman, snap)")
}

func dedupe(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	var out []string
	for _, s := range in {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}
