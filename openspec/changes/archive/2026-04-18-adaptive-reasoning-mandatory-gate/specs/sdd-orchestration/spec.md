---
openspec_delta:
  base_sha: "4f1f561c45279de1a5d2a1cd068453cf60b510d5952707a147ad175587b57b72"
  base_path: "openspec/specs/sdd-orchestration/spec.md"
  base_captured_at: "2026-04-18T05:10:30Z"
  generator: sdd-spec
  generator_version: 1
---

# Delta for SDD Orchestration

## MODIFIED Requirements

### Requirement: Result Contract Extension
The orchestrator result contract MUST include `chosen_mode` and `mode_rationale` fields.
(Previously: Result contract only included status, summary, artifacts, etc.)

#### Scenario: Successful result processing
- GIVEN a sub-agent response with `Mode: 2. Why: High risk.`
- WHEN the orchestrator processes the result
- THEN the result envelope MUST contain `chosen_mode: "2"` and `mode_rationale: "High risk."`.

### Requirement: Mode Field Validation and Re-prompt
The orchestrator MUST validate the presence of the mode declaration and re-prompt the sub-agent exactly once if missing.

#### Scenario: Missing mode declaration
- GIVEN a sub-agent response missing the `Mode: {n}` line
- WHEN the orchestrator validates the result
- THEN it MUST send a re-prompt message requesting the mode declaration.

#### Scenario: Fallback after second failure
- GIVEN a sub-agent has failed to provide a mode declaration after a re-prompt
- WHEN the orchestrator processes the second response
- THEN it MUST record a fallback to Mode 1 in Engram and proceed.
