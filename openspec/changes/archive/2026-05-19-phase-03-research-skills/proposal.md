# Proposal: Phase 3 - Skill Registry v3 Tiers + Researcher Universal + bash-expert/fish

## Intent
Implement a declarative "Context Kubernetes" model for the Skill Registry (v3) using Tiers to eliminate context bloat caused by `bridge:always`. Centralize all research routing into a single `researcher` agent, preventing duplication across orchestrators. Extend `bash-expert` natively to support Fish shell. Ensure `mcp-notebooklm-orchestrator` is loaded on-demand, not globally.

## Scope
### In Scope
- Skill Registry v3 — 3 Tiers (Tier 1: Foundation, Tier 2: Context Activated, Tier 3: On-Demand).
- `_shared/foundation.md` — Go installer merges Tier 1 skills into a single auto-generated file.
- `skill-resolver.md` v3.0 — Orchestrators use this to conditionally inject Tier 2 skills.
- Researcher Universal v2.0: single delegation point for ALL investigation (Tier 3).
- `bash-expert` extended to `bash-expert-fish` (Tier 1).
- Go: `internal/skill/registry/generator.go` and `generator_test.go` for manifest and foundation generation.

### Out of Scope
- Odoo-specific research patterns (Phase 9).
- MCP server configuration changes (Phase 10).

## Impact
- Eliminates context overhead of loading unused skills (like notebooklm) into every agent.
- Declarative `.atl/skill-manifest.yaml` simplifies skill state management.
- Fish shell users get first-class support.
- Orchestrators are leaner without duplicated research routing logic.
