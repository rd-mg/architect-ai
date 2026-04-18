# OpenSpec State Recovery Runbook

If `architect-ai sdd-status` reports that a `state.yaml` file is invalid, follow these steps to restore correctness.

## Common Errors

### 1. `schema_version unsupported`
- **Cause**: The file was created by a newer version of `architect-ai` or manually edited with a wrong version.
- **Fix**: Check your `architect-ai version`. If it's old, upgrade. If the file is wrong, set `schema_version: 1`.

### 2. `change_name mismatch`
- **Cause**: The `change_name` field in `state.yaml` does not match the name of the directory it resides in.
- **Fix**: Rename the directory or update the `change_name` field to match. Kebab-case is mandatory.

### 3. `required field missing`
- **Cause**: An agent crashed mid-write or a manual edit missed a field (e.g., `updated_at`, `status`).
- **Fix**: Add the missing field. Use RFC 3339 format for timestamps (e.g., `2026-04-18T12:00:00Z`).

### 4. `value not in allowed enum`
- **Cause**: Typo in `status` or `artifact_store`.
- **Fix**: Correct the value. 
  - Status: `pending`, `in_progress`, `completed`, `skipped`, `failed`.
  - Store: `engram`, `openspec`, `hybrid`, `none`.

## Manual Reset (Last Resort)

If the file is severely corrupted and you know the current state:
1. Delete `state.yaml`.
2. Run `architect-ai sdd-status {change-name}` (it will report missing).
3. Re-propose or manually recreate a minimal valid `state.yaml` following the canonical example in `internal/assets/skills/_shared/openspec-convention.md`.
