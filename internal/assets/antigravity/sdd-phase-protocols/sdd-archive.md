<!-- architect-ai:prompt-caching-anchor:start -->
# SDD Phase Protocol: sdd-archive
Project: architect-ai
Adapter: Antigravity
Version: 1.1
<!-- architect-ai:prompt-caching-anchor:end -->

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
   - Research audit (NotebookLM status)
   - **Deviation Log**: Table of (Planned | Actual | Rationale). If none, state "None".
   - **Lessons Learned**: ≥ 2 items documenting what worked, what failed, and what to avoid next time.
   - Any open follow-ups
   - Step 3c: Entity Tag Extraction (scan artifacts and extract named entities).
     **Format**:
     ## Entity Index
     modules: ...
     types: ...
     functions: ...
     files: ...
     services: ...
     concepts: ...
     risks: ...
3b. Eval Gate Check: Verify NO tasks classified as HIGH risk lack an explicit eval step (peer review, manual validation, test evidence) in the mem_search observation history. If HIGH risk tasks lack evidence, STOP and return blocked.
   
4. **Persistence (Learned Patterns)**: Save lessons and patterns to `project/{project}/lessons`. Search `knowledge/_global/skill/{skill-name}/learned-patterns`. If found, `mem_update` with appended patterns and incremented version. If not, `mem_save` new patterns.
4. If OpenSpec mode: move change directory to archive/ folder
5. Update DAG state to "archived"


## Empirical Verification Loop (+++Empirical)
- **MANDATORY**: Before concluding, you MUST perform an empirical verification of your findings/artifacts.
- Examples: run a script, check a file, verify a tool output, or perform a manual check of the logic.
- Record the evidence in the `empirical_proof` field of the return handshake.

## Artifact Store: {mode}

## Persistence (MANDATORY)
mem_save(
  title: "sdd/{change-name}/archive-report",
  topic_key: "sdd/{change-name}/archive-report",
  type: "archive-summary",
  project: "{project}",
  content: "{archive summary markdown with entity index at bottom}"
)

## Size Budget: 200 words max

## Return Envelope & Compliance per sdd-phase-common.md (Sections A-F)
```

## Deviation Logging (MANDATORY)
Compare the final implementation against the original Specs and Design:
1. **Technical Debt**: List any shortcuts or temporary solutions implemented.
2. **Spec Deviations**: List features that differ from the spec or were not implemented.
3. **Rationale**: Briefly explain WHY these deviations occurred.

Save to Engram: `sdd/{project}/deviations`.

## Result Processing

- Update state: `verified` → `archived`
- Next recommended: `none` (change is closed)
- Report completion to user in LITE caveman style

## Failure Handling

- If verify-report is missing → return `blocked`, route to sdd-verify first
- If verdict is NEEDS CHANGES → return `blocked`, don't archive incomplete work
- If required artifacts missing → return `partial`, list what's missing
