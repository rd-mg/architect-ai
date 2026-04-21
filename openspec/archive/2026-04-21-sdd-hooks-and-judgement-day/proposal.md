# Proposal: 05-sdd-hooks-and-judgement-day

## Intent
Implement and harden the SDD Hook system (`before_model` and `after_model`) and integrate the "Judgement Day" audit gate into the `sdd-verify` phase across all supported IDE and CLI orchestrators.

## Scope
- **Orchestrator Assets**: Update 11 orchestrator assets in `internal/assets/` to include hook logic.
- **Odoo Overlay**: Update `internal/assets/overlays/odoo-development-skill/sdd-supplements/verify-odoo.md` to include Judgement Day Gate.
- **Phase Protocols**: Potential updates to `sdd-verify.md` if global gate logic is required.

## Success Criteria
- [ ] All 11 orchestrators define `before_model` hook with collision check, state injection, and error context.
- [ ] All 11 orchestrators define `after_model` hook for automatic persistence of state, patterns, and briefs.
- [ ] Odoo overlay includes the Judgement Day Gate protocol.
- [ ] State machine handles `FAIL` verdict from Judgement Day by returning to `sdd-design`.

## Approach
1. **Hooks Specification**: Define the precise logic for `before_model` (collision check) and `after_model` (persistence).
2. **Orchestrator Update**: Apply the hooks logic to all orchestrator files using a template-based approach to ensure consistency.
3. **Judgement Day Integration**: Add the audit gate to the Odoo overlay's verification supplement.
4. **State Machine hardening**: Ensure the orchestrator logic correctly routes Judgement Day failures.

## Risks
- **Hook Complexity**: Over-complex collision checks could increase latency or token consumption.
- **Divergence**: Manually updating 11 files risks inconsistency if not carefully checked.
