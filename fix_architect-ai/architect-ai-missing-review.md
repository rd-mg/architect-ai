# architect-ai — Missing Review & Gap Analysis

**Date:** 2026-06-27  
**Scope:** Everything not yet fully analysed or planned across all 7 batches  
**Status:** Extends the existing full-analysis.md and fix-plan.md

> Severity: 🔴 Critical · 🟠 High · 🟡 Medium · 🟢 Low

---

## 1. L0 Routing Emission Labels — Fully Inverted 🔴

### The Bug

In `internal/assets/_shared/architect-identity.md` (the actual L0 thinking-agent
prompt), the routing emission labels are **backwards**:

```
# What the file currently says:
Mode A: "[L0→L1a] SDD intent: {intent}. Routing to SDD Orchestrator."
Mode B: "[L0→L1b] {intent}. Routing to General Orchestrator."

# What it SHOULD say:
Mode A: "[L0→L1b] SDD intent: {intent}. Routing to SDD Orchestrator."
Mode B: "[L0→L1a] {intent}. Routing to General Orchestrator."
```

The same inversion exists in ALL agent-specific orchestrator identity blocks — any
file that emits `[L0→L1x]` routing signals uses the wrong layer number.

Confirmed affected lines across batches:
- Line 9934-9935: `architect-identity.md` ← canonical source (WRONG)
- Line 11716: `sdd-orchestrator.md` emits `[L0→L1a]` for SDD (WRONG)
- Line 14213, 15173, 17791: Various agent orchestrators (WRONG)

The NEW V4 design output (from `orchestrator-redesign-L0-L1a-L1b.md`) correctly
uses L0→L1b for SDD and L0→L1a for General. The old codebase has the opposite.

**Fix — all affected files:**

```bash
# Batch fix across all orchestrator/identity files:
rg "\[L0→L1a\] SDD" internal/assets/ -l | xargs sed -i 's/\[L0→L1a\] SDD/[L0→L1b] SDD/g'
rg "\[L0→L1b\].*General\|Routing to General" internal/assets/ -l | \
  xargs sed -i 's/\[L0→L1b\]\(.*\)Routing to General/[L0→L1a]\1Routing to General/g'
```

Then verify manually — automated sed on routing labels can miss context-specific variants.

**eintegrate check (add to E-35):**
```go
if checkAnyFile(root, "architect-identity.md", "[L0→L1a] SDD") {
    errs = append(errs, "E-35: architect-identity.md emits [L0→L1a] for SDD (should be L1b)")
}
```

---

## 2. Phase Protocols Never Individually Reviewed 🟠

### 2.1 `sdd-design.md` — Missing CodeGraph + No-Stub Enforcement

**What exists (correct):**
- Architecture diagram mandatory ✅
- ADR table mandatory ✅
- YAGNI Gate mandatory ✅
- Poka-Yoke Checklist mandatory ✅
- Open Questions blocks apply ✅
- Alternative designs section ✅

**What is MISSING:**

**a) CodeGraph for design-time boundary validation:**

```markdown
## Step 0b: Semantic Boundary Analysis (BEFORE any design decisions)

IF codegraph available:
  codegraph_context(query: "{change_topic}", maxNodes: 25)
  → Understand EXISTING call chains before designing new ones
  
  For each affected module in Affected Areas (from proposal):
    codegraph_impact(nodeId: "{module_entrypoint}", depth: 3)
    → Confirm blast radius matches proposal's Affected Areas
    → If blast radius larger than expected → flag as design risk
    → If blast radius smaller → proposal may have over-scoped

  codegraph_trace(entry: "{primary_entrypoint}")
  → Understand existing flow BEFORE designing replacement/extension
```

**b) No-Stub Enforcement in Mandatory Sections:**

Add to `## Result Processing`:
```markdown
- **Stub Gate**: Reject if ANY mandatory section contains:
  "To be designed", "TBD", "TODO", "N/A" (without justification),
  or is empty. Every section must be substantively complete.
- **Open Questions Gate**: If Open Questions is non-empty, set status to `partial`.
  Orchestrator must surface questions to user BEFORE routing to sdd-tasks.
```

**c) Missing: sequential-thinking for architectural alternatives:**

The design phase should mandate sequential_thinking with branch exploration:
```markdown
## Step 0a: Architectural Branching (MANDATORY)
Call sequential_thinking with ≥ 2 branchId values before any design decision.
Explore at minimum: (A) extend existing pattern, (B) introduce new abstraction.
Document rationale for chosen branch in ADR table.
```

**d) Missing: Architecture Guardrails reference:**

Add to mandatory sections:
```markdown
- **Architecture Guardrails Compliance**: Validate design against all 5 Constitution rules.
  For each rule: state whether design complies and why.
  If any rule is violated: set status to `blocked`. Do NOT proceed to sdd-tasks.
```

---

### 2.2 `sdd-propose.md` — CodeGraph for Blast Radius, Viability Score Calibration

**What exists (correct):**
- Viability Score (1-15) blocks at < 8 ✅
- Pre-mortem section ✅
- Open Assumptions table ✅
- Hypothesis Branching with sequential_thinking ✅

**What is MISSING:**

**a) CodeGraph for "Affected Areas" verification:**

Currently: "Affected Areas (concrete file paths where possible)" — uses developer intuition.

```markdown
## Step 0b: Blast Radius Verification (if codegraph available)

For each suspected affected area:
  codegraph_callers(nodeId: "{suspected_entrypoint}")
  → Confirms actual inbound callers (not guessed)
  
  codegraph_impact(nodeId: "{suspected_entrypoint}", depth: 2)
  → Verifies blast radius BEFORE committing to scope

Record actual codegraph output as the Affected Areas list.
This makes the Pre-mortem "weakest dependency" analysis evidence-based.
```

**b) Viability Score formula needs clarification:**

Current formula: Sum(Complexity 1-5, Clarity 1-5, Tooling 1-5). Max = 15. Block at < 8.

Problem: A change with perfect Clarity (5) and Tooling (5) but extreme Complexity (1)
scores 11 — not blocked — despite being maximally complex. The scoring doesn't penalize
extreme complexity enough.

**Fix:** Change to a weighted formula:
```markdown
Viability Score = (Clarity × 2) + (Tooling × 2) + (Complexity × 1)
Max = 25. Block if: Score < 14 OR Complexity ≤ 1.
```
This ensures high complexity is a blocking factor regardless of other scores.

---

### 2.3 `sdd-spec.md` — Missing Gaps

**What exists (correct):**
- FMEA table for external I/O ✅
- Sad-path BDD for severity ≥ 3 ✅
- UI FSM (stateDiagram-v2) ✅
- Accessibility Contract ✅

**What is MISSING:**

**a) Explicit no-stub enforcement:**

```markdown
## Result Processing — Stub Gate (MANDATORY)

Reject spec if ANY capability section contains:
- "TODO", "TBD", "PLACEHOLDER", "unclear", "to be determined"
- Empty Preconditions, Behavior, or Error Handling sections
- Unmeasurable success criteria ("the system should work correctly")

Set status: blocked. List each stub location. Route back to sub-agent.
```

**b) Cross-spec consistency check:**

When a spec has multiple capabilities, ensure they don't contradict each other:
```markdown
## Cross-Capability Consistency Check

For each pair of capabilities:
  - Do their Invariants conflict? (e.g., Capability A says "X always true", B says "X can be false")
  - Do their Preconditions assume different system states?
  - If conflict found → set status: blocked, escalate to user
```

**c) Engram prior-spec lookup:**

```markdown
## Step 0: Prior Spec Check (BEFORE writing new spec)

mem_search("sdd/{project}/spec", project: "{project}")
→ If similar spec found: review it. Extract any reusable invariants or patterns.
  DO NOT repeat what's already specified elsewhere. Cross-reference instead.
```

---

### 2.4 `sdd-tasks.md` — Missing Traceability and Spec Coverage Enforcement

**What exists (correct):**
- Hierarchical numbering ✅
- Atomic task requirement (< 30 min) ✅
- Execution Graph (≥ 5 tasks) ✅
- Risk classification ✅
- Vertical Slice Organization ✅

**What is MISSING:**

**a) Spec coverage enforcement — every capability must have ≥ 1 task:**

```markdown
## Spec Coverage Check (MANDATORY before persisting)

Load: sdd/{change-name}/spec → mem_get_observation(id)
Extract: all capability names from spec

For each capability:
  - Find ≥ 1 task with acceptance criterion traceable to this capability
  - IF no matching task → ADD a task for it. Do NOT skip.
  - Log: "Spec capability '{name}' covered by task(s): {N.N, N.N}"

IF any capability has zero tasks → status: partial (not complete).
```

**b) Traceability links in task format:**

Update the task format to include `Spec-ref`:
```markdown
- [ ] {number} {action verb} {target}
      Acceptance: {condition}
      Spec-ref: {capability name from sdd-spec}   ← ADD THIS
      Depends-on: {comma separated task numbers, or NONE}
      Parallel-safe: {true|false}
      Risk: LOW | MEDIUM | HIGH
      Risk-reason: {required only when HIGH}
```

**c) 530-word budget is too tight:**

A spec with 8 capabilities, each needing 3-4 tasks, plus the execution graph, plus
documentation tasks generates 500+ words in task content alone. The budget excludes
setup tasks, migration tasks, and documentation tasks.

**Fix:** Change to:
```markdown
## Size Budget: 800 words max (was 530 — insufficient for changes with ≥ 5 capabilities)
```

---

### 2.5 `sdd-onboard.md` — Missing Setup Steps

**What exists (correct):**
- Orchestrator-driven (not sub-agent) ✅
- Socratic questions ✅
- Phase walkthrough ✅
- Resource pointers ✅

**What is MISSING:**

**a) Project tooling setup during onboarding:**

```markdown
## Step 1b: Project Tooling Setup (MANDATORY — run during onboarding)

Before walking through phases:
1. Run sdd-init (already in Step 1 ✅)
2. Verify skill-registry:
   - mem_search("skill-registry", project) → if not found, run skill-registry build
   - Show user: "X skills loaded. Active: {list}"
3. Initialize CodeGraph (if available):
   - Run: codegraph init -i --quiet
   - Confirm: "Code graph initialized. Semantic search active."
4. Detect Odoo overlay (if applicable):
   - If IS_ODOO detected → activate overlay, show: "Odoo {version} overlay active"
5. Confirm Engram connection:
   - mem_context() → show recent session history (or "No prior sessions")
```

**b) User preference persistence:**

Currently onboarding saves `execution_mode` and `artifact_store` but not:
```markdown
Add to Persistence:
mem_save(
  topic_key: "sdd-onboard/{project}/preferences",
  content: {
    execution_mode: {mode},
    artifact_store: {mode},
    delivery_strategy: {strategy},  ← add this
    test_runner: {runner},           ← add this
    preferred_model: {model}         ← add this
  }
)
```

---

### 2.6 `sdd-init.md` — Missing `openspec/` Creation and `codegraph init`

**What exists (correct):**
- Pre-flight validation ✅
- Build/test/lint health checks ✅
- Artifact store mode detection ✅
- Test baseline persistence ✅
- Surface mapping ✅

**What is MISSING:**

**a) `openspec/` directory creation — critical gap:**

`sdd-spec.md` writes to `openspec/changes/{change-name}/specs/{domain}/spec.md`.
But NO phase protocol or Go code creates this directory structure. The spec write
will fail with "no such file or directory" unless the directory exists.

```markdown
### Step 4b: OpenSpec Directory Initialization (IF artifact_mode ≠ none AND ≠ engram)

IF artifact_store = openspec OR artifact_store = hybrid:
  Create directory structure:
  openspec/
    changes/{change-name}/
      specs/
    specs/
    archive/
  
  Create openspec/.gitignore:
  archive/
  
  Notify: "openspec/ initialized at {project_root}/openspec/"
```

**b) `codegraph init` as part of project initialization:**

```markdown
### Step 8b: Code Graph Initialization (if codegraph available)

IF codegraph tool available in verified tool list:
  Run: codegraph init -i --quiet
  → Builds semantic index of project codebase
  → Required for codegraph_context/callers/impact in sdd-explore
  Note to user: "Code graph initialized. Semantic analysis active for sdd-explore."
  
  Persist: mem_save("sdd-init/{project}/codegraph", {initialized: true, timestamp: ...})
```

---

## 3. `architect-identity.md` — Additional Gaps 🔴

Beyond the L0 routing label inversion (§1):

### 3.1 Spanish "listo" in `mem_session_summary` trigger

In `internal/assets/claude/general-orchestrator.md` (line ~16135):
```
Before ending a session or saying "done" / "that's it" (including Spanish "listo"), 
call mem_session_summary
```

And in `sdd-orchestrator.md` (line ~18768):
```
Before ending a session or saying "done" / "that's it" (including Spanish equivalent "listo")
```

🟠 **Fix:** Remove Spanish trigger word from BOTH files. The session summary should
trigger on English phrases only:
```markdown
## Session End Protocol (MANDATORY)

Trigger: User says "done", "that's all", "finished", "close session", 
         "wrap up", "goodbye", "thanks" — or conversation shows natural end.
(NOT "listo" — removed per Spanish regionalism policy)
```

### 3.2 `architect-identity.md` Still Contains `iniciar sdd` and `haceme un sdd`

```python
# Natural language (regex): \b(use sdd|start sdd|sdd mode|spec-driven|iniciar sdd|haceme un sdd)\b
```

Both `iniciar sdd` (Spanish: start sdd) and `haceme un sdd` (Rioplatense: make me an sdd)
should be removed. The English-only routing table should be:

```python
# Natural language (regex): \b(use sdd|start sdd|begin sdd|sdd mode|spec-driven|let's sdd)\b
```

---

## 4. `apply-continuity.md` — Schema and Logic Gaps 🟡

### 4.1 Missing `partial` Task Status

Current statuses: `pending | running | completed | failed`

Missing: `partial` — for tasks that started execution but couldn't complete (e.g.,
tests failing, dependency missing). Without this, an interrupted task appears as
`running` or `failed` with no way to track what was partially done.

**Fix:** Add `partial` status with required `partial_reason` field:
```yaml
tasks:
  - id: "task 12"
    description: "Create state template"
    status: "partial"
    partial_reason: "Template created but tests failed — needs retry"
    completed_at: ""
    started_at: "2026-06-22T10:00:00Z"
```

### 4.2 Missing Git Branch Reference

The continuity file doesn't record the apply branch name. If the project has multiple
in-flight apply branches, resumption cannot determine which branch to checkout.

**Fix:** Add `apply_branch` field:
```yaml
change_name: "phase-02-sdd-worktrees"
apply_branch: "apply/phase-02-sdd-worktrees"  ← ADD
started_at: "2026-06-22T00:52:04Z"
```

### 4.3 Missing Task Checksum Verification

When resuming, the task list in `.atl/apply-progress.yaml` should be compared against
the current task list in Engram. If tasks have been modified since the progress file
was created, resuming could apply outdated tasks.

**Fix:** Add checksum field:
```yaml
tasks_checksum: "sha256:{hash_of_tasks_artifact}"
```

At resumption: `sha256(current_tasks_artifact) == tasks_checksum`? If not → warn user
before resuming, show diff.

---

## 5. `rollback-harness.md` — Fragile Implementation 🟡

### 5.1 Python Regex on YAML is Fragile

The rollback script uses Python regex to modify `sdd-state.yaml`:
```python
content = re.sub(
    r'(  sdd-apply:.*?status: )\"(running|failed)\"',
    r'\1\"pending\"',
    content, flags=re.DOTALL
)
```

This regex will silently fail if:
- Indentation differs (2 spaces assumed)
- YAML keys are quoted differently
- Multi-line values cause `DOTALL` to match too much

**Fix:** Replace with a proper YAML parser:
```python
import yaml, sys
with open('.atl/sdd-state.yaml') as f:
    state = yaml.safe_load(f)

if 'sdd-apply' in state.get('phases', {}):
    if state['phases']['sdd-apply'].get('status') in ('running', 'failed'):
        state['phases']['sdd-apply']['status'] = 'pending'

with open('.atl/sdd-state.yaml.tmp', 'w') as f:
    yaml.dump(state, f, default_flow_style=False)
os.rename('.atl/sdd-state.yaml.tmp', '.atl/sdd-state.yaml')
```

Or (more robust) implement this in Go:
```go
// internal/app/rollback.go
func ResetApplyStatus(stateYAMLPath string) error {
    // Use gopkg.in/yaml.v3 (already likely in go.mod)
    // Parse → modify phase.status → write atomically
}
```

### 5.2 `architect-ai restore {filename}` Command May Not Exist

The rollback script says `architect-ai restore CLAUDE.md` but looking at `internal/app/app.go`
and the commands list, there is no `restore` subcommand.

**Fix:** Add `restore` subcommand to `internal/app/app.go`:
```go
case "restore":
    return app.RunRestore(args[1:])
```

Or: rename references to the actual backup restore mechanism (TUI → Backups screen).

### 5.3 No Agent-Specific MCP Rollback

If MCP config is corrupted for a specific agent, the rollback only says
"Regenerate from template: `architect-ai mcp --regenerate`". This command may not exist.

**Fix:** Add to rollback table:
```
| Claude MCP corrupted | Regenerate | architect-ai install --agent claude-code --component mcp --dry-run |
| VSCode MCP corrupted | Regenerate | architect-ai install --agent vscode-copilot --component mcp       |
```

---

## 6. `sdd-tasks.md` → `sdd-apply.md` Handoff Gap 🟠

### 6.1 `sdd-tasks.md` Does Not Gate Apply Authorization Explicitly

`sdd-tasks.md` ends with:
```
- Next recommended: sdd-apply
```

But there is no explicit "IMPLEMENTATION AUTHORIZED" sentinel in the tasks artifact.
The `sdd-apply.md` pre-apply gate checks for `tasks.md explicitly authorizes implementation phase`.

**Fix:** Add to `sdd-tasks.md` persistence:
```markdown
## Authorization Sentinel (MANDATORY — last line of tasks artifact)

```
IMPLEMENTATION_AUTHORIZED: true
authorized_at: {ISO_8601_timestamp}
authorized_tasks: {total_task_count}
authorized_by: sdd-tasks-{version}
```

sdd-apply will HALT if this sentinel is missing.
```

### 6.2 `sdd-apply.md` Cross-Phase Reference Check Is Too Slow If Run Inline

The Pre-Apply Completeness Gate (FIX-09) checks every task against spec capabilities.
For changes with 20+ tasks and 10+ capabilities, this creates 200+ comparison operations
inline in the sub-agent context.

**Fix:** Add to the Pre-Apply Completeness Gate:
```markdown
### Performance Guard
IF tasks_count × capabilities_count > 100:
  Use ctx_batch_execute to run cross-reference checks in parallel batches
  ctx_batch_execute([
    "mem_get_observation({spec_id}) | grep 'Capability:'",
    "mem_get_observation({tasks_id}) | grep 'Spec-ref:'"
  ])
  Compare outputs programmatically
```

---

## 7. Cross-Agent Protocol Parity 🟠

### 7.1 Agent Variant Completeness Matrix

Not all agents have all 9 SDD phase protocols. Current verified state:

| Phase | Claude | Generic | Gemini | OpenCode | Cursor | Kiro | Codex | GGA | Windsurf |
|-------|--------|---------|--------|----------|--------|------|-------|-----|---------|
| sdd-init | ✅ | ✅ | ? | ? | ? | ? | ? | ? | ? |
| sdd-onboard | ✅ | ✅ | ? | ? | ? | ? | ? | ? | ? |
| sdd-explore | ✅ | ✅ | ✅ | ? | ? | ? | ? | ? | ? |
| sdd-propose | ✅ | ✅ | ? | ? | ? | ? | ? | ? | ? |
| sdd-spec | ✅ | ✅ | ? | ? | ? | ? | ? | ? | ? |
| sdd-design | ✅ | ✅ | ? | ? | ? | ? | ? | ? | ? |
| sdd-tasks | ✅ | ✅ | ? | ? | ? | ? | ? | ? | ? |
| sdd-apply | ✅ | ✅ | ✅ | ? | ✅ | ? | ? | ? | ? |
| sdd-verify | ✅ | ✅ | ? | ? | ? | ? | ? | ? | ? |
| sdd-archive | ✅ | ✅ | ? | ? | ? | ? | ? | ? | ? |

🟠 **Action required:** Verify every cell marked `?`. Missing phase protocols should
either be created (copying from generic) or explicitly documented as "uses generic".

### 7.2 Phase Protocol Divergence Check

Even where both exist, Claude and Generic versions may have diverged. Key divergence
risks:

- `sdd-apply.md`: Size budget differs (Claude: variable, Generic: 900 words)
- `sdd-design.md`: YAGNI Gate present in Claude — verify it exists in Generic
- `sdd-propose.md`: Viability Score in Claude — verify it exists in Generic
- `sdd-tasks.md`: Execution Graph in Claude — verify in Generic

**Action:** Run diff on all corresponding file pairs and align.

### 7.3 OpenCode Phase Protocol Location

OpenCode uses `~/.config/opencode/prompts/sdd/sdd-{phase}.md` (shared across profiles).
The phase files in `internal/assets/opencode/` should match the Claude/Generic content.
But OpenCode has sub-agent entries in `opencode.json` that reference these files.

Verify: does `internal/assets/opencode/` contain ALL 9 phase protocols?

---

## 8. Go Module Dependencies for New Code 🟡

### 8.1 `fsnotify` for L3 Hot-Reload (FIX-26 from common plan)

```go
// In FIX from L3 hot-reload, we use:
import "github.com/fsnotify/fsnotify"
```

This package is NOT in the current `go.mod`. Must add:
```bash
go get github.com/fsnotify/fsnotify@v1.7.0
go mod tidy
```

### 8.2 `gopkg.in/yaml.v3` for YAML manipulation

Used in rollback harness fix (§5.1). May already be present (bubbletea dependencies
often pull it in). Verify:
```bash
grep "gopkg.in/yaml" go.mod
```

### 8.3 `golang.org/x/sync/errgroup` for B1 Fix

Confirmed present at line 32558. No action needed.

### 8.4 New package: `internal/agents/antigravity-cli`

Must add to `internal/agents/` and ensure it is part of the Go build:
```bash
# After creating the package:
go build ./internal/agents/antigravity-cli/...
go test ./internal/agents/antigravity-cli/...
```

---

## 9. TUI Dead-End Routes 🟡

### 9.1 `ScreenUninstallConfirm` Missing Forward Route (SRE F-15)

The SRE audit finding F-15 identified this. Tests for `RenderUninstallConfirm` exist
but the `linearRoutes` navigation map may be missing the forward route from
`ScreenUninstallConfirm` to the actual uninstall execution screen.

**Action:** In `internal/tui/` (wherever `linearRoutes` is defined), verify:
```go
linearRoutes[ScreenUninstallConfirm] = ScreenUninstallResult
// AND reverse:
linearRoutes[ScreenUninstallResult] = ScreenUninstallConfirm
```

If missing, add the forward route.

---

## 10. `mem_session_summary` — Session End Enforcement 🟡

### 10.1 No Auto-Trigger Mechanism

Session summary is triggered only when user says "done" / "that's all" etc. There is
no mechanism to:
- Auto-trigger when context window reaches 80% saturation
- Trigger on conversation timeout
- Trigger at end of sdd-archive phase (natural SDD session end)

**Fix:** Add to the circuit-breaker and context-guardian:

```markdown
## Auto Session Summary Triggers (MANDATORY)

Trigger mem_session_summary automatically when:
1. Context saturation ≥ 80% (D4 ≥ 2) AND any SDD phase just completed
2. sdd-archive phase returns `status: complete` (natural SDD end)
3. sdd-verify returns APPROVED verdict (another natural SDD milestone)
4. Orchestrator detects D4 = 3 (saturated) → summarize BEFORE compaction

Parameters: Include `goal`, `accomplished`, `next_steps`, `key_files`, 
            `open_sdd_change` (if mid-SDD), `current_phase`.
```

### 10.2 Spanish "listo" in Trigger Phrase (Already noted in §3.1)

Remove from both `general-orchestrator.md` and `sdd-orchestrator.md`.

---

## 11. `openspec/` Directory Creation Gap 🔴

Neither the sdd-init phase protocol NOR any Go code creates the `openspec/`
directory structure required by `openspec` and `hybrid` artifact modes.

`sdd-spec.md` attempts to write to:
```
openspec/changes/{change-name}/specs/{domain}/spec.md
```

If `openspec/` does not exist, this will fail with a filesystem error.

### Fix Option A (prompt-level — in sdd-init.md):

Already addressed in §2.6 Fix (a) above. Add to sdd-init Step 4b.

### Fix Option B (Go-level — in install pipeline):

```go
// internal/components/sdd/inject.go
// When sdd component is installed, also create the openspec/ structure:
func EnsureOpenSpecDirs(projectRoot string) error {
    dirs := []string{
        filepath.Join(projectRoot, "openspec", "changes"),
        filepath.Join(projectRoot, "openspec", "specs"),
        filepath.Join(projectRoot, "openspec", "archive"),
    }
    for _, dir := range dirs {
        if err := os.MkdirAll(dir, 0o755); err != nil {
            return fmt.Errorf("create openspec dir %s: %w", dir, err)
        }
    }
    // Create openspec/.gitignore
    gitignore := filepath.Join(projectRoot, "openspec", ".gitignore")
    return os.WriteFile(gitignore, []byte("archive/\n"), 0o644)
}
```

**Recommendation:** Both options — Go creates the root structure at install; prompt
ensures subdirectory for each change-name is created at sdd-init time.

---

## 12. `sdd-design.md` Missing `+++Economic` Posture 🟡

`sdd-design.md` specifies `+++Critical + +++Systemic`. But the YAGNI Gate (already
mandatory in the design) is inherently an +++Economic analysis (cost/value trade-off
of abstractions). The posture should include +++Economic when evaluating YAGNI:

**Fix:** Add conditional posture:
```markdown
## Cognitive Posture
+++Critical + +++Systemic — Architecture needs both rigor and system view.

## Conditional Posture
+++Economic (if design involves ≥ 3 abstraction decisions):
  Apply when filling YAGNI Gate table. For each proposed abstraction:
  evaluate cost of implementation vs cost of deferral.
```

Note: This does NOT violate the "Max 2 postures" invariant — it's a secondary posture
for a specific sub-task within the design phase, not a third primary posture.

---

## 13. Windsurf, Codex, GGA — Phase Protocol Gaps 🟡

### 13.1 Windsurf: No Phase Protocols in Assets

Windsurf uses `internal/assets/windsurf/` but only has `sdd-orchestrator.md`. The 9
phase protocols are missing. Windsurf is a solo-agent (no sub-agent delegation) so
all phases run inline in the main conversation.

**Action:** Create `internal/assets/windsurf/sdd-phase-protocols/` with all 9 phases,
adapting from generic but with Windsurf-specific notes:
- Note Plan Mode vs Code Mode for each phase
- Use Windsurf Workflows for `sdd-new.md` ✅ (already exists)
- All phases run inline (no Task tool delegation)
- Context window: 200K (Cascade) — context management less critical but still needed

### 13.2 Codex: Solo-Agent, No Sub-Agent Delegation

Codex has `~/.codex/agents.md` (system prompt) but NO `sdd-phase-protocols/`.
Codex runs all 9 phases inline. The cumulative context across 9 inline phases will
approach the context limit for complex changes.

**Action:** Create `internal/assets/codex/sdd-phase-protocols/` with:
- All 9 phase protocols ✅
- Explicit context budget per phase (leave 40% for tool outputs)
- After each phase: `mem_session_summary` to compact (then `mem_context` to restore)
- "Amnesia-safe" SDD: every phase starts with `mem_search("sdd-init/{project}")` to
  restore full context without relying on conversation history

### 13.3 GGA: Provider Switcher — Which Model Gets Which Phase?

GGA routes to different AI backends. The GGA orchestrator must specify per-phase model
assignments that match GGA's routing capabilities.

**Action:** Verify `internal/assets/gga/general-orchestrator.md` and
`sdd-orchestrator.md` have GGA-specific model routing tables. If GGA is routing
to Claude: Claude's model assignments apply. If to Gemini: Gemini's apply.

---

## 14. Cognitive Posture in `sdd-design.md` Sub-Agents 🟡

The design phase currently uses a single sub-agent with `+++Critical + +++Systemic`.
For large architectural changes (D1 = 3, multiple subsystems), a single sub-agent
context becomes saturated before completing all mandatory sections.

**Recommendation:** For D1 ≥ 3, split design into two sub-agent calls:

```markdown
## Design Sub-Agent Split (when D1 ≥ 3)

Call 1 — Architecture (+++Critical):
  Focus: Module/component boundaries, ADR table, Alternative designs
  Output: partial design artifact

Call 2 — Constraints (+++Systemic):
  Focus: Data flow, Interface contracts, Error propagation, YAGNI Gate, Poka-Yoke
  Input: partial design from Call 1
  Output: complete design artifact

Persist merged result.
```

---

## 15. `sdd-ff` Fast-Forward Mode — Completeness Gate 🟡

The `sdd-ff` mode (fast-forward) bypasses full SDD. But it still reaches `sdd-apply`.
Does `sdd-ff` pass through the Pre-Apply Completeness Gate (FIX-09)?

Looking at the existing `sdd-ff` protocol: it skips to apply directly with minimal
spec/design context. The completeness gate would ALWAYS fail for sdd-ff (no formal spec,
no FMEA table).

**Fix:** Add sdd-ff override to the Pre-Apply Completeness Gate:
```markdown
## Pre-Apply Completeness Gate

IF change_origin == "sdd-ff":
  Apply SIMPLIFIED gate:
  - [ ] Change is ≤ 3 files
  - [ ] Change is additive-only (no behavior modification)
  - [ ] Existing tests cover the changed area
  - [ ] No external API or schema change
  
  IF simplified gate passes → PROCEED (skip full spec/design checks)
  IF simplified gate fails → ROUTE to full SDD. sdd-ff is not eligible.

IF change_origin == "sdd" (full pipeline):
  Apply FULL gate (as in FIX-09)
```

---

## 16. Summary — Complete Missing Items List

| # | Issue | Severity | Files Affected |
|---|-------|---------|----------------|
| 1 | L0 routing emission labels inverted (`[L0→L1a]` for SDD) | 🔴 | `architect-identity.md`, all orchestrators |
| 2a | `sdd-design.md` missing CodeGraph boundary analysis | 🟠 | `sdd-design.md` (all agents) |
| 2b | `sdd-design.md` missing sequential_thinking for alternatives | 🟡 | `sdd-design.md` |
| 2c | `sdd-design.md` missing Architecture Guardrails compliance check | 🟠 | `sdd-design.md` |
| 2d | `sdd-design.md` missing no-stub enforcement in Result Processing | 🟠 | `sdd-design.md` |
| 3a | `sdd-propose.md` missing CodeGraph for Affected Areas | 🟠 | `sdd-propose.md` |
| 3b | `sdd-propose.md` Viability Score formula doesn't penalize complexity | 🟡 | `sdd-propose.md` |
| 4a | `sdd-spec.md` missing explicit no-stub enforcement | 🟠 | `sdd-spec.md` |
| 4b | `sdd-spec.md` missing cross-capability consistency check | 🟡 | `sdd-spec.md` |
| 4c | `sdd-spec.md` missing prior spec lookup in Engram | 🟡 | `sdd-spec.md` |
| 5a | `sdd-tasks.md` missing spec coverage enforcement (all capabilities) | 🔴 | `sdd-tasks.md` |
| 5b | `sdd-tasks.md` missing `Spec-ref` traceability field in task format | 🟠 | `sdd-tasks.md` |
| 5c | `sdd-tasks.md` 530-word budget too tight | 🟡 | `sdd-tasks.md` |
| 6a | `sdd-onboard.md` missing project tooling setup (codegraph init, skill-registry) | 🟠 | `sdd-onboard.md` |
| 6b | `sdd-onboard.md` missing delivery_strategy + test_runner in persistence | 🟡 | `sdd-onboard.md` |
| 7a | `sdd-init.md` missing `openspec/` directory creation | 🔴 | `sdd-init.md`, `sdd/inject.go` |
| 7b | `sdd-init.md` missing `codegraph init` step | 🟠 | `sdd-init.md` |
| 8 | `architect-identity.md` L0 routing labels inverted (§1) | 🔴 | `architect-identity.md` |
| 9 | Spanish "listo" in `mem_session_summary` trigger | 🟠 | `general-orchestrator.md`, `sdd-orchestrator.md` |
| 10 | `apply-continuity.md` missing `partial` status, `apply_branch`, checksum | 🟡 | `apply-continuity.md` |
| 11 | `rollback-harness.md` Python regex on YAML is fragile | 🟡 | `rollback-harness.md` |
| 12 | `rollback-harness.md` `architect-ai restore` command may not exist | 🟠 | `app.go` |
| 13 | `sdd-tasks.md` IMPLEMENTATION_AUTHORIZED sentinel missing | 🟠 | `sdd-tasks.md` |
| 14 | `sdd-apply.md` cross-phase check too slow inline for large changes | 🟡 | `sdd-apply.md` |
| 15 | Agent phase protocol parity matrix — many agents unverified | 🟠 | All agent assets |
| 16 | `fsnotify` not in go.mod (needed for L3 hot-reload) | 🟡 | `go.mod` |
| 17 | TUI `ScreenUninstallConfirm` missing forward route | 🟡 | TUI router |
| 18 | No auto-trigger for `mem_session_summary` (only manual "done") | 🟡 | All orchestrators |
| 19 | `openspec/` directory never created by any code | 🔴 | `sdd/inject.go`, `sdd-init.md` |
| 20 | `sdd-design.md` missing `+++Economic` for YAGNI evaluation | 🟡 | `sdd-design.md` |
| 21 | Windsurf has no sdd-phase-protocols/ directory | 🟠 | `internal/assets/windsurf/` |
| 22 | Codex has no sdd-phase-protocols/ directory | 🟠 | `internal/assets/codex/` |
| 23 | GGA model routing per phase not verified | 🟡 | GGA orchestrators |
| 24 | Large D1≥3 design needs two sub-agent calls (context saturation) | 🟡 | `sdd-design.md` |
| 25 | `sdd-ff` Pre-Apply gate always fails (no spec/FMEA in fast-forward) | 🟠 | `sdd-apply.md`, `sdd-ff` |

---

## 17. Fix Plan Additions for Missing Items

These extend Sprint 3 or form a Sprint 4:

### Sprint 3 Additions (Week 6)

| Fix ID | Description | Files |
|--------|-------------|-------|
| FIX-31 | Fix L0 routing emission labels (L1a↔L1b swap) | All orchestrators + `architect-identity.md` |
| FIX-32 | Add `openspec/` dir creation to `sdd-init.md` + `sdd/inject.go` | `sdd-init.md`, `inject.go` |
| FIX-33 | Add `IMPLEMENTATION_AUTHORIZED` sentinel to `sdd-tasks.md` | `sdd-tasks.md` (all agents) |
| FIX-34 | Remove Spanish "listo" from `mem_session_summary` triggers | `general-orchestrator.md`, `sdd-orchestrator.md` |
| FIX-35 | Add `partial` status + `apply_branch` + checksum to `apply-continuity.md` | `_shared/apply-continuity.md` |
| FIX-36 | Fix `rollback-harness.md` YAML parsing (Python → yaml.safe_load) | `_shared/rollback-harness.md` |
| FIX-37 | Add `architect-ai restore` subcommand to `app.go` | `internal/app/app.go` |
| FIX-38 | Add `fsnotify` to `go.mod` | `go.mod` |
| FIX-39 | Fix TUI `ScreenUninstallConfirm` forward route | TUI router file |

### Sprint 4 — Phase Protocol Completion (Week 7–8)

| Fix ID | Description | Files |
|--------|-------------|-------|
| FIX-40 | Add CodeGraph to `sdd-design.md` (all agents) | `sdd-design.md` × 8 agents |
| FIX-41 | Add no-stub enforcement to `sdd-design.md` Result Processing | `sdd-design.md` × 8 agents |
| FIX-42 | Add Architecture Guardrails compliance check to `sdd-design.md` | `sdd-design.md` × 8 agents |
| FIX-43 | Add CodeGraph to `sdd-propose.md` Affected Areas | `sdd-propose.md` × 8 agents |
| FIX-44 | Fix Viability Score formula (weighted) in `sdd-propose.md` | `sdd-propose.md` × 8 agents |
| FIX-45 | Add no-stub + cross-capability consistency to `sdd-spec.md` | `sdd-spec.md` × 8 agents |
| FIX-46 | Add spec coverage enforcement + Spec-ref field to `sdd-tasks.md` | `sdd-tasks.md` × 8 agents |
| FIX-47 | Fix 530→800 word budget in `sdd-tasks.md` | `sdd-tasks.md` × 8 agents |
| FIX-48 | Add project tooling setup to `sdd-onboard.md` | `sdd-onboard.md` × 8 agents |
| FIX-49 | Add `codegraph init` step to `sdd-init.md` | `sdd-init.md` × 8 agents |
| FIX-50 | Add auto `mem_session_summary` triggers to all orchestrators | All orchestrators |
| FIX-51 | Create `sdd-phase-protocols/` for Windsurf (9 phases from generic) | `internal/assets/windsurf/` |
| FIX-52 | Create `sdd-phase-protocols/` for Codex (9 phases, solo-agent) | `internal/assets/codex/` |
| FIX-53 | Add sdd-ff simplified pre-apply gate to `sdd-apply.md` | `sdd-apply.md` × 8 agents |
| FIX-54 | Verify and fix agent phase protocol parity matrix (all agents) | All agent asset dirs |
| FIX-55 | Add `EnsureOpenSpecDirs()` to `internal/components/sdd/inject.go` | `sdd/inject.go` |
