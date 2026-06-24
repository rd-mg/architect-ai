package state

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/rd-mg/architect-ai/internal/model"
)

// mu protects concurrent access to the state file on disk.
// All exported Read/Write callers go through this lock.
var mu sync.RWMutex

const stateDir = ".architect-ai"
const stateFile = "state.json"

// ModelAssignmentState is the JSON-serialisable form of a provider+model pair
// used by OpenCode-style model assignments. It mirrors model.ModelAssignment
// but lives in the state package to avoid an import cycle.
type ModelAssignmentState struct {
	ProviderID string `json:"provider_id"`
	ModelID    string `json:"model_id"`
}

// InstallState holds the persisted user selections from the last install run.
type InstallState struct {
	InstalledAgents []string `json:"installed_agents"`

	// ClaudeModelAssignments maps SDD phase names (e.g. "sdd-explore") to a
	// Claude model alias ("opus", "sonnet", "haiku"). Persisted so that
	// `architect-ai sync` preserves the user's model choices instead of falling
	// back to the "balanced" preset every time.
	ClaudeModelAssignments map[string]string `json:"claude_model_assignments,omitempty"`

	// KiroModelAssignments maps SDD phase names to a Claude model alias for
	// Kiro IDE specifically. Persisted independently from ClaudeModelAssignments
	// so Kiro and Claude Code model choices survive across sync runs.
	KiroModelAssignments map[string]string `json:"kiro_model_assignments,omitempty"`

	// ModelAssignments maps sub-agent names to provider/model pairs (OpenCode).
	ModelAssignments map[string]ModelAssignmentState `json:"model_assignments,omitempty"`
}

// Path returns the absolute path to the state file for the given home directory.
func Path(homeDir string) string {
	return filepath.Join(homeDir, stateDir, stateFile)
}

// AgentManifestPath returns the absolute path to the manifest file for the
// given home directory, project root, and agent.
func AgentManifestPath(homeDir, projectRoot string, agent model.AgentID) string {
	slug := "global"
	if projectRoot != "" {
		slug = hexSlug(projectRoot)
	}
	return filepath.Join(homeDir, stateDir, "managed", slug, string(agent)+".json")
}

func hexSlug(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])[:12]
}

func Read(homeDir string) (InstallState, error) {
	mu.RLock()
	defer mu.RUnlock()

	data, err := os.ReadFile(Path(homeDir))
	if err != nil {
		return InstallState{}, err
	}
	var s InstallState
	if err := json.Unmarshal(data, &s); err != nil {
		return InstallState{}, err
	}
	return s, nil
}

func Write(homeDir string, s InstallState) error {
	mu.Lock()
	defer mu.Unlock()

	dir := filepath.Join(homeDir, stateDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(Path(homeDir), append(data, '\n'), 0o644)
}
// DiscoverManifests returns a map of AgentID -> []string (absolute paths to manifest files)
// by scanning the managed directory in the given home directory.
func DiscoverManifests(homeDir string) (map[model.AgentID][]string, error) {
	mu.RLock()
	defer mu.RUnlock()

	managedDir := filepath.Join(homeDir, stateDir, "managed")
	_, err := os.Stat(managedDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	result := make(map[model.AgentID][]string)
	err = filepath.WalkDir(managedDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".json" {
			agentID := model.AgentID(strings.TrimSuffix(filepath.Base(path), ".json"))
			result[agentID] = append(result[agentID], path)
		}
		return nil
	})
	return result, err
}
