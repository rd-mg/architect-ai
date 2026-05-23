# Phase DAG Enforcement Protocol

Every SDD phase agent MUST perform Phase DAG checking as first execution step.

## Enforcement Rules

1. **Locate state file**: Look for `.atl/sdd-state.yaml` in project root. If missing:
   - Exit `BLOCKED` (Exit Code 1). Error: `sdd-state.yaml not found. Run /sdd-init first.`
2. **Verify Prerequisites**: Read target phase's `requires` array from `.atl/sdd-state.yaml`.
   For each prerequisite, verify `status` is `completed`. If not:
   - Exit `BLOCKED` (Exit Code 1). Error: `Blocked: Prerequisite '{phase}' is '{status}', not 'completed'.`
3. **Atomic Phase Transitions**:
   - Phase start → update status to `running`
   - Phase success → update status to `completed`, set `completed_at` to UTC timestamp (RFC 3339)
   - Phase failure → update status to `failed`
   - All writes atomic: `.atl/sdd-state.yaml.tmp` → rename, protected by `.atl/sdd-state.yaml.lock`
