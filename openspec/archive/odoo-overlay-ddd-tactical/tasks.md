# Tasks: Odoo Overlay DDD Tactical Patterns

## Phase 1: Asset Preparation
- [ ] Create skill bundle directory: `internal/assets/overlays/odoo-development-skill/skills/patterns-ddd/`
- [ ] Implement `SKILL.md` with 6 tactical patterns (Aggregate, VO, Service, Event, Repository, Spec)
- [ ] Include Odoo 19 `odoo.fields.Domain` guidance in the Spec pattern section
- [ ] Add 4 "Anti-patterns" (Non-applicable DDD concepts in Odoo)

## Phase 2: Supplement Wiring
- [ ] Update `internal/assets/overlays/odoo-development-skill/sdd-supplements/design-odoo.md` to link `patterns-ddd`
- [ ] Update `internal/assets/overlays/odoo-development-skill/sdd-supplements/apply-odoo.md` to link `patterns-ddd`

## Phase 3: Registration & Verification
- [ ] Run `architect-ai overlay refresh odoo-development-skill`
- [ ] Verify `patterns-ddd` appears in `.atl/skill-registry.md` under Overlay Skills
- [ ] Verify trigger works: run a mock design task for a new Odoo model and check if the skill is loaded

## Phase 4: Archiving
- [ ] Rebase delta specs into main `openspec/specs/odoo-ddd-tactical/spec.md` (if any remain)
- [ ] Archive change
