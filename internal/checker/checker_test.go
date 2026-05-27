package checker

import (
	"errors"
	"testing"
)

func TestCheck_Passed(t *testing.T) {
	c := Check{Name: "pass", Run: func() error { return nil }}
	r := c.RunCheck()
	if r.Status != StatusPassed {
		t.Errorf("expected StatusPassed, got %s", r.Status)
	}
	if r.Name != "pass" {
		t.Errorf("expected name 'pass', got %s", r.Name)
	}
	if r.Message != "" {
		t.Errorf("expected empty message, got %s", r.Message)
	}
}

func TestCheck_Failed(t *testing.T) {
	c := Check{Name: "fail", Run: func() error { return errors.New("boom") }}
	r := c.RunCheck()
	if r.Status != StatusFailed {
		t.Errorf("expected StatusFailed, got %s", r.Status)
	}
	if r.Message != "boom" {
		t.Errorf("expected message 'boom', got %s", r.Message)
	}
}

func TestCheck_NilRun_Skipped(t *testing.T) {
	c := Check{Name: "skip"}
	r := c.RunCheck()
	if r.Status != StatusSkipped {
		t.Errorf("expected StatusSkipped, got %s", r.Status)
	}
	if r.Message != "check not implemented" {
		t.Errorf("expected 'check not implemented', got %s", r.Message)
	}
}

func TestRunAll(t *testing.T) {
	c1 := Check{Name: "a", Run: func() error { return nil }}
	c2 := Check{Name: "b", Run: func() error { return errors.New("err") }}
	c3 := Check{Name: "c"}

	results := RunAll(c1, c2, c3)
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	if results[0].Status != StatusPassed {
		t.Errorf("expected check 'a' to be passed, got %s", results[0].Status)
	}
	if results[1].Status != StatusFailed {
		t.Errorf("expected check 'b' to be failed, got %s", results[1].Status)
	}
	if results[2].Status != StatusSkipped {
		t.Errorf("expected check 'c' to be skipped, got %s", results[2].Status)
	}
}

func TestRunAll_Empty(t *testing.T) {
	results := RunAll()
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestSummarize(t *testing.T) {
	results := []Result{
		{Status: StatusPassed},
		{Status: StatusFailed},
		{Status: StatusPassed},
		{Status: StatusSkipped},
	}
	s := Summarize(results)
	if s.Passed != 2 {
		t.Errorf("expected 2 passed, got %d", s.Passed)
	}
	if s.Failed != 1 {
		t.Errorf("expected 1 failed, got %d", s.Failed)
	}
	if s.Skipped != 1 {
		t.Errorf("expected 1 skipped, got %d", s.Skipped)
	}
}

func TestSummarize_Empty(t *testing.T) {
	s := Summarize(nil)
	if s.Passed != 0 || s.Failed != 0 || s.Skipped != 0 {
		t.Errorf("expected zero summary, got %+v", s)
	}
}

func TestAllPassed_True(t *testing.T) {
	if !AllPassed([]Result{{Status: StatusPassed}, {Status: StatusPassed}}) {
		t.Error("expected true when all passed")
	}
}

func TestAllPassed_False_WhenFailed(t *testing.T) {
	if AllPassed([]Result{{Status: StatusPassed}, {Status: StatusFailed}}) {
		t.Error("expected false when one failed")
	}
}

func TestAllPassed_False_WhenSkipped(t *testing.T) {
	if AllPassed([]Result{{Status: StatusPassed}, {Status: StatusSkipped}}) {
		t.Error("expected false when one skipped")
	}
}

func TestAllPassed_Empty(t *testing.T) {
	if !AllPassed(nil) {
		t.Error("expected true for empty results")
	}
}

func TestError_NoFailures(t *testing.T) {
	results := []Result{
		{Name: "a", Status: StatusPassed},
		{Name: "b", Status: StatusPassed},
	}
	if err := Error(results); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestError_NoFailures_SkippedIncluded(t *testing.T) {
	results := []Result{
		{Name: "a", Status: StatusPassed},
		{Name: "b", Status: StatusSkipped},
	}
	if err := Error(results); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestError_WithFailures(t *testing.T) {
	results := []Result{
		{Name: "a", Status: StatusPassed},
		{Name: "b", Status: StatusFailed, Message: "boom"},
		{Name: "c", Status: StatusFailed, Message: "bam"},
	}
	err := Error(results)
	if err == nil {
		t.Fatal("expected error")
	}
	msg := err.Error()
	if len(msg) == 0 {
		t.Error("expected non-empty error message")
	}
}

func TestError_Empty(t *testing.T) {
	if err := Error(nil); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}
