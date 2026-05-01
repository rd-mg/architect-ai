# SDD Spec — Odoo Context

## Acceptance Criteria Format (Odoo)

Every AC in an Odoo spec MUST be verifiable by `sdd-verify`. Use this template:

### AC Template
- **Given**: {state of Odoo database / records / config}
- **When**: {action — user action, API call, cron trigger, or install step}
- **Then**: {observable outcome — field value, view state, log entry, or migration result}
- **Version gate**: {versions this AC applies to, e.g., "v18+ only"}
- **Verifiable by**: {rg pattern | test method | manual step}

### Mandatory ACs for Schema Changes
If ANY model field is added/removed/renamed, the spec MUST include:
- AC: migration script exists at `migrations/{version}/`
- AC: required fields have default values or post-migrate population
- AC: `ir.model.access.csv` entry added for any new model

### Mandatory ACs for View Changes
- AC: XPath targets tested against actual view XML ID
- AC: No `attrs=` used (v17+), no `<tree>` (v18+)
- AC: OWL component version matches Odoo version

### Security ACs (ALWAYS required)
- AC: Access rule covers all CRUD operations for affected model
- AC: sudo() calls documented with justification comment

## Spec Size Budget
≤ 300 words for Odoo-specific section.
The spec supplement is SHORT — Odoo adds precision, not volume.

## Manifest Spec Entry
Every spec MUST declare:
```markdown
## Manifest Impact
- version bump: {X.Y.Z.W → X.Y.Z'.W'} (Z for features, W for fixes)
- data[] order changes: {list if any}
- new dependencies: {list if any}
```
