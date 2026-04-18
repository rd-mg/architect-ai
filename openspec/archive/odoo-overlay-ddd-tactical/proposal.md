# Proposal: Odoo Overlay DDD Tactical Patterns

## Intent

Directive #2: Provide DDD tactical guidance translated to Odoo-native patterns. 
Teams using Architect-AI in Odoo contexts need clear mapping between DDD concepts 
(Aggregates, Value Objects, Domain Services) and Odoo ORM decorators/models 
to maintain architectural integrity without fighting the framework.

## Scope

### In Scope
- Create `internal/assets/overlays/odoo-18/patterns-ddd/SKILL.md` with 6 applicable patterns.
- Port skill to Odoo 17 (`odoo-17/`) with version-specific link adjustments.
- Update Odoo overlay manifests for v17 and v18 to include `patterns-ddd`.
- Extend `design-odoo.md` and `apply-odoo.md` supplements to reference the new skill.

### Out of Scope
- Go-side code changes (Pure asset/manifest updates).
- Non-tactical DDD patterns (Bounded Contexts, Context Mapping).
- Patterns that clash with Odoo (Builders, Unit of Work, Repositories).

## Capabilities

### New Capabilities
- odoo-ddd-tactical: Guidance on implementing DDD tactical patterns within Odoo ORM.

### Modified Capabilities
- odoo-overlay-registration: Manifest system must now account for DDD-specific skill assets.

## Approach

1. **Asset Creation**: Write the SKILL.md for odoo-18 first as the reference implementation.
2. **Version Porting**: Mirror to odoo-17 with minor adjustments.
3. **Manifest Integration**: Add entries to `registry_entries` in both version manifests.
4. **Supplement Wiring**: Add cross-references in SDD supplements to ensure agents load the DDD skill during design/apply phases.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/assets/overlays/odoo-18/patterns-ddd/` | New | Primary skill asset |
| `internal/assets/overlays/odoo-17/patterns-ddd/` | New | Ported skill asset |
| `internal/assets/overlays/odoo-18/manifest.yaml` | Modified | Register new skill |
| `internal/assets/overlays/odoo-17/manifest.yaml` | Modified | Register new skill |
| `.atl/overlays/odoo-*/sdd-supplements/` | Modified | Update design/apply cross-refs |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Pattern confusion | Low | Clear "Non-applicable" section in SKILL.md with Odoo-native rationales. |
| Manifest breaking change | Low | Ensure TOPIC-05 dependencies are met before final apply. |

## Rollback Plan

Revert manifest changes and delete the `patterns-ddd` directories. Revert supplement edits.

## Dependencies

- TOPIC-05 (Skill Registry Manifests) must be merged/available.

## Success Criteria

- [ ] `patterns-ddd/SKILL.md` exists for v17 and v18.
- [ ] Manifests correctly list `patterns-ddd` as a registered skill.
- [ ] SDD supplements reference the DDD skill in Design/Apply phases.
- [ ] Code examples in skills are syntactically valid Odoo Python.
