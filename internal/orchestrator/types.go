// Package orchestrator provides runtime dispatch for Adaptive Reasoning Gate integration.
// It computes D1-D5 dimension scores from task context, validates sub-agent response headers,
// and persists scoring decisions to Engram.
package orchestrator

// Dims48 is a fixed-size array representing D1-D5 dimension scores (each 0-3).
// D5 (Security) is computed server-side; agents only declare D1-D4.
type Dims48 [5]int

// TaskContext describes a task to be scored by CognitiveScorer.
type TaskContext struct {
	PhaseName         string // e.g. "sdd-apply", "sdd-design"
	ChangeID          string // e.g. "wire-cognitive-mode-runtime"
	FileCount         int    // number of files affected
	CrossPackage      bool   // true if change spans multiple packages
	HasSpecs          bool   // true if specs exist and are complete
	AttemptCount      int    // number of prior attempts for this phase (0 = first)
	ContextEstimateKB int    // estimated context size in KB

	// Economic/Empirical posture triggers
	IsCostSensitive   bool // task involves cost/ROI/quota decisions
	IsMeasurementTask bool // task needs benchmarks/measurements

	// D5: Security dimension
	IsSecuritySensitive bool // task involves security-sensitive operations
}

// ValidationResult represents the outcome of validating a sub-agent's response header
// against the expected mode computed by CognitiveScorer.
type ValidationResult struct {
	Matched       bool     // true if declared mode matches expected mode
	DeclaredMode  int      // mode declared in the sub-agent response (0 if parse failed)
	ExpectedMode  int      // mode computed by CognitiveScorer
	Postures      []string // postures computed by CognitiveScorer
	ExpectedDims  Dims48   // dimensions used to compute expected mode
	DeclaredDims  Dims48   // dimensions declared in the sub-agent response
	Err           error    // parse error, if any
	ForceMode3    bool     // circuit breaker: escalate to Mode 3
	RePromptCount int      // number of re-prompts issued (0 or 1)
}
