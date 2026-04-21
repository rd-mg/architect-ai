# Specification: 08-odoo-supplements-v2-hardening

## Requirement: Design Collision Guard
- `design-odoo.md` MUST mandate a pre-design collision check using `mem_search` (Layer 0/before_model).
- Agents MUST justify inheritance vs modification based on discovered global decisions.

## Requirement: Judgement Day Integration
- `verify-odoo.md` MUST include the Judgement Day Gate as a final, non-optional step for new modules.
- Verification verdict MUST be gated by Judgement Day PASS/FAIL.

## Requirement: High-Risk Model Map
- `domain-map.md` MUST list models at high risk for performance and integrity issues:
    - `res.partner`
    - `account.move`
    - `account.move.line`
    - `stock.move`
    - `res.users`
- For each high-risk model, a "Conflict Check" protocol MUST be defined (e.g., "NEVER use direct SQL", "ALWAYS use _inherit").
