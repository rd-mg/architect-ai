# Specification: 06-force-english-agents

## Requirement: Strict English Mandate
- ALL sub-agents MUST reason and respond in English.
- ALL generated artifacts (specs, designs, tasks, reports) MUST be in English.
- This mandate OVERRIDES any language detection from user prompts.

## Requirement: Orchestrator Alignment
- The orchestrator (Architect) MUST respond in English for this project, following the user's explicit preference for a "Global English" environment.

## Requirement: Enforcement Mechanism
- The orchestrator MUST inject a "Language Mandate" block at the top of every sub-agent delegation prompt.
- Sub-agent return envelopes MUST be validated for English content.
