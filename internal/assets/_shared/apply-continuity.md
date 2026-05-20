# Apply Continuity Protocol

To prevent full restarts when `sdd-apply` is interrupted, progress is tracked at the task level using `.atl/apply-progress.yaml`.

## Configuration Structure

The `.atl/apply-progress.yaml` file tracks the progress of the active change's tasks:

```yaml
change_name: "phase-02-sdd-worktrees"
started_at: "2026-05-20T00:52:04Z"
updated_at: "2026-05-20T00:52:04Z"
tasks:
  - id: "task 10"
    description: "Implement state writer"
    status: "completed"  # pending | running | completed | failed
    completed_at: "2026-05-20T00:53:50Z"
  - id: "task 11"
    description: "Write tests for state writer"
    status: "completed"
    completed_at: "2026-05-20T00:53:51Z"
  - id: "task 12"
    description: "Create state template"
    status: "pending"
    completed_at: ""
```

## Resumption Logic

1. **Check Existing Progress**: When `sdd-apply` starts, look for `.atl/apply-progress.yaml`.
2. **Validate Change Name**:
   - If the file exists and `change_name` matches the current change, load it.
   - If `change_name` does not match, archive the old file to `.atl/apply-progress.yaml.old` and initialize a fresh state.
3. **Execution**:
   - Skip all tasks marked `completed`.
   - Resume execution from the first task marked `pending` or `running`.
4. **Atomic Checkpointing**:
   - Update `.atl/apply-progress.yaml` immediately and atomically after each task completes.
   - Report progress output formatting: `[apply] T{N}/{M} complete — {description} | {lines_added} lines`.
