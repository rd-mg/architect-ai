# Spec Delta: SDD Orchestration Contract

## Requirement: Terminology Split
- The system must distinguish between `SDD Bootstrap` (CLI environment setup) and `SDD Init Analysis` (AI phase execution).
- The `sdd-init` CLI command must be documented as the "Bootstrap" layer.
- The `sdd-init` skill/phase must be documented as the "Analysis" layer.

## Requirement: State Separation
- CLI Bootstrap must be represented by a persistent marker (e.g., `.atl/state/bootstrap.json` or equivalent).
- AI Init Analysis must be represented by a separate persistent marker (e.g., `.atl/state/init-analysis.json` or equivalent).
- The SDD Guard logic (`EnsureSDDReady`) must be updated to require both markers where full analysis is expected.


## Requirement: Result Contract Extension
The orchestrator result contract MUST include `chosen_mode` and `mode_rationale` fields.
(Previously: Result contract only included status, summary, artifacts, etc.)

#### Scenario: Successful result processing
- GIVEN a sub-agent response with `Mode: 2. Why: High risk.`
- WHEN the orchestrator processes the result
- THEN the result envelope MUST contain `chosen_mode: "2"` and `mode_rationale: "High risk."`.

## Requirement: Mode Field Validation and Re-prompt
The orchestrator MUST validate the presence of the mode declaration and re-prompt the sub-agent exactly once if missing.

#### Scenario: Missing mode declaration
- GIVEN a sub-agent response missing the `Mode: {n}` line
- WHEN the orchestrator validates the result
- THEN it MUST send a re-prompt message requesting the mode declaration.

#### Scenario: Fallback after second failure
- GIVEN a sub-agent has failed to provide a mode declaration after a re-prompt
- WHEN the orchestrator processes the second response
- THEN it MUST record a fallback to Mode 1 in Engram and proceed.

## Verification
- `TestRunSddInit_WritesBootstrapMarker` must pass.
- `TestSDDAnalysis_WritesAnalysisMarker` must pass.
- `TestAdaptiveReasoningGateInjected` must pass for all orchestrators.
