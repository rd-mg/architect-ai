package pipeline

type Stage string

const (
	StagePrepare  Stage = "prepare"
	StageApply    Stage = "apply"
	StageRollback Stage = "rollback"
)

type Step interface {
	ID() string
	Run() error
}

type RollbackStep interface {
	Step
	Rollback() error
}

// FailurePolicy controls how the runner behaves when a step fails.
type FailurePolicy int

const (
	// StopOnError stops execution at the first failed step (default).
	StopOnError FailurePolicy = iota
	// ContinueOnError continues executing remaining steps, collecting all errors.
	ContinueOnError
)

// ProgressEvent is emitted by the runner as each step starts and completes.
type ProgressEvent struct {
	StepID      string
	Stage       Stage
	Status      StepStatus
	Notes       string
	Err         error
	ParallelIdx int // index within concurrent group (0 for sequential execution)
	GroupSize   int // total steps in this group (1 for sequential execution)
}

// ProgressFunc is a callback invoked for every step lifecycle event.
// When used with RunGroup, the callback may be called concurrently from
// multiple goroutines. Callers must ensure their ProgressFunc is concurrent-safe
// (e.g., by sending events on a channel that the TUI model reads).
type ProgressFunc func(ProgressEvent)

// StepGroup is a set of steps that can execute concurrently.
// All steps in a group must be free of write conflicts with each other.
// Groups themselves are executed sequentially (group[0] → group[1] → ...).
type StepGroup struct {
	Steps []Step
}

// SingleGroup wraps a flat slice of steps into a single StepGroup.
// This is the backward-compatible path: steps execute sequentially within
// one group (no fan-out).
func SingleGroup(steps ...Step) StepGroup {
	return StepGroup{Steps: steps}
}

// StagePlan contains ordered step groups for each stage.
// Prepare groups run first, then Apply groups.
// Within each group, steps execute concurrently.
// Groups themselves run sequentially.
type StagePlan struct {
	Prepare []StepGroup
	Apply   []StepGroup
}
