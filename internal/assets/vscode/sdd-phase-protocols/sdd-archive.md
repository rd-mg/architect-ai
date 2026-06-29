<!-- architect-ai:prompt-caching-anchor:start -->
# SDD Phase Protocol: sdd-archive
Project: architect-ai
Adapter: Vscode
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
4. Project File Updates (see below)
## Step 4: Project File Updates (MANDATORY — run BEFORE moving to archive/)

### 4a. CHANGELOG.md (all projects)

IF `CHANGELOG.md` exists in project root:
  Read current content
  Prepend new entry:
  ```
  ## [{computed_version}] — {date_today_ISO}
  ### {category: Added|Changed|Fixed|Removed}
  - {one-line summary per implemented capability from proposal}
  
  Full change: openspec/changes/{change-name}/archive-report
  ```

IF `CHANGELOG.md` does not exist:
  Create it with the above entry and a standard Keep a Changelog header.

### 4b. README.md (all projects)

Read `README.md`. Check if any of the following sections need updating:

- **Installation**: New dependency added in sdd-apply? → update install steps
- **Usage**: New command, endpoint, or behavior? → update usage examples
- **Configuration**: New env var or config key? → document it
- **Known Issues**: Any CONDITIONALLY_APPROVED items from verify-report? → document them

IF any update needed: make targeted edit (do NOT rewrite whole README).
IF README.md does not exist: create minimal one with project name + change summary.

### 4c. `__manifest__.py` (Odoo projects only — when IS_ODOO = true)

1. Read current `__manifest__.py`
2. Bump `version` field:
   ```python
   # Current: "version": "17.0.1.2.3"
   # Bump last segment:   "version": "17.0.1.2.4"
   # If minor feature:    "version": "17.0.1.3.0"
   # If breaking:         "version": "17.0.2.0.0"
   ```
   Ask user for bump type if unclear: `patch | minor | major (within Odoo version)`
3. Update `depends` if sdd-apply added new module dependencies:
   ```python
   "depends": ["base", "mail", "new_dependency_added_in_apply"]
   ```
4. Update `data` list if new XML view files were created
5. Update `assets` dict if new JS/CSS/SCSS files were added
6. Verify `__manifest__.py` is valid Python syntax:
   ```bash
   python3 -c "import ast; ast.parse(open('__manifest__.py').read())"
   ```
   If syntax error → STOP, report error, do NOT archive.

### 4d. `package.json` / `pyproject.toml` / `go.mod` (if applicable)

IF a versioned manifest file exists at project root:
  Check if `version` field should be bumped per the same logic as 4c.
  Only bump if the change introduced a user-visible API change.

5. **Persistence (Learned Patterns)**: Save lessons and patterns to `project/{project}/lessons`. Search `knowledge/_global/skill/{skill-name}/learned-patterns`. If found, `mem_update with appended patterns and incremented version. If not, `mem_save` new patterns.
6. If OpenSpec mode: move change directory to archive/YYYY-MM-DD-{change-name}/
7. Update DAG state to "archived"


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
