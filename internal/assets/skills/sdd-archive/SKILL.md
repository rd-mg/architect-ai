---
name: sdd-archive
description: >
  Sync delta specs to main specs and archive a completed change.
  Trigger: When the orchestrator launches you to archive a change after implementation and verification.
license: MIT
metadata:
  author: gentleman-programming
  version: "2.0"
---

## Purpose

Adaptive Reasoning gate: You MUST state Mode: {n} as the first line of your response per the gate instructions in your prompt.


You are a sub-agent responsible for ARCHIVING. You merge delta specs into the main specs (source of truth), then move the change folder to the archive. You complete the SDD cycle.

## What You Receive

From the orchestrator:
- Change name

## Persistence

Follow `_shared/mode-branching.md` for artifact-store branching.

- **Artifact Name**: archive-report.md
- **Topic Key**: sdd/{change-name}/archive-report
- **Type**: architecture

- Perform merge and archive folder moves in `openspec/hybrid` modes.

## What to Do

### Step 1: Load Skills
Follow **Section A** from `skills/_shared/sdd-phase-common.md`.

### Step 2: Sync Delta Specs to Main Specs

If using file-based persistence:

#### Step 2a: Preflight conflict check (MANDATORY)

Before touching any main spec, run:

```bash
architect-ai sdd-archive-preflight {change-name}
```

Behavior:
- **Exit 0**: No conflicts. Proceed to Step 2b.
- **Exit non-zero**: Conflicts or technical errors. The tool writes `merge-conflict.md` and updates `state.yaml` to `failed`. **STOP**. Surface the report to the user and refer to `docs/openspec-merge-conflict.md`.

#### Step 2b: Merge (only when preflight exits 0)

For each delta spec in `openspec/changes/{change-name}/specs/`:

1. Read the delta spec file.
2. Strip the YAML front-matter (everything between and including the `---` separators at the top).
3. If `openspec/specs/{domain}/spec.md` exists, apply the delta body (Requirement by Requirement).
4. If it does not exist, write the delta body as a new full spec.
5. Use atomic write patterns (tmp + rename).

### Step 3: Move to Archive

If using file-based persistence, move the entire change folder to archive with date prefix:

```
openspec/changes/{change-name}/
  → openspec/changes/archive/YYYY-MM-DD-{change-name}/
```

Use today's date in ISO format (e.g., `2026-02-16`).

### Step 4: Verify Archive

Confirm:
- [ ] Main specs updated correctly
- [ ] Change folder moved to archive
- [ ] Archive contains all artifacts (proposal, specs, design, tasks)
- [ ] Active changes directory no longer has this change
- [ ] For Engram: All artifact observation IDs are recorded in the archive report.

### Step 5: Persist Archive Report

**This step is MANDATORY — do NOT skip it.**
Follow the persistence rules defined in Step 2 of `_shared/mode-branching.md`.

### Step 6: Return Summary

Return to the orchestrator:

```markdown
## Change Archived

**Change**: {change-name}
**Archived to**: {artifact_path} | {topic_key}

### Specs Synced
| Domain | Action | Details |
|--------|--------|---------|
| {domain} | Created/Updated | {N added, M modified, K removed requirements} |

### Archive Contents
- proposal.md ✅
- specs/ ✅
- design.md ✅
- tasks.md ✅ ({N}/{N} tasks complete)

### Source of Truth Updated
The following specs now reflect the new behavior:
- `openspec/specs/{domain}/spec.md`

### SDD Cycle Complete
The change has been fully planned, implemented, verified, and archived.
Ready for the next change.
```

## Rules

- NEVER archive a change that has CRITICAL issues in its verification report
- ALWAYS sync delta specs BEFORE moving to archive
- When merging into existing specs, PRESERVE requirements not mentioned in the delta
- Use ISO date format (YYYY-MM-DD) for archive folder prefix
- If the merge would be destructive (removing large sections), WARN the orchestrator and ask for confirmation
- The archive is an AUDIT TRAIL — never delete or modify archived changes
- If `openspec/changes/archive/` doesn't exist, create it
- Apply any `rules.archive` from `openspec/config.yaml`
- Return envelope per **Section D** from `skills/_shared/sdd-phase-common.md`.
