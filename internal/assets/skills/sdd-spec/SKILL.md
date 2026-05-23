---
name: sdd-spec
description: >
  Write specifications with requirements and scenarios (delta specs for changes).
  Trigger: When the orchestrator launches you to write or update specs for a change.
license: MIT
metadata:
  author: rd-mg
  version: "2.0"
---

## Purpose

Adaptive Reasoning gate: You MUST state Mode: {n} as the first line of your response per the gate instructions in your prompt.

Produce delta specs — requirements and scenarios describing what's ADDED, MODIFIED, or REMOVED from system behavior.

## Input

Orchestrator provides: change name.

## Persistence

Follow `_shared/mode-branching.md`.

- **Artifact Name**: spec.md
- **Topic Key**: sdd/{change-name}/spec
- **Type**: architecture

## Steps

### Step 1: Load Skills
Follow **Section A** from `skills/_shared/sdd-phase-common.md`.

### Step 2: Identify Affected Domains

Read proposal's **Capabilities section** — primary contract:

```
FOR EACH "New Capabilities" entry:
├── New FULL spec: openspec/specs/<capability-name>/spec.md
└── Write complete spec (not delta)

FOR EACH "Modified Capabilities" entry:
├── DELTA spec: openspec/changes/{change-name}/specs/<capability-name>/spec.md
└── Read existing openspec/specs/<capability-name>/spec.md first
```

Fallback: infer from "Affected Areas" if no Capabilities section.

### Step 3: Read Existing Specs

If `openspec/specs/{domain}/spec.md` exists, read it. Delta specs describe CHANGES.

### Step 3b: Failure Analysis (FMEA) — MANDATORY for Behavioral Changes

```markdown
FOR EACH requirement:
├── Identify failure mode
├── Determine effect
├── Rate Severity (1-5, 5 = data loss/crash)
└── Define Prevention/Mitigation
```

Pure refactor without behavioral impact → state "FMEA: Not applicable (Refactor)" in summary.

### Step 3c: UI State Analysis (FSM) — MANDATORY for Complex UI (>3 states)

```markdown
FSM:
├── States: Idle, Loading, Error, Success...
├── Triggers: user click, API response...
└── Transitions: Current + Trigger = Next
```

### Step 3d: Accessibility (WCAG) — MANDATORY for UI

```markdown
WCAG:
├── Keyboard: operable via keyboard only?
├── ARIA: proper roles and labels (aria-*)?
└── Focus: focus management handled?
```

### Step 4: Write Delta Specs

File-based persistence structure:

```
openspec/changes/{change-name}/
├── proposal.md
└── specs/
    └── {domain}/
        └── spec.md          ← Delta spec
```

#### MODIFIED Requirements Workflow (CRITICAL)

```
1. Locate requirement in openspec/specs/{domain}/spec.md
2. COPY ENTIRE requirement block — from "### Requirement:" through ALL scenarios
3. PASTE under "## MODIFIED Requirements"
4. EDIT copy to reflect new behavior
5. Add "(Previously: {one-line summary})" under requirement text

Why copy-full-then-edit?
→ Archive step REPLACES requirement in main specs with your MODIFIED block
→ Partial block = lost scenarios at archive time
→ Adding NEW behavior without changing existing → use ADDED, not MODIFIED
```

#### Delta Spec Format

```markdown
# Delta for {Domain}

## FMEA (Failure Mode and Effects Analysis)

| Failure Mode | Effect | Severity | Mitigation |
|--------------|--------|----------|------------|
| {mode}       | {eff}  | {1-5}    | {mit}      |

## FSM (UI States)
> Only for UI components with >3 states.

| Current State | Trigger | Next State | Action |
|---------------|---------|------------|--------|
| {state}       | {trig}  | {next}     | {act}  |

## Accessibility (WCAG)
> Only for UI changes.

- **Keyboard**: {How keyboard nav handled}
- **Screen Reader**: {ARIA roles and labels}
- **Focus**: {Focus management strategy}

## ADDED Requirements

### Requirement: {Name}

{Description using RFC 2119: MUST, SHALL, SHOULD, MAY}

The system {MUST/SHALL/SHOULD} {do something}.

#### Scenario: {Happy path}

- GIVEN {precondition}
- WHEN {action}
- THEN {expected outcome}

#### Scenario: {Sad path}

- GIVEN {error-triggering precondition}
- WHEN {action}
- THEN {expected error outcome}

#### Scenario: {Edge case}

- GIVEN {precondition}
- WHEN {action}
- THEN {expected outcome}

## MODIFIED Requirements

### Requirement: {Existing Name}

{Full updated requirement text — replaces existing entirely}
(Previously: {one-line summary})

#### Scenario: {Unchanged — keep if still valid}

- GIVEN {precondition}
- WHEN {action}
- THEN {outcome}

#### Scenario: {Updated or new}

- GIVEN {updated precondition}
- WHEN {updated action}
- THEN {updated outcome}

## REMOVED Requirements

### Requirement: {Name}

(Reason: {why deprecated/removed})
```

#### Delta front-matter (mandatory)

Every delta spec MUST begin with YAML front-matter. Compute `base_sha` as SHA-256 of `openspec/specs/{domain}/spec.md`, or `"0"` if new:

```yaml
---
openspec_delta:
  base_sha: "<SHA-256 of openspec/specs/{domain}/spec.md, or '0' if new>"
  base_path: "openspec/specs/{domain}/spec.md"
  base_captured_at: "<now RFC3339 UTC>"
  generator: sdd-spec
  generator_version: 1
---

# ...delta content below...
```

Do NOT omit any field. Archive step refuses merge without valid front-matter.

#### For NEW Specs (No Existing Spec)

Write FULL spec:

```markdown
# {Domain} Specification

## Purpose

Adaptive Reasoning gate: You MUST state Mode: {n} as the first line of your response per the gate instructions in your prompt.

{High-level domain description.}

## Requirements

### Requirement: {Name}

The system {MUST/SHALL/SHOULD} {behavior}.

#### Scenario: {Name}

- GIVEN {precondition}
- WHEN {action}
- THEN {outcome}
```

### Step 5: Persist Artifact

**MANDATORY — do NOT skip.** Follow persistence rules in Step 2 of `_shared/mode-branching.md`.

### Step 6: Return Summary

```markdown
## Specs Created

**Change**: {change-name}

### Specs Written
| Domain | Type | Requirements | Scenarios |
|--------|------|-------------|-----------|
| {domain} | Delta/New | {N added, M modified, K removed} | {total} |

### Coverage
- Happy paths: {covered/missing}
- Edge cases: {covered/missing}
- Error states: {covered/missing}

### Next Step
Ready for design (sdd-design). If design exists, ready for tasks (sdd-tasks).
```

## Rules

- ALWAYS Given/When/Then format for scenarios
- ALWAYS RFC 2119 keywords (MUST, SHALL, SHOULD, MAY)
- Read proposal's **Capabilities section** first — tells exactly which spec files to create
- If existing specs exist → DELTA specs (ADDED/MODIFIED/REMOVED)
- If no existing specs → FULL spec
- Every requirement: ≥1 Happy Path scenario
- Every behavioral requirement: ≥1 Sad Path scenario
- Every delta with behavioral changes: FMEA table
- Every UI spec: Accessibility (WCAG) section
- Every complex UI (>3 states): FSM table
- Scenarios must be TESTABLE
- NO implementation details — specs describe WHAT, not HOW
- **MODIFIED requirements: FULL block** — copy entire requirement + all scenarios, then edit. Partial blocks lose content at archive.
- Adding new behavior without changing existing → use ADDED, not MODIFIED
- Apply `rules.specs` from `openspec/config.yaml`
- **Size budget**: ≤650 words. Requirement tables over narrative. Each scenario: 3-5 lines max.
- Return envelope per **Section D** from `skills/_shared/sdd-phase-common.md`.

## RFC 2119 Keywords Quick Reference

| Keyword | Meaning |
|---------|---------|
| **MUST / SHALL** | Absolute requirement |
| **MUST NOT / SHALL NOT** | Absolute prohibition |
| **SHOULD** | Recommended, exceptions possible with justification |
| **SHOULD NOT** | Not recommended, may be acceptable with justification |
| **MAY** | Optional |
