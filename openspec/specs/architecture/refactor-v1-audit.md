# Phase 0: Discovery & Baseline Audit

## Objective
Establish baseline measurements for all identified bottlenecks before any code changes. Pure audit — NO code modifications permitted.

## Tasks

### 0A — Audit Pipeline Sequential Execution
**Files**: `internal/pipeline/runner.go`, `internal/pipeline/stages.go`, `internal/pipeline/orchestrator.go`
**Output**: `audit/pipeline-sequential.md`
- Verify: `runner.go:Run()` spawns goroutines per step but awaits sequentially
- Verify: No `StepGroup` or `RunGroup` parallelism exists  
- Count: N steps in typical install flow
- Measure: Sequential vs potential parallel time (theoretical)

### 0B — Audit Skill Registry Serial I/O
**Files**: `internal/cli/skill_registry.go`, `.atl/skill-registry.md`, `.agent/skills/_shared/skill-resolver.md`
**Output**: `audit/skill-registry-io.md`
- Verify: `WriteLocalSkillRegistry` calls `collectUserSkills` → `collectProjectSkills` → `collectOverlayContent` serially
- Verify: No errgroup or concurrent fs walk
- Count: Total skills + overlay entries in typical project
- Measure: Serial vs parallel time estimate

### 0C — Audit General Orchestrator Routing
**Files**: `.agent/skills/_shared/general-orchestrator.md`
**Output**: `audit/orchestrator-routing.md`
- Verify: Tool Availability Check runs before SDD forwarding (double probe with SDD orchestrator)
- Verify: Routing Table dispatch flow
- Count: RPC calls per cold start
- Identify: Early-exit fork opportunity

### 0D — Audit State Management
**Files**: `internal/state/`, `internal/app/app.go`
**Output**: `audit/state-management.md`
- Verify: No mutex/sync primitive protecting `state.Read`/`state.Write`
- Verify: Read-modify-write vulnerability
- Identify: All callers of state package
- Assess: Race risk after Phase 1 parallelization

### 0E — Audit Context Saturation
**Files**: `.agent/skills/_shared/general-orchestrator.md`, `.agent/skills/_shared/sdd-orchestrator.md`
**Output**: `audit/context-saturation.md`
- Verify: `{content from ...}` expansions in sub-agent templates
- Count: Token overhead per expansion (estimate)
- Measure: Total waste across 6-8 sub-agent calls

### 0F — Audit SDD Orchestrator Phase Graph
**Files**: `.agent/skills/_shared/sdd-orchestrator.md`
**Output**: `audit/sdd-phase-graph.md`
- Verify: Session-Setup Triplet probe overlap with General Orchestrator
- Verify: Phase Dependency Graph correctness
- Identify: Parallel dispatch opportunities not enforced
- Map: Full probe sequence from cold start to first delegation

### 0S — Baseline Summary
**Depends On**: 0A, 0B, 0C, 0D, 0E, 0F
**Output**: `audit/baseline-summary.md`
- Aggregate all audit findings
- Establish baseline metrics per MASTER-PLAN table
- Flag any risks or deviations from plan
