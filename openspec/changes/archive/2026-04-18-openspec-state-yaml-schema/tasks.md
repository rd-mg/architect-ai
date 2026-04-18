# Tasks: TOPIC-09 — OpenSpec `state.yaml` Schema + Validator

- [x] **Phase 1: Documentation**
    - [x] Update `internal/assets/skills/_shared/openspec-convention.md` with V1 schema
- [x] **Phase 2: Core Implementation**
    - [x] Create `internal/components/openspec/state.go` (Structs, Load, Save, Validate)
    - [x] Create `internal/components/openspec/state_test.go` (Unit tests)
    - [x] Run unit tests and ensure green
- [x] **Phase 3: CLI Integration**
    - [x] Create `internal/cli/sdd_status.go` (`RunSDDStatus` implementation)
    - [x] Register `sdd-status` in `internal/app/app.go`
    - [x] Verify `architect-ai sdd-status --help` works
- [x] **Phase 4: Verification & Polish**
    - [x] Create valid/invalid fixtures and verify `sdd-status` behavior
    - [x] Ensure `openspec-state-recovery.md` is complete and linked
    - [x] Final full test suite run
