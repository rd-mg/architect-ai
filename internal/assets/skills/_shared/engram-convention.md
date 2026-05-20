# Engram Artifact Convention (reference documentation)

NOTE: Critical engram calls (`mem_search`, `mem_save`, `mem_get_observation`) are inlined directly in each skill's SKILL.md. This document is supplementary reference — sub-agents do NOT need to read it to function.

## Naming Rules

ALL SDD artifacts persisted to Engram MUST follow this deterministic naming:

```
title:     sdd/{change-name}/{artifact-type}
topic_key: sdd/{change-name}/{artifact-type}
type:      architecture
project:   {detected or current project name}
scope:     project
```

### Knowledge Roots (Global & External)

Artifacts that are NOT tied to a specific change use a hierarchical knowledge root:

| Root | Topic Key Pattern | Description |
|------|-------------------|-------------|
| `knowledge/_global/skill/` | `knowledge/_global/skill/{name}` | Global skill definitions and rules |
| `knowledge/external/` | `knowledge/{domain}/external/{topic}` | External research (NotebookLM, Context7) |

`{domain}` identifies the broad area of knowledge (e.g., `odoo`, `architecture`, `vendor-playbook`). For NotebookLM, this is the sanitized notebook name.

---

### Artifact Types

| Artifact Type | Produced By | Description |
|---------------|-------------|-------------|
| `explore` | sdd-explore | Exploration analysis |
| `proposal` | sdd-propose | Change proposal |
| `spec` | sdd-spec | Delta specifications (all domains concatenated) |
| `design` | sdd-design | Technical design |
| `tasks` | sdd-tasks | Task breakdown |
| `apply-progress` | sdd-apply | Implementation progress (one per batch) |
| `verify-report` | sdd-verify | Verification report |
| `archive-report` | sdd-archive | Archive closure with lineage |
| `state` | orchestrator | DAG state for recovery after compaction |

Exception: `sdd-init` uses `sdd-init/{project-name}` as both title and topic_key.

### State Artifact

```
mem_save(
  title: "sdd/{change-name}/state",
  topic_key: "sdd/{change-name}/state",
  type: "architecture",
  project: "{project}",
  content: "change: {change-name}\nphase: {last-phase}\nartifact_store: engram\nartifacts:\n  proposal: true\n  specs: true\n  design: false\n  tasks: false\ntasks_progress:\n  completed: []\n  pending: []\nlast_updated: {ISO date}"
)
```

Recovery: `mem_search("sdd/{change-name}/state")` → `mem_get_observation(id)` → parse YAML → restore state.

## Recovery Protocol (2 steps)

```
Step 1: mem_search(query: "sdd/{change-name}/{artifact-type}", project: "{project}") → truncated preview + ID
Step 2: mem_get_observation(id: {observation-id}) → complete content
```

When retrieving multiple artifacts, group all searches first, then all retrievals:

```
STEP A — SEARCH (get IDs only):
  mem_search(query: "sdd/{change-name}/proposal", ...) → save ID
  mem_search(query: "sdd/{change-name}/spec", ...) → save ID
  mem_search(query: "sdd/{change-name}/design", ...) → save ID

STEP B — RETRIEVE FULL CONTENT (mandatory):
  mem_get_observation(id: {proposal_id})
  mem_get_observation(id: {spec_id})
  mem_get_observation(id: {design_id})
```

Loading project context:
```
mem_search(query: "sdd-init/{project}", project: "{project}") → get ID
mem_get_observation(id) → full project context
```

## Writing Artifacts

Standard write:
```
mem_save(
  title: "sdd/{change-name}/{artifact-type}",
  topic_key: "sdd/{change-name}/{artifact-type}",
  type: "architecture",
  project: "{project}",
  content: "{full markdown content}"
)
```

Concrete example — saving a proposal for `add-dark-mode`:
```
mem_save(
  title: "sdd/add-dark-mode/proposal",
  topic_key: "sdd/add-dark-mode/proposal",
  type: "architecture",
  project: "my-app",
  content: "## Proposal\n\nAdd dark mode toggle..."
)
```

Update existing artifact (when you have the observation ID):
```
mem_update(id: {observation-id}, content: "{updated full content}")
```

Use `mem_update` when you have the exact ID. Use `mem_save` with same `topic_key` for upserts.

### Browsing All Artifacts for a Change

```
mem_search(query: "sdd/{change-name}/", project: "{project}")
→ Returns all artifacts for that change
```

## Project Name Resolution (engram v1.11.0+)

Engram auto-detects the project name from the git remote at MCP startup. The `--project` flag and `ENGRAM_PROJECT` env var can override detection. All project names are normalized to lowercase and trimmed.

If the agent saves a memory under a project name that doesn't match existing observations, engram warns about potential name drift. Use `mem_merge_projects` (MCP tool) or `engram projects consolidate` (CLI) to merge variants.

## Upsert Behavior

Same `topic_key` + `project` + `scope` → UPDATE (overwrite), not INSERT. Previous content is lost — `revision_count` increments but old content is NOT saved. This is by design — engram is working memory, not an audit trail. For iteration history or team collaboration, use `openspec` or `hybrid` mode.

## Why This Convention

- Deterministic titles → recovery works by exact match
- `topic_key` → enables upserts without duplicates
- `sdd/` prefix → namespaces all SDD artifacts
- Two-step recovery → search previews are always truncated; `mem_get_observation` is the only way to get full content
- Lineage → archive-report includes all observation IDs for complete traceability

## Engram Tool Distribution v3.0 [MANDATORY for all agents]

### Tool Distribution Matrix

| Tool | L0 architect | L1 sdd-orch | L1 gen-orch | L2 SDD phases | L2 non-SDD | Justification |
|---|---|---|---|---|---|---|
| `mem_search` | ✅ | ✅ | ✅ | ✅ | ✅ | Base universal — no exceptions |
| `mem_save` | ✅ | ✅ | ✅ | ✅ | ✅ | Base universal |
| `mem_get_observation` | ✅ | ✅ | ✅ | ✅ | ✅ | Base universal — always after search |
| `mem_suggest_topic_key` | ✅ | ✅ | ✅ | ✅ only sdd-archive/sdd-design | ✅ researcher | Prevents key drift — low cost |
| `mem_session_summary` | ✅ | ✅ | ✅ | ✅ sdd-archive ONLY | ❌ | Structured session close |
| `mem_timeline` | ❌ | ✅ sdd-orch | ❌ | ✅ sdd-verify, sdd-archive | ❌ | Chronological change audit |
| `mem_update` | ❌ | ✅ | ❌ | ✅ sdd-apply (progress) | ❌ | In-place update without new key |
| `mem_context` | ✅ L0 ONLY | ❌ | ❌ | ❌ | ❌ | Compact resume for session start |
| `mem_current_project` | ✅ | ✅ | ✅ | ❌ | ❌ | Project identity at orchestrator level |
| `mem_delete` | ❌ NEVER | ❌ | ❌ | ❌ | ❌ | Risk of irreversible data loss |
| `mem_judge` | ❌ | ❌ | ❌ | ❌ | ❌ | CLI only — never in agents |
| `mem_compare` | ❌ | ❌ | ❌ | ❌ | ❌ | CLI only — never in agents |
| `mem_merge_projects` | ❌ | ❌ | ❌ | ❌ | ❌ | CLI only — never in agents |

### Universal (all agents receive these in every prompt)
- `mem_search` — find relevant knowledge
- `mem_save` — persist findings
- `mem_get_observation` — read full document (always call after search)
- `mem_suggest_topic_key` — prevent key drift (call BEFORE every new mem_save)

### Orchestrators only (L0 and L1)
- `mem_current_project` — establish project identity
- `mem_session_summary` — structured session close
- `mem_context(limit: 5)` — compact resume for L0 only
- `mem_timeline` — L1a sdd-orchestrator + sdd-verify + sdd-archive
- `mem_update` — L1a sdd-orchestrator + sdd-apply (progress update only)

### NEVER in any agent (CLI only)
- `mem_delete` — irreversible, CLI-only
- `mem_judge` — CLI-only
- `mem_compare` — CLI-only
- `mem_merge_projects` — CLI-only

### ByteRover Loading Order
- **Level 1 — Working Memory (session scope)**: `mem_context(limit: 5)` at session start
- **Level 2 — Episodic Memory (project scope)**: `mem_search("sdd/{change_name}/spec")` → load only what current phase needs
- **Level 3 — Semantic Memory (knowledge scope)**: NEVER load Odoo guides eagerly. Lazy load via `mem_search('odoo {version} {topic}')`
- **Level 4 — Archive Memory (historical scope)**: `sdd/{change}/archive`

