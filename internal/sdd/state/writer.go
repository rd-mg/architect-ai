package state

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type PhaseState struct {
	Status      string   `yaml:"status"`
	CompletedAt string   `yaml:"completed_at"`
	Artifacts   []string `yaml:"artifacts"`
	Requires    []string `yaml:"requires"`
}

type CircuitBreakerState struct {
	Enabled         bool           `yaml:"enabled"`
	MaxAttempts     int            `yaml:"max_attempts"`
	AttemptCounts   map[string]int `yaml:"attempt_counts"`
	AbandonedPhases []string       `yaml:"abandoned_phases"`
}

type SDDState struct {
	Version          string                 `yaml:"version"`
	ChangeName       string                 `yaml:"change_name"`
	Project          string                 `yaml:"project"`
	StartedAt        string                 `yaml:"started_at"`
	ArtifactStore    string                 `yaml:"artifact_store"`
	ExecutionMode    string                 `yaml:"execution_mode"`
	DeliveryStrategy string                 `yaml:"delivery_strategy"`
	TDDMode          bool                   `yaml:"tdd_mode"`
	Phases           map[string]PhaseState  `yaml:"phases"`
	CircuitBreaker   CircuitBreakerState    `yaml:"circuit_breaker"`
}

type ApplyTask struct {
	ID            string   `yaml:"id"`
	Description   string   `yaml:"description"`
	Status        string   `yaml:"status"`
	CompletedAt   string   `yaml:"completed_at"`
	FilesModified []string `yaml:"files_modified"`
	CommitHash    string   `yaml:"commit_hash"`
	LinesAdded    int      `yaml:"lines_added"`
	LinesDeleted  int      `yaml:"lines_deleted"`
}

type ApplyTotals struct {
	Completed    int `yaml:"completed"`
	Running      int `yaml:"running"`
	Pending      int `yaml:"pending"`
	Failed       int `yaml:"failed"`
	LinesAdded   int `yaml:"lines_added"`
	LinesDeleted int `yaml:"lines_deleted"`
}

type ApplyProgress struct {
	Version          string       `yaml:"version"`
	ChangeName       string       `yaml:"change_name"`
	ApplyBranch      string       `yaml:"apply_branch"`
	StartedAt        string       `yaml:"started_at"`
	LastUpdated      string       `yaml:"last_updated"`
	ArtifactStore    string       `yaml:"artifact_store"`
	DeliveryStrategy string       `yaml:"delivery_strategy"`
	CurrentSlice     int          `yaml:"current_slice"`
	TotalSlices      int          `yaml:"total_slices"`
	Tasks            []ApplyTask  `yaml:"tasks"`
	Totals           ApplyTotals  `yaml:"totals"`
}

// InitialState generates the starter sdd-state.yaml content
func InitialState(changeName, project, artifactStore, executionMode, deliveryStrategy string) string {
	nowStr := time.Now().UTC().Format(time.RFC3339)
	return fmt.Sprintf(`# .atl/sdd-state.yaml — AUTO-MANAGED by architect-ai
# Do not edit manually. Use /sdd-* commands to update phases.
version: "3.0"
change_name: %q
project: %q
started_at: %q
artifact_store: %q
execution_mode: %q
delivery_strategy: %q
tdd_mode: false

phases:
  sdd-init:     { status: "completed", completed_at: %q, artifacts: [] }
  sdd-onboard:  { status: "pending", completed_at: "", artifacts: [], requires: [] }
  sdd-explore:  { status: "pending", completed_at: "", artifacts: [], requires: [] }
  sdd-propose:  { status: "pending", completed_at: "", artifacts: [], requires: ["sdd-explore"] }
  sdd-spec:     { status: "pending", completed_at: "", artifacts: [], requires: ["sdd-propose"] }
  sdd-design:   { status: "pending", completed_at: "", artifacts: [], requires: ["sdd-spec"] }
  sdd-tasks:    { status: "pending", completed_at: "", artifacts: [], requires: ["sdd-design"] }
  sdd-apply:    { status: "pending", completed_at: "", artifacts: [], requires: ["sdd-tasks"], apply_branch: "", current_slice: 0, total_slices: 1 }
  sdd-verify:   { status: "pending", completed_at: "", artifacts: [], requires: ["sdd-apply"] }
  sdd-archive:  { status: "pending", completed_at: "", artifacts: [], requires: ["sdd-verify"] }

circuit_breaker:
  enabled: true
  max_attempts: 3
  attempt_counts: {}
  abandoned_phases: []
`,
		changeName, project, nowStr, artifactStore, executionMode, deliveryStrategy, nowStr,
	)
}

// WriteSddState writes sdd-state.yaml atomically with lock
func WriteSddState(atDir, content string) error {
	stateFile := filepath.Join(atDir, "sdd-state.yaml")
	tmpFile := stateFile + ".tmp"
	lockFile := stateFile + ".lock"

	if info, err := os.Stat(lockFile); err == nil {
		if time.Since(info.ModTime()) > 30*time.Second {
			os.Remove(lockFile)
		} else {
			return fmt.Errorf("state file is locked — another process is writing")
		}
	}
	if err := os.WriteFile(lockFile, []byte(fmt.Sprintf("%d", os.Getpid())), 0644); err != nil {
		return fmt.Errorf("acquire lock: %w", err)
	}
	defer os.Remove(lockFile)

	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("write tmp: %w", err)
	}
	return os.Rename(tmpFile, stateFile)
}

// ValidateStateYAML does basic structural validation
func ValidateStateYAML(atDir string) []string {
	data, err := os.ReadFile(filepath.Join(atDir, "sdd-state.yaml"))
	if err != nil {
		return []string{fmt.Sprintf("sdd-state.yaml not found: %v", err)}
	}
	content := string(data)
	var issues []string
	for _, field := range []string{"version:", "change_name:", "project:", "artifact_store:", "execution_mode:", "delivery_strategy:", "circuit_breaker:"} {
		if !strings.Contains(content, field) {
			issues = append(issues, fmt.Sprintf("missing field: %s", field))
		}
	}
	return issues
}

// ParseSddState parses yaml content into SDDState struct
func ParseSddState(content string) (*SDDState, error) {
	var s SDDState
	err := yaml.Unmarshal([]byte(content), &s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// CheckPrerequisites returns nil if all prerequisites for target phase are completed, otherwise returns error
func (s *SDDState) CheckPrerequisites(phase string) error {
	pState, exists := s.Phases[phase]
	if !exists {
		return fmt.Errorf("phase %s not found", phase)
	}
	for _, req := range pState.Requires {
		reqState, reqExists := s.Phases[req]
		if !reqExists {
			return fmt.Errorf("prerequisite phase %s for target %s not found", req, phase)
		}
		if reqState.Status != "completed" {
			return fmt.Errorf("prerequisite phase %s is %s, not completed", req, reqState.Status)
		}
	}
	return nil
}

// RecordAttempt increments the attempt count for a phase and checks if circuit breaker trips
// Returns true if tripped, error if any issues
func (s *SDDState) RecordAttempt(phase string) (bool, error) {
	if s.CircuitBreaker.AttemptCounts == nil {
		s.CircuitBreaker.AttemptCounts = make(map[string]int)
	}
	s.CircuitBreaker.AttemptCounts[phase]++
	attempts := s.CircuitBreaker.AttemptCounts[phase]
	if s.CircuitBreaker.Enabled && attempts >= s.CircuitBreaker.MaxAttempts {
		pState, exists := s.Phases[phase]
		if exists {
			pState.Status = "abandoned"
			s.Phases[phase] = pState
		}
		s.CircuitBreaker.AbandonedPhases = append(s.CircuitBreaker.AbandonedPhases, phase)
		return true, nil
	}
	return false, nil
}
