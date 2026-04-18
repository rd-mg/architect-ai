# Proposal: OpenSpec Mode-Branching Consolidation

## Intent
Consolidate redundant persistence branching logic across 10 SDD skills into a single shared resource to eliminate drift and reduce maintenance overhead.

## Scope
- Create `internal/assets/skills/_shared/mode-branching.md`.
- Refactor 10 `SKILL.md` files (sdd-propose, sdd-spec, sdd-design, sdd-tasks, sdd-apply, sdd-verify, sdd-archive, sdd-explore, sdd-onboard, sdd-init) to reference the shared block.

## Approach (Option B)
- **Centralization**: Move the canonical 50-line persistence logic (covering engram, openspec, hybrid, and none modes) to `_shared/mode-branching.md`.
- **Referencing**: Replace the inline blocks in each skill with a 3-line reference and skill-specific artifact metadata.

## Success Criteria
- `_shared/mode-branching.md` exists and contains the canonical persistence logic.
- 10 `SKILL.md` files reference the shared file.
- No functional regression in how skills handle different persistence modes.
- `rg "mode-branching.md" internal/assets/skills/sdd-*/SKILL.md | wc -l` returns 10.
