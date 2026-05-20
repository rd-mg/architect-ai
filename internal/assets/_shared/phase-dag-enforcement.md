# Phase DAG Enforcement Protocol

Every SDD phase agent MUST perform Phase DAG checking as its first execution step.

## Enforcement Rules

1. **Locate state file**: Look for `.atl/sdd-state.yaml` in the project root. If missing:
   - Exit with status `BLOCKED`.
   - Error: `sdd-state.yaml not found. Run /sdd-init first.`
   - Exit Code: 1.
2. **Verify Prerequisites**: Read the target phase's `requires` array from `.atl/sdd-state.yaml`.
   - For each prerequisite phase listed, verify its `status` is `completed`.
   - If any prerequisite is not `completed` (e.g. `pending`, `running`, `failed`):
     - Exit with status `BLOCKED`.
     - Error: `Blocked: Prerequisite '{phase}' is '{status}', not 'completed'.`
     - Exit Code: 1.
3. **Atomic Phase Transitions**:
   - When a phase begins execution, update its status to `running`.
   - When a phase finishes execution successfully, update its status to `completed` and set `completed_at` to the current UTC timestamp (RFC 3339).
   - If a phase fails, update its status to `failed`.
   - Ensure all state file writes are atomic (write to `.atl/sdd-state.yaml.tmp` first, then rename/move, protected by `.atl/sdd-state.yaml.lock`).
