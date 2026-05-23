# Apply Continuity Protocol

Prevent full restarts when `sdd-apply` interrupted. Progress tracked at task level via `.atl/apply-progress.yaml`.

## Configuration

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

1. **Check Existing**: On `sdd-apply` start, look for `.atl/apply-progress.yaml`.
2. **Validate Change Name**:
   - File exists + `change_name` matches → load it.
   - `change_name` does not match → archive to `.atl/apply-progress.yaml.old`, init fresh.
3. **Execution**:
   - Skip all tasks marked `completed`.
   - Resume from first `pending` or `running` task.
4. **Atomic Checkpointing**:
   - Update `.atl/apply-progress.yaml` atomically after each task completes.
   - Progress output: `[apply] T{N}/{M} complete — {description} | {lines_added} lines`.
