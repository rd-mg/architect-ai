package system

import (
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// ConfigState records the filesystem presence of an agent's global config directory.
// All known registry agents are always represented — Exists=false for absent dirs.
// This contract is consumed by the TUI detection screen and install/validate flows.
//
// For CLI agents, BinaryName and BinaryFound provide a second signal: a config
// directory created solely by architect-ai should not count as "installed" unless
// the binary is also on PATH. IDE agents (no binary) rely on directory presence
// alone.
type ConfigState struct {
	Agent       string
	Path        string
	Exists      bool
	IsDirectory bool
	BinaryName  string // empty for IDE agents (no PATH check)
	BinaryFound bool   // true when BinaryName is non-empty and exec.LookPath succeeds
}

// knownAgentConfigDirs enumerates the per-agent config roots used by ScanConfigs
// for presence scanning as (agentID, path) pairs. This is a compatibility shim
// that mirrors the adapter registry's full set without importing the agents
// package (which would create an import cycle: system ← agents ← system).
//
// Most entries mirror Adapter.GlobalConfigDir(). Kiro is an intentional
// exception: we scan `~/.kiro` (managed artifacts root) instead of
// `%APPDATA%/kiro/User` (settings root) due to Kiro's split-root layout.
//
// When a new agent is added to the registry, its entry must also be added here
// until the import cycle is resolved and ScanConfigs can delegate directly to
// agents.DiscoverInstalled.
func knownAgentConfigDirs(homeDir string) []ConfigState {
	return []ConfigState{
		{Agent: "claude-code", Path: filepath.Join(homeDir, ".claude"), BinaryName: "claude"},
		{Agent: "opencode", Path: filepath.Join(homeDir, ".config", "opencode"), BinaryName: "opencode"},
		{Agent: "kilocode", Path: filepath.Join(homeDir, ".config", "kilo"), BinaryName: "kilo"},

		{Agent: "cursor", Path: filepath.Join(homeDir, ".cursor")},
		{Agent: "vscode-copilot", Path: vscodeCopilotGlobalConfigDir(homeDir)},
		{Agent: "codex", Path: filepath.Join(homeDir, ".codex"), BinaryName: "codex"},
		{Agent: "antigravity", Path: filepath.Join(homeDir, ".gemini", "antigravity"), BinaryName: "antigravity"},
		{Agent: "antigravity-cli", Path: filepath.Join(homeDir, ".gemini", "antigravity-cli", "plugins", "architect-ai"), BinaryName: "agy"},
		{Agent: "windsurf", Path: filepath.Join(homeDir, ".codeium", "windsurf")},
		{Agent: "qwen-code", Path: filepath.Join(homeDir, ".qwen"), BinaryName: "qwen"},
		{Agent: "kiro-ide", Path: filepath.Join(homeDir, ".kiro"), BinaryName: "kiro"},
	}
}

// vscodeCopilotGlobalConfigDir returns ~/.copilot, the GlobalConfigDir used by
// the vscode-copilot adapter across all platforms. The vscode adapter's
// SystemPromptDir and SettingsPath are OS-dependent, but GlobalConfigDir is not.
func vscodeCopilotGlobalConfigDir(homeDir string) string {
	return filepath.Join(homeDir, ".copilot")
}

// ScanConfigs returns the presence state of every known managed agent's global
// config directory. All agents are always represented in the result; Exists and
// IsDirectory reflect the actual filesystem state at call time.
//
// This is a compatibility shim: it preserves the ConfigState contract for TUI
// and validation callers while the canonical discovery (agents.DiscoverInstalled)
// is used by sync and upgrade flows. Full delegation is deferred until the
// system ← agents import cycle is resolved (follow-up change).
func ScanConfigs(homeDir string) []ConfigState {
	states := knownAgentConfigDirs(homeDir)

	for idx := range states {
		info, err := os.Stat(states[idx].Path)
		if err != nil {
			continue
		}

		states[idx].Exists = true
		states[idx].IsDirectory = info.IsDir()

		// For CLI agents, also check if the binary is on PATH.
		// A config directory created by architect-ai alone is not
		// sufficient evidence that the agent is actually installed.
		if states[idx].BinaryName != "" {
			_, err := exec.LookPath(states[idx].BinaryName)
			states[idx].BinaryFound = err == nil
		}
	}

	return states
}

// SessionState represents the cached tool-availability state forwarded from L0
// to L1a/L1b orchestrators. It is persisted to Engram under the topic key
// "session-state/{project}/tools" and validated by age before reuse.
//
// The schema mirrors the JSON shape saved by the orchestrator session-state
// cache logic in the prompt layer (thinking-agent.md / general-orchestrator.md).
type SessionState struct {
	Tools     map[string]bool `json:"tools"`
	Timestamp time.Time       `json:"timestamp"`
	Project   string          `json:"project"`
}

// IsValid reports whether the session state is still within the maxAge window.
// A zero Timestamp is always considered invalid.
func (s SessionState) IsValid(maxAge time.Duration) bool {
	return !s.Timestamp.IsZero() && time.Since(s.Timestamp) < maxAge
}
