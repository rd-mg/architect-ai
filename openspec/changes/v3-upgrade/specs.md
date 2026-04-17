# Specs: architect-ai V3 Upgrade

## Scenario 1: Unified Adaptive Reasoning
- GIVEN the adaptive-reasoning skill is loaded
- WHEN a task requires adversarial review (judgment-day)
- THEN the skill SHOULD provide inline reasoning prompts without delegating to a separate skill file.

## Scenario 2: Cognitive Postures
- GIVEN the cognitive-mode skill is active
- WHEN the orchestrator is in 'sdd-explore' phase
- THEN it SHOULD adopt the 'Investigator' posture.

## Scenario 3: Odoo Version Bundling
- GIVEN a project with Odoo 18.0
- WHEN 'architect-ai overlay install odoo-development-skill' is run
- THEN only the 'patterns-agnostic' and 'patterns-18' bundles SHOULD be installed in .agent/skills/.
