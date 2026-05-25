# Specification: Phase 13 — Harness Infrastructure

## Requirements

### 1. Session YAML — `.atl/session.yaml`
- MUST persist: `execution_mode`, `delivery_strategy`, `artifact_store`, `tdd_mode`, `active_change`, `active_phase`, `active_agent`, MCP availability, and session history (last 5).
- L0 MUST read `session.yaml` on startup and NEVER re-ask preferences if they exist.
- On session end: update `last_updated`, append to history (keep last 5), record `summary_key`.
- User can override at any time: "set mode automatic" → updates immediately.
- Atomic writes: write to `.tmp` then `os.Rename()`.

### 2. Backup Harness
- MUST create timestamped backup BEFORE any mutation of critical files: `opencode.json`, `CLAUDE.md`, `GEMINI.md`, `.github/copilot-instructions.md`, `.antigravity/agent.md`, `.vscode/mcp.json`, `.claude/settings.json`, `.gemini/settings.json`, `.atl/sdd-state.yaml`, `.atl/session.yaml`, `.atl/skill-manifest.yaml`.
- Retention: max 5 backups per file, auto-purge older on new backup.
- Location: `.atl/backups/` (gitignored).
- Format: `{filename}.{YYYYMMDD-HHMMSS}.bak`.
- Go implementation: `backup.Manager` with `BackupBefore(targetPath)`, `Restore(targetPath)`, `ListBackups(basename)`.

### 3. Rollback Harness
- `rollback-apply.sh {change_name}`: checkout original branch, preserve apply branch for inspection, reset sdd-state.yaml to pending, archive apply-progress.yaml.
- `architect-ai restore {filename}`: restore most recent backup from `.atl/backups/`.
- Decision table for: sdd-apply failed, sync corruption, pre-commit partial state, installer partial write.

### 4. Component Dependency Tree
- `resolve-task-order.py`: topological sort via Kahn's algorithm on `apply-progress.yaml` tasks.
- Detect circular dependencies → exit 1 with error.
- Output: numbered execution order with depends_on and parallel_with info.

### 5. Skill Resolution Feedback Protocol
- Every sub-agent Result Contract MUST include `skill_resolution` with: `status`, `skills_used`, `tier_2_activated`, `foundation_hash`, `fallback_reason`.
- Status values: `paths-injected` (ok), `fallback-registry`, `fallback-path`, `none`.
- On non-`paths-injected`: log to Engram, rebuild skill-manifest + foundation.md, retry ONCE.
- If retry also fails and D5≥2: BLOCKED + escalate. If D5<2: WARN + proceed.

### 6. Golden Tests
- `AssertGoldenJSON(t, name, actual)`: compare generated JSON against `.golden` file.
- `AssertGoldenMD(t, name, actual)`: compare generated Markdown against `.golden` file.
- `UPDATE_GOLDEN=1` env var to regenerate golden files.
- Directory: `.atl/tests/golden/{platform}/` for all 5 platforms.
- E2E tests: `test-sdd-flow.sh`, `test-rollback.sh`, `test-review-workload.sh`.

### 7. Go Session Generator
- `session.InitialSession(project)` returns starter YAML content.
- `session.WriteSession(atDir, content)` writes atomically.
- `session.InstallRollbackScripts(atDir)` writes `rollback-apply.sh`, `resolve-task-order.py`, `backup-before-mutate.sh`.

## Scenarios

### Scenario 1: Session YAML Persistence
**Given** user selects `execution_mode="automatic"` in session 1.
**Then** `.atl/session.yaml` contains `execution_mode: "automatic"`.
**When** session 2 starts.
**Then** L0 reads session.yaml, does NOT ask for execution_mode again.

### Scenario 2: Backup Before Sync
**Given** `architect-ai sync` mutates CLAUDE.md.
**Then** `.atl/backups/CLAUDE.md.{timestamp}.bak` created BEFORE mutation with original content.

### Scenario 3: Restore After Corruption
**Given** CLAUDE.md corrupted to `"{CORRUPT JSON"`.
**When** `architect-ai restore CLAUDE.md` runs.
**Then** CLAUDE.md restored to last backup content. No `.restore.tmp` left behind.

### Scenario 4: Task Dependency Order
**Given** T1 and T2 independent, T3 depends on T1+T2, T4 depends on T3.
**When** `python3 .atl/scripts/resolve-task-order.py` runs.
**Then** T1 and T2 appear before T3, T3 before T4.

### Scenario 5: Skill Resolution Auto-Correction
**Given** sdd-verify returns `skill_resolution.status = "fallback-registry"`.
**Then** orchestrator re-reads skill-manifest.yaml, regenerates foundation.md, retries sdd-verify ONCE.

### Scenario 6: Golden Test Update/Verify
**Given** `UPDATE_GOLDEN=1 go test ./internal/testing/golden/...`
**Then** golden files created.
**When** `go test` runs again without UPDATE_GOLDEN.
**Then** tests pass (golden files match).

## Verification Criteria

| Test | Input | Expected | PASS if |
|---|---|---|---|
| Session persistence | Set automatic, restart | No re-ask | Mode question once across 2 sessions |
| Backup created | sync CLAUDE.md | Backup exists before mutation | Timestamp backup + original content |
| Restore works | Corrupt + restore | Original content back | No .tmp left |
| Task order | T1∥T2→T3→T4 | Topological order | T1,T2 before T3 before T4 |
| Skill feedback | fallback-registry | Re-read manifest, retry | Second attempt paths-injected |
| Golden tests | UPDATE + verify | Match | No failures on second run |
