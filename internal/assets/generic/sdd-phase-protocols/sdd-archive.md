# Phase Protocol: sdd-archive

## Dependencies
- **Reads**: all artifacts for the change
- **Writes**: `archive-report` artifact; marks change as archived

## Cognitive Posture
None — mechanical close-out.

## Model
haiku — simple copy and state update

## Sub-Agent Launch Template

```
## Project Standards (auto-resolved)
{matching compact rules}

## Available Tools
{verified tool list}

## Phase: sdd-archive

Task: Close out the change "{change-name}". This is a mechanical phase.

## Procedure
1. Read verify-report. Confirm verdict is APPROVED or CONDITIONALLY APPROVED
2. If verdict is NEEDS CHANGES or UNRESOLVED → STOP and return `blocked`
3. Generate archive summary:
   - Change name
   - Start date, end date
   - Proposal summary
   - Outcome summary (what shipped)
   - Tasks completed count
   - Verification verdict
   - Any open follow-ups
4. If OpenSpec mode: move change directory to archive/ folder
5. Update DAG state to "archived"

## Artifact Store: {mode}

## Persistence (MANDATORY)
mem_save(
  title: "sdd/{change-name}/archive-report",
  topic_key: "sdd/{change-name}/archive-report",
  type: "archive-summary",
  project: "{project}",
  content: "{archive summary markdown}"
)

## Size Budget: 200 words max

## Return Envelope per sdd-phase-common.md Section D
```

## Result Processing

- Update state: `verified` → `archived`
- Next recommended: `none` (change is closed)
- Report completion to user in LITE caveman style

## Failure Handling

- If verify-report is missing → return `blocked`, route to sdd-verify first
- If verdict is NEEDS CHANGES → return `blocked`, don't archive incomplete work
- If required artifacts missing → return `partial`, list what's missing
