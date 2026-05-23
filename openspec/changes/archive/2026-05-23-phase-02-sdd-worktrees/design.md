# Design: FASE 2 — SDD v3: Phase DAG Enforced + Result Contract + Circuit Breaker + Apply Continuity

> **Source of Truth:** `/home/rdmachadog/Documents/fix_achitect-ai/02-phase-sdd.md`

## Architecture & Code Implementations

## 2.1 `.atl/sdd-state.yaml` — Fuente de Verdad del Phase DAG

### Formato (generado por sdd-init, actualizado por cada fase)

```yaml
# .atl/sdd-state.yaml
# AUTO-MANAGED by architect-ai SDD phases — do not edit manually
# Each phase MUST read this file before executing. Phase DAG is ENFORCED here.
version: "3.0"
change_name: ""           # Set by sdd-init
project: ""               # Set by sdd-init
started_at: ""            # ISO8601
artifact_store: "hybrid"  # engram | openspec | hybrid | none
execution_mode: "interactive"   # interactive | automatic
delivery_strategy: "ask-on-risk"  # ask-on-risk | auto-chain | single-pr | exception-ok
tdd_mode: false           # true = RED→GREEN→TRIANGULATE→REFACTOR enforced

phases:
  sdd-init:
    status: "pending"     # pending | running | completed | failed | skipped
    completed_at: ""
    artifacts: []

  sdd-onboard:
    status: "pending"
    completed_at: ""
    artifacts: []

  sdd-explore:
    status: "pending"
    completed_at: ""
    artifacts: []
    requires: []           # No prerequisites

  sdd-propose:
    status: "pending"
    completed_at: ""
    artifacts: []
    requires: ["sdd-explore"]

  sdd-spec:
    status: "pending"
    completed_at: ""
    artifacts: []
    requires: ["sdd-propose"]

  sdd-design:
    status: "pending"
    completed_at: ""
    artifacts: []
    requires: ["sdd-spec"]

  sdd-tasks:
    status: "pending"
    completed_at: ""
    artifacts: []
    requires: ["sdd-design"]
    review_workload:
      estimated_lines: 0
      budget_risk: "low"
      chained_prs_recommended: false

  sdd-apply:
    status: "pending"
    completed_at: ""
    artifacts: []
    requires: ["sdd-tasks"]
    apply_branch: ""        # git branch used for isolation
    current_slice: 0        # for auto-chain: which slice is running
    total_slices: 1

  sdd-verify:
    status: "pending"
    completed_at: ""
    artifacts: []
    requires: ["sdd-apply"]

  sdd-archive:
    status: "pending"
    completed_at: ""
    artifacts: []
    requires: ["sdd-verify"]

circuit_breaker:
  enabled: true
  max_attempts: 3           # Per phase
  attempt_counts: {}        # { "sdd-apply": 2, "sdd-verify": 1 }
  abandoned_phases: []      # Phases that hit max_attempts
```

### Reglas de Escritura Atómica

TODOS los agentes que escriben `sdd-state.yaml` DEBEN seguir el patrón:
```bash
# Atomic write — NEVER write directly to sdd-state.yaml
TMPFILE=".atl/sdd-state.yaml.tmp"
LOCKFILE=".atl/sdd-state.yaml.lock"

# Acquire lock (fail if exists for > 30 seconds)
if [ -f "${LOCKFILE}" ]; then
  LOCK_AGE=$(( $(date +%s) - $(stat -c %Y "${LOCKFILE}" 2>/dev/null || echo 0) ))
  [ "${LOCK_AGE}" -gt 30 ] && rm -f "${LOCKFILE}"  # stale lock
fi
(set -C; echo $$ > "${LOCKFILE}") 2>/dev/null || { echo "STATE LOCKED — retry in 5s" >&2; exit 1; }

# Write to tmp
cat > "${TMPFILE}" << 'EOF'
{yaml content}
EOF

# Atomic rename
mv "${TMPFILE}" ".atl/sdd-state.yaml"
rm -f "${LOCKFILE}"
```

---

## 2.2 Phase DAG Enforcement Protocol

### `internal/assets/_shared/phase-dag-enforcement.md`

```markdown
## Phase DAG Enforcement Protocol [MANDATORY — First action of every SDD phase agent]

Every SDD phase agent MUST execute this check BEFORE any other action.

### Step 1: Read sdd-state.yaml
```bash
STATE_FILE=".atl/sdd-state.yaml"
if [ ! -f "${STATE_FILE}" ]; then
  echo "BLOCKED: sdd-state.yaml not found. Run /sdd-init first." >&2
  exit 1
fi
```

### Step 2: Parse prerequisites for THIS phase
```bash
# For sdd-apply specifically:
PHASE="sdd-apply"
REQUIRES=("sdd-tasks")  # from sdd-state.yaml phases.sdd-apply.requires

# Check each prerequisite
for req_phase in "${REQUIRES[@]}"; do
  REQ_STATUS=$(grep -A2 "  ${req_phase}:" "${STATE_FILE}" | grep "status:" | awk '{print $2}' | tr -d '"')
  if [ "${REQ_STATUS}" != "completed" ]; then
    echo "BLOCKED: Prerequisite '${req_phase}' is '${REQ_STATUS}', not 'completed'" >&2
    echo "Cannot execute ${PHASE} until ${req_phase} completes." >&2
    exit 1
  fi
done
```

### Step 3: Check circuit breaker
```bash
MAX_ATTEMPTS=3
CURRENT_ATTEMPTS=$(grep -A1 "  ${PHASE}:" "${STATE_FILE}" | grep "attempt_counts" || echo 0)
# Parse integer from YAML
ATTEMPTS=$(echo "${CURRENT_ATTEMPTS}" | grep -o '[0-9]*' | head -1)
ATTEMPTS=${ATTEMPTS:-0}

if [ "${ATTEMPTS}" -ge "${MAX_ATTEMPTS}" ]; then
  echo "CIRCUIT BREAKER: ${PHASE} has failed ${ATTEMPTS} times. STATUS: ABANDONED" >&2
  echo "Manual intervention required. Check: cat .atl/sdd-state.yaml" >&2
  exit 2  # Exit code 2 = Ralph Loop circuit breaker (not 0=success, not 1=error)
fi
```

### Step 4: Mark phase as "running"
Update sdd-state.yaml:
- phases.{phase}.status = "running"
- phases.{phase}.started_at = ISO8601 timestamp

### Step 5: Execute phase logic (your actual work)

### Step 6: On SUCCESS — Mark completed
- phases.{phase}.status = "completed"
- phases.{phase}.completed_at = ISO8601 timestamp
- phases.{phase}.artifacts = list of created artifact keys

### Step 6b: On FAILURE — Increment circuit breaker
- phases.{phase}.status = "failed"
- circuit_breaker.attempt_counts.{phase} += 1
- If attempt_counts.{phase} >= max_attempts: STATUS = ABANDONED
```

---

## 2.3 Result Contract — JSON Schema + Validator

### Formato estándar de Result Contract

**TODOS** los agentes SDD devuelven este JSON como ÚLTIMO output:

```json
{
  "status": "completed|failed|blocked|abandoned",
  "phase": "sdd-explore",
  "change_name": "add-user-auth",
  "executive_summary": "Explored codebase. Found auth service in /src/auth/. No existing token handling.",
  "artifacts": [
    "sdd/add-user-auth/explore",
    "sdd/add-user-auth/explore-risks"
  ],
  "next_recommended": "sdd-propose",
  "risks": [
    "No existing auth tests — TDD will require building from zero",
    "JWT library outdated — may need upgrade"
  ],
  "skill_resolution": {
    "status": "paths-injected|fallback-registry|fallback-path|none",
    "skills_used": ["ripgrep", "bash-expert", "architecture-guardrails"],
    "fallback_reason": null
  },
  "review_workload": null,
  "attempt_number": 1,
  "blocked_reason": null
}
```

### Result Contract Validator — Shell Script

```bash
#!/usr/bin/env bash
# .atl/scripts/validate-result-contract.sh
# Usage: echo "$RESULT_JSON" | bash .atl/scripts/validate-result-contract.sh
# Exit 0 = valid, Exit 1 = invalid (orchestrator must handle)

set -euo pipefail

RESULT=$(cat)

# Required fields
REQUIRED_FIELDS=("status" "phase" "executive_summary" "artifacts" "next_recommended" "risks" "skill_resolution")

for field in "${REQUIRED_FIELDS[@]}"; do
  VALUE=$(echo "${RESULT}" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('${field}','__MISSING__'))" 2>/dev/null)
  if [ "${VALUE}" = "__MISSING__" ]; then
    echo "RESULT CONTRACT VIOLATION: missing field '${field}'" >&2
    exit 1
  fi
done

# Status must be valid enum
STATUS=$(echo "${RESULT}" | python3 -c "import sys,json; print(json.load(sys.stdin).get('status',''))")
case "${STATUS}" in
  completed|failed|blocked|abandoned) ;;
  *)
    echo "RESULT CONTRACT VIOLATION: invalid status '${STATUS}'" >&2
    echo "Valid: completed|failed|blocked|abandoned" >&2
    exit 1
    ;;
esac

# skill_resolution must have status field
SKILL_STATUS=$(echo "${RESULT}" | python3 -c "
import sys,json
d=json.load(sys.stdin)
sr=d.get('skill_resolution',{})
print(sr.get('status','__MISSING__') if isinstance(sr,dict) else '__MISSING__')
")
if [ "${SKILL_STATUS}" = "__MISSING__" ]; then
  echo "RESULT CONTRACT VIOLATION: skill_resolution.status missing" >&2
  exit 1
fi

echo "RESULT CONTRACT: VALID"
exit 0
```

### SDD Orchestrator — Result Contract Check

En el prompt de `sdd-orchestrator.md`, añadir:

```markdown
## Result Contract Validation [MANDATORY — after every phase completion]

When a phase agent reports completion, BEFORE proceeding to next phase:

1. Capture the last JSON output from the phase agent
2. Run validation:
   ```bash
   echo "${PHASE_OUTPUT}" | bash .atl/scripts/validate-result-contract.sh
   ```
3. If validation fails:
   - Emit LITE: "[sdd-orchestrator] Result Contract violation from {phase}. Requesting retry."
   - Increment circuit_breaker.attempt_counts.{phase} in sdd-state.yaml
   - If attempts >= 3: STATUS = ABANDONED, stop

4. If skill_resolution.status = "fallback-registry" or "none":
   - Emit LITE: "WARN: {phase} ran without correct skill injection. Re-reading skill registry."
   - Re-run skill resolution for that phase
   - Retry the phase once

5. If validation passes: update sdd-state.yaml, proceed to next phase
```

---

## 2.4 Apply Continuity — `.atl/apply-progress.yaml`

### Formato

```yaml
# .atl/apply-progress.yaml
# Checkpoint file for sdd-apply. Enables resume after failure.
version: "2.0"
change_name: ""
apply_branch: ""
started_at: ""
last_updated: ""
artifact_store: "hybrid"
delivery_strategy: "ask-on-risk"
current_slice: 1
total_slices: 1

tasks:
  - id: "T1"
    description: "Add token model with validation"
    status: "completed"    # pending | running | completed | failed
    completed_at: "2026-05-17T10:30:00Z"
    files_modified: ["src/auth/token.go", "src/auth/token_test.go"]
    commit_hash: "abc123"
    lines_added: 45
    lines_deleted: 0

  - id: "T2"
    description: "Wire token validation into login endpoint"
    status: "running"
    files_modified: []
    commit_hash: ""
    lines_added: 0
    lines_deleted: 0

  - id: "T3"
    description: "Add integration tests for full auth flow"
    status: "pending"
    files_modified: []
    commit_hash: ""
    lines_added: 0
    lines_deleted: 0

totals:
  completed: 1
  running: 1
  pending: 1
  failed: 0
  lines_added: 45
  lines_deleted: 0
```

### Apply Continuity Protocol

```markdown
## Apply Continuity Protocol [sdd-apply MANDATORY — start of EVERY apply session]

### Step 1: Check for existing apply-progress.yaml
```bash
PROGRESS_FILE=".atl/apply-progress.yaml"
if [ -f "${PROGRESS_FILE}" ]; then
  echo "Resuming from previous apply session..."
  CHANGE=$(grep "^change_name:" "${PROGRESS_FILE}" | awk '{print $2}' | tr -d '"')
  echo "Change: ${CHANGE}"
  echo "Progress: $(grep -c 'status: "completed"' "${PROGRESS_FILE}") tasks completed"
fi
```

### Step 2: Load task list
If apply-progress.yaml exists AND change_name matches current sdd-state.yaml:
→ Load task list from apply-progress.yaml
→ SKIP tasks with status: "completed" (they're DONE — do not re-execute)
→ Resume from first task with status: "pending" or "running"

If apply-progress.yaml does NOT exist:
→ Read tasks from Engram: mem_get_observation("sdd/{change_name}/tasks")
→ Initialize apply-progress.yaml with all tasks as "pending"
→ Create apply branch (from Phase 02 Branch Protocol)

If change_name mismatch:
→ WARN: "apply-progress.yaml is for a different change. Archiving old progress."
→ Rename to apply-progress.yaml.{timestamp}.bak
→ Start fresh for current change

### Step 3: Update task status atomically
For each task transition (pending→running, running→completed, etc.):
→ Update ONLY the specific task entry (YAML field level, not full file rewrite)
→ Update totals section
→ Update last_updated timestamp

### Step 4: Report progress to user (LITE)
After each task completes:
"[apply] T{N}/{M} complete — {description} | {lines_added} lines"

### Step 5: On full slice completion
If delivery_strategy = "auto-chain":
→ Verify: git diff --stat | sum lines ≤ 400
→ Commit with Work Unit Commits protocol
→ Emit: "[apply] Slice {N}/{M} ready. Check .atl/apply-progress.yaml for details."
→ Pause (interactive) or continue (automatic)
```

---

## 2.5 SDD Init Guard + Artifact Store Mode

### SDD Init Guard (L0 responsibility, Phase 01 reference)

```markdown
## SDD Init Guard [L0 architect — check before ANY SDD command]

Before routing ANY SDD intent to sdd-orchestrator:

```bash
STATE_FILE=".atl/sdd-state.yaml"
if [ ! -f "${STATE_FILE}" ]; then
  # sdd-init has not run for this project
  emit LITE: "[L0] SDD context not initialized. Running sdd-init first."
  # Route to sdd-init FIRST, then resume original SDD intent
  PENDING_INTENT="${original_user_message}"
  delegate to sdd-init
  # After sdd-init completes, continue with original intent
else
  INIT_STATUS=$(grep "sdd-init:" -A2 "${STATE_FILE}" | grep "status:" | awk '{print $2}')
  if [ "${INIT_STATUS}" != '"completed"' ]; then
    emit LITE: "[L0] SDD init in progress or failed. Check .atl/sdd-state.yaml"
  fi
fi
```

### Artifact Store Mode Selection [sdd-init responsibility]

During sdd-init, ask ONCE:

```
Artifact store mode?
  [e] engram (default) — store in local SQLite database
  [o] openspec — store as Markdown files in openspec/ directory
  [h] hybrid — store in both (recommended for team projects with git)
  [n] none — volatile session memory only (no persistence)
```

Write answer to `sdd-state.yaml`:
```yaml
artifact_store: "hybrid"  # cached for all phases
```

Pass `artifact_store` in every delegation. Each phase uses it to determine WHERE to save.
```

---

## 2.6 Circuit Breaker — Ralph Loop Prevention

### Fuente: "The Ralph Loop: Long-Running AI Agents" + "Zero-Step Thinking" (Early Exits)

```markdown
## Circuit Breaker Protocol [ALL SDD phase agents + sdd-orchestrator]

### What is a Ralph Loop?
An agent that fails repeatedly, retrying the same broken approach N times, wasting all API quota.
Exit Code 2 = "I've tried enough. Stop me."

### Circuit Breaker Rules

**Per-phase attempt limit**: 3 attempts maximum.

After attempt 1 (failed):
- Emit LITE: "[{phase}] Attempt 1 failed: {reason}. Trying alternative approach."
- Change approach: if approach A failed → try approach B
- Increment circuit_breaker.attempt_counts.{phase} in sdd-state.yaml

After attempt 2 (failed):
- Emit LITE: "[{phase}] Attempt 2 failed: {reason}. Escalating."
- Request additional context from orchestrator
- Increment counter

After attempt 3 (failed):
- DO NOT retry
- Emit LITE: "[{phase}] CIRCUIT BREAKER: 3 attempts failed. STATUS: ABANDONED"
- Write to sdd-state.yaml:
  ```yaml
  circuit_breaker:
    abandoned_phases: ["sdd-apply"]
  ```
- Return Result Contract:
  ```json
  {
    "status": "abandoned",
    "phase": "{phase}",
    "executive_summary": "3 attempts failed. Manual intervention required.",
    "risks": ["See .atl/apply-progress.yaml for partial state"],
    "skill_resolution": {"status": "paths-injected"},
    "attempt_number": 3,
    "blocked_reason": "Circuit breaker triggered after 3 failed attempts. Root cause: {last_error}"
  }
  ```
- Exit with code 2 (not 0 or 1)

### Orchestrator Response to Circuit Breaker (exit code 2)

When sdd-orchestrator receives STATUS: ABANDONED:
1. Emit LITE: "[sdd-orchestrator] {phase} abandoned. Pausing SDD cycle."
2. Save current state: mem_session_summary
3. Emit diagnostic to user:
   ```
   SDD BLOCKED: {phase} could not complete after 3 attempts.
   Root cause: {blocked_reason}
   Preserved: {list of completed phases so far}
   Next action options:
     1. Fix the blocking issue and run /sdd-continue
     2. Skip {phase} with risk: /sdd-ff {next_phase} --skip-risk
     3. Abandon change: /sdd-archive --status=abandoned
   ```
4. STOP — do not continue SDD cycle
```

---

## 2.7 git Branch Isolation (Simplified from worktree to temp branch)

Referencia a Phase 02 v2 existente — confirmar protocolo:

```markdown
## Apply Branch Protocol [sdd-apply MANDATORY]

Before ANY file modification:

```bash
ORIGINAL=$(git rev-parse --abbrev-ref HEAD 2>/dev/null) || { echo "WARN: not a git repo"; exit 0; }
CHANGE=$(grep "change_name:" .atl/sdd-state.yaml | awk '{print $2}' | tr -d '"')
APPLY_BRANCH="apply/$(echo "${CHANGE}" | tr '[:upper:] ./_' '[:lower:]----')"

# Verify clean state
git diff --quiet && git diff --cached --quiet || {
  echo "ERROR: Uncommitted changes. Commit or stash first." >&2; exit 1
}

# Create apply branch
git branch -D "${APPLY_BRANCH}" 2>/dev/null || true
git checkout -b "${APPLY_BRANCH}"

# Update sdd-state.yaml
# phases.sdd-apply.apply_branch = "${APPLY_BRANCH}"
echo "ISOLATED on ${APPLY_BRANCH}"
```

On success: fast-forward merge → delete apply branch.
On failure: preserve apply branch, return to original. STATUS: FAILED.
```

---

## 2.8 Semantic Audit Protocol (Odoo-aware, multi-language)

### Updated patterns table for rg in sdd-verify

```markdown
## Semantic Audit Patterns by Language [Step 0 in sdd-verify]

### Go
```bash
rg -l "func {function_name}" --type go        # existence
rg "func {function_name}\(" --type go -A 3    # signature
rg "_{field}\" " --type go                     # struct field (JSON tag)
```

### Python (Odoo + generic)
```bash
rg -l "def {function_name}" --type py          # existence
rg "'{field}' = fields\." --type py            # Odoo field
rg "class {ClassName}" --type py               # class existence
rg "route.*\"{path}\"" --type py               # Odoo controller route
```

### JavaScript / OWL (Odoo frontend)
```bash
rg -l "{FunctionName}" --type js               # JS function
rg "\"name\": \"{model}\"" --type xml          # OWL template
rg "<t t-name=\"{component_name}\"" --type xml # OWL component
```

### XML (Odoo views)
```bash
rg -l "id=\"{view_id}\"" --type xml            # view existence
rg "model=\"{model_name}\"" --type xml         # model reference
rg "<list\b" --type xml || rg "<tree\b"        # check for deprecated <tree
```

### Negative assertions (MUST NOT exist)
```bash
rg -l "{forbidden_pattern}" . && echo "VIOLATION:{forbidden_pattern}" || echo "CLEAN"
# Example: check no <tree> in Odoo v18+ views
rg -l "<tree " --type xml && echo "VIOLATION:deprecated-tree-tag" || echo "CLEAN"
```

### No-TBD Linter (mandatory in sdd-verify)
```bash
rg -i "\bTBD\b|\bto be determined\b|\bfixme\b" openspec/changes/{change_name}/ \
  && echo "VIOLATION: TBD/FIXME found in SDD artifacts" \
  && exit 1 \
  || echo "CLEAN: no TBD/FIXME"
```
```

---

## 2.9 Go Installer Changes (minimal — installer only)

```go
// internal/sdd/state/writer.go
// Generates the initial sdd-state.yaml during `architect-ai sdd-init`
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
  sdd-tasks:    { status: "pending", completed_at: "", artifacts: [], requires: ["sdd-design"], review_workload: { estimated_lines: 0, budget_risk: "low", chained_prs_recommended: false } }
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

// WriteSddState writes sdd-state.yaml with atomic pattern
func WriteSddState(atDir, content string) error {
    stateFile := filepath.Join(atDir, "sdd-state.yaml")
    tmpFile := stateFile + ".tmp"
    lockFile := stateFile + ".lock"

    // Check for stale lock
    if info, err := os.Stat(lockFile); err == nil {
        if time.Since(info.ModTime()) > 30*time.Second {
            os.Remove(lockFile)
        } else {
            return fmt.Errorf("state file is locked — another process is writing")
        }
    }

    // Create lock
    if err := os.WriteFile(lockFile, []byte(fmt.Sprintf("%d", os.Getpid())), 0644); err != nil {
        return fmt.Errorf("acquire lock: %w", err)
    }
    defer os.Remove(lockFile)

    // Write to tmp
    if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
        return fmt.Errorf("write tmp: %w", err)
    }

    // Atomic rename
    return os.Rename(tmpFile, stateFile)
}

// ValidateStateYAML does basic structural validation
func ValidateStateYAML(atDir string) []string {
    stateFile := filepath.Join(atDir, "sdd-state.yaml")
    data, err := os.ReadFile(stateFile)
    if err != nil {
        return []string{fmt.Sprintf("sdd-state.yaml not found: %v", err)}
    }

    content := string(data)
    var issues []string

    requiredFields := []string{
        "version:", "change_name:", "project:", "artifact_store:",
        "execution_mode:", "delivery_strategy:", "circuit_breaker:",
    }
    for _, field := range requiredFields {
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

    required := []string{
        "change_name:", "project:", "artifact_store:", "execution_mode:",
        "delivery_strategy:", "circuit_breaker:", "sdd-init:", "sdd-apply:",
        "requires:", "tdd_mode:",
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

    // Verify file exists
    stateFile := filepath.Join(dir, "sdd-state.yaml")
    if _, err := os.Stat(stateFile); os.IsNotExist(err) {
        t.Error("sdd-state.yaml not created")
    }

    // Verify tmp file cleaned up
    if _, err := os.Stat(stateFile + ".tmp"); !os.IsNotExist(err) {
        t.Error("tmp file should be removed after atomic write")
    }

    // Verify lock file cleaned up
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

func TestValidateStateYAML_ValidFile(t *testing.T) {
    dir := t.TempDir()
    content := InitialState("test", "proj", "hybrid", "interactive", "ask-on-risk")
    WriteSddState(dir, content)

    issues := ValidateStateYAML(dir)
    if len(issues) > 0 {
        t.Errorf("valid state.yaml should have no issues, got: %v", issues)
    }
}
```

---

## Criterios de Verificación

### Test 1: Phase DAG Enforcement
```
Setup: .atl/sdd-state.yaml with sdd-design status: "pending"
Input: sdd-apply agent starts
Expected: Phase DAG check fires: "BLOCKED: sdd-design is 'pending', not 'completed'"
Expected: Exit code 1 (not 2 = circuit breaker, not 0 = success)
PASS if: sdd-apply does NOT execute any file modifications
```

### Test 2: Circuit Breaker Activation
```
Setup: sdd-apply attempts = 2 in sdd-state.yaml
Input: sdd-apply fails a 3rd time
Expected: Circuit breaker fires: "CIRCUIT BREAKER: 3 attempts failed. STATUS: ABANDONED"
Expected: Exit code 2
Expected: sdd-state.yaml updated: abandoned_phases includes "sdd-apply"
Expected: Result Contract returned with status: "abandoned"
PASS if: sdd-orchestrator receives ABANDONED and pauses SDD cycle
```

### Test 3: Apply Continuity Resume
```
Setup: apply-progress.yaml with T1 status: "completed", T2 status: "pending"
Input: sdd-apply starts new session
Expected: Reads apply-progress.yaml
Expected: T1 SKIPPED (not re-executed)
Expected: Starts from T2
PASS if: git log shows no new commits for T1's already-done work
```

### Test 4: Result Contract Validation
```
Input: Phase agent returns JSON without "skill_resolution" field
Expected: validate-result-contract.sh exits 1
Expected: sdd-orchestrator logs "Result Contract violation"
Expected: Phase retried (attempt count incremented)
PASS if: Orchestrator does NOT proceed to next phase with invalid contract
```

### Test 5: SDD Init Guard
```
Setup: .atl/sdd-state.yaml does NOT exist
Input: User sends "/sdd-explore my feature"
Expected: L0 detects missing state file
Expected: Routes to sdd-init FIRST
Expected: After sdd-init: resumes with sdd-explore
PASS if: No sdd-explore executes without sdd-init completion
```

### Test 6: No-TBD Linter
```
Setup: openspec/changes/auth-feature/spec.md contains "## Auth Flow\nTBD: decide token format"
Input: sdd-verify runs
Expected: No-TBD linter fires: "VIOLATION: TBD found in SDD artifacts"
Expected: sdd-verify returns status: "failed"
PASS if: sdd-verify does NOT pass with TBD content in spec artifacts
```

---

## Resultados Esperados

| Métrica | Antes (v2) | Después (v3) |
|---|---|---|
| Phase DAG enforcement | Prompt-only (bypassable) | ✅ YAML check (sdd-state.yaml) |
| Circuit Breaker | ❌ No existía | ✅ 3 intentos → exit code 2 → ABANDONED |
| Result Contract | Implícito en prompts | ✅ JSON schema + shell validator |
| Apply Continuity | Parcial (Engram hint) | ✅ apply-progress.yaml con task-level tracking |
| Artifact Store Mode | Asumido por defecto | ✅ Preguntado 1x en sdd-init, cacheado en YAML |
| TBD/TODO enforcement | En sdd-verify (parcial) | ✅ Shell linter en sdd-verify Step 0 |
| state.yaml concurrencia | No protegida | ✅ Lock file + atomic write |
| SDD Init Guard | ❌ No existía | ✅ L0 verifica sdd-state.yaml antes de routing |
