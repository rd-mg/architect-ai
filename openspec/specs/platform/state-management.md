# Spec: State Management (OpenSpec)

## Context
Implementing canonical state.yaml schema and validator for SDD governance.

## Requirements

### Requirement: Schema Versioning
- **ID**: STATE-01
- **Statement**: The `state.yaml` file MUST include a `schema_version` field.
- **Validation**: Current version MUST be `1`.

### Requirement: Validation Invariants
- **ID**: STATE-02
- **Statement**: The system MUST enforce 12 critical invariants (I1..I12) including change name matching, timestamp order, and DAG cycle detection.
- **Validation**: `architect-ai sdd-status` MUST report errors for any violation.

### Requirement: Atomic Writes
- **ID**: STATE-03
- **Statement**: All writes to `state.yaml` MUST be atomic.
- **Validation**: System writes to `.tmp` and uses `rename`.
