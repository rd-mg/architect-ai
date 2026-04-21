# Design: 08-odoo-supplements-v2-hardening

## Component: design-odoo.md Update
Inject a new section `## Global Collision Check (MANDATORY)` after the intro.
- Instruction: Call `mem_search` for global decisions or existing patterns for the target model.
- Requirement: If a collision is found, transition to `+++Autoreason-lite` posture.

## Component: verify-odoo.md Update
Update the `Deterministic Checklist` to include:
- `[ ] Judgement Day Gate executed and PASSED (Required for COMPLETE status)`.

## Component: domain-map.md Update
Add section `## High-Risk Models & Conflict Checklist`.
- Table: Model, Risk, Verification Mandatory.
- Models: res.partner (Locking/Integrity), account.move (Performance/N+1), stock.move (Accounting consistency).

## Implementation Strategy
- Multi-replace edits to the 3 supplement files.
