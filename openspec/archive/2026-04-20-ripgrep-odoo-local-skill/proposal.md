# Proposal: 04-ripgrep-odoo-local-skill

## Scope
- **In**: 
  - Implementation of a new Architect-AI skill: `ripgrep-odoo`.
  - Integration into the project's skill registry.
  - Support for local Odoo monorepo discovery (CE, EE, OCA, o-spreadsheet, OWL).
- **Out**: 
  - Remote Odoo instance searching.
  - Auto-discovery of Odoo paths (paths are assumed to be standard as per source material).

## Approach
- Create a new skill file following the Architect-AI skill specification.
- Use the provided `rg` patterns for backend (ORM), frontend (OWL), and spreadsheet logic.
- Update `.atl/skill-registry.md` to include the new skill.
- Implement as a local-only capability to provide high-fidelity evidence for Odoo tasks.

## Affected Areas
- `.atl/skill-registry.md` (Update)
- `internal/assets/overlays/odoo-development-skill/skills/ripgrep-odoo/SKILL.md` (New)

## Rollback Plan
- Revert changes to `.atl/skill-registry.md`.
- Delete `internal/assets/overlays/odoo-development-skill/skills/ripgrep-odoo/` directory.

## Success Criteria
- `ripgrep-odoo` skill is registered and discoverable.
- Sub-agents can use the skill to locate Odoo core logic.
- Verification tests pass for skill loading.

## Capabilities
- **New**: `odoo-discovery` (local).
- **Modified**: `skill-registry`.
