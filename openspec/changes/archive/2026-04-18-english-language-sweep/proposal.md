# Proposal: English Language Sweep

## Intent
Remove Spanish language leaks from framework voice while preserving trigger phrases for UX.

## Scope
- internal/assets/ (all agents)
- docs/ and top-level documentation
- CI/CD pipeline (linting)
- scripts/ (new lint script)

## Approach
Define a language policy, perform a manual sweep of Spanish voice, and implement a CI lint guard to prevent regression.

## Risks
- Potential for false positives in linting
- Regression in UX for Spanish-speaking users if trigger phrases are accidentally removed
