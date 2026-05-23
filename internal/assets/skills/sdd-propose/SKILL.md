---
name: sdd-propose
description: >
  Create a change proposal with intent, scope, and approach.
  Trigger: When the orchestrator launches you to create or update a proposal for a change.
license: MIT
metadata:
  author: rd-mg
  version: "2.0"
---

## Purpose

Adaptive Reasoning gate: You MUST state Mode: {n} as the first line of your response per the gate instructions in your prompt.

Create structured `proposal.md` inside change folder from exploration analysis or direct user input.

## Input

Orchestrator provides: change name, exploration analysis (from sdd-explore) or user description.

## Persistence

Follow `_shared/mode-branching.md`.

- **Artifact Name**: proposal.md
- **Topic Key**: sdd/{change-name}/proposal
- **Type**: architecture

Never force `openspec/` creation unless user requested file-based persistence or mode `hybrid`.

## Steps

### Step 0: Hypothesis Branching (Sequential Thinking)
**MANDATORY**: Call `sequential_thinking` with ≥2 branches (`branchId`) before committing.

### Step 1: Load Skills
Follow **Section A** from `skills/_shared/sdd-phase-common.md`.

### Step 2: Create Change Directory and Initial State

File-based persistence: create change folder structure:

```
openspec/changes/{change-name}/
├── proposal.md
└── state.yaml       (initial state)
```

**Write `state.yaml` initial skeleton** (atomic write: tmp + rename):
```yaml
schema_version: 1
change_name: {change-name}
created_at: {now_rfc3339}
updated_at: {now_rfc3339}
artifact_store: {mode}

phases:
  sdd-explore: { status: skipped }
  sdd-propose: { status: completed, completed_at: {now_rfc3339}, artifact: proposal.md }
  sdd-spec:    { status: pending, depends_on: [sdd-propose] }
  sdd-design:  { status: pending, depends_on: [sdd-spec] }
  sdd-tasks:   { status: pending, depends_on: [sdd-design] }
  sdd-apply:   { status: pending, depends_on: [sdd-tasks] }
  sdd-verify:  { status: pending, depends_on: [sdd-apply] }
  sdd-archive: { status: pending, depends_on: [sdd-verify] }
```

### Step 3: Read Existing Specs

If `openspec/specs/` has relevant specs, read them.

### Step 4: Write proposal.md

```markdown
# Proposal: {Change Title}

## Intent

{What problem? Why? Be specific about user need or technical debt.}

## Scope

### In Scope
- {Concrete deliverable 1}
- {Concrete deliverable 2}

### Out of Scope
- {What we're explicitly NOT doing}

## Capabilities

> CONTRACT between proposal and specs phases. sdd-spec reads this to know which spec files to create.
> Research `openspec/specs/` first.

### New Capabilities
<!-- Each becomes openspec/specs/<name>/spec.md. kebab-case names. Empty if none. -->
- `<capability-name>`: <brief description>

### Modified Capabilities
<!-- Existing capabilities whose REQUIREMENTS change. Delta spec per entry. -->
- `<existing-capability-name>`: <what requirement changes>

## Approach

{High-level technical approach. Reference exploration if available.}

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `path/to/area` | New/Modified/Removed | {What changes} |

## Pre-mortem (Risks)

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| {Risk} | Low/Med/High | {Mitigation} |

## Viability Score: {N}/10
**Rationale**: {1-2 sentences on current codebase readiness.}

## Rollback Plan

{How to revert. Be specific.}

## Dependencies

- {External dependency or prerequisite, if any}

## Success Criteria

- [ ] {Measurable outcome}
- [ ] {Measurable outcome}
```

### Step 5: Persist Artifact

**MANDATORY — do NOT skip.** Follow persistence rules in Step 2 of `_shared/mode-branching.md`.

### Step 6: Return Summary

```markdown
**Status**: success
**Summary**: Proposal created for `{change-name}`.
**Pre-mortem**: Viability {N}/10. Top risk: {risk} ({Likelihood}).
**Artifacts**: {artifact_path} | {topic_key}
**Next**: sdd-spec or sdd-design
```

## Rules

- In `openspec` mode, ALWAYS create `proposal.md`
- If change directory already has proposal, READ then UPDATE
- Keep proposal CONCISE
- Pre-mortem section: ≥2 risks
- Viability Score with rationale
- Rollback plan
- Success criteria
- Concrete file paths in "Affected Areas"
- Apply `rules.proposal` from `openspec/config.yaml`
- **ALWAYS fill Capabilities section** — contract with sdd-spec. Research `openspec/specs/` for correct names.
- New Capabilities → new `openspec/specs/<name>/spec.md`
- Modified Capabilities → delta spec in change folder
- Pure refactor/config change → explicitly write "None" under both sub-sections
- **Size budget**: ≤450 words. Bullet points and tables over prose.
- Return envelope per **Section D** from `skills/_shared/sdd-phase-common.md`.
