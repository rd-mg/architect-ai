# Proposal: architect-ai V3 Upgrade

## Intent
Upgrade the core architecture to support unified reasoning modes, cognitive postures, and a consolidated Odoo overlay to reduce token footprint and improve reasoning quality.

## Scope
- internal/assets/skills/: Unified reasoning, cognitive modes, MCP persistence.
- internal/assets/claude/: V3 orchestrator and phase protocols.
- internal/assets/overlays/odoo-development-skill/: Pattern consolidation and bundle logic.
- internal/cli/overlay.go: Patch for versioned pattern support.

## Approach
1. Phase 0: Stabilization (Green CI, purge Spanish, remove duplicates).
2. Phase 1: Core Architecture (Unified adaptive-reasoning, new cognitive-mode skill).
3. Phase 2: Odoo Overlay Restructure (Mechanical consolidation into versioned bundles).
4. Phase 3: Polish & Validation (Remaining orchestrator cores, doc updates).

## Rollback Plan
- Restore SKILL.md and persona backups.
- Revert overlay.go patch.
- Reinstall loose Odoo pattern files from V2 package.
