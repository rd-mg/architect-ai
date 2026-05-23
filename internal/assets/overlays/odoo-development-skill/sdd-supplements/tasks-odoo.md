# SDD Tasks — Odoo Context

## Atomic Task Definition (Odoo)

One Odoo atomic task = one independently committable unit:

| Component | Atomic Unit | Example Task Title |
|-----------|-------------|-------------------|
| Model | One `.py` file (all fields/methods for one model) | "Add approval.request model with state field" |
| View | One XML file per model | "Add form/list views for approval.request" |
| Security | One `ir.model.access.csv` entry set | "Add ACL entries for approval.request (user/manager)" |
| Migration | One version folder with pre/post scripts | "Add migration script v18.0.1.0.0→18.0.2.0.0" |
| Controller | One `controllers/{name}.py` | "Add /api/approval/request endpoint" |
| Test | One `tests/test_{feature}.py` | "Add unit tests for approval workflow" |
| Manifest | Bump version + update data[] | "Bump version to 18.0.2.0.0, add views to data[]" |

## Task Dependencies (Odoo)

Always sequence: Security → Model → Migration → Views → Controller → Tests → Manifest

NEVER batch model + views in one task — different risk profiles.

## Manifest Task (ALWAYS last)

Every task batch MUST end with "Bump manifest version" task. Never optional.

## Test Task Pairing

STRICT TDD: Every model/controller task paired with test task immediately after.

## Forbidden Task Patterns
- "Implement approval module" — too coarse; violates atomicity
- "Fix bug and add feature" — two tasks masquerading as one
- "Write all views" — must be split per model
