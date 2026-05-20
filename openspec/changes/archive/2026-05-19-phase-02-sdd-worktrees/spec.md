# Spec: Phase 2 - SDD v3: Phase DAG + Circuit Breaker + Result Contract + Apply Continuity


## Requirements

### 1. Apply Branch Protocol [sdd-apply MANDATORY]
Git isolation for all file modifications during `sdd-apply`. No remote required. No PR required. Works in any git repo.
- **Pre-flight:**
  ```bash
  set -euo pipefail
  ORIGINAL_BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null) || {
    echo "WARN: not a git repo — skipping branch isolation"
    echo "GIT_AVAILABLE=false"
    exit 0
  }
  CHANGE_NAME="${1:?change_name required}"
  APPLY_BRANCH="apply/$(echo "${CHANGE_NAME}" | tr '[:upper:] ./_' '[:lower:]----')"
  echo "original=${ORIGINAL_BRANCH} apply=${APPLY_BRANCH}"
  ```
- **Setup:**
  ```bash
  # Verify clean state
  git diff --quiet && git diff --cached --quiet || {
    echo "ERROR: Uncommitted changes. Commit or stash first." >&2; exit 1
  }
  # Clean stale apply branch if exists
  git branch -D "${APPLY_BRANCH}" 2>/dev/null || true
  # Create and switch
  git checkout -b "${APPLY_BRANCH}"
  echo "ISOLATED on ${APPLY_BRANCH} — original branch is safe"
  ```
- **During Apply:** Every task = ≥1 commit. Format: `type(scope): description`. Tests run in this branch.
- **On Success:** Try `--ff-only` first, fallback to `--no-ff`. Delete apply branch.
  ```bash
  git checkout "${ORIGINAL_BRANCH}"
  if git merge --ff-only "${APPLY_BRANCH}" 2>/dev/null; then
    echo "MERGED via fast-forward — clean history"
  else
    git merge --no-ff "${APPLY_BRANCH}" -m "feat(${CHANGE_NAME}): sdd-apply complete — merged from ${APPLY_BRANCH}"
    echo "MERGED via merge commit"
  fi
  git branch -d "${APPLY_BRANCH}"
  ```
- **On Failure:** Branch preserved.
  ```bash
  git checkout "${ORIGINAL_BRANCH}"
  echo "BLOCKED — ${APPLY_BRANCH} preserved for inspection"
  ```

### 2. Odoo Project Detection & Routing
- Detection: `__manifest__.py`, `odoo` in `requirements.txt`/`pyproject.toml`, or `odoo_version` in `.atl/config.yaml`.
- Version Extraction:
  ```bash
  ODOO_VERSION=$(python3 -c "
  import ast, re, sys
  txt = open('${MANIFEST}').read()
  try:
      d = ast.literal_eval(txt)
      v = d.get('version','')
      print(v.split('.')[0] if v else 'unknown')
  except:
      m = re.search(r'[\"\\']version[\"\\']\\s*:\\s*[\"\\'](\d+)\\.', txt)
      print(m.group(1) if m else 'unknown')
  " 2>/dev/null || echo "unknown")
  ```
- Save to Engram: `project/meta/odoo-detection`.
- Load Odoo skills and supplements for each SDD phase. Set Odoo Research Order. Enable Cross-Agent Odoo Calling Patterns.

### 3. Semantic Audit Protocol (sdd-verify Step 0)
Execute BEFORE test runner.
1. Retrieve `spec` and `design` contracts from Engram.
2. Parse assertions (MUST/SHALL, BDD, FMEA, signatures, endpoints).
3. Verify via `rg`.
4. Score and Gate:
   - `score = 0.0` → PASS
   - `0 < score ≤ 0.10` → WARN
   - `score > 0.10` → REJECT
5. Persist Result to `sdd/{change_name}/semantic-audit`.

### 4. FMEA + FSM Gate in sdd-design
MANDATORY EXIT CONDITIONS:
A. **FMEA Matrix:** Required for every component. RPN > 15 → explicit mitigation. RPN > 20 → +++Adversarial verify.
B. **FSM Diagram:** Required if stateful (mermaid `stateDiagram-v2`).
C. **Pre-Test Contract (5 items):** Interface, Input/Output, Error, Side-effect, Invariant.

### 5. sdd-hotfix (Micro-Ciclo)
- Trigger: `/sdd-hotfix {desc}`
- Gate: ≤ 3 files, no public API changes, no schema migrations, D5 < 2, urgency stated.
- Cycle: 1. explore-lite 2. propose-lite 3. apply-branch 4. verify-lite (no semantic audit) 5. archive-lite.

### 6. Phase DAG Enforcement — `.atl/sdd-state.yaml` [ALL sdd-* agents MANDATORY]

Every SDD phase agent MUST read `.atl/sdd-state.yaml` as its first action and verify all prerequisites are `completed` before executing.

- If file missing → emit `BLOCKED: sdd-state.yaml not found. Run /sdd-init first.` → exit 1.
- If prerequisite phase status ≠ `completed` → emit `BLOCKED: Prerequisite '{phase}' is '{status}', not 'completed'.` → exit 1.
- Phase transitions: `pending → running → completed | failed`. Writes are ATOMIC (tmp + rename + lock).
- Circuit breaker state tracked here: `circuit_breaker.attempt_counts.{phase}`.

Required fields in `sdd-state.yaml`: `version`, `change_name`, `project`, `artifact_store`, `execution_mode`, `delivery_strategy`, `circuit_breaker`.

Phase DAG (required dependency order):
```
sdd-explore → sdd-propose → sdd-spec → sdd-design → sdd-tasks → sdd-apply → sdd-verify → sdd-archive
```

### 7. Result Contract — JSON Envelope [ALL sdd-* agents MANDATORY]

Every phase agent MUST emit a validated JSON block as its LAST output:

```json
{
  "status": "completed|failed|blocked|abandoned",
  "phase": "sdd-explore",
  "change_name": "string",
  "executive_summary": "string",
  "artifacts": ["string"],
  "next_recommended": "string",
  "risks": ["string"],
  "skill_resolution": {
    "status": "paths-injected|fallback-registry|fallback-path|none",
    "skills_used": ["string"],
    "fallback_reason": null
  },
  "attempt_number": 1,
  "blocked_reason": null
}
```

Orchestrator validates via `.atl/scripts/validate-result-contract.sh`. Invalid contract → increment circuit breaker + retry. After 3 failures → STATUS: ABANDONED.

If `skill_resolution.status` = `fallback-registry` or `none` → re-run skill resolution + retry phase once.

### 8. Circuit Breaker — Ralph Loop Prevention [ALL sdd-* agents + sdd-orchestrator]

- Max 3 attempts per phase.
- After attempt 1 fail: change approach, increment `circuit_breaker.attempt_counts.{phase}`.
- After attempt 2 fail: escalate + request additional context.
- After attempt 3 fail: exit code 2, write STATUS: ABANDONED to `sdd-state.yaml`, return Result Contract with `status: abandoned`.
- sdd-orchestrator on exit code 2: save state, emit diagnostic, STOP (do not advance).
- User options after ABANDONED: fix + `/sdd-continue`, skip with risk `/sdd-ff`, or `/sdd-archive --status=abandoned`.

### 9. Apply Continuity — `.atl/apply-progress.yaml` [sdd-apply MANDATORY]

Task-level checkpoint file written by `sdd-apply` to enable resume after failure or interruption.

- If `apply-progress.yaml` exists for current `change_name` → load, skip `completed` tasks, resume from first `pending`.
- If missing → initialize from `sdd/{change_name}/tasks` with all tasks `pending`.
- If `change_name` mismatch → archive old file + start fresh.
- Each task transition is atomic (field-level update, not full rewrite).
- Report progress to user: `[apply] T{N}/{M} complete — {description} | {lines_added} lines`.

## Scenarios
### Test 1 — Branch Aislamiento
**Given** `sdd-apply`, `change_name="auth-feature"`.
**Then** `git branch` shows `apply/auth-feature` BEFORE file edit. All commits on apply, main unchanged.

### Test 2 — Fast-Forward vs Merge Commit
**Given** clean completion.
**Then** if main linear, `merge --ff-only`. If diverged, `merge --no-ff`. Apply branch deleted.

### Test 3 — Odoo Detection in sdd-init
**Given** `my_module/__manifest__.py` version `18.0.1.0.0`.
**When** `sdd-init`.
**Then** `IS_ODOO=true`, `ODOO_VERSION=18`, skills loaded, `init-odoo.md` injected.

### Test 4 — L3 Odoo Agent Cross-Call
**Given** Odoo v18 project, exploring `sale.order`.
**Then** delegates to `odoo-context-gatherer` utilizing local Odoo source via `rg`.

### Test 5 — Semantic Audit REJECT
**Given** Spec requires `compute_total(self)`. Code has `calculate_total(self)`.
**Then** `MISMATCH:func:compute_total`. Score > 0.1 → STATUS: REJECTED.

### Test 6 — Phase DAG Blocks Out-of-Order Execution
**Given** `sdd-state.yaml` with `sdd-design: pending`.
**When** `sdd-tasks` is invoked.
**Then** STATUS: BLOCKED — `sdd-design not completed`. `sdd-tasks` does not execute.

### Test 7 — Result Contract Validation
**Given** Phase agent emits JSON with missing `executive_summary` field.
**Then** `validate-result-contract.sh` exits 1. Orchestrator increments attempt count + retries.

### Test 8 — Circuit Breaker Abandons After 3 Failures
**Given** `sdd-apply` fails attempt 1, 2, 3.
**Then** exit code 2. `circuit_breaker.abandoned_phases` = `["sdd-apply"]`. No further retry.

### Test 9 — Apply Continuity Resume
**Given** `apply-progress.yaml` with T1 `completed`, T2 `running`, T3 `pending`.
**When** `sdd-apply` restarts.
**Then** T1 skipped. Resumes from T2. T3 executed after T2.

## Expected Results
- Git isolation: `apply/` branch + fast-forward merge. Zero remote dependency.
- Odoo auto-detection via `__manifest__.py`.
- Odoo skills/supplements auto-loaded.
- Semantic audit: spec-vs-code via `rg` Step 0.
- FMEA + FSM gate mandatory in `sdd-design`.
- Phase DAG enforced via `sdd-state.yaml` — bypassing any phase blocked at source.
- Result Contract validated by orchestrator after every phase.
- Circuit Breaker prevents Ralph Loops — 3 attempts max per phase.
- Apply Continuity: resume from last completed task, no full restart needed.
