# Proposal: 08-odoo-supplements-v2-hardening

## Intent
Harden the Odoo-specific SDD supplements (`design-odoo.md`, `verify-odoo.md`, and `domain-map.md`) by integrating version 2 architectural mandates, specifically collision detection references and high-risk model auditing.

## Scope
- **design-odoo.md**: Add `before_model` collision check requirement.
- **verify-odoo.md**: Explicitly mandate Judgement Day Gate execution in the verification workflow.
- **domain-map.md**: Add a "High-Risk Models & Conflicts" section to track models requiring strict oversight (res.partner, account.move, etc.).

## Success Criteria
- [ ] `design-odoo.md` instructs the agent to check Engram for collisions before designing.
- [ ] `verify-odoo.md` checklist includes the Judgement Day Gate.
- [ ] `domain-map.md` contains a list of high-risk models with verification criteria.

## Approach
1. **Refactor design-odoo**: Inject collision check logic into the design protocol.
2. **Refactor verify-odoo**: Formalize the Judgement Day step in the deterministic checklist.
3. **Refactor domain-map**: Add the DDD High-Risk Model list.
