# Engram Convention v3.0

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
- `mem_save_prompt` — CLI-only edge case

### mem_suggest_topic_key Standard Flow
```
BEFORE EVERY NEW mem_save:
  1. suggestion = mem_suggest_topic_key(query: "{what you're saving}")
  2. Check existing_keys for similarity:
     - > 0.85 similarity → mem_update existing key
     - 0.5-0.85 → decide: same concept (update) or refinement (new key /v2)
     - < 0.5 → mem_save with suggested_key
  3. Validate key format: lowercase, hyphens, forward slashes only
  4. NEVER save with malformed key
```

### ByteRover Loading Order
```
Level 1 (always): mem_context(limit:5) at session start → working memory
Level 2 (on demand): mem_search → mem_get_observation → episodic memory
Level 3 (lazy): knowledge/odoo-*/reference/* → semantic memory (search, don't preload)
Level 4 (archive): sdd/{change}/archive → historical context
```
