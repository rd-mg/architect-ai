# Spec Delta: SDD Orchestration Contract

## Requirement: Terminology Split
- The system must distinguish between `SDD Bootstrap` (CLI environment setup) and `SDD Init Analysis` (AI phase execution).
- The `sdd-init` CLI command must be documented as the "Bootstrap" layer.
- The `sdd-init` skill/phase must be documented as the "Analysis" layer.

## Requirement: State Separation
- CLI Bootstrap must be represented by a persistent marker (e.g., `.atl/state/bootstrap.json` or equivalent).
- AI Init Analysis must be represented by a separate persistent marker (e.g., `.atl/state/init-analysis.json` or equivalent).
- The SDD Guard logic (`EnsureSDDReady`) must be updated to require both markers where full analysis is expected.

## Verification
- `TestRunSddInit_WritesBootstrapMarker` must pass.
- `TestSDDAnalysis_WritesAnalysisMarker` must pass.
