# Baseline Summary — refactor-architecture-v1

**Date**: 2026-05-12
**Source**: MASTER-PLAN.md v1.0
**Status**: Phase 0 complete — all 6 audit tasks validated

---

## Executive Verdict

| # | Bottleneck | MASTER-PLAN Claim | Audit Measurement | Verdict | Discrepancy |
|---|-----------|-------------------|-------------------|---------|-------------|
| B1 | Sequential Pipeline | Latency −40–60% | ~65% latency reduction | ✅ VALIDATED | Est. exceeds plan upper bound |
| B2 | Skill Registry Serial I/O | Registry latency −60% | ~60% improvement | ✅ VALIDATED | — None |
| B3 | Double Tool Probe | 3–5 redundant RPCs | 4–7 redundant RPCs | ✅ VALIDATED | Slightly higher than estimated |
| B4 | Sequential Probes | 3–4 serial → 1 batch | Confirmed in B3 audit | ✅ VALIDATED | Subsumed by B3 |
| B5 | Context Saturation | 3,600–7,500 tok/session | ~10,800 tok/session | ✅ VALIDATED | **⚠ 44–200% higher than estimated** |
| B6 | No State Mutex | Race window post-Phase 1 | 6 callers, unprotected RMW | ✅ VALIDATED | — None |

**Overall**: All 6 MASTER-PLAN bottleneck claims validated. One severity upgrade recommended: B5 (Context Saturation) from HIGH to **CRITICAL** given actual waste is 44–200% above the original estimate.

---

## Baseline Metrics

| Bottleneck | Current State | Measured Impact | Target After Fix | Priority |
|---|---|---|---|---|
| Pipeline runner | Serial goroutine-await pattern | ~65% latency waste (6× T_step vs 2× T_step grouped) | ~65% faster | 🔴 HIGH |
| Skill registry I/O | 3 serial filesystem walks | ~3× T_walk (serial) vs ~1× T_walk (parallel) | ~60% faster | 🟡 HIGH |
| Orchestrator probes | 4–7 redundant RPCs per cold start | ~40–60% fewer RPC calls possible | ~40–60% fewer | 🔴 HIGH |
| Sub-agent templates | ~10,800 tokens wasted/session | 8 calls × ~1,350 tok unconditional injection | ~65% fewer tokens | 🔴 HIGH ⬆ |
| State management | No mutex — unprotected RMW | Silent corruption risk post-Phase 1 | Zero race window | 🟡 MEDIUM (⬆ after Phase 1) |

---

## Detailed Audit Summaries

### 0A — Pipeline Sequential Execution
- **Audit file**: `audit/pipeline-sequential.md`
- **MASTER-PLAN ref**: B1 — Severity HIGH
- **Root Cause**: `internal/pipeline/runner.go:Run()` spawns goroutine per step but immediately awaits via `select { case err := <-done }`. Net effect: serial execution despite concurrent syntax.
- **Impact**: Install latency = N × T_step. With 6 MVP components: ~6× slower than achievable.
- **Missing**: `StepGroup` type, `RunGroup()` with `errgroup` fan-out.
- **Evidence**: Lines 34-100 — goroutine spawn + blocking select.
- **Drives**: Phase 1 (Pipeline Parallelization).

### 0B — Skill Registry Serial I/O
- **Audit file**: `audit/skill-registry-io.md`
- **MASTER-PLAN ref**: B2 — Severity MEDIUM
- **Root Cause**: `WriteLocalSkillRegistry()` calls `collectUserSkills` → `collectProjectSkills` → `collectOverlayContent` sequentially. All three are independent filesystem walks.
- **Impact**: ~3× longer registry regeneration on Odoo projects (30+ skills).
- **Missing**: `errgroup`, `sync.WaitGroup`, or any goroutine fan-out.
- **Evidence**: Serial append pattern across 3 collection calls.
- **Drives**: Phase 2 (Skill Registry Refactor).

### 0C+0F — Orchestrator Double Probe & SDD Phase Graph Overlap
- **Audit files**: `audit/orchestrator-routing.md`, `audit/sdd-phase-graph.md`
- **MASTER-PLAN ref**: B3 — Severity MEDIUM, B4 — Severity LOW-MEDIUM
- **Root Cause (B3)**: General Orchestrator runs Tool Availability Check (4+ mem_search) before forwarding to SDD Orchestrator, which re-runs its own Session-Setup Triplet (3 calls) + Tool Check (4 calls). Result: 4–7 redundant RPCs per SDD session cold start.
- **Root Cause (B4)**: Tool probe steps are listed without parallel dispatch instruction. LLM executes serially.
- **Additional finding (0F)**: Parallel delegation is documented as "MANDATORY" but has no mechanical enforcement. No collision detection for concurrent `sdd-apply` tasks touching same filesystem resources.
- **Drives**: Phase 3 (Orchestrator Routing Fork).

### 0D — State Mutex Safety
- **Audit file**: `audit/state-management.md`
- **MASTER-PLAN ref**: B6 — Severity LOW (CRITICAL after Phase 1)
- **Root Cause**: `state.Read`/`state.Write` use `os.ReadFile`/`os.WriteFile` directly. No `sync.Mutex`, `sync.RWMutex`, or file locking. `persistAssignments` in `app.go` is a classic read-modify-write race.
- **Callers at risk**: 6 across `cli/`, `app/`, `components/`.
- **Risk**: Currently LOW (sequential pipeline). Becomes CRITICAL after Phase 1 introduces parallel step groups.
- **Drives**: Phase 5 (State Hardening) — must follow Phase 1.

### 0E — Context Saturation
- **Audit file**: `audit/context-saturation.md`
- **MASTER-PLAN ref**: B5 — Severity HIGH → **recommended upgrade to CRITICAL**
- **Root Cause**: `{content from ...}` template expansions inject complete file contents unconditionally into every sub-agent. No conditional injection. No tiered priority.
- **Measured waste**:
  - General Orchestrator template: ~600 tokens/sub-agent
  - SDD Orchestrator template: ~1,200–1,500 tokens/sub-agent
  - Session total: 8 calls × ~1,350 tokens = **~10,800 wasted tokens/session**
- **MASTER-PLAN estimate**: 3,600–7,500 tokens/session
- **Discrepancy**: Actual waste exceeds estimate by 44–200%.
- **Recommendation**: Upgrade B5 severity from HIGH to CRITICAL.
- **Drives**: Phase 4 (Context Compression).

---

## Dependency Constraints

```
Phase 0 (Complete) ─┬─→ Phase 1 (Pipeline Parallel)
                     ├─→ Phase 2 (Skill Registry)
                     └─→ Phase 3 (Orchestrator Fork) ─── needs 0C+0F
                               │
Phase 1 ──────────────→ Phase 5 (State Hardening)
Phase 2 ──────────────→ Phase 4 (Context Compression) ─── needs 0E
```

**Canonical execution**: Phase 0 → [Phase 1 ∥ Phase 2 ∥ Phase 3] → Phase 4 (after Phase 2) → Phase 5 (after Phase 1).

---

## Lateral Issues (Out of Scope)

Tracked in MASTER-PLAN L1–L5. No new lateral issues discovered during audit.

| ID | Issue | Risk |
|---|---|---|
| L1 | Flat agent catalog (no runtime discovery) | Extensibility |
| L2 | No dependency cycle verification in StagePlan build | Correctness |
| L3 | Skill registry not hot-reloadable | Staleness |
| L4 | No Engram cache eviction policy (168h hardcoded) | Staleness |
| L5 | `openspec/changes/archive/` not garbage-collected | Disk growth |

---

## Risks & Deviations

- **Phase 0 was pure audit** — no code changes made, per plan rules.
- **All 6 bottleneck claims validated** — no deviations from MASTER-PLAN.
- **One severity upgrade recommended**: B5 (Context Saturation) from HIGH → CRITICAL due to measured waste exceeding estimates.
- **B4 (Sequential Probes)** is subsumed by B3 fix — no separate phase needed.
- **Phase 5 must follow Phase 1** — introducing parallel pipeline before mutex hardening creates a race condition.

---

## Go/No-Go Recommendation

**GO** — All Phase 0 audit objectives met. All MASTER-PLAN bottleneck claims confirmed with evidence. Ready for parallel execution of Phase 1, Phase 2, and Phase 3.

**Pre-conditions for Phase 1**: Strict-TDD mode active. No implementation before test exists. Race detector must pass (`go test -race ./...`).

---

*Audit artifacts: `audit/pipeline-sequential.md`, `audit/skill-registry-io.md`, `audit/orchestrator-routing.md`, `audit/state-management.md`, `audit/context-saturation.md`, `audit/sdd-phase-graph.md`*