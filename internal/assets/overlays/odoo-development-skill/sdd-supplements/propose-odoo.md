# SDD Propose — Odoo Context

Follow IN ADDITION to standard sdd-propose.

## Odoo Version
{See shared preamble — version cached from sdd-init.}

## Scope Framing

Every Odoo proposal MUST specify:

- **Target Odoo version(s)**: v14, v15, v16, v17, v18, v19, or combination
- **Module scope**: new module, extension of existing, or modification
- **Impact domain(s)**: Sales, Inventory, Accounting, HR, CRM, Purchase, MRP, Website, POS, Project
- **Edition**: Community only, Enterprise only, or both

## Module Naming

If NEW module proposed, follow naming convention:

### Client Modules
Format: `{client_prefix}_{core_app}_{descriptive_name}`
Example: `acme_sale_custom_approval`

### Internal Modules
Format: `{org_prefix}_{core_app}_{descriptive_name}`
Example: `cudio_api_connector` (Cudio-specific — see cudio-dod-addendum)

Rules:
- Only lowercase letters, numbers, underscores
- Folder name MUST match technical name
- Manifest `name` field: "{Customer} | {Human Title}"

## Manifest Planning

In proposal, specify what `__manifest__.py` will contain:

```markdown
## Manifest Plan
- name: "{Customer} | {Module Title}"
- version: {new-version} (bumped from {old-version})
- category: {category}
- depends: [{list of dependencies}]
- external_dependencies: {if any}
- data: [{list of XML/CSV files to load, in correct order}]
- license: {OPL-1 | LGPL-3 | AGPL-3}
```

## Don't-Reinvent-the-Wheel Validation

Before proposing new code, cite exploration findings:

```markdown
## Reuse Analysis
- Checked Odoo core: {finding}
- Checked OCA: {finding}
- Decision: {inherit-extend | reference-only | new-from-scratch}
- Justification: {1-2 sentences}
```

If proposing new-from-scratch without reuse analysis, REJECT and return to sdd-explore.

## Rollback Plan

In Odoo, rollback is trickier than "git revert" because of:
- Database migrations already run
- Data created by new feature
- User-configured settings

Rollback plan MUST address:
1. **Code rollback**: how to undo code deployment
2. **Data rollback**: what migration scripts run in reverse
3. **Setting rollback**: which `ir.config_parameter` entries to clean up
4. **Module uninstall behavior**: does uninstall leave orphan data?

## Capabilities Section

Define capabilities with Odoo-specific precision:

```markdown
## Capabilities

### Capability 1: {name}
- Purpose: {1 sentence}
- Models affected: {list of model names}
- New models: {list with _name values}
- Fields added: {list of (model, field, type)}
- Views modified: {list of (view type, xml id)}
- Security impact: {new groups, access rules, record rules}
```

## Size Budget

Respect 450-word limit from standard sdd-propose protocol. Odoo proposals tend to sprawl — compress aggressively.

## Boundaries

- Do NOT propose implementation code in this phase (that's sdd-apply's job)
- Do NOT commit to specific migration SQL (that's sdd-design's detail)
- Do NOT skip reuse analysis section
