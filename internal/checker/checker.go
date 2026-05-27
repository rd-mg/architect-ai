// Package checker provides a lightweight, composable check/validation utility.
// Each Check is a named function that returns either passed, failed, or skipped.
// Use RunAll to batch-execute checks and Summarize / Error to inspect results.
package checker

import "fmt"

// Status represents the outcome of a single check.
type Status string

const (
	StatusPassed  Status = "passed"
	StatusFailed  Status = "failed"
	StatusSkipped Status = "skipped"
)

// Result holds the outcome of a single check.
type Result struct {
	Name    string
	Status  Status
	Message string
}

// Check defines a single checkable condition.
type Check struct {
	Name string
	Run  func() error
}

// RunCheck executes the check and returns a Result.
func (c Check) RunCheck() Result {
	if c.Run == nil {
		return Result{Name: c.Name, Status: StatusSkipped, Message: "check not implemented"}
	}
	if err := c.Run(); err != nil {
		return Result{Name: c.Name, Status: StatusFailed, Message: err.Error()}
	}
	return Result{Name: c.Name, Status: StatusPassed}
}

// RunAll executes multiple checks and returns all results in order.
func RunAll(checks ...Check) []Result {
	results := make([]Result, 0, len(checks))
	for _, c := range checks {
		results = append(results, c.RunCheck())
	}
	return results
}

// Summary counts results by status.
type Summary struct {
	Passed  int
	Failed  int
	Skipped int
}

// Summarize produces a Summary from a slice of results.
func Summarize(results []Result) Summary {
	var s Summary
	for _, r := range results {
		switch r.Status {
		case StatusPassed:
			s.Passed++
		case StatusFailed:
			s.Failed++
		case StatusSkipped:
			s.Skipped++
		}
	}
	return s
}

// AllPassed returns true only when every result has StatusPassed.
func AllPassed(results []Result) bool {
	for _, r := range results {
		if r.Status != StatusPassed {
			return false
		}
	}
	return true
}

// Error returns a single error summarizing all failed checks.
// Returns nil when no checks failed (passed and skipped are ignored).
func Error(results []Result) error {
	var failed []Result
	for _, r := range results {
		if r.Status == StatusFailed {
			failed = append(failed, r)
		}
	}
	if len(failed) == 0 {
		return nil
	}
	var s string
	for i, r := range failed {
		if i > 0 {
			s += "; "
		}
		s += fmt.Sprintf("%s: %s", r.Name, r.Message)
	}
	return fmt.Errorf("%d check(s) failed: %s", len(failed), s)
}
