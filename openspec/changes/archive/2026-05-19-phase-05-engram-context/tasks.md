# Tasks: Phase 5 - Engram Context Guardian v3: ByteRover + Branch A/B + Skill Tiers

## Status: COMPLETED

## Implementation Tasks

### 5.1 Engram Tool Tiering (per agent level)
- [x] Define tool allowlists per layer in `internal/assets/_shared/engram-convention.md`:
  - [x] L0 architect: `mem_current_project`, `mem_context`, `mem_search`, `mem_get_observation`, `mem_save`, `mem_suggest_topic_key`, `mem_session_summary`
  - [x] L1a sdd-orchestrator: full set above + `mem_timeline`, `mem_update`
  - [x] L1b general-orchestrator: `mem_search`, `mem_get_observation`, `mem_save`, `mem_suggest_topic_key`, `mem_current_project`, `mem_session_summary`
  - [x] L2 researcher: `mem_search`, `mem_get_observation`, `mem_save`, `mem_suggest_topic_key`
  - [x] L2 solver/ideator/generalist: `mem_search`, `mem_get_observation`, `mem_save`
  - [x] ByteRover Protocol: implement Working, Episodic, Semantic, Archive memory loading order.

### 5.2 Context Guardian v3 & mem_suggest_topic_key
- [x] `internal/assets/skills/context-guardian/SKILL.md` — v3 update
  - [x] Implement Branch A (Engram persistence checkpoint) before /compact.
  - [x] Implement Branch B (context-mode buffer) for large output (>10KB).
  - [x] Manual summary protocol for VSCode Copilot / Antigravity (no native compress).
  - [x] Auto-trigger on D4 >= 2 or token usage > 50%.
- [x] `mem_suggest_topic_key` Collision Protocol:
  - [x] BEFORE EVERY NEW `mem_save`: check similarity.
  - [x] >0.85 similarity → use `mem_update` on existing key.

### 5.3 Context-Mode MCP Evaluation
- [x] Keep context-mode: YES — Complementary to Engram (session vs persistent).
- [x] Add graceful fallback if context-mode unavailable (never block execution).
- [x] Document policy: never use ctx_index as substitute for mem_save.

### 5.4 mem_session_summary Protocol
- [x] Standardize template (Goal / Instructions / Discoveries / Accomplished / Next Steps / Relevant Files).
- [x] Add to every L0/L1 agent's SKILL.md as mandatory SESSION CLOSE PROTOCOL.

### Go Implementation — Idempotent Indexing
- [x] `internal/skill/odoo/indexer.go`
  - [x] `OdooIndexer` struct with `EngramClient` and `SaveIdempotent`
  - [x] Parallel `IndexAll` execution via goroutines (`errgroup`)
  - [x] Topic key generation for Semantic Memory (ByteRover L3)
- [x] `internal/skill/odoo/indexer_test.go`
  - [x] `TestIndexAll_Idempotent` (verifies UPDATE over duplicate INSERT)
  - [x] `TestIndexAll_PartialFailure`
  - [x] `TestIndexAll_TopicKeyFormat`

