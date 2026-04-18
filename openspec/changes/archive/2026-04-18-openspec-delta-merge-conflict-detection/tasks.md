# Tasks: TOPIC-10 — OpenSpec Delta-Merge Conflict Detection

- [x] **Phase 2: Core Implementation**
    - [x] Create `internal/components/openspec/merge.go` (SHA, parser, check logic)
    - [x] Create `internal/components/openspec/merge_test.go` (Unit tests)
    - [x] Verify with `go test`
- [x] **Phase 3: CLI & Integration**
    - [x] Create `internal/cli/sdd_archive_preflight.go`
    - [x] Register in `internal/app/app.go`
    - [x] Update `sdd-spec` SKILL.md (front-matter emission)
    - [x] Update `sdd-archive` SKILL.md (preflight gate)
- [x] **Phase 4: Verification & Docs**
    - [x] Create `docs/openspec-merge-conflict.md` (Runbook)
    - [x] Integration test: simulate conflict and verify `merge-conflict.md` generation
    - [x] Final sweep and archive
