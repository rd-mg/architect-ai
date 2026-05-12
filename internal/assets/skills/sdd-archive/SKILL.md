---
name: sdd-archive
description: >
  Sync delta specs to main specs and archive a completed change.
  Trigger: When the orchestrator launches you to archive a change after implementation and verification.
license: MIT
metadata:
  author: rd-mg
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

#### Step 2c: Scan for Deviations (MANDATORY)

Compare the final implementation against the `design.md`.

1. Check if all files in the implementation are listed in the design's "File Changes" table.
2. Check if any major architecture decisions were changed.
3. Record findings in the `## Deviation Log` section of the archive report.

### Step 3: Move to Archive

If using file-based persistence, move the entire change folder to archive with date prefix:

```
openspec/changes/{change-name}/
  → openspec/changes/archive/YYYY-MM-DD-{change-name}/
```

Use today's date in ISO format (e.g., `2026-02-16`).

### Step 4: Commit Changes

Commit all repository changes introduced by this change:

1. Stage all modified files (specs, code, tests, assets).
2. Write a conventional commit message:
   ```
   feat(scope): summary of the change

   - What changed (high-level)
   - Why it changed (motivation)
   - Closes openspec/changes/{change-name}
   ```
3. Adhere strictly to conventional commit formatting (type(scope): description).
4. Do NOT include "Co-Authored-By" or AI attribution.

### Step 5: Update Documentation

Update project documentation to reflect the completed change:

1. **README.md**: If the change introduces user-facing features, update the relevant sections.
2. **CHANGELOG.md**: Append an entry under the current version/unreleased section:
   ```markdown
   ## [Unreleased]

   ### Added|Changed|Fixed|Removed
   - Summary of the change (link to openspec change if public)
   ```
3. **AGENTS.md** (if applicable): Update any agent-facing instructions affected by the change.

### Step 6: Verify Archive

Confirm:
- [ ] Main specs updated correctly
- [ ] Change folder moved to archive
- [ ] Archive contains all artifacts (proposal, specs, design, tasks)
- [ ] Active changes directory no longer has this change
- [ ] Changes committed to repository with conventional commit message
- [ ] README.md and CHANGELOG.md updated (if applicable)
- [ ] For Engram: All artifact observation IDs are recorded in the archive report.

### Step 7: Persist Archive Report

**This step is MANDATORY — do NOT skip it.**
Follow the persistence rules defined in Step 2 of `_shared/mode-branching.md`.

### Step 8: Return Summary

Return to the orchestrator:

```markdown
## Change Archived

**Change**: {change-name}
**Archived to**: {artifact_path} | {topic_key}

### Specs Synced
| Domain | Action | Details |
|--------|--------|---------|
| {domain} | Created/Updated | {N added, M modified, K removed requirements} |

### Deviation Log
{List any implementation deviations from design, or "None identified."}

### Archive Contents
- proposal.md 
- specs/ 
- design.md 
- tasks.md  ({N}/{N} tasks complete)

### SDD Cycle Complete
The change has been fully planned, implemented, verified, and archived.
```

## Rules

- ALWAYS include a `### Deviation Log` in the archive report
- NEVER archive a change that has CRITICAL issues in its verification report
- ALWAYS sync delta specs BEFORE moving to archive
- When merging into existing specs, PRESERVE requirements not mentioned in the delta
- Use ISO date format (YYYY-MM-DD) for archive folder prefix
- If the merge would be destructive (removing large sections), WARN the orchestrator and ask for confirmation
- The archive is an AUDIT TRAIL — never delete or modify archived changes
- If `openspec/changes/archive/` doesn't exist, create it
- Apply any `rules.archive` from `openspec/config.yaml`
- Return envelope per **Section D** from `skills/_shared/sdd-phase-common.md`.
