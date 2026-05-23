# Proposal: Phase 5 - Engram Context Guardian v3: ByteRover + Branch A/B + Skill Tiers

## Intent
Expand Engram tool availability and implement Context Guardian v3 using the ByteRover Hierarchical Context Pattern (4 levels) and a unified Branch A/B strategy for context overflow. Introduce collision handling for `mem_suggest_topic_key` and ensure idempotent indexing for Odoo knowledge nodes via a Go-based indexer. Ensure persistent memory survives context pressure through proactive save checkpoints and automated compression hooks.

## Scope
### In Scope
- Engram tool tier matrix: universal vs orchestrator-only tools.
- ByteRover loading protocol: Working (L1), Episodic (L2), Semantic (L3), Archive (L4).
- `mem_suggest_topic_key` collision handling protocol.
- Context Guardian v3: Branch A (Engram persistence) + Branch B (context-mode session buffer).
- Manual summary fallback for platforms without native compression.
- Go implementation: `internal/skill/odoo/indexer.go` for idempotent SaveIdempotent.

### Out of Scope
- Engram server-side changes (memory storage, FTS5 engine).
- New MCP tool creation in Engram itself.

## Impact
- Context pressure triggers automatic compression and checkpointing instead of silent degradation.
- Key drift in Engram is eliminated via collision handling.
- Semantic memory (Odoo guides) is lazy-loaded (ByteRover L3) to save tokens.
- Idempotent indexing prevents duplicate entries during registry updates.
