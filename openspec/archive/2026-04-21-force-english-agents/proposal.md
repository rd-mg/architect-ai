# Proposal: 06-force-english-agents

## Intent
Enforce a strict English-only policy for all agent reasoning, communication, and artifacts, regardless of the input language from the user.

## Scope
- **All 11 Orchestrator Assets**: Update `internal/assets/*/sdd-orchestrator.md` to mandate English for all sub-delegations.
- **Phase Protocols**: Update global phase instructions to include the language mandate.
- **Architect Persona**: Transition the orchestrator interface to English per user request.

## Success Criteria
- [ ] Sub-agents always reason and respond in English.
- [ ] Artifacts (Specs, Designs, Tasks) are generated in English only.
- [ ] Orchestrator responses are in English even when prompted in Spanish.

## Approach
1. Update `sdd-orchestrator.md` files to include a `## Language Mandate` section in the launch template.
2. Update `sdd-phase-common.md` (if applicable) or core protocols to reinforce the requirement.
