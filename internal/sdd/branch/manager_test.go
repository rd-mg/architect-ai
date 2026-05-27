package branch_test

import (
	"os/exec"
	"strings"
	"testing"

	branch "github.com/rd-mg/architect-ai/internal/sdd/branch"
)

func gitConfig(dir, key, value string) error {
	return exec.Command("git", "-C", dir, "config", key, value).Run()
}

func initTempRepo(t *testing.T, dir string) {
	t.Helper()
	if err := exec.Command("git", "-C", dir, "init").Run(); err != nil {
		t.Fatalf("git init: %v", err)
	}
	// Minimal git config needed for merge commits.
	if err := gitConfig(dir, "user.name", "test"); err != nil {
		t.Fatalf("git config user.name: %v", err)
	}
	if err := gitConfig(dir, "user.email", "test@test.com"); err != nil {
		t.Fatalf("git config user.email: %v", err)
	}
	// Create an initial commit so HEAD resolves.
	if err := exec.Command("git", "-C", dir, "commit", "--allow-empty", "-m", "init").Run(); err != nil {
		t.Fatalf("git commit: %v", err)
	}
}

func TestNewNoGitRepo(t *testing.T) {
	m, err := branch.New("/nonexistent")
	if err != nil {
		t.Fatalf("New on nonexistent dir returned error: %v", err)
	}
	if m.OriginalBranch != "" {
		t.Errorf("Expected empty OriginalBranch for non-git dir, got %q", m.OriginalBranch)
	}
}

func TestCreateApplyBranchOnNonGitRepo(t *testing.T) {
	m, err := branch.New(t.TempDir())
	if err != nil {
		t.Fatalf("New on non-git dir returned error: %v", err)
	}
	result, err := m.CreateApplyBranch("test-change")
	if err != nil {
		t.Fatalf("CreateApplyBranch on non-git dir returned error: %v", err)
	}
	if result.GitAvailable {
		t.Error("Expected GitAvailable=false for non-git dir")
	}
}

func TestNewInGitRepo(t *testing.T) {
	dir := t.TempDir()
	initTempRepo(t, dir)

	// Create and checkout a named branch so we can verify OriginalBranch.
	exec.Command("git", "-C", dir, "checkout", "-b", "main").Run()

	m, err := branch.New(dir)
	if err != nil {
		t.Fatalf("New in git repo returned error: %v", err)
	}
	if m.RepoDir != dir {
		t.Errorf("RepoDir = %q, want %q", m.RepoDir, dir)
	}
	if m.OriginalBranch != "main" {
		t.Errorf("OriginalBranch = %q, want %q", m.OriginalBranch, "main")
	}
}

func TestSanitizeViaBranchName(t *testing.T) {
	dir := t.TempDir()
	initTempRepo(t, dir)

	m, err := branch.New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	result, err := m.CreateApplyBranch("My Feature/1.0_test")
	if err != nil {
		t.Fatalf("CreateApplyBranch: %v", err)
	}
	if !strings.Contains(result.BranchName, "my-feature-1-0") {
		t.Errorf("BranchName = %q, expected to contain %q", result.BranchName, "my-feature-1-0")
	}

	// Clean up: checkout original and delete apply branch.
	exec.Command("git", "-C", dir, "checkout", m.OriginalBranch).Run()
	exec.Command("git", "-C", dir, "branch", "-D", result.BranchName).Run()
}

func TestMergeBack(t *testing.T) {
	dir := t.TempDir()
	initTempRepo(t, dir)
	exec.Command("git", "-C", dir, "checkout", "-b", "main").Run()

	m, err := branch.New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// Create an apply branch.
	result, err := m.CreateApplyBranch("merge-test")
	if err != nil {
		t.Fatalf("CreateApplyBranch: %v", err)
	}
	if !result.GitAvailable {
		t.Fatal("Expected GitAvailable=true in git repo")
	}

	// Make a change on the apply branch.
	if err := exec.Command("git", "-C", dir, "commit", "--allow-empty", "-m", "apply change").Run(); err != nil {
		t.Fatalf("commit on apply branch: %v", err)
	}

	// Merge back.
	if err := m.MergeBack(result, "merge-test"); err != nil {
		t.Fatalf("MergeBack: %v", err)
	}

	// Verify we're back on the original branch.
	out, err := exec.Command("git", "-C", dir, "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		t.Fatalf("git rev-parse: %v", err)
	}
	current := strings.TrimSpace(string(out))
	if current != "main" {
		t.Errorf("After MergeBack on %q, want %q", current, "main")
	}

	// Verify apply branch is deleted.
	if exec.Command("git", "-C", dir, "rev-parse", "--verify", "apply/merge-test").Run() == nil {
		t.Error("Apply branch should have been deleted after merge")
	}
}

func TestMergeBackNonGitResult(t *testing.T) {
	m, err := branch.New("/nonexistent")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	r := &branch.Result{GitAvailable: false}
	if err := m.MergeBack(r, "test"); err != nil {
		t.Errorf("MergeBack on non-git result should succeed: %v", err)
	}
}

func TestAbandon(t *testing.T) {
	dir := t.TempDir()
	initTempRepo(t, dir)
	exec.Command("git", "-C", dir, "checkout", "-b", "main").Run()

	m, err := branch.New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	result, err := m.CreateApplyBranch("abandon-test")
	if err != nil {
		t.Fatalf("CreateApplyBranch: %v", err)
	}

	// Abandon returns to original WITHOUT deleting apply branch.
	if err := m.Abandon(result); err != nil {
		t.Fatalf("Abandon: %v", err)
	}

	// Verify we're back on main.
	out, err := exec.Command("git", "-C", dir, "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		t.Fatalf("git rev-parse: %v", err)
	}
	current := strings.TrimSpace(string(out))
	if current != "main" {
		t.Errorf("After Abandon on %q, want %q", current, "main")
	}

	// Verify apply branch is preserved.
	if exec.Command("git", "-C", dir, "rev-parse", "--verify", "apply/abandon-test").Run() != nil {
		t.Error("Apply branch should be preserved after Abandon")
	}

	// Clean up.
	exec.Command("git", "-C", dir, "branch", "-D", result.BranchName).Run()
}

func TestAbandonNonGitResult(t *testing.T) {
	m, err := branch.New("/nonexistent")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	r := &branch.Result{GitAvailable: false}
	if err := m.Abandon(r); err != nil {
		t.Errorf("Abandon on non-git result should succeed: %v", err)
	}
}
