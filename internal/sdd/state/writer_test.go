package state

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestInitialState_ContainsRequiredFields(t *testing.T) {
	content := InitialState("auth-feature", "myproject", "hybrid", "interactive", "ask-on-risk")
	required := []string{
		"change_name:",
		"project:",
		"artifact_store:",
		"execution_mode:",
		"delivery_strategy:",
		"circuit_breaker:",
		"sdd-init:",
		"sdd-apply:",
		"requires:",
		"tdd_mode:",
	}
	for _, field := range required {
		if !strings.Contains(content, field) {
			t.Errorf("missing field in generated state: %s", field)
		}
	}
}

func TestInitialState_InitMarkedCompleted(t *testing.T) {
	content := InitialState("test", "proj", "engram", "automatic", "auto-chain")
	if !strings.Contains(content, `sdd-init:     { status: "completed"`) {
		t.Error("sdd-init should be marked completed in initial state")
	}
}

func TestWriteSddState_Atomic(t *testing.T) {
	dir := t.TempDir()
	content := InitialState("test", "proj", "hybrid", "interactive", "ask-on-risk")
	if err := WriteSddState(dir, content); err != nil {
		t.Fatalf("WriteSddState: %v", err)
	}
	stateFile := filepath.Join(dir, "sdd-state.yaml")
	if _, err := os.Stat(stateFile); os.IsNotExist(err) {
		t.Error("sdd-state.yaml not created")
	}
	if _, err := os.Stat(stateFile + ".tmp"); !os.IsNotExist(err) {
		t.Error("tmp file should be removed after atomic write")
	}
	if _, err := os.Stat(stateFile + ".lock"); !os.IsNotExist(err) {
		t.Error("lock file should be removed after write")
	}
}

func TestValidateStateYAML_MissingFile(t *testing.T) {
	issues := ValidateStateYAML(t.TempDir())
	if len(issues) == 0 {
		t.Error("should report issues for missing sdd-state.yaml")
	}
}

func TestCheckPrerequisites_BlockedOutOfOrder(t *testing.T) {
	content := InitialState("test-dag", "proj", "hybrid", "interactive", "ask-on-risk")
	s, err := ParseSddState(content)
	if err != nil {
		t.Fatalf("ParseSddState: %v", err)
	}

	// sdd-spec requires sdd-propose. Since sdd-propose is pending, checking spec should fail.
	err = s.CheckPrerequisites("sdd-spec")
	if err == nil {
		t.Error("expected error checking prerequisites for sdd-spec when sdd-propose is pending")
	}

	// sdd-onboard has no requirements, should succeed.
	err = s.CheckPrerequisites("sdd-onboard")
	if err != nil {
		t.Errorf("expected no error for sdd-onboard, got: %v", err)
	}

	// Mark sdd-propose as completed and check again.
	// Note sdd-propose also requires sdd-explore. Let's make both completed.
	explore := s.Phases["sdd-explore"]
	explore.Status = "completed"
	s.Phases["sdd-explore"] = explore

	propose := s.Phases["sdd-propose"]
	propose.Status = "completed"
	s.Phases["sdd-propose"] = propose

	err = s.CheckPrerequisites("sdd-spec")
	if err != nil {
		t.Errorf("expected no error for sdd-spec after completion of prerequisites, got: %v", err)
	}
}

func TestResultContractValidation_ShellScript(t *testing.T) {
	// Create a temporary file with a valid Result Contract JSON
	dir := t.TempDir()
	validJSON := `{
		"status": "completed",
		"phase": "sdd-explore",
		"change_name": "test-change",
		"executive_summary": "Done",
		"artifacts": ["file1.go"],
		"next_recommended": "sdd-propose",
		"risks": [],
		"skill_resolution": {
			"status": "paths-injected",
			"skills_used": ["ripgrep"],
			"fallback_reason": null
		},
		"attempt_number": 1,
		"blocked_reason": null
	}`
	validFile := filepath.Join(dir, "valid.json")
	if err := os.WriteFile(validFile, []byte(validJSON), 0644); err != nil {
		t.Fatalf("WriteFile valid.json: %v", err)
	}

	// Create a temporary file with an invalid Result Contract JSON (missing status)
	invalidJSON := `{
		"phase": "sdd-explore",
		"change_name": "test-change",
		"executive_summary": "Done",
		"artifacts": ["file1.go"],
		"next_recommended": "sdd-propose",
		"risks": [],
		"skill_resolution": {
			"status": "paths-injected",
			"skills_used": ["ripgrep"],
			"fallback_reason": null
		},
		"attempt_number": 1,
		"blocked_reason": null
	}`
	invalidFile := filepath.Join(dir, "invalid.json")
	if err := os.WriteFile(invalidFile, []byte(invalidJSON), 0644); err != nil {
		t.Fatalf("WriteFile invalid.json: %v", err)
	}

	// Find the validation script. It should be in the root .atl/scripts/validate-result-contract.sh
	// Since we are running tests from internal/sdd/state/, let's walk up to find it
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	scriptPath := filepath.Join(wd, "../../../.atl/scripts/validate-result-contract.sh")

	// Verify script exists
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		t.Skipf("Skip result contract script validation test: script not found at %s", scriptPath)
	}

	// Run with valid JSON
	// We can use exec.Command to run it
	importExec := func(script string, jsonFile string) error {
		importOS := os.NewFile(0, "stdin")
		defer importOS.Close()
		cmd := strings.Join([]string{script, jsonFile}, " ")
		importCmd := os.Getenv("SHELL")
		if importCmd == "" {
			importCmd = "/bin/bash"
		}
		// Run shell command
		importProc, err := os.StartProcess(importCmd, []string{importCmd, "-c", cmd}, &os.ProcAttr{
			Files: []*os.File{importOS, os.Stdout, os.Stderr},
		})
		if err != nil {
			return err
		}
		state, err := importProc.Wait()
		if err != nil {
			return err
		}
		if !state.Success() {
			return fmt.Errorf("exit status %d", state.ExitCode())
		}
		return nil
	}

	if err := importExec(scriptPath, validFile); err != nil {
		t.Errorf("validation script failed on valid JSON: %v", err)
	}

	if err := importExec(scriptPath, invalidFile); err == nil {
		t.Error("validation script succeeded on invalid JSON, expected failure")
	}
}

func TestCircuitBreaker_MaxAttempts(t *testing.T) {
	content := InitialState("cb-test", "proj", "engram", "automatic", "auto-chain")
	s, err := ParseSddState(content)
	if err != nil {
		t.Fatalf("ParseSddState: %v", err)
	}

	// Initially, attempts should be 0. Let's record attempts.
	tripped, err := s.RecordAttempt("sdd-apply")
	if err != nil {
		t.Fatalf("RecordAttempt: %v", err)
	}
	if tripped {
		t.Error("should not trip on attempt 1")
	}
	if s.CircuitBreaker.AttemptCounts["sdd-apply"] != 1 {
		t.Errorf("expected 1 attempt, got %d", s.CircuitBreaker.AttemptCounts["sdd-apply"])
	}

	tripped, err = s.RecordAttempt("sdd-apply")
	if tripped {
		t.Error("should not trip on attempt 2")
	}

	// 3rd attempt should trip — max_attempts is 3, trip on >= 3.
	tripped, err = s.RecordAttempt("sdd-apply")
	if !tripped {
		t.Error("should trip on attempt 3")
	}

	if s.Phases["sdd-apply"].Status != "abandoned" {
		t.Errorf("expected phase status to be 'abandoned', got %s", s.Phases["sdd-apply"].Status)
	}
	if len(s.CircuitBreaker.AbandonedPhases) != 1 || s.CircuitBreaker.AbandonedPhases[0] != "sdd-apply" {
		t.Errorf("expected abandoned phases list to contain sdd-apply, got: %v", s.CircuitBreaker.AbandonedPhases)
	}
}

func TestApplyContinuity_Resume(t *testing.T) {
	// Test parsing and structure of ApplyProgress
	progressYAML := `change_name: "test-change"
started_at: "2026-05-20T00:52:04Z"
updated_at: "2026-05-20T00:52:04Z"
tasks:
  - id: "task 1"
    description: "First task"
    status: "completed"
    completed_at: "2026-05-20T00:53:50Z"
  - id: "task 2"
    description: "Second task"
    status: "pending"
    completed_at: ""
`
	var p ApplyProgress
	err := yaml.Unmarshal([]byte(progressYAML), &p)
	if err != nil {
		t.Fatalf("Unmarshal ApplyProgress: %v", err)
	}

	if p.ChangeName != "test-change" {
		t.Errorf("expected change_name test-change, got %s", p.ChangeName)
	}
	if len(p.Tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(p.Tasks))
	}
	if p.Tasks[0].Status != "completed" || p.Tasks[1].Status != "pending" {
		t.Errorf("task statuses incorrect: %s, %s", p.Tasks[0].Status, p.Tasks[1].Status)
	}
}
