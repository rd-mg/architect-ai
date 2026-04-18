# Adaptive Reasoning Gate Specification

## Purpose
Enforce mandatory reasoning depth classification before sub-agent execution.

## Requirements

### Requirement: Structural Injection
The system MUST inject the adaptive reasoning gate content between the cognitive posture and project standards blocks in the sub-agent launch template.

#### Scenario: Sub-agent launch
- GIVEN a sub-agent is about to be launched
- WHEN the orchestrator builds the prompt
- THEN the `Adaptive Reasoning (MANDATORY)` section MUST be present before `Project Standards`.

### Requirement: Mode Declaration
Sub-agents MUST declare their chosen mode in the first line of their response using the format `Mode: {n}. Why: {reason}.`.

#### Scenario: Compliant response
- GIVEN a sub-agent has received the gate instruction
- WHEN it produces output
- THEN the first non-blank line MUST match the mode declaration pattern.

### Requirement: Mode Parsing Tolerance
The orchestrator MUST tolerate mode declarations within the first 5 lines of the response to account for minor model preambles.

#### Scenario: Preamble in response
- GIVEN a sub-agent response with a one-sentence preamble
- WHEN the orchestrator parses the output
- THEN it MUST find and validate the `Mode: {n}` line if present within the first 5 lines.

## Invariants
- The gate content MUST be sourced from `internal/assets/skills/_shared/adaptive-reasoning-gate.md`.
- No Go code changes are required for the classifier itself.
