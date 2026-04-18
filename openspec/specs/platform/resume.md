# Platform Spec: Apply-Progress Continuity

This specification defines the requirements for symmetric task-level progress resumption across all Architect-AI persistence modes.

## Requirements

### Requirement: Mode-Aware Resumption
- **ID**: RESUME-01
- **Statement**: Orchestrators MUST branch their apply-continuity check based on the active `artifact_store`.
- **Validation**: `rg` for mode-specific branches in orchestrator files.

### Requirement: Hybrid Authority
- **ID**: RESUME-02
- **Statement**: In `hybrid` mode, the `apply-progress.md` file on disk IS authoritative over Engram.
- **Validation**: Orchestrator instructions explicitly state "FILESYSTEM WINS".

### Requirement: Verification Guard
- **ID**: RESUME-03
- **Statement**: Orchestrators MUST block `sdd-verify` launch if `sdd-apply` status is `in_progress` or `failed`.
- **Validation**: Attempting to verify while apply is incomplete results in an orchestrator refusal.
