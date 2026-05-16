package branch

import (
    "fmt"
    "os/exec"
    "strings"
)

// Manager handles git branch lifecycle for sdd-apply isolation
type Manager struct {
    RepoDir        string
    OriginalBranch string
}

// Result holds branch creation outcome
type Result struct {
    BranchName     string
    OriginalBranch string
    GitAvailable   bool
}

// New creates a Manager, auto-detecting current branch.
// Returns a no-op manager (GitAvailable=false) if git is absent.
func New(repoDir string) (*Manager, error) {
    if !gitAvailable() {
        return &Manager{RepoDir: repoDir}, nil
    }
    out, err := exec.Command("git", "-C", repoDir,
        "rev-parse", "--abbrev-ref", "HEAD").Output()
    if err != nil {
        return &Manager{RepoDir: repoDir}, nil // not a git repo — silent no-op
    }
    return &Manager{
        RepoDir:        repoDir,
        OriginalBranch: strings.TrimSpace(string(out)),
    }, nil
}

// CreateApplyBranch creates the isolation branch before any file edits.
func (m *Manager) CreateApplyBranch(changeName string) (*Result, error) {
    if !gitAvailable() || m.OriginalBranch == "" {
        return &Result{GitAvailable: false}, nil
    }
    if err := m.verifyClean(); err != nil {
        return nil, fmt.Errorf("repo dirty: %w", err)
    }

    branchName := "apply/" + sanitize(changeName)

    // Drop stale branch silently
    exec.Command("git", "-C", m.RepoDir, "branch", "-D", branchName).Run() //nolint:errcheck

    if out, err := exec.Command("git", "-C", m.RepoDir,
        "checkout", "-b", branchName).CombinedOutput(); err != nil {
        return nil, fmt.Errorf("create branch %s: %w — %s", branchName, err, out)
    }

    return &Result{
        BranchName:     branchName,
        OriginalBranch: m.OriginalBranch,
        GitAvailable:   true,
    }, nil
}

// MergeBack merges the apply branch to original.
// Tries --ff-only first; falls back to --no-ff merge commit.
// Deletes apply branch on success.
func (m *Manager) MergeBack(r *Result, changeName string) error {
    if !r.GitAvailable {
        return nil
    }
    if out, err := exec.Command("git", "-C", m.RepoDir,
        "checkout", r.OriginalBranch).CombinedOutput(); err != nil {
        return fmt.Errorf("checkout %s: %w — %s", r.OriginalBranch, err, out)
    }

    // Fast-forward attempt
    if exec.Command("git", "-C", m.RepoDir,
        "merge", "--ff-only", r.BranchName).Run() == nil {
        return m.deleteBranch(r.BranchName)
    }

    // Merge commit fallback
    msg := fmt.Sprintf("feat(%s): sdd-apply complete — merged from %s", changeName, r.BranchName)
    if out, err := exec.Command("git", "-C", m.RepoDir,
        "merge", "--no-ff", r.BranchName, "-m", msg).CombinedOutput(); err != nil {
        return fmt.Errorf("merge %s: %w — %s", r.BranchName, err, out)
    }

    return m.deleteBranch(r.BranchName)
}

// Abandon returns to original branch WITHOUT merging (apply failed).
// The apply branch is preserved for inspection.
func (m *Manager) Abandon(r *Result) error {
    if !r.GitAvailable {
        return nil
    }
    out, err := exec.Command("git", "-C", m.RepoDir,
        "checkout", r.OriginalBranch).CombinedOutput()
    if err != nil {
        return fmt.Errorf("abandon checkout: %w — %s", err, out)
    }
    return nil // branch intentionally preserved
}

func (m *Manager) verifyClean() error {
    for _, extra := range []string{"", "--cached"} {
        args := []string{"-C", m.RepoDir, "diff", "--quiet"}
        if extra != "" {
            args = append(args, extra)
        }
        if exec.Command("git", args...).Run() != nil {
            return fmt.Errorf("uncommitted changes present")
        }
    }
    return nil
}

func (m *Manager) deleteBranch(name string) error {
    out, err := exec.Command("git", "-C", m.RepoDir, "branch", "-d", name).CombinedOutput()
    if err != nil {
        return fmt.Errorf("delete %s: %w — %s", name, err, out)
    }
    return nil
}

func gitAvailable() bool { return exec.Command("git", "--version").Run() == nil }

func sanitize(s string) string {
    return strings.ToLower(
        strings.NewReplacer("/", "-", " ", "-", ".", "-", "_", "-").Replace(s))
}
