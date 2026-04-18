# Design: Odoo Overlay DDD Tactical Patterns

## Technical Approach

We are introducing a new architectural guidance skill (`patterns-ddd`) for Odoo developers. This skill will be part of the `odoo-development-skill` overlay. By placing it in the `skills/` directory of the overlay, it will be automatically detected and registered by the `architect-ai overlay install` and `skill-registry` systems.

## Architecture Decisions

### Decision: Placement of DDD Patterns
**Choice**: Create a dedicated `patterns-ddd` skill bundle in `internal/assets/overlays/odoo-development-skill/skills/`.
**Alternatives considered**: Add to `patterns-agnostic`.
**Rationale**: DDD patterns are specialized and shouldn't clutter the agnostic reference. Separation allows for cleaner version porting and specific triggering.

### Decision: Supplement Integration
**Choice**: Explicitly reference `patterns-ddd` in `sdd-supplements/design-odoo.md` and `apply-odoo.md`.
**Alternatives considered**: Rely on auto-triggering alone.
**Rationale**: Explicitly mentioning the skill in Odoo-specific supplements ensures that agents working in Odoo contexts are prompted to use it, reinforcing the "Directive #2" mandate.

## Data Flow

    Sub-agent Starts ──→ sdd-supplements/design-odoo.md
                                │
                                └──→ Loads patterns-ddd/SKILL.md (via Skill Registry)
                                            │
                                            └──→ Design uses Aggregate/Value Object patterns

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/assets/overlays/odoo-development-skill/skills/patterns-ddd/SKILL.md` | Create | Primary DDD tactical guidance for Odoo. |
| `internal/assets/overlays/odoo-development-skill/sdd-supplements/design-odoo.md` | Modify | Add cross-reference to patterns-ddd. |
| `internal/assets/overlays/odoo-development-skill/sdd-supplements/apply-odoo.md` | Modify | Add cross-reference to patterns-ddd. |

## Interfaces / Contracts

The new skill MUST use the following metadata to ensure correct triggering:
```yaml
name: patterns-ddd
Trigger: "When implementing domain logic, aggregates, invariants, value objects, or complex business rules in Odoo."
globs: "**/*.{py,xml}"
```

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Asset | SKILL.md Syntax | Verify markdown structure and code block validity. |
| Integration | Registration | Run `architect-ai overlay refresh odoo-development-skill` and verify `.atl/skill-registry.md` includes `patterns-ddd`. |

## Migration / Rollout

No data migration required. The skill becomes available immediately upon overlay refresh.

## Open Questions

- [ ] Should we also create version-specific ports (patterns-ddd-17, patterns-ddd-18)?
  - *Decision*: Start with one agnostic bundle as DDD principles are largely version-stable across Odoo 17/18.
