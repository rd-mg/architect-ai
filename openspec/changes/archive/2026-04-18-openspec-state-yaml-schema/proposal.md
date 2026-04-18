# TOPIC-09 — OpenSpec `state.yaml` Schema + Validator

Canonicalize the `state.yaml` file used by OpenSpec for phase-level state tracking. This ensures that session recovery and archive operations are reliable and that malformed state is detected immediately rather than causing silent drift.

## User Review Required

> [!IMPORTANT]
> **Hard Gate on Corruption**: The validator will act as a hard gate. If `state.yaml` is malformed, the CLI and Agents will refuse to proceed. Users will need to follow a manual recovery runbook.

## Proposed Changes

### [OpenSpec Convention]

#### [MODIFY] [openspec-convention.md](file:///home/rdmachadog/gitproj/architect-ai/internal/assets/skills/_shared/openspec-convention.md)
- Append the `## state.yaml Schema (V1)` section defining fields, enums, and the atomic write requirement.

---

### [Go Components]

#### [NEW] [state.go](file:///home/rdmachadog/gitproj/architect-ai/internal/components/openspec/state.go)
- Define `State`, `Phase`, and `Metering` structs with YAML tags.
- Implement `Load`, `Save` (atomic), and `Validate` functions.
- Define typed errors for specific invariant violations.

#### [NEW] [state_test.go](file:///home/rdmachadog/gitproj/architect-ai/internal/components/openspec/state_test.go)
- Unit tests for validation logic, round-trip serialization, and error conditions.

---

### [CLI]

#### [NEW] [sdd_status.go](file:///home/rdmachadog/gitproj/architect-ai/internal/cli/sdd_status.go)
- Implement `RunSDDStatus` to read and validate `state.yaml` for a given change.
- Print a status table on success or a detailed error on failure.

#### [MODIFY] [app.go](file:///home/rdmachadog/gitproj/architect-ai/internal/app/app.go)
- Register the `sdd-status` subcommand.

---

### [Documentation]

#### [NEW] [openspec-state-recovery.md](file:///home/rdmachadog/gitproj/architect-ai/docs/openspec-state-recovery.md)
- Runbook for manual state recovery if validation fails.

## Verification Plan

### Automated Tests
- `go test ./internal/components/openspec/... -v`
- `go test ./internal/cli/...` (for command registration)

### Manual Verification
1. Run `architect-ai sdd-status {change-name}` on a valid change.
2. Manually corrupt a `state.yaml` (e.g., change `schema_version` to 2 or remove a required field) and verify that `sdd-status` fails with a clear error.
3. Verify that `sdd-init` (in next phase) creates a valid initial `state.yaml`.
