# Design: Phase 2 - SDD v3: Phase DAG + Circuit Breaker + Result Contract + Apply Continuity


## Architecture
SDD v3 introduces mechanical enforcement of the Phase DAG via `.atl/sdd-state.yaml`. Previously, phase ordering was documented in prompts only — bypassable under context pressure. v3 makes it structural: each phase agent reads the YAML state file, checks prerequisites are `completed`, and exits with STATUS: BLOCKED if not.

v2 features retained: Temp Branch + Fast-Forward Merge isolation for `sdd-apply`, Odoo auto-detection via `__manifest__.py`, and Semantic Audit Step 0 in `sdd-verify`.

New v3 components:
- **`.atl/sdd-state.yaml`**: Single source of truth for SDD phase lifecycle. Managed by `internal/sdd/state/writer.go`.
- **Result Contract JSON**: Validated envelope returned by every phase agent. Schema enforced by `.atl/scripts/validate-result-contract.sh`.
- **Circuit Breaker**: Max 3 attempts per phase tracked in `sdd-state.yaml.circuit_breaker.attempt_counts`. Exit code 2 = ABANDONED.
- **`.atl/apply-progress.yaml`**: Task-level checkpoint for `sdd-apply`. Resume from last completed task without full restart.

## FODA Matrix

| | Detail |
|---|---|
| **F** | Temp branch = universal git isolation, zero external deps. sdd-state.yaml makes DAG enforcement mechanical, not prompt-based. |
| **O** | FF merge keeps linear history. Branch preserved on failure = natural audit trail. Circuit Breaker saves API quota. Apply Continuity saves full restarts. |
| **D** | Without remote, apply branch has no off-machine backup. Mitigate: `engram sync` before sdd-apply. |
| **A** | If original branch diverges during apply (parallel dev), FF fails — auto-fallback to no-FF merge commit. If sdd-state.yaml is corrupted, all phase DAG enforcement fails — atomic write + lock prevents this. |

## Go Implementation

### Branch Manager (`internal/sdd/branch/manager.go`)
```go
package branch

import (
    "fmt"
    "os/exec"
    "strings"
)

type Manager struct {
    RepoDir        string
    OriginalBranch string
}

type Result struct {
    BranchName     string
    OriginalBranch string
    GitAvailable   bool
}

func New(repoDir string) (*Manager, error) {
    if !gitAvailable() {
        return &Manager{RepoDir: repoDir}, nil
    }
    out, err := exec.Command("git", "-C", repoDir, "rev-parse", "--abbrev-ref", "HEAD").Output()
    if err != nil {
        return &Manager{RepoDir: repoDir}, nil
    }
    return &Manager{RepoDir: repoDir, OriginalBranch: strings.TrimSpace(string(out))}, nil
}

func (m *Manager) CreateApplyBranch(changeName string) (*Result, error) {
    if !gitAvailable() || m.OriginalBranch == "" {
        return &Result{GitAvailable: false}, nil
    }
    if err := m.verifyClean(); err != nil {
        return nil, fmt.Errorf("repo dirty: %w", err)
    }
    branchName := "apply/" + sanitize(changeName)
    exec.Command("git", "-C", m.RepoDir, "branch", "-D", branchName).Run() //nolint:errcheck
    if out, err := exec.Command("git", "-C", m.RepoDir, "checkout", "-b", branchName).CombinedOutput(); err != nil {
        return nil, fmt.Errorf("create branch %s: %w — %s", branchName, err, out)
    }
    return &Result{BranchName: branchName, OriginalBranch: m.OriginalBranch, GitAvailable: true}, nil
}

func (m *Manager) MergeBack(r *Result, changeName string) error {
    if !r.GitAvailable { return nil }
    if out, err := exec.Command("git", "-C", m.RepoDir, "checkout", r.OriginalBranch).CombinedOutput(); err != nil {
        return fmt.Errorf("checkout %s: %w — %s", r.OriginalBranch, err, out)
    }
    if exec.Command("git", "-C", m.RepoDir, "merge", "--ff-only", r.BranchName).Run() == nil {
        return m.deleteBranch(r.BranchName)
    }
    msg := fmt.Sprintf("feat(%s): sdd-apply complete — merged from %s", changeName, r.BranchName)
    if out, err := exec.Command("git", "-C", m.RepoDir, "merge", "--no-ff", r.BranchName, "-m", msg).CombinedOutput(); err != nil {
        return fmt.Errorf("merge %s: %w — %s", r.BranchName, err, out)
    }
    return m.deleteBranch(r.BranchName)
}

func (m *Manager) Abandon(r *Result) error {
    if !r.GitAvailable { return nil }
    out, err := exec.Command("git", "-C", m.RepoDir, "checkout", r.OriginalBranch).CombinedOutput()
    if err != nil {
        return fmt.Errorf("abandon checkout: %w — %s", err, out)
    }
    return nil
}

func (m *Manager) verifyClean() error {
    for _, extra := range []string{"", "--cached"} {
        args := []string{"-C", m.RepoDir, "diff", "--quiet"}
        if extra != "" { args = append(args, extra) }
        if exec.Command("git", args...).Run() != nil {
            return fmt.Errorf("uncommitted changes present")
        }
    }
    return nil
}

func (m *Manager) deleteBranch(name string) error {
    out, err := exec.Command("git", "-C", m.RepoDir, "branch", "-d", name).CombinedOutput()
    if err != nil { return fmt.Errorf("delete %s: %w — %s", name, err, out) }
    return nil
}

func gitAvailable() bool { return exec.Command("git", "--version").Run() == nil }

func sanitize(s string) string {
    return strings.ToLower(strings.NewReplacer("/", "-", " ", "-", ".", "-", "_", "-").Replace(s))
}
```

### Odoo Detector (`internal/project/odoo/detector.go`)
```go
package odoo

import (
    "bufio"
    "os"
    "path/filepath"
    "regexp"
    "strings"
)

type Info struct {
    IsOdoo          bool
    Version         string
    ManifestPath    string
    OverlayInstalled bool
    AvailableAgents []string
    AvailableSkills []string
}

var versionRe = regexp.MustCompile(`["']version["']\s*:\s*["'](\d+)\.`)

func Detect(projectDir string) (*Info, error) {
    info := &Info{}
    if mp := findFile(projectDir, "__manifest__.py"); mp != "" {
        info.IsOdoo = true
        info.ManifestPath = mp
        info.Version = extractVersion(mp)
    }
    if !info.IsOdoo {
        for _, f := range []string{"requirements.txt", "pyproject.toml"} {
            if containsLine(filepath.Join(projectDir, f), "odoo") {
                info.IsOdoo = true
                break
            }
        }
    }
    if !info.IsOdoo { return info, nil }
    overlayDir := filepath.Join(projectDir, ".atl", "overlays", "odoo-development-skill")
    if _, err := os.Stat(overlayDir); err == nil {
        info.OverlayInstalled = true
        info.AvailableAgents = listMDs(filepath.Join(overlayDir, "agents"))
    }
    return info, nil
}

func findFile(root, name string) string {
    var found string
    filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error { //nolint:errcheck
        if err != nil { return nil }
        skip := map[string]bool{".git": true, "node_modules": true, "__pycache__": true}
        if d.IsDir() && skip[d.Name()] { return filepath.SkipDir }
        if !d.IsDir() && d.Name() == name { found = p; return filepath.SkipAll }
        return nil
    })
    return found
}

func extractVersion(path string) string {
    f, _ := os.Open(path)
    if f == nil { return "unknown" }
    defer f.Close()
    s := bufio.NewScanner(f)
    for s.Scan() {
        if m := versionRe.FindStringSubmatch(s.Text()); len(m) > 1 { return m[1] }
    }
    return "unknown"
}

func containsLine(path, prefix string) bool {
    f, _ := os.Open(path)
    if f == nil { return false }
    defer f.Close()
    s := bufio.NewScanner(f)
    for s.Scan() {
        if strings.HasPrefix(strings.ToLower(strings.TrimSpace(s.Text())), prefix) { return true }
    }
    return false
}

func listMDs(dir string) []string {
    entries, _ := os.ReadDir(dir)
    var names []string
    for _, e := range entries {
        if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
            names = append(names, strings.TrimSuffix(e.Name(), ".md"))
        }
    }
    return names
}
```

### Detector Tests (`internal/project/odoo/detector_test.go`)
```go
package odoo

import (
    "os"
    "path/filepath"
    "testing"
)

func TestDetect_EmptyDir_NotOdoo(t *testing.T) {
    info, _ := Detect(t.TempDir())
    if info.IsOdoo { t.Error("empty dir must not be Odoo") }
}

func TestDetect_Manifest_V18(t *testing.T) {
    dir := t.TempDir()
    mod := filepath.Join(dir, "sale_custom")
    os.MkdirAll(mod, 0755)
    os.WriteFile(filepath.Join(mod, "__manifest__.py"), []byte(`{'name':'Sale Custom','version':'18.0.1.0.0','depends':['sale']}`), 0644)

    info, err := Detect(dir)
    if err != nil { t.Fatal(err) }
    if !info.IsOdoo || info.Version != "18" {
        t.Errorf("expected Odoo v18, got is_odoo=%v ver=%s", info.IsOdoo, info.Version)
    }
}

func TestDetect_Requirements(t *testing.T) {
    dir := t.TempDir()
    os.WriteFile(filepath.Join(dir, "requirements.txt"), []byte("odoo>=18.0\nrequests\n"), 0644)
    info, _ := Detect(dir)
    if !info.IsOdoo { t.Error("must detect Odoo from requirements.txt") }
}

func TestExtractVersion_Variants(t *testing.T) {
    dir := t.TempDir()
    path := filepath.Join(dir, "__manifest__.py")
    cases := []struct{ content, want string }{
        {`{'version': '19.0.1.0.0'}`, "19"},
        {`{"version": "17.0.2.1.0"}`, "17"},
        {`{'name': 'test'}`, "unknown"},
    }
    for _, tc := range cases {
        os.WriteFile(path, []byte(tc.content), 0644)
        got := extractVersion(path)
        if got != tc.want { t.Errorf("extractVersion content=%q got=%q want=%q", tc.content, got, tc.want) }
    }
}
```

### SDD State Writer (`internal/sdd/state/writer.go`) — v3

```go
package state

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "time"
)

// InitialState generates the starter sdd-state.yaml content
func InitialState(changeName, project, artifactStore, executionMode, deliveryStrategy string) string {
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
        changeName, project,
        time.Now().UTC().Format(time.RFC3339),
        artifactStore, executionMode, deliveryStrategy,
        time.Now().UTC().Format(time.RFC3339),
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
```

```go
// internal/sdd/state/writer_test.go
package state

import (
    "os"
    "path/filepath"
    "strings"
    "testing"
)

func TestInitialState_ContainsRequiredFields(t *testing.T) {
    content := InitialState("auth-feature", "myproject", "hybrid", "interactive", "ask-on-risk")
    required := []string{"change_name:", "project:", "artifact_store:", "execution_mode:", "delivery_strategy:", "circuit_breaker:", "sdd-init:", "sdd-apply:", "requires:", "tdd_mode:"}
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
```

## Key Decisions
- **sdd-state.yaml as enforcement mechanism**: Prompts alone cannot prevent phase bypassing under context pressure. Mechanical YAML state readable by all agents closes this gap without requiring a running server.
- **Atomic state writes**: Concurrent phase agents (or agent crash mid-write) cannot corrupt the state file. Lock + tmp + rename pattern is universally supported.
- **Circuit Breaker exit code 2**: Distinguishes `ABANDONED` from `FAILED` (1) and `SUCCESS` (0). Orchestrator routes differently for each.
- **apply-progress.yaml at task level**: Full `sdd-apply` restarts from T0 waste all prior work. Task-level checkpointing is the minimum viable resume unit.
