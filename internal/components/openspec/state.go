// Package openspec provides runtime helpers for the OpenSpec persistence mode:
// state.yaml read/write/validate, delta-spec merge (TOPIC-10), and spec index
// generation (TOPIC-13). All functions are pure Go; no agent calls.
package openspec

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// SchemaVersion is the state.yaml schema understood by this build.
const SchemaVersion = 1

// ValidPhases is the closed set of SDD phase names. Order matters for
// presentation but not for validation.
var ValidPhases = []string{
	"sdd-explore", "sdd-propose", "sdd-spec", "sdd-design",
	"sdd-tasks", "sdd-apply", "sdd-verify", "sdd-archive",
}

// ValidStatuses is the closed set of phase status values.
var ValidStatuses = []string{
	"pending", "in_progress", "completed", "skipped", "failed",
	"abandoned", "infrastructure_blocked", "partially_completed",
	"rollback_in_progress",
}

// ValidArtifactStores is the closed set of persistence mode values.
var ValidArtifactStores = []string{
	"engram", "openspec", "hybrid", "none",
}

// State is the in-memory representation of state.yaml V1.
type State struct {
	SchemaVersion   int               `yaml:"schema_version"`
	ChangeName      string            `yaml:"change_name"`
	CreatedAt       time.Time         `yaml:"created_at"`
	UpdatedAt       time.Time         `yaml:"updated_at"`
	ArtifactStore   string            `yaml:"artifact_store"`
	Phases          map[string]*Phase `yaml:"phases"`
	Metering        *Metering         `yaml:"metering,omitempty"`
	CircuitBreaker  CircuitBreaker    `yaml:"circuit_breaker"`
}

// Phase is per-phase state.
type Phase struct {
	Status      string         `yaml:"status"`
	StartedAt   *time.Time     `yaml:"started_at,omitempty"`
	CompletedAt *time.Time     `yaml:"completed_at,omitempty"`
	Error       string         `yaml:"error,omitempty"`
	Artifact    string         `yaml:"artifact,omitempty"`
	Artifacts   []string       `yaml:"artifacts,omitempty"`
	Model       string         `yaml:"model,omitempty"`
	DependsOn   []string       `yaml:"depends_on,omitempty"`
	Meta        map[string]any `yaml:"meta,omitempty"` // loose; unknown fields allowed
}

// Metering is the optional session-metering rollup.
type Metering struct {
	TotalTokens      int     `yaml:"total_tokens"`
	Sessions         int     `yaml:"sessions"`
	EstimatedCostUSD float64 `yaml:"estimated_cost_usd"`
}

// ResetEvent records a manual circuit-breaker reset for audit purposes.
type ResetEvent struct {
	Phase             string `yaml:"phase"`
	Timestamp         string `yaml:"timestamp"`
	Justification     string `yaml:"justification"`
	ResetBy           string `yaml:"reset_by"`
	PriorAttemptCount int    `yaml:"prior_attempt_count"`
}

// CircuitBreaker tracks attempt counts and state for the CB protocol.
type CircuitBreaker struct {
	Enabled            bool           `yaml:"enabled"`
	MaxAttempts        int            `yaml:"max_attempts"`
	AttemptCounts      map[string]int `yaml:"attempt_counts"`
	InfraAttemptCounts map[string]int `yaml:"infra_attempt_counts"`
	ResetEvents        []ResetEvent   `yaml:"reset_events"`
	AbandonedPhases    []string       `yaml:"abandoned_phases"`
}

// Typed errors for CLI UX. Export everything the CLI needs to switch on.
var (
	ErrSchemaVersion      = errors.New("schema_version unsupported")
	ErrMissingField       = errors.New("required field missing")
	ErrTimestampOrder     = errors.New("updated_at < created_at")
	ErrEnum               = errors.New("value not in allowed enum")
	ErrUnknownPhase       = errors.New("phases contains unknown key")
	ErrDanglingDepends    = errors.New("depends_on points to missing phase")
	ErrCycle              = errors.New("phases graph has a cycle")
	ErrChangeNameMismatch = errors.New("change_name does not match parent directory")
)

var kebabRe = regexp.MustCompile(`^[a-z][a-z0-9]*(-[a-z0-9]+)*$`)

// Load reads and validates state.yaml at the given path.
// On success returns a fully-populated *State.
// On failure returns nil + a typed error.
func Load(path string) (*State, error) {
	bs, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var s State
	if err := yaml.Unmarshal(bs, &s); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	// Change name must match parent directory name.
	parent := filepath.Base(filepath.Dir(path))
	if err := Validate(&s, parent); err != nil {
		return nil, err
	}
	return &s, nil
}

// Save writes state.yaml atomically (unique-tmp + rename + O_EXCL lock).
//
// Atomicity guarantee:
//   - Uses a per-PID+nanosecond .tmp path so concurrent writers never share
//     the same temporary file (eliminates the TOCTOU collision on .tmp).
//   - Uses O_CREATE|O_EXCL for the lockfile so only one writer can hold
//     the lock at a time (kernel-level atomicity, no Stat+Write race).
//   - Rename is atomic on all POSIX filesystems (Linux, macOS).
//
// Precondition:  path is writable; s is a valid *State.
// Postcondition: path contains the marshalled s; lockfile removed.
func Save(path string, s *State) error {
	s.UpdatedAt = time.Now().UTC()
	parent := filepath.Base(filepath.Dir(path))
	if err := Validate(s, parent); err != nil {
		return err
	}
	out, err := yaml.Marshal(s)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	lockPath := path + ".lock"
	lf, lockErr := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if lockErr != nil {
		if os.IsExist(lockErr) {
			info, statErr := os.Stat(lockPath)
			if statErr == nil && time.Since(info.ModTime()) > 30*time.Second {
				_ = os.Remove(lockPath)
				return Save(path, s)
			}
			return fmt.Errorf("state file locked by another process (%s)", lockPath)
		}
		return fmt.Errorf("acquire lock: %w", lockErr)
	}
	_, _ = fmt.Fprintf(lf, "%d", os.Getpid())
	_ = lf.Close()
	defer func() { _ = os.Remove(lockPath) }()

	tmp := fmt.Sprintf("%s.%d.%d.tmp", path, os.Getpid(), time.Now().UnixNano())
	defer func() { _ = os.Remove(tmp) }()

	if err := os.WriteFile(tmp, out, 0o644); err != nil {
		return fmt.Errorf("write tmp: %w", err)
	}
	if f, ferr := os.OpenFile(tmp, os.O_RDWR, 0o644); ferr == nil {
		_ = f.Sync()
		_ = f.Close()
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("rename: %w", err)
	}
	return nil
}

// Validate enforces I1..I11. The caller passes the parent folder name so we
// can enforce I2 (change_name matches folder) without re-reading filesystem.
func Validate(s *State, folderName string) error {
	// I1
	if s.SchemaVersion != SchemaVersion {
		return fmt.Errorf("%w: got %d, want %d", ErrSchemaVersion, s.SchemaVersion, SchemaVersion)
	}
	// I2
	if s.ChangeName == "" {
		return fmt.Errorf("%w: change_name", ErrMissingField)
	}
	if !kebabRe.MatchString(s.ChangeName) {
		return fmt.Errorf("change_name %q not kebab-case", s.ChangeName)
	}
	if folderName != "" && folderName != s.ChangeName {
		return fmt.Errorf("%w: folder %q vs state %q", ErrChangeNameMismatch, folderName, s.ChangeName)
	}
	// I3
	if s.CreatedAt.IsZero() {
		return fmt.Errorf("%w: created_at", ErrMissingField)
	}
	if s.UpdatedAt.IsZero() {
		return fmt.Errorf("%w: updated_at", ErrMissingField)
	}
	if s.UpdatedAt.Before(s.CreatedAt) {
		return ErrTimestampOrder
	}
	// I4
	if !inSet(s.ArtifactStore, ValidArtifactStores) {
		return fmt.Errorf("%w: artifact_store=%q", ErrEnum, s.ArtifactStore)
	}
	// I5, I6
	for name, ph := range s.Phases {
		if !inSet(name, ValidPhases) {
			return fmt.Errorf("%w: %q", ErrUnknownPhase, name)
		}
		if ph == nil {
			return fmt.Errorf("%w: phases.%s is null", ErrMissingField, name)
		}
		if !inSet(ph.Status, ValidStatuses) {
			return fmt.Errorf("%w: phases.%s.status=%q", ErrEnum, name, ph.Status)
		}
		// I7
		if ph.Status == "completed" && ph.CompletedAt == nil {
			return fmt.Errorf("%w: phases.%s.completed_at", ErrMissingField, name)
		}
		// I8
		if ph.Status == "in_progress" && ph.StartedAt == nil {
			return fmt.Errorf("%w: phases.%s.started_at", ErrMissingField, name)
		}
		// Failed requires error string.
		if ph.Status == "failed" && ph.Error == "" {
			return fmt.Errorf("%w: phases.%s.error", ErrMissingField, name)
		}
		// I9
		for _, dep := range ph.DependsOn {
			if !inSet(dep, ValidPhases) {
				return fmt.Errorf("%w: phases.%s.depends_on=%q", ErrEnum, name, dep)
			}
			if _, ok := s.Phases[dep]; !ok {
				return fmt.Errorf("%w: %s -> %s", ErrDanglingDepends, name, dep)
			}
		}
	}
	// I10 — cycle check (defence-in-depth; topology is fixed)
	if err := detectCycle(s.Phases); err != nil {
		return err
	}
	return nil
}

// detectCycle runs Kahn's algorithm. If any nodes remain after processing,
// there is a cycle.
func detectCycle(phases map[string]*Phase) error {
	// inDeg[n] is the number of dependencies n has.
	inDeg := map[string]int{}
	for name, ph := range phases {
		inDeg[name] = len(ph.DependsOn)
	}

	queue := []string{}
	for n, d := range inDeg {
		if d == 0 {
			queue = append(queue, n)
		}
	}

	sort.Strings(queue) // deterministic
	visited := 0
	for len(queue) > 0 {
		n := queue[0]
		queue = queue[1:]
		visited++

		// Decrement in-degree of all nodes that depend on n.
		for name, ph := range phases {
			for _, dep := range ph.DependsOn {
				if dep == n {
					inDeg[name]--
					if inDeg[name] == 0 {
						queue = append(queue, name)
					}
				}
			}
		}
	}

	if visited < len(phases) {
		return ErrCycle
	}
	return nil
}

func inSet(v string, set []string) bool {
	for _, x := range set {
		if v == x {
			return true
		}
	}
	return false
}

func (s *State) RecordDomainAttempt(phase string) (tripped bool) {
	if s.CircuitBreaker.AttemptCounts == nil {
		s.CircuitBreaker.AttemptCounts = make(map[string]int)
	}
	s.CircuitBreaker.AttemptCounts[phase]++

	maxAttempts := s.CircuitBreaker.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 3
	}
	if s.CircuitBreaker.Enabled && s.CircuitBreaker.AttemptCounts[phase] >= maxAttempts {
		if ph, ok := s.Phases[phase]; ok {
			ph.Status = "abandoned"
			s.Phases[phase] = ph
		}
		s.CircuitBreaker.AbandonedPhases = append(s.CircuitBreaker.AbandonedPhases, phase)
		return true
	}
	return false
}

func (s *State) RecordInfraAttempt(phase string, infraMax int) (infraTripped bool) {
	if s.CircuitBreaker.InfraAttemptCounts == nil {
		s.CircuitBreaker.InfraAttemptCounts = make(map[string]int)
	}
	if infraMax <= 0 {
		infraMax = 5
	}
	s.CircuitBreaker.InfraAttemptCounts[phase]++
	return s.CircuitBreaker.InfraAttemptCounts[phase] >= infraMax
}

func (s *State) ResetPhase(phase, justification, resetBy string) error {
	if strings.TrimSpace(justification) == "" {
		return fmt.Errorf("reset requires a non-empty justification")
	}

	priorResets := 0
	for _, ev := range s.CircuitBreaker.ResetEvents {
		if ev.Phase == phase {
			priorResets++
		}
	}
	if priorResets >= 3 {
		return fmt.Errorf("phase %q has been manually reset %d times", phase, priorResets)
	}

	prior := s.CircuitBreaker.AttemptCounts[phase]
	s.CircuitBreaker.ResetEvents = append(s.CircuitBreaker.ResetEvents, ResetEvent{
		Phase:             phase,
		Timestamp:         time.Now().UTC().Format(time.RFC3339),
		Justification:     strings.TrimSpace(justification),
		ResetBy:           resetBy,
		PriorAttemptCount: prior,
	})

	if s.CircuitBreaker.AttemptCounts == nil {
		s.CircuitBreaker.AttemptCounts = make(map[string]int)
	}
	s.CircuitBreaker.AttemptCounts[phase] = 0

	var newAbandoned []string
	for _, p := range s.CircuitBreaker.AbandonedPhases {
		if p != phase {
			newAbandoned = append(newAbandoned, p)
		}
	}
	s.CircuitBreaker.AbandonedPhases = newAbandoned

	if ph, ok := s.Phases[phase]; ok {
		ph.Status = "pending"
		s.Phases[phase] = ph
	}
	return nil
}
