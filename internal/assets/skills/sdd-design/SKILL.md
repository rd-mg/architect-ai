---
name: sdd-design
description: >
  Create technical design document with architecture decisions and approach.
  Trigger: When the orchestrator launches you to write or update the technical design for a change.
license: MIT
metadata:
  author: rd-mg
  version: "2.0"
---

## Purpose

Adaptive Reasoning gate: You MUST state Mode: {n} as the first line of your response per the gate instructions in your prompt.

Sub-agent for TECHNICAL DESIGN. Takes proposal and specs, produces `design.md` capturing HOW change will be implemented — architecture decisions, data flow, file changes, technical rationale.

## What You Receive

From orchestrator:
- Change name

## Persistence

Follow `_shared/mode-branching.md` for artifact-store branching.

- **Artifact Name**: design.md
- **Topic Key**: sdd/{change-name}/design
- **Type**: architecture

## What to Do

### Step 1: Load Skills
Follow **Section A** from `skills/_shared/sdd-phase-common.md`.

### Step 2: Read Codebase

Before designing, read actual code affected:
- Entry points and module structure
- Existing patterns and conventions
- Dependencies and interfaces
- Test infrastructure (if any)

### Step 3: Poka-Yoke Analysis (MANDATORY)

Defensive design analysis preventing common implementation errors.

```markdown
FOR EACH architectural decision:
├── Identify potential mistake modes.
├── How does design prevent those mistakes?
└── List specific checklist items for implementer.
```

### Step 4: Write design.md

If file-based persistence:

```
openspec/changes/{change-name}/
├── proposal.md
├── specs/
└── design.md              ← You create this
```

#### Design Document Format

```markdown
# Design: {Change Title}

## Technical Approach

{Concise description of overall technical strategy.}

## Poka-Yoke Checklist (Mistake-Proofing)

- [ ] **State Invalidation**: {How prevented}
- [ ] **Dependency Loop**: {How avoided}
- [ ] **Resource Leak**: {How handled}
- [ ] **Boundary Errors**: {How validated}

## Architecture Decisions

### Decision: {Decision Title}

**Choice**: {What we chose}
**Alternatives considered**: {What we rejected}
**Rationale**: {Why this choice over alternatives}

## Data Flow

{How data moves through system for this change.}

    Component A ──→ Component B ──→ Component C
         │                              │
         └──────── Store ───────────────┘

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `path/to/new-file.ext` | Create | {What this file does} |
| `path/to/existing.ext` | Modify | {What changes and why} |
| `path/to/old-file.ext` | Delete | {Why removed} |

## Interfaces / Contracts

{Define new interfaces, API contracts, type definitions, data structures.
Use code blocks with project's language.}

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | {What} | {How} |
| Integration | {What} | {How} |
| E2E | {What} | {How} |

## Migration / Rollout

{If change requires data migration, feature flags, or phased rollout, describe plan.
If not applicable, state "No migration required."}

## Open Questions

- [ ] {Unresolved technical question}
- [ ] {Decision needing team input}
```

### Step 4: Persist Artifact

**MANDATORY — do NOT skip.**
Follow persistence rules in Step 2 of `_shared/mode-branching.md`.

### Step 5: Return Summary

Return to orchestrator:

```markdown
## Design Created

**Change**: {change-name}
**Location**: {artifact_path} | {topic_key}

### Summary
- **Approach**: {one-line technical approach}
- **Key Decisions**: {N decisions documented}
- **Files Affected**: {N new, M modified, K deleted}
- **Testing Strategy**: {unit/integration/e2e coverage planned}

### Open Questions
{List unresolved questions, or "None"}

### Next Step
Ready for tasks (sdd-tasks).
```

## Rules

- ALWAYS include `## Poka-Yoke Checklist` section
- ALWAYS read actual codebase before designing — never guess
- Every decision MUST have rationale (the "why")
- Include concrete file paths, not abstract descriptions
- Use project's ACTUAL patterns and conventions, not generic best practices
- If codebase uses pattern different from recommendation, note it but FOLLOW existing pattern unless change specifically addresses it
- Keep ASCII diagrams simple — clarity over beauty
- Apply any `rules.design` from `openspec/config.yaml`
- If open questions BLOCK design, say so clearly — don't guess
- **Size budget**: Design artifact MUST be under 800 words. Architecture decisions as tables (option | tradeoff | decision). Code snippets only for non-obvious patterns.
- Return envelope per **Section D** from `skills/_shared/sdd-phase-common.md`
