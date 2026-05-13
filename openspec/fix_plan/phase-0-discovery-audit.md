# Phase 0 — Discovery & Baseline Audit

> **Cognitive Mode**: +++Forensic +++Systemic +++Empirical  
> **CCLD Tag**: `[PHASE-0][DISCOVERY][AUDIT]`  
> **Status**: READY — Execute before any refactoring  
> **Estimated Duration**: 1–2 sessions  

---

## 0.1 Objective

Establish a verified, evidence-backed baseline of the `architect-ai` architecture **before** any code changes. Every bottleneck claim in Phases 1–5 must be traceable back to a finding in this phase. No refactoring without a corresponding audit entry.

---

## 0.2 Scope Inventory

### 0.2.1 Go Runtime Layer (Binary)

| Package | File(s) | Audit Focus |
|---|---|---|
| `internal/pipeline` | `orchestrator.go`, `runner.go`, `stages.go`, `rollback.go` | Sequential step execution; no goroutine fan-out |
| `internal/planner` | `graph.go`, `resolver.go`, `order.go`, `review.go`, `types.go` | Topological sort producing fixed sequential order; soft-ordering constraints |
| `internal/cli` | `install.go`, `skill_registry.go` | Registry collection sequential I/O; install flag parsing |
| `internal/app` | `app.go`, `skills_cmd.go`, `selfupdate.go` | `tuiExecute` → `Orchestrator.Execute` monolithic call path; state R/W inline |
| `internal/catalog` | `agents.go`, `components.go`, `skills.go` | Static slice catalogs; no lazy resolution |
| `internal/state` | `state.go` (inferred) | Read-merge-write pattern; concurrent access risk |
| `internal/backup` | `snapshot.go`, `manifest.go`, `compression.go` | I/O-heavy operations; potential for async with context |
| `internal/agents/*` | `adapter.go` (vscode, windsurf) | Agent metering; per-adapter synchrony |

### 0.2.2 Prompt / Orchestration Layer (Markdown)

| File | Audit Focus |
|---|---|
| `.agent/skills/_shared/general-orchestrator.md` | Routing fork logic; parallel delegation rules; sub-agent template verbosity |
| `.agent/skills/_shared/sdd-orchestrator.md` | Phase dependency graph; session-setup triplet; tool availability probes |
| `.agent/skills/_shared/skill-resolver.md` | Resolution protocol; context injection pattern |
| `.agent/skills/_shared/persistence-contract.md` | Engram/openspec write patterns; state duplication |
| `.agent/skills/_shared/research-routing.md` | 5-step escalation; mode-based restrictions |
| `.atl/skill-registry.md` | Flat markdown index; full-file read required per agent invocation |

---

## 0.3 Audit Tasks

### Task A — Pipeline Sequential Bottleneck Measurement

**Goal**: Prove that `Runner.Run()` executes all steps serially even when steps have no data dependency.

**Steps**:
1. `rg "for _, step := range steps" internal/pipeline/runner.go` — confirm sequential loop, no goroutine fan-out.
2. `rg "errgroup\|sync\.WaitGroup\|go func" internal/pipeline/` — confirm zero concurrency primitives present.
3. `rg "StagePlan\|BuildRealStagePlan" internal/cli/` — trace how plan is populated and confirm all `Apply` steps are independent for component installs.
4. Document: which step IDs are independent (no write conflict on same file)?
5. Baseline: add timing instrumentation to `Runner.Run()` via test — measure time delta for N=6 independent steps (Engram, SDD, Skills, Context7, Persona, Permissions) on a dev install.

**Expected Finding**: N×T_step latency where T_step ≈ identical. Target post-refactor: max(T_steps) + scheduling overhead.

**Evidence Record**: `audit/pipeline-sequential.md`

---

### Task B — Skill Registry Kernel I/O Audit

**Goal**: Quantify serial I/O overhead in `WriteLocalSkillRegistry`.

**Steps**:
1. `rg "collectUserSkills\|collectProjectSkills\|collectOverlayContent" internal/cli/skill_registry.go` — confirm sequential calls.
2. `rg "os.ReadFile\|filepath.Walk\|os.ReadDir" internal/cli/skill_registry.go` — enumerate all blocking I/O calls.
3. Map each collect function to its I/O operations: file walks, SKILL.md reads, frontmatter parses.
4. Measure: how many files does each function read on a typical Odoo project? (check `.atl/skill-registry.md` content width — ~30 skill entries).
5. Confirm: are the three collect calls independent? (no shared state mutation → yes, based on code reading).
6. Check `deduplicateSkills` — is it append-safe for concurrent input? (no: uses single `allSkills` slice → needs mutex or channel fanin).

**Expected Finding**: 3 serial I/O traversals collectable in parallel; `deduplicateSkills` needs sync wrapper.

**Evidence Record**: `audit/skill-registry-io.md`

---

### Task C — Orchestrator Routing Fork Analysis

**Goal**: Map the exact delegation path for SDD vs non-SDD intents and find where the fork is delayed or missing.

**Steps**:
1. Read `general-orchestrator.md` → Routing Table section.
2. Read `sdd-orchestrator.md` → Intent Resolution section.
3. Map the routing path:
   - User message → General Orchestrator receives it
   - General Orchestrator scans `Routing Table`
   - On SDD match: confirm it routes to `sdd-orchestrator` **immediately** or continues general setup first
   - Find: does the General Orchestrator run `Tool Availability Check` BEFORE routing to SDD Orchestrator? (Yes → redundant; SDD Orchestrator runs its own check)
4. Find: are `Tool Availability Check` probes (`mem_search × 3`) launched sequentially? (Yes — no parallel dispatch in current prompts)
5. Find: does the sub-agent template embed full SKILL.md content or compact rules? (Full `{content of ...}` references — context saturation risk)
6. Find: `Session-Setup Triplet` — does the General Orchestrator's triplet conflict with SDD Orchestrator's triplet? (Yes — both run independent tool probes)

**Expected Finding**: Double tool-probe overhead on SDD path; sequential probe calls; redundant session setup.

**Evidence Record**: `audit/orchestrator-routing.md`

---

### Task D — State Management Concurrent-Safety Audit

**Goal**: Identify shared-state race conditions in `app.go` and `state.go`.

**Steps**:
1. `rg "state.Read\|state.Write" internal/app/app.go` — find all R/W call sites.
2. Confirm: are `loadPersistedAssignments` and `persistAssignments` called from concurrent goroutines? (Currently no — but `tuiExecute` calls both synchronously post-install. Risk emerges when parallel pipeline is added.)
3. `rg "sync\." internal/app/app.go internal/state/` — confirm no mutex on state file access.
4. `rg "race" internal/app/app_test.go` — check if `-race` tests exist.
5. Check `backup.DeleteBackup`, `backup.RenameBackup`, `backup.TogglePin` — are these called from TUI event handlers that could fire concurrently?

**Expected Finding**: No mutex on state file; race window exists post-pipeline-parallelization.

**Evidence Record**: `audit/state-management.md`

---

### Task E — Context Window Saturation Analysis

**Goal**: Measure token overhead of current sub-agent prompt injection.

**Steps**:
1. Extract the `Sub-Agent Launch Template` from `general-orchestrator.md`.
2. Identify all `{content of ...}` placeholder expansions.
3. Count approximate tokens for each injected block:
   - `_shared/context-mode-routing-policy.md` — estimate size
   - `skills/_shared/general-phase-common.md` — estimate size
   - Mandatory skills compact rules (4 skills × compact rules length)
4. Find: does the template inject the same blocks for every sub-agent call, even when sub-agent doesn't use them? (Yes — Mandatory Skills are unconditional)
5. Find: is there a diff-based injection (only inject changed/new context)? (No)
6. Find: is there a `context-guardian` enforcement mechanism in the current sub-agent template? (Listed as skill but not verified as active enforcement)

**Expected Finding**: 800–2000 redundant tokens per sub-agent call; 5–8 sub-agents per SDD workflow = 4K–16K redundant tokens per session.

**Evidence Record**: `audit/context-saturation.md`

---

### Task F — SDD Phase Dependency Graph Validation

**Goal**: Confirm which SDD phases are truly sequential vs parallelizable.

**Steps**:
1. Extract dependency graph from `sdd-orchestrator.md`:
   ```
   proposal → specs → tasks → apply → verify → archive
               ↑
            design
   ```
2. For each phase transition, document:
   - Input artifacts required
   - Output artifacts produced
   - Whether input/output overlap with any sibling phase
3. Identify parallelizable opportunities:
   - `sdd-explore` of multiple modules → parallel (confirmed in orchestrator: "Multiple file explorations → parallel")
   - `sdd-spec` for unrelated features → parallel (confirmed)
   - Tests + static analysis during `sdd-verify` → parallel (confirmed)
   - But: `sdd-apply` tasks modifying the same files → NEVER parallel (confirmed)
4. Find: does the current `sdd-orchestrator.md` `Parallel Delegation` section actually enforce parallel dispatch for the identified cases? Or is it aspirational?

**Expected Finding**: Parallelization is documented in the prompt but not mechanically enforced; orchestrator still decides at runtime without a static dispatch table.

**Evidence Record**: `audit/sdd-phase-graph.md`

---

## 0.4 Audit Completion Criteria

All 6 tasks must produce an evidence record before Phase 1 begins.

| Task | Evidence Record | Status |
|---|---|---|
| A — Pipeline Sequential | `audit/pipeline-sequential.md` | [ ] |
| B — Skill Registry I/O | `audit/skill-registry-io.md` | [ ] |
| C — Orchestrator Routing | `audit/orchestrator-routing.md` | [ ] |
| D — State Concurrent-Safety | `audit/state-management.md` | [ ] |
| E — Context Saturation | `audit/context-saturation.md` | [ ] |
| F — SDD Phase Graph | `audit/sdd-phase-graph.md` | [ ] |

---

## 0.5 Systemic Anti-Patterns Identified (Pre-Audit Hypothesis)

These will be validated or invalidated by the audit tasks above.

| Anti-Pattern | Location | Impact | Phase to Fix |
|---|---|---|---|
| Sequential step execution, no fan-out | `pipeline/runner.go` | High — install latency scales linearly with component count | Phase 1 |
| Serial I/O in skill registry collection | `cli/skill_registry.go` | Medium — cold start latency on Odoo projects with 30+ skills | Phase 2 |
| Double session-setup on SDD routing | `general-orchestrator.md` + `sdd-orchestrator.md` | Medium — 2× tool probe calls on every SDD session | Phase 3 |
| No routing fork — always General Orchestrator first | `general-orchestrator.md` | Medium — intent detection latency on every message | Phase 3 |
| Full SKILL.md injection per sub-agent | Sub-agent templates | High — token burn on large context windows | Phase 4 |
| No mutex on state R/W | `internal/state/`, `app.go` | Low now, Critical after Phase 1 | Phase 5 |
| Flat skill registry (full-file read) | `.atl/skill-registry.md` | Medium — no indexed lookup | Phase 2 |
| Sequential tool probes in orchestrators | Both orchestrators | Low-Medium — 3–4 sequential RPC calls on startup | Phase 3 |

---

## 0.6 Output Artifacts

- `audit/pipeline-sequential.md`
- `audit/skill-registry-io.md`
- `audit/orchestrator-routing.md`
- `audit/state-management.md`
- `audit/context-saturation.md`
- `audit/sdd-phase-graph.md`
- `audit/baseline-summary.md` — consolidated findings with severity ranking

---

## 0.7 Sub-Agent Delegation Map

```
[PHASE-0 ORCHESTRATOR]
    │
    ├── [A] rg-forensic-agent    → pipeline/runner.go sequential analysis
    ├── [B] rg-io-agent          → cli/skill_registry.go I/O mapping
    ├── [C] md-analysis-agent    → general-orchestrator.md + sdd-orchestrator.md routing
    ├── [D] rg-state-agent       → state R/W call sites, mutex scan
    ├── [E] token-count-agent    → sub-agent template expansion analysis
    └── [F] graph-agent          → SDD phase dependency validation
```

Tasks A, B, D can launch in parallel (all independent rg scans).  
Tasks C, E, F can launch in parallel after A/B (need doc content).  
`audit/baseline-summary.md` produces after all 6 complete.
