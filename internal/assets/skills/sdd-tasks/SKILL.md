---
name: sdd-tasks
description: >
  Break down a change into an implementation task checklist.
  Trigger: When the orchestrator launches you to create or update the task breakdown for a change.
license: MIT
metadata:
  author: rd-mg
  version: "2.0"
---

## Cognitive Posture

**+++Pragmatic + +++Economic**. Ship smallest working change (Pragmatic) within budget (Economic).

If no budget declared, estimate from historical metering: `mem_search(query: "metering/{project}")`. Default: ≤10 tasks, ≤2000 tokens per task-description. Flag tasks exceeding share of budget.

## Purpose

Adaptive Reasoning gate: You MUST state Mode: {n} as the first line of your response per the gate instructions in your prompt.

Create `tasks.md` — concrete, actionable implementation steps from proposal, specs, and design. Organized by phase.

## Input

Orchestrator provides: change name.

## Persistence

Follow `_shared/mode-branching.md`.

- **Artifact Name**: tasks.md
- **Topic Key**: sdd/{change-name}/tasks
- **Type**: architecture

## Steps

### Step 1: Load Skills
Follow **Section A** from `skills/_shared/sdd-phase-common.md`.

### Step 2: Analyze Design

Identify:
- Files to create/modify/delete
- Dependency order
- Testing requirements per component

### Step 2b: Graphing Dependencies (MANDATORY)

Visualize implementation flow with Mermaid:

```mermaid
graph TD
  P1[Phase 1: Foundation] --> P2[Phase 2: Core]
  P2 --> P3[Phase 3: Integration]
  P3 --> P4[Phase 4: Testing]
```

Rules:
1. Use task IDs (1.1, 2.1) or Phase names as nodes
2. `-->` arrows = hard dependencies (X before Y)
3. Identify Critical Path

### Step 3: Write tasks.md

File-based structure:

```
openspec/changes/{change-name}/
├── proposal.md
├── specs/
├── design.md
└── tasks.md               ← You create this
```

#### Task File Format

```markdown
# Tasks: {Change Title}

## Dependency Graph

```mermaid
{Graph from Step 2b}
```

## Phase 1: {Phase Name} (Foundation)

- [ ] 1.1 {Concrete action — what file, what change}
- [ ] 1.2 {Concrete action}
- [ ] 1.3 {Concrete action}

## Phase 2: {Phase Name} (Core)

- [ ] 2.1 {Concrete action}
- [ ] 2.2 {Concrete action}
- [ ] 2.3 {Concrete action}

## Phase 3: {Phase Name} (Testing)

- [ ] 3.1 {Write tests for ...}
- [ ] 3.2 {Write tests for ...}

## Phase 4: {Phase Name} (Cleanup)

- [ ] 4.1 {Update docs/comments}
- [ ] 4.2 {Remove temporary code}
```

### Task Writing Rules

| Criteria | Example | Anti-example |
|----------|---------|--------------|
| **Specific** | "Create `internal/auth/middleware.go` with JWT validation" | "Add auth" |
| **Actionable** | "Add `ValidateToken()` method to `AuthService`" | "Handle tokens" |
| **Verifiable** | "Test: `POST /login` returns 401 without token" | "Make sure it works" |
| **Small** | One file or one logical unit | "Implement the feature" |

### Phase Organization

```
Phase 1: Foundation / Infrastructure
  └─ New types, interfaces, DB changes, config
  └─ Things other tasks depend on

Phase 2: Core Implementation
  └─ Main logic, business rules
  └─ The meat of the change

Phase 3: Integration / Wiring
  └─ Connect components, routes, UI

Phase 4: Testing
  └─ Unit, integration, e2e tests
  └─ Verify against spec scenarios

Phase 5: Cleanup
  └─ Documentation, dead code removal, polish
```

### Step 4: Persist Artifact

**MANDATORY — do NOT skip.** Follow persistence rules in Step 2 of `_shared/mode-branching.md`.

### Step 5: Return Summary

```markdown
## Tasks Created

**Change**: {change-name}
**Location**: {artifact_path} | {topic_key}

### Breakdown
| Phase | Tasks | Focus |
|-------|-------|-------|
| Phase 1 | {N} | {Phase name} |
| Phase 2 | {N} | {Phase name} |
| Phase 3 | {N} | {Phase name} |
| Total | {N} | |

### Implementation Order
{Brief description of recommended order and why}

### Next Step
Ready for implementation (sdd-apply).
```

## Rules

- ALWAYS include Mermaid `## Dependency Graph`
- ALWAYS reference concrete file paths in tasks
- Tasks ordered by dependency — Phase 1 tasks must not depend on Phase 2
- Testing tasks reference specific spec scenarios
- Each task completable in ONE session; split if too big
- Hierarchical numbering: 1.1, 1.2, 2.1, 2.2...
- NEVER vague tasks like "implement feature" or "add tests"
- Apply `rules.tasks` from `openspec/config.yaml`
- For TDD: RED task (write failing test) → GREEN task (make it pass) → REFACTOR task (clean up)
- **Size budget**: ≤530 words. Each task: 1-2 lines max. Checklist format, not paragraphs.
- Return envelope per **Section D** from `skills/_shared/sdd-phase-common.md`.
