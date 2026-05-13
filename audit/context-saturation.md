# Audit: Context Saturation & Injection Waste

## Expansion Inventory
| Pattern | File | Conditional? | Est. Tokens |
|---|---|---|---|
| {content of _shared/context-mode-routing-policy.md} | Both | No | 300 |
| {content from skills/_shared/general-phase-common.md} | Gen | No | 300 |
| {content of _shared/research-routing.md} | SDD | No | 300 |
| {instructions from sdd-phase-protocols/{phase}.md} | SDD | Yes | 300 |
| [content of .atl/overlays/odoo-18/sdd-supplements/verify-odoo.md] | SDD | Yes | 300 |

## Per-Agent Waste
- General Orchestrator template: 600 tokens
- SDD Orchestrator template: 1200-1500 tokens

## Session Waste
- 8 sub-agent calls × ~1350 tokens = ~10,800 wasted tokens/session
- Fix: tiered injection (compact rules only for relevant skill), conditional expansions

## Verdict
MASTER-PLAN Phase 4 claim VALIDATED
