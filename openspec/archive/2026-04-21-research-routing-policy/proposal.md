# Proposal: 07-research-routing-policy

## Intent
Formalize and implement the Research-Routing Policy across all Architect-AI orchestrators. The policy enforces a "least-cost-first" search order (Engram -> ripgrep-odoo -> Context7 -> NotebookLM -> Web) and integrates mode-based restrictions to optimize token usage and accuracy.

## Scope
- **All 11 Orchestrator Assets**: Update `internal/assets/*/sdd-orchestrator.md` to include the Research-Routing Policy (Layer 5).
- **Odoo Overlay**: Hardcode local search paths to `~/gitproj/odoo/` for the `ripgrep-odoo` skill and supplements.
- **Adaptive Reasoning Integration**: Implement mode-based source restrictions in the routing logic.

## Success Criteria
- [ ] Orchestrators mandate Engram search before any external lookup.
- [ ] Sub-agents follow the 5-step routing protocol.
- [ ] Odoo-specific searches target `~/gitproj/odoo/`.
- [ ] Mode 3 (Context Saturated) correctly blocks expensive/variable research sources.

## Approach
1. **Routing Definition**: Define the Research-Routing Policy block (Layer 5) for orchestrator prompts.
2. **Orchestrator Update**: Inject the policy and mode restriction table into all 11 orchestrator assets.
3. **Odoo Path Hardening**: Update `ripgrep-odoo` skill and supplements to use the canonical Odoo path.
