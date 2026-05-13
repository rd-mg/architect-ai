# Architect-AI Refactoring Master Plan
## Parallelized, Tool-Native Multi-Agent Architecture

> **Lead Role**: Lead AI Architect, Context Engineer, Senior Go Developer  
> **Cognitive Mode**: +++Systemic +++Forensic +++Empirical  
> **CCLD Tag**: `[MASTER-PLAN][ARCHITECT-AI][MULTI-AGENT][PARALLEL]`  
> **Version**: 1.0 — 2026-05-12  
> **Repository**: `architect-ai`

---

## Executive Summary

`architect-ai` is a Go TUI/CLI installer for AI agent environments. The system has two concurrent execution layers:

- **Runtime layer (Go)**: `pipeline/runner.go` executes N install steps **sequentially** even when steps are independent. `cli/skill_registry.go` performs 3 I/O collection passes **serially**. `internal/state` has no mutex protection.
- **Orchestration layer (Markdown prompts)**: The General Orchestrator runs its own Tool Availability Check before forwarding SDD intents — creating a **double probe** overhead. Parallel delegation is documented as "MANDATORY" but has no mechanical enforcement. Sub-agent templates inject full file content, burning 600–940 tokens of redundant context per agent call.

This plan eliminates all identified bottlenecks across 6 phases, prioritized by impact-to-effort ratio.

---

## System Architecture Map

```
┌─────────────────────────────────────────────────────────────────┐
│                      USER (IDE / CLI)                           │
└─────────────────────┬───────────────────────────────────────────┘
                       │
┌─────────────────────▼───────────────────────────────────────────┐
│              ORCHESTRATION LAYER (Markdown Prompts)             │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  General Orchestrator (.agent/skills/_shared/)          │   │
│  │  ├── ROUTER GATE ←── [Phase 3: add early-exit fork]    │   │
│  │  ├── Tool Probe  ←── [Phase 3: parallelize probes]     │   │
│  │  ├── Routing Table (Solver/Ideator/Researcher/General)  │   │
│  │  └── Sub-Agent Template ←── [Phase 4: tiered injection] │   │
│  └─────────────────┬───────────────────────────────────────┘   │
│                    │ SDD intent → direct forward                │
│  ┌─────────────────▼───────────────────────────────────────┐   │
│  │  SDD Orchestrator (.agent/skills/_shared/)              │   │
│  │  ├── Session-Setup Triplet ←── [Phase 3: dedup probes] │   │
│  │  ├── Phase Dependency Graph                             │   │
│  │  ├── Parallel Dispatch Table ←── [Phase 3: enforce]    │   │
│  │  └── Protocol Progressive Loading ←── [Phase 4: guard] │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
│  Skill Registry (.atl/skill-registry.md)                       │
│  └── Flat format ←── [Phase 2: indexed with Quick Index]       │
└─────────────────────┬───────────────────────────────────────────┘
                       │ CLI commands (architect-ai install/sync/update)
┌─────────────────────▼───────────────────────────────────────────┐
│                 RUNTIME LAYER (Go Binary)                       │
│                                                                 │
│  cmd/architect-ai/main.go                                       │
│  └── internal/app/app.go (RunArgs → tuiExecute → Orchestrator) │
│                           ↑                                     │
│  internal/planner/                                              │
│  ├── resolver.go    (TopologicalSort → OrderedComponents)       │
│  ├── graph.go       (MVPGraph: Engram→SDD→Skills dependency)    │
│  └── order.go       (applySoftOrdering: Persona before Engram)  │
│                           ↑                                     │
│  internal/pipeline/                                             │
│  ├── orchestrator.go   (Execute: Prepare → Apply → Rollback)    │
│  ├── runner.go         ←── [Phase 1: SEQUENTIAL → PARALLEL]    │
│  ├── stages.go         ←── [Phase 1: add StepGroup type]       │
│  └── rollback.go       (reverse-order rollback)                 │
│                           ↑                                     │
│  internal/cli/                                                  │
│  ├── install.go        (BuildRealStagePlan → StagePlan)         │
│  └── skill_registry.go ←── [Phase 2: SERIAL → PARALLEL I/O]   │
│                           ↑                                     │
│  internal/state/        ←── [Phase 5: add mutex Manager]       │
│  internal/catalog/      (agents.go, components.go, skills.go)   │
│  internal/backup/       ←── [Phase 5: context-aware snapshot]  │
└─────────────────────────────────────────────────────────────────┘
```

---

## Phase Overview

| Phase | Name | Layer | Priority | Impact | Effort | Depends On |
|---|---|---|---|---|---|---|
| 0 | Discovery & Baseline Audit | Both | 🔴 CRITICAL | Baseline | Low | — |
| 1 | Pipeline Parallelization | Go | 🔴 HIGH | Latency −40–60% | Medium | Phase 0A |
| 2 | Skill Registry Refactor | Go + MD | 🟡 HIGH | Registry latency −60% | Low-Med | Phase 0B |
| 3 | Orchestrator Routing Fork | MD Prompts | 🔴 HIGH | SDD overhead −40% | Medium | Phase 0C+F |
| 4 | Context Compression | MD Prompts | 🟡 HIGH | Token burn −65% | Low-Med | Phase 0E, 2 |
| 5 | State Hardening | Go | 🟡 MEDIUM | Safety (post Phase 1) | Low | Phase 1, 0D |

---

## Dependency Graph

```
Phase 0 (Audit)
    ├── Task A ──────────────────────────────→ Phase 1 (Pipeline Parallel)
    ├── Task B ──────────────────────────────→ Phase 2 (Skill Registry)
    ├── Task C + F ──────────────────────────→ Phase 3 (Orchestrator Fork)
    ├── Task E ──────────────────────────────→ Phase 4 (Context Compression)
    └── Task D ──────────────────────────────→ Phase 5 (State Hardening)
                                                   ↑
Phase 1 ─────────────────────────────────────────────
Phase 2 ──────────────────────────────────→ Phase 4
```

**Execution Order**: Phase 0 → [Phase 1, Phase 2, Phase 3] in parallel → Phase 4 (after 2) → Phase 5 (after 1).

---

## Critical Bottlenecks Found

### B1 — Sequential Pipeline Runner [SEVERITY: HIGH]
- **File**: `internal/pipeline/runner.go:Run()`
- **Root Cause**: `for _, step := range steps { select { case err := <-done: ... } }` — goroutine spawned per step but awaited immediately. Net effect: serial.
- **Impact**: Install time = N × T_step. With 6 MVP components: ~6× slower than possible.
- **Fix**: Phase 1 — `StepGroup` type + `RunGroup()` with errgroup fan-out.

### B2 — Serial I/O in Skill Registry [SEVERITY: MEDIUM]
- **File**: `internal/cli/skill_registry.go:WriteLocalSkillRegistry()`
- **Root Cause**: `collectUserSkills` → `collectProjectSkills` → `collectOverlayContent` called sequentially; all are independent filesystem walks.
- **Impact**: 3× longer registry regeneration on Odoo projects with 30+ skills.
- **Fix**: Phase 2 — errgroup concurrent fan-out; indexed registry format.

### B3 — Double Tool Probe on SDD Path [SEVERITY: MEDIUM]
- **Files**: `general-orchestrator.md`, `sdd-orchestrator.md`
- **Root Cause**: General Orchestrator runs Tool Availability Check (3× mem_search) before forwarding to SDD Orchestrator, which runs its own Session-Setup Triplet (3× mem_search).
- **Impact**: 3–5 redundant RPC calls per SDD session start.
- **Fix**: Phase 3 — Router Gate early-exit; forwarded session state.

### B4 — Sequential Tool Probes [SEVERITY: LOW-MEDIUM]
- **Files**: Both orchestrators
- **Root Cause**: Tool probe steps are listed without parallel dispatch instruction. LLM executes serially.
- **Impact**: 3–4 sequential RPC calls instead of 1 parallel batch.
- **Fix**: Phase 3 — explicit parallel probe dispatch instruction.

### B5 — Context Window Saturation [SEVERITY: HIGH]
- **Files**: Sub-agent templates in both orchestrators
- **Root Cause**: `{content from ...}` expansions inject full files into every sub-agent. No conditional injection. No tiered priority.
- **Impact**: 600–940 tokens overhead per sub-agent × 6–8 agents = 3,600–7,500 wasted tokens/session.
- **Fix**: Phase 4 — tiered injection; targeted compact rule extraction.

### B6 — No State Mutex [SEVERITY: LOW (CRITICAL after Phase 1)]
- **Files**: `internal/state/`, `internal/app/app.go`
- **Root Cause**: `state.Read` / `state.Write` have no mutex. Read-modify-write pattern vulnerable to race after parallel pipeline.
- **Impact**: Silent state corruption on concurrent TUI actions post-Phase 1.
- **Fix**: Phase 5 — `state.Manager` with `sync.Mutex`; `Merge` atomic operation.

---

## Lateral Systemic Issues (Beyond Initial Scope)

These were discovered during the architectural investigation and are tracked here for future phases:

### L1 — Flat Agent Catalog (no runtime discovery)
`internal/catalog/agents.go` is a hardcoded `[]Agent` slice. Adding a new agent requires a Go recompile. Future: make catalog extensible from a YAML config file.

### L2 — No Dependency Cycle Detection at StagePlan Build Time
`planner/graph.go:TopologicalSort` detects cycles in the component graph, but `BuildRealStagePlan` doesn't verify that `StepGroup` ordering respects the topological sort. A programmer error could create a plan where group[1] depends on group[2]. Add a verification step in `NewOrchestrator`.

### L3 — Skill Registry Not Hot-Reloadable
`.atl/skill-registry.md` is read once per session. If a skill is added/updated mid-session, the orchestrator uses stale compact rules. Future: add file watcher + session invalidation signal.

### L4 — No Engram Cache Eviction Policy
The `research-routing.md` 168h (7-day) cache TTL is hardcoded in the prompt. No mechanism to invalidate Engram entries when underlying files change (e.g., after `architect-ai sync` updates a SKILL.md). Future: add a `sync` hook that writes an invalidation marker to Engram.

### L5 — `openspec/changes/` Not Garbage-Collected
Archived changes in `openspec/changes/archive/` accumulate indefinitely. The `sdd-archive-preflight` CLI command exists but no automatic retention policy. Future: add `--max-archive-age` flag to `cleanup` command.

---

## Refactoring Rules (Enforced Across All Phases)

1. **No behavior changes** in Phase 0 (audit only).
2. **Backward compatibility**: All public Go APIs maintain their signatures via wrapper functions during transition. Breaking changes require a new function name + deprecation comment.
3. **Race detector clean**: Every modified Go package must pass `go test -race ./...` before merge.
4. **TDD gate**: No implementation code written before test exists (strict-TDD mode implied by project standards).
5. **Sync everything**: Any prompt change to `_shared/` must be synced to all `internal/assets/*/` variants via `architect-ai sync`.
6. **Sub-agent delegation**: Orchestrator does not execute domain-specific tasks inline. All write operations delegated to specialized sub-agents.
7. **CCLD in handoffs**: All sub-agent prompts use CCLD for internal reasoning sections.

---

## Files Master Index

### Phase 0 — Audit Output
- `audit/pipeline-sequential.md`
- `audit/skill-registry-io.md`
- `audit/orchestrator-routing.md`
- `audit/state-management.md`
- `audit/context-saturation.md`
- `audit/sdd-phase-graph.md`
- `audit/baseline-summary.md`

### Phase 1 — Go Pipeline
- `internal/pipeline/stages.go` (MODIFY)
- `internal/pipeline/runner.go` (MODIFY)
- `internal/pipeline/orchestrator.go` (MODIFY)
- `internal/pipeline/runner_test.go` (CREATE)
- `internal/cli/install.go` (MODIFY)

### Phase 2 — Skill Registry
- `internal/cli/skill_registry.go` (MODIFY)
- `internal/cli/skill_registry_test.go` (MODIFY)
- `.atl/skill-registry.md` (REGENERATE)
- `.agent/skills/_shared/skill-resolver.md` (MODIFY)

### Phase 3 — Orchestrator Routing
- `.agent/skills/_shared/general-orchestrator.md` (MODIFY)
- `.agent/skills/_shared/sdd-orchestrator.md` (MODIFY)
- `internal/assets/*/general-orchestrator.md` (SYNC ALL)
- `internal/assets/*/sdd-orchestrator.md` (SYNC ALL)

### Phase 4 — Context Compression
- `.agent/skills/_shared/general-orchestrator.md` (MODIFY)
- `.agent/skills/_shared/sdd-orchestrator.md` (MODIFY)
- `.agent/skills/_shared/sdd-phase-common.md` (MODIFY)
- `internal/assets/*/general-orchestrator.md` (SYNC ALL)

### Phase 5 — State Hardening
- `internal/state/manager.go` (CREATE)
- `internal/state/manager_test.go` (CREATE)
- `internal/app/app.go` (MODIFY)
- `internal/cli/install.go` (MODIFY)

---

## Total Expected Impact

| Metric | Current | After All Phases | Improvement |
|---|---|---|---|
| Install latency (6 components) | ~6 × T_step | ~2 × T_step (grouped) | ~65% faster |
| Skill registry regeneration (Odoo) | ~3 × T_walk | ~1 × T_walk (parallel) | ~60% faster |
| SDD session cold-start probes | 6–8 RPC calls | 2–3 RPC calls | ~60% fewer |
| Sub-agent context overhead | 600–940 tokens | 215–285 tokens | ~65% fewer |
| State race window | Exists post-Phase 1 | Zero | Safe |
| Test coverage on pipeline | Partial | Full with race detector | +Race safe |

---

## Phase Plan File Index

| File | Phase |
|---|---|
| `phase-0-discovery-audit.md` | Discovery & Baseline Audit |
| `phase-1-pipeline-parallelization.md` | Go Pipeline Parallelization |
| `phase-2-skill-registry-refactor.md` | Skill Registry Kernel Refactor |
| `phase-3-orchestrator-routing-fork.md` | Orchestrator Routing Fork & Parallel Probes |
| `phase-4-context-compression.md` | Context Compression & Injection Optimization |
| `phase-5-state-hardening.md` | State Management Hardening |
