# Persistence Contract (shared across all SDD skills)

## Mode Resolution

Orchestrator passes `artifact_store.mode` with one of: `engram | openspec | hybrid | none`.

Orchestrator ASKs user which mode when `/sdd-new`, `/sdd-ff`, or `/sdd-continue` is invoked first time in session. Choice cached for session.

Default (if user doesn't specify): if Engram available → `engram`. Otherwise → `none`.

## Mode Roles

- **`engram`**: Working memory between sessions. Upserts overwrite — no iteration history. Local only, not shareable.
- **`openspec`**: Source of truth. Files in repo, git history, team-shareable, full audit trail.
- **`hybrid`**: Both — files for team + engram for recovery. Higher token cost.
- **`none`**: Ephemeral. Lost when conversation ends.

### Mode Comparison

| Capability | `engram` | `openspec` | `hybrid` | `none` |
|------------|----------|------------|----------|--------|
| Cross-session recovery |  |  (needs git) |  |  |
| Compaction survival |  |  |  |  |
| Shareable with team |  (local DB) |  (committed files) |  (files) |  |
| Full iteration history |  (upsert overwrites) |  (git history) |  (files + git) |  |
| Audit trail (archive) | Partial (report only) |  (full folder) |  (both) |  |
| Project files created | Never | Yes | Yes | Never |

### `engram` mode limitation

Engram uses `topic_key`-based upserts. Re-running a phase for same change **overwrites** previous version — no revision history kept. Archive phase saves summary report, not full artifact folder. For iteration history or team collaboration, use `openspec` or `hybrid`.

## Behavior Per Mode

| Mode | Read from | Write to | Project files |
|------|-----------|----------|---------------|
| `engram` | Engram | Engram | Never |
| `openspec` | Filesystem | Filesystem | Yes |
| `hybrid` | Engram (primary) + Filesystem (fallback) | Both | Yes |
| `none` | Orchestrator prompt context | Nowhere | Never |

### Hybrid Mode

Persists every artifact to BOTH Engram and OpenSpec:
- Engram: cross-session recovery, compaction survival, deterministic search
- OpenSpec: human-readable files, version-controllable artifacts

Write to Engram (per `engram-convention.md`) AND to filesystem (per `openspec-convention.md`) for every artifact.

Read priority: Engram first; fall back to filesystem if Engram returns no results.
Write: both writes MUST succeed for operation to be complete.
Token cost: hybrid consumes MORE tokens per operation. Use only when you need both cross-session persistence AND local file artifacts.

## State Persistence (Orchestrator)

Orchestrator persists DAG state after each phase transition for SDD recovery after compaction.

| Mode | Persist State | Recover State |
|------|--------------|---------------|
| `engram` | `mem_save(topic_key: "sdd/{change-name}/state")` | `mem_search("sdd/*/state")` → `mem_get_observation(id)` |
| `openspec` | Write `openspec/changes/{change-name}/state.yaml` | Read `openspec/changes/{change-name}/state.yaml` |
| `hybrid` | Both: `mem_save` AND write `state.yaml` | Engram first; filesystem fallback |
| `none` | Not possible — warn user | Not possible |

## Common Rules

- `none` → do NOT create or modify project files; return results inline only
- `engram` → do NOT write project files; persist to Engram and return observation IDs
- `openspec` → write files ONLY to paths defined in `openspec-convention.md`
- `hybrid` → persist to BOTH Engram AND filesystem; follow both conventions
- NEVER force `openspec/` creation unless orchestrator explicitly passed `openspec` or `hybrid`
- If unsure which mode, default to `none`

## Sub-Agent Context Rules

Sub-agents launch with fresh context and NO access to orchestrator's instructions or memory protocol.

Who reads, who writes:
- Non-SDD (general task): orchestrator searches engram, passes summary in prompt; sub-agent saves discoveries via `mem_save`
- SDD (phase with dependencies): sub-agent reads artifacts directly from backend; sub-agent saves its artifact
- SDD (phase without dependencies, e.g. explore): nobody reads; sub-agent saves its artifact

Why this split:
- Orchestrator reads for non-SDD: knows what context is relevant; sub-agents doing own searches waste tokens on irrelevant results
- Sub-agents read for SDD: SDD artifacts are large; inlining in orchestrator prompt would consume entire context window
- Sub-agents always write: they have complete detail; nuance lost by time results flow back to orchestrator

## Orchestrator Prompt Instructions for Sub-Agents

Non-SDD:
```
PERSISTENCE (MANDATORY):
If you make important discoveries, decisions, or fix bugs, you MUST save them to engram before returning:
  mem_save(title: "{short description}", type: "{decision|bugfix|discovery|pattern}",
           project: "{project}", content: "{What, Why, Where, Learned}")
Do NOT return without saving what you learned. This is how the team builds persistent knowledge across sessions.
```

SDD (with dependencies):
```
Artifact store mode: {engram|openspec|hybrid|none}
Read these artifacts before starting (search returns truncated previews):
  mem_search(query: "sdd/{change-name}/{type}/", project: "{project}") → get ID
  mem_get_observation(id: {id}) → full content (REQUIRED)

PERSISTENCE (MANDATORY — do NOT skip):
After completing your work, you MUST call:
  mem_save(
    title: "sdd/{change-name}/{artifact-type}/main",
    topic_key: "sdd/{change-name}/{artifact-type}/main",
    type: "architecture",
    project: "{project}",
    content: "{your full artifact markdown}"
  )
If you return without calling mem_save, the next phase CANNOT find your artifact and the pipeline BREAKS.
```

SDD (no dependencies):
```
Artifact store mode: {engram|openspec|hybrid|none}

PERSISTENCE (MANDATORY — do NOT skip):
After completing your work, you MUST call:
  mem_save(
    title: "sdd/{change-name}/{artifact-type}/main",
    topic_key: "sdd/{change-name}/{artifact-type}/main",
    type: "architecture",
    project: "{project}",
    content: "{your full artifact markdown}"
  )
If you return without calling mem_save, the next phase CANNOT find your artifact and the pipeline BREAKS.
```

## Skill Registry

Orchestrator pre-resolves compact rules from skill registry and injects as `## Project Standards (auto-resolved)` in launch prompt. Sub-agents do NOT read registry or individual SKILL.md files — rules arrive pre-digested.

To generate/update: run `skill-registry` skill, or `sdd-init`.

Sub-agent skill loading: check for `## Project Standards (auto-resolved)` block in prompt — if present, follow those rules. If not present, check for `SKILL: Load` instructions as fallback. If neither exists, proceed without — this is not an error.

## Detail Level

Orchestrator may pass `detail_level`: `concise | standard | deep`. Controls output verbosity but does NOT affect what gets persisted — always persist full artifact.
