# Walkthrough: TOPIC-09 — OpenSpec `state.yaml` Schema + Validator

I have implemented the canonical OpenSpec state management system. This ensures that every change has a versioned, validated state file that prevents system drift and allows for robust agent recovery.

## Changes Made

### 1. Core State Logic
Created [state.go](file:///home/rdmachadog/gitproj/architect-ai/internal/components/openspec/state.go) which handles:
- **V1 Schema Enforcement**: Validates all required fields and enum values.
- **DAG Validation**: Implements Kahn's algorithm to ensure the phase dependency graph is cycle-free.
- **Atomic Persistence**: Uses a write-to-tmp-then-rename pattern to prevent file corruption during crashes.

### 2. CLI Command
Implemented `architect-ai sdd-status`:
- Provides a summary of active changes.
- Validates and renders a detailed phase table for a specific change.
- Integrates with the main application entry point.

### 3. Governance & Skills
- **Shared Conventions**: Updated [openspec-convention.md](file:///home/rdmachadog/gitproj/architect-ai/internal/assets/skills/_shared/openspec-convention.md) with 12 strict invariants (I1..I12).
- **Sub-Agent Protocols**: Updated [sdd-phase-common.md](file:///home/rdmachadog/gitproj/architect-ai/internal/assets/skills/_shared/sdd-phase-common.md) and [sdd-propose.md](file:///home/rdmachadog/gitproj/architect-ai/internal/assets/claude/sdd-phase-protocols/sdd-propose.md) to mandate state maintenance and validation.
- **Bootstrapping**: Updated [sdd-propose skill](file:///home/rdmachadog/gitproj/architect-ai/internal/assets/skills/sdd-propose/SKILL.md) to initialize `state.yaml` on new change creation.

### 4. Recovery
Created [docs/openspec-state-recovery.md](file:///home/rdmachadog/gitproj/architect-ai/docs/openspec-state-recovery.md) to guide users through manual repair of corrupted state files.

## Verification Results

### Automated Tests
Ran unit tests for the openspec component:
```bash
go test ./internal/components/openspec/... -v
```
**Result**: `PASS` (All 12 tests passing, including cycle detection and round-trip persistence).

### Manual Audit
Verified `sdd-status` output with valid and invalid fixtures:
- **Listing**: Successfully lists folders in `openspec/changes/`.
- **Validation**: Correctly identifies missing files, corrupt YAML, and schema violations.
- **Rendering**: Produces a clean, sorted table of phase statuses.

## Next Steps
The platform now has the governance layer required for the next topics:
- **TOPIC-10**: Delta-Spec Merge Logic (utilizing the validated state).
- **TOPIC-13**: OpenSpec Registry (utilizing the state index).
