# Implementation Plan — TOPIC-10: OpenSpec Delta-Merge Conflict Detection

Implementing a SHA-256 based conflict detection system for OpenSpec delta specs to prevent silent data loss during concurrent or delayed changes.

## User Review Required

> [!IMPORTANT]
> This change introduces a **hard gate** for `sdd-archive`. Any change that modifies a spec that has changed since the delta was written will fail archival until rebased.

> [!NOTE]
> Delta specs will now include YAML front-matter. This is transparent to most markdown renderers but required for SDD tooling.

## Proposed Changes

### [OpenSpec Core Component]

#### [MODIFY] [merge.go](file:///home/rdmachadog/gitproj/architect-ai/internal/components/openspec/merge.go) [NEW]
- Implement `SHAOfFile` (SHA-256 hex).
- Implement `ReadDeltaFrontMatter` and `WriteDeltaFrontMatter`.
- Implement `CheckConflict` logic comparing delta's `base_sha` with main-spec's current SHA.
- Implement `WriteConflictReport` for generating human-readable recovery docs.

#### [MODIFY] [merge_test.go](file:///home/rdmachadog/gitproj/architect-ai/internal/components/openspec/merge_test.go) [NEW]
- Unit tests for SHA computation, front-matter parsing, and conflict detection.
- Coverage for new-capability sentinel (`base_sha: "0"`).

---

### [CLI Layer]

#### [MODIFY] [sdd_archive_preflight.go](file:///home/rdmachadog/gitproj/architect-ai/internal/cli/sdd_archive_preflight.go) [NEW]
- Subcommand to run `CheckConflict` in dry-run mode.
- Writes `merge-conflict.md` and fails `state.yaml` archive phase on mismatch.

#### [MODIFY] [app.go](file:///home/rdmachadog/gitproj/architect-ai/internal/app/app.go)
- Register `sdd-archive-preflight` subcommand.

---

### [Agent Skills & Documentation]

#### [MODIFY] [sdd-spec/SKILL.md](file:///home/rdmachadog/gitproj/architect-ai/internal/assets/skills/sdd-spec/SKILL.md)
- Update instructions to compute and prepend `base_sha` front-matter to every delta.

#### [MODIFY] [sdd-archive/SKILL.md](file:///home/rdmachadog/gitproj/architect-ai/internal/assets/skills/sdd-archive/SKILL.md)
- Update instructions to run `sdd-archive-preflight` before merge.
- Update merge step to strip front-matter before writing to main spec.

#### [MODIFY] [openspec-merge-conflict.md](file:///home/rdmachadog/gitproj/architect-ai/docs/openspec-merge-conflict.md) [NEW]
- Recovery runbook for users encountering conflicts.

## Verification Plan

### Automated Tests
- `go test ./internal/components/openspec/...`
- Integration test: manual simulation of concurrent changes and verification that preflight fails as expected.

### Manual Verification
1. Create a change, modify a spec.
2. Manually modify the main spec on disk.
3. Run `architect-ai sdd-archive-preflight` and verify it detects the conflict and writes the report.
