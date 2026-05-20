# Spec: Phase 5 - Engram Context Guardian v3: ByteRover + Branch A/B + Skill Tiers

## Requirements
- Engram tools MUST be distributed by agent level (see Tool Distribution Table below).
- `mem_suggest_topic_key` MUST be called BEFORE every new `mem_save` and handle collision (>0.85 similarity = update).
- `mem_delete`, `mem_judge`, `mem_compare`, `mem_merge_projects` MUST NEVER be available to any agent — CLI only.
- ByteRover Hierarchical Context Pattern MUST be used for memory loading (Levels 1-4).
- Context Guardian v3 MUST implement unified Branch A (Engram persistence) and Branch B (context-mode buffer) strategies.
- Context Guardian MUST add Trigger 4: `D4 >= 2` from Adaptive Reasoning Gate.
- context-mode MUST auto-trigger for commands producing > 10KB output (Branch B).
- context-mode unavailability MUST NEVER block execution — graceful fallback with WARN.
- Idempotent Indexing for Odoo Knowledge Nodes MUST be used to prevent duplicate entries during skill registry updates.

## Engram Tool Distribution Table

## 5.1 Engram Tool Distribution Table

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

## Usage Protocols

### `mem_suggest_topic_key` — Collision Handling Protocol

```
BEFORE EVERY NEW mem_save:
1. suggestion = mem_suggest_topic_key(query: "{description}", project: current_project)

2. IF suggestion.existing_keys contains semantically similar entry:
   CASE "semantically identical" (> 0.85 similarity):
     → Use mem_update(existing_key, new_content)
     → Do NOT create new key (avoid duplicate)
     
   CASE "related but different" (0.5 - 0.85 similarity):
     → If same concept: mem_update
     → If refinement: use new key with suffix: {existing_key}/v2
     
   CASE "tangentially related" (< 0.5 similarity):
     → Use the suggested new key and mem_save

3. ELSE: Use suggested.recommended_key (or taxonomy format) and mem_save

4. Validate key format (lowercase, hyphens, slashes). Never save with malformed key.
```

### ByteRover Loading Protocol

```
Level 1 — Working Memory (session scope)
  mem_context(limit: 5) at session start

Level 2 — Episodic Memory (project scope)
  mem_search("sdd/{change_name}/spec") → load only what current phase needs

Level 3 — Semantic Memory (knowledge scope)
  NEVER load Odoo guides eagerly. Lazy load via mem_search('odoo {version} {topic}')

Level 4 — Archive Memory (historical scope)
  sdd/{change}/archive
```

### `mem_session_summary` — Session Close Protocol

```
On session end (user says "wrap up", "done", "session close") OR at sdd-archive:

mem_session_summary(
  project: current_project,
  goal: "{what was attempted}",
  accomplished: "{what was completed — files/functions changed}",
  decisions_made: ["{key decision 1}", "{key decision 2}"],
  next_steps: ["{next action 1}", "{next action 2}"],
  key_files: ["{file1}", "{file2}"],
  blockers: ["{any open blockers}"]
)

DO NOT call mid-session — only at session close or archive phase.
```

### `mem_update` — In-Place Update Protocol

```
Use ONLY when updating an existing, previously-saved observation.
DO NOT use for new observations — use mem_save for those.

# Pattern: sdd-apply progress update
existing = mem_search("sdd/{change_name}/apply-progress")
IF existing:
  mem_update(
    topic_key: "sdd/{change_name}/apply-progress",
    content: merged_progress  # MERGE new state with existing, not replace
  )
ELSE:
  mem_save("sdd/{change_name}/apply-progress", initial_progress)

Rule: ALWAYS read existing content via mem_get_observation BEFORE updating.
Never overwrite — always merge state.
```

### `mem_timeline` — Chronological Audit

```
Shows all observations in chronological order for the current project.
Use to reconstruct the sequence of decisions during a change.

# In sdd-verify: verify decision timeline is consistent
timeline = mem_timeline(
  project: current_project,
  filter: "sdd/{change_name}/",
  limit: 20
)
# Review: does spec → design → tasks → apply sequence make sense?
# Flag if design was saved AFTER apply (process violation)

# In sdd-archive: include timeline in final report
timeline_summary = mem_timeline(project: current_project, filter: "sdd/{change_name}/")
mem_save("sdd/{change_name}/archive", {
  ...,
  decision_timeline: timeline_summary
})
```

## Context Guardian v3 — Branch A/B Unified Protocol

```markdown
### Auto-Trigger Conditions (ANY of these fires the guardian)
Trigger 1: token_usage > 50% of context window
Trigger 2: compaction detected (skill_resolution: fallback reported)
Trigger 3: sub-agent reports context loss in Result Contract
Trigger 4: D4 >= 2 in Adaptive Reasoning Gate assessment
Trigger 5: 3+ exploratory reads in same context window (long-session rule)

### Branch A: Engram Persistence (PRIMARY)
Goal: Durable cross-session memory. Zero information loss. LCM-compliant.
1. Generate checkpoint BEFORE compress via mem_save("session/context-pack/{project}/{timestamp}")
2. Execute compress command (/compact or /compress)
3. Reload working memory via mem_context(limit: 3)
* If Branch A fails (Engram down), WARN and attempt Branch B.

### Branch B: context-mode MCP Buffer (SECONDARY)
Goal: Transparent session buffering for large outputs.
Scope: Session only.
Use ctx_execute() for outputs > 10KB. Do NOT use for architectural decisions.

### Manual Summary Protocol (VSCode Copilot / Antigravity)
When no compress command available AND Engram checkpoint saved:
1. Emit LITE instructing user: "Context limit approaching... Start a new chat and say: 'resume {change_name} from Engram'"
2. Halt session.
```

### Platform Compress Command Matrix

| Platform | Has compress? | Command | Auto-call? |
|---|---|---|---|
| OpenCode | ✅ | `/compact` | YES |
| Claude Code | ✅ | `/compact` | YES |
| Gemini CLI | ✅ | `/compress` | YES |
| VSCode Copilot | ❌ | — | NO → manual summary |
| Antigravity | ❌ | — | NO → manual summary |

### Updated Auto-Trigger Rules (v3)
```
Trigger 1 (unchanged): token usage > 50% of context window
Trigger 2 (unchanged): compaction detected (skill_resolution: fallback)
Trigger 3 (unchanged): sub-agent reports context loss
Trigger 4 (NEW): D4 >= 2 (high context pressure from Adaptive Reasoning Gate)
```

### Hook Detection

```bash
# .atl/hooks/on_context_pressure.sh (generated by architect-ai install)
#!/usr/bin/env bash
# Called by context-guardian when threshold reached

set -euo pipefail

PLATFORM="${ARCHITECT_PLATFORM:-unknown}"
ENGRAM_BIN="${ENGRAM_BIN:-engram}"

# Step 1: Try Engram checkpoint
"${ENGRAM_BIN}" checkpoint save --project "${ARCHITECT_PROJECT}" 2>/dev/null \
  && echo "CHECKPOINT: saved to Engram" \
  || echo "WARN: Engram checkpoint failed"

# Step 2: Platform compress
case "${PLATFORM}" in
  opencode|claude) echo "/compact" ;;
  gemini)          echo "/compress" ;;
  *)               echo "MANUAL_SUMMARY_REQUIRED" ;;
esac
```

## context-mode — Verdict and Improvements

### Verdict: MAINTAIN — Complementary to Engram

| Feature | context-mode | Engram |
|---|---|---|
| Purpose | Anti-flooding of large outputs in session | Persistent memory cross-session |
| Persistence | Session only (transient) | Durable (SQLite + Git sync) |
| Search | `ctx_search` (FTS in session) | `mem_search` (FTS5 in SQLite) |
| Overhead | MCP server (network latency) | Local (minimal) |
| Versioning | ❌ | ✅ engram sync → Git |
| Use case | Script that returns 50KB of output | Architecture decision from last week |

### context-mode Improvements

```
## Auto-trigger threshold (NEW)
IF command output > 10KB estimated:
  USE ctx_execute() or ctx_batch_execute() instead of direct command

Detection heuristic:
- git log --all --oneline → potentially large → ctx_execute
- rg pattern . (project root, unfiltered) → potentially large → ctx_execute
- npm install / pip install → output > 100 lines → ctx_execute
- pytest / go test (full suite) → use ctx_execute for output capture

## Fallback graceful (NEW — when context-mode server unavailable)
IF context_mode_available = false:
  For large commands: pipe to head -50 and note "[TRUNCATED - run ctx_execute for full]"
  For web fetches: use web_search tool snippets instead of full HTML
  Continue without blocking — add WARN to response

NEVER block execution because context-mode is unavailable.

## Policy reminder (unchanged)
context-mode = session buffer only
Engram = durable architecture memory
NEVER use ctx_index as substitute for mem_save.
NEVER use ctx_search as substitute for mem_search (no cross-session).
```

## Engram Convention Asset — `_shared/engram-convention.md` Patch

```markdown
## Engram Tool Distribution v3.0 [MANDATORY for all agents]

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
```
Level 1 (always): mem_context(limit:5) at session start → working memory
Level 2 (on demand): mem_search → mem_get_observation → episodic memory
Level 3 (lazy): knowledge/odoo-*/reference/* → semantic memory (search, don't preload)
Level 4 (archive): sdd/{change}/archive → historical context
```
```

## Scenarios

### Scenario 1: mem_suggest_topic_key Prevents Duplicates
**Given** Engram has `"sdd/auth-feature/design"` with OAuth design.
**When** agent tries to save "authentication design for auth feature".
**Then** `mem_suggest_topic_key` MUST return `existing_key = "sdd/auth-feature/design"`.
**And** agent MUST use `mem_update` instead of `mem_save`.
**And** no duplicate taxonomic entries for same concept.

### Scenario 2: Context Guardian Compress in OpenCode
**Given** OpenCode, no hook configured, context at 52%.
**When** token threshold triggers.
**Then** context-guardian MUST fire → `/compact` executed.
**And** LITE emitted: `"Context at 52%. Running /compact."`.
**And** `/compact` invoked automatically without user prompt.

### Scenario 3: Context Guardian Manual Summary in VSCode
**Given** VSCode Copilot, no hook, no native compress, context at 55%.
**When** token threshold triggers.
**Then** manual_summary_protocol MUST execute.
**And** Engram entry saved at `"session/context-pack/{timestamp}"`.
**And** user instructed to start new chat with "resume from Engram".

### Scenario 4: context-mode Auto-Trigger
**Given** command that produces > 10KB output (e.g., `git log --all --oneline` on large repo).
**When** agent needs to run it.
**Then** MUST redirect to `ctx_execute` instead of direct shell command.
**And** only stdout relevant portion enters context.

### Scenario 5: Idempotent Indexing (Odoo)
**Input**: architect-ai skill-registry run twice on same Odoo project.
**Expected**: Second run updates existing Engram entries (no duplicate INSERTS).
**PASS if**: No duplicate keys in Engram after double indexing.

### Scenario 6: ByteRover Lazy Loading
**Input**: sdd-explore on Odoo project — needs to understand OWL components.
**Expected**: Agent calls `mem_search("owl component lifecycle odoo 18")` FIRST.
**PASS if**: No Context7 call when Engram cache is warm.

## Expected Results

| Metric | Before | After |
|---|---|---|
| Key drift en Engram | ⚠️ Frecuente | ✅ mem_suggest_topic_key con collision handling |
| Context overflow strategy | ❌ Solo compress | ✅ Branch A (persist) + Branch B (buffer) unified |
| Engram checkpoint | ❌ No | ✅ Obligatorio antes de /compact o /compress |
| Odoo guides overhead | ~15,000 tokens | ✅ ~70 tokens (ByteRover lazy FTS5) |
| Idempotencia skill-registry| ❌ Posibles duplicados | ✅ SaveIdempotent con goroutines |
| mem_delete en agentes | ⚠️ Posible | ✅ Removido de todos los agentes |
| VSCode/Antigravity | ❌ Sin instrucciones | ✅ Manual summary + new chat protocol |
