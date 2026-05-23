---
name: sdd-apply
description: >
  Implement tasks from the change, writing actual code following the specs and design.
  Trigger: When the orchestrator launches you to implement one or more tasks from a change.
license: MIT
metadata:
  author: rd-mg
  version: "3.0"
---

## Purpose

Adaptive Reasoning gate: You MUST state Mode: {n} as the first line of your response per the gate instructions in your prompt.

Sub-agent for IMPLEMENTATION. Receives tasks from `tasks.md`, implements them writing code. Follows specs and design strictly.

## What You Receive

From orchestrator:
- Change name
- Specific task(s) to implement (e.g., "Phase 1, tasks 1.1-1.3")

## Persistence

Follow `_shared/mode-branching.md` for artifact-store branching.

- **Artifact Name**: apply-progress.md
- **Topic Key**: sdd/{change-name}/apply-progress
- **Type**: architecture

- Update `tasks.md` with `[x]` marks in `openspec/hybrid` modes.

## What to Do

### Step 1: Load Skills
Follow **Section A** from `skills/_shared/sdd-phase-common.md`.

### Step 2: Read Context

Before writing ANY code:
1. Read specs — understand WHAT code must do
2. Read design — understand HOW to structure code
3. Read existing code in affected files — understand current patterns
4. Check project's coding conventions from `config.yaml`

#### Step 2b: Read Previous Apply-Progress (Symmetric Resumption)

MUST check for prior progress before starting. Follow retrieval rules in Step 1 of `_shared/mode-branching.md`.

- **Action**: Parse retrieved progress, skip completed tasks, start from first incomplete task. If orchestrator provided specific progress text in prompt, use as immediate seed but still verify against store.

### Step 3: Read Testing Capabilities and Resolve Mode

```
Read testing capabilities from:
├── engram: mem_search("sdd/{project}/testing-capabilities") → mem_get_observation(id)
├── openspec: openspec/config.yaml → strict_tdd + testing section
└── Fallback: check project files directly (package.json, go.mod, etc.)

Resolve mode:
├── IF strict_tdd: true AND test runner exists
│   └── STRICT TDD MODE → Load strict-tdd.md
│       (read: skills/sdd-apply/strict-tdd.md)
│
├── IF strict_tdd: false OR no test runner
│   └── STANDARD MODE → use Step 4 (no TDD module loaded)
│
└── Cache resolved mode for return summary
```

**Key principle**: If Strict TDD Mode is not active, ZERO TDD instructions loaded. `strict-tdd.md` never read, never processed, never consumes tokens.

#### Hard Gate (Strict TDD Only)

If Strict TDD Mode active (orchestrator injection or self-discovery):
- MUST produce **TDD Cycle Evidence** table in apply-progress artifact
- Each task row MUST have: RED → GREEN → REFACTOR columns
- Task completed WITHOUT writing tests first → mark FAILED in evidence table
- Verify phase WILL reject work if TDD Evidence table missing or incomplete

**No silent fallback.** If resolved Strict TDD as active, follow it or report failure. Do NOT quietly switch to Standard Mode.

### Step 3b: TDD Prerequisite Lock (MANDATORY)

**HALT execution.** Before writing/modifying ANY implementation code, verify ALL:

1. TDD specifications defined: change's `spec.md` contains testable acceptance criteria and scenario tables.
2. Design approved: change's `design.md` exists, marked complete.
3. Tasks authorize implementation: current task(s) in `tasks.md` explicitly belong to implementation phase (e.g., "Phase 3 — Implementation").

**If any prerequisite missing:**
- STOP immediately.
- Do NOT write code.
- Return to `sdd-orchestrator` with: `PREREQUISITE MISSING: {which}. Cannot proceed with sdd-apply until resolved.`

### Step 4: Implement Tasks (Standard Workflow)

Used when Strict TDD Mode is NOT active:

```
FOR EACH TASK:
├── Read task description
├── Read relevant spec scenarios (acceptance criteria)
├── Read design decisions (constrain approach)
├── Read existing code patterns (match project style)
├── Write code
├── Mark task complete [x] in tasks.md
└── Note issues or deviations
```

### Step 5: Mark Tasks Complete

Update `tasks.md` — change `- [ ]` to `- [x]`:

```markdown
## Phase 1: Foundation

- [x] 1.1 Create `internal/auth/middleware.go` with JWT validation
- [x] 1.2 Add `AuthConfig` struct to `internal/config/config.go`
- [ ] 1.3 Add auth routes to `internal/server/server.go`  ← still pending
```

### Step 6: Persist Progress

**MANDATORY — do NOT skip.**
Follow persistence rules in Step 2 of `_shared/mode-branching.md`.

#### Merge Protocol (MANDATORY)
When saving `apply-progress`, MUST merge new work with ALL prior history. Do NOT overwrite.
1. **CUMULATIVE ARTIFACT**: Final `apply-progress` MUST include ALL previously completed tasks (copy status + evidence) PLUS new completions. Represents *current total state* of implementation.
2. **STORE SYNC**: Follow branching rules in `_shared/mode-branching.md`.
3. **TASK UPDATES**: Update `tasks.md` (openspec/hybrid) and tasks observation (engram/hybrid) with `[x]` marks for newly completed tasks.

### Step 7: Return Summary

Return to orchestrator:

```markdown
## Implementation Progress

**Change**: {change-name}
**Mode**: {Strict TDD | Standard}

### Completed Tasks
- [x] {task 1.1 description}
- [x] {task 1.2 description}

### Files Changed
| File | Action | What Was Done |
|------|--------|---------------|
| `path/to/file.ext` | Created | {brief description} |
| `path/to/other.ext` | Modified | {brief description} |

{IF Strict TDD Mode → include TDD Cycle Evidence table from strict-tdd.md}

### Deviations from Design
{List deviations from design.md and why. "None — implementation matches design." if none.}

### Issues Found
{Problems discovered during implementation. "None." if none.}

### Remaining Tasks
- [ ] {next task}
- [ ] {next task}

### Status
{N}/{total} tasks complete. {Ready for next batch / Ready for verify / Blocked by X}
```

## Rules

- ALWAYS read specs before implementing — specs are acceptance criteria
- ALWAYS follow design decisions — don't freelance
- ALWAYS match existing code patterns and conventions
- In `openspec` mode, mark tasks complete in `tasks.md` AS you go, not at end
- If design is wrong or incomplete, NOTE IT in return summary — don't silently deviate
- If task blocked by unexpected issue, STOP and report back
- NEVER implement unassigned tasks
- Skill loading handled in Step 1 — follow loaded skills strictly
- Apply any `rules.apply` from `openspec/config.yaml`
- If Strict TDD Mode active (Step 3), load `strict-tdd.md` and follow its cycle INSTEAD of Step 4
- When Strict TDD active, `strict-tdd.md` rules OVERRIDE Step 4 entirely
- Return envelope per **Section D** from `skills/_shared/sdd-phase-common.md`
