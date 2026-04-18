# Tasks: Adaptive Reasoning Mandatory Gate

## Phase 1: Foundation (Setup)

- [ ] 1.1 Create `internal/assets/skills/_shared/adaptive-reasoning-gate.md`. Acceptance: File exists with the canonical gate text.
- [ ] 1.2 Identify all 10 orchestrator files in `internal/assets/`. Acceptance: List of paths confirmed.

## Phase 2: Orchestrator Core Implementation

- [ ] 2.1 Add injection markers and gate content to `internal/assets/*/sdd-orchestrator.md`. Acceptance: Markers wrap the gate text in all 10 files.
- [ ] 2.2 Extend result contract in `internal/assets/*/sdd-orchestrator.md` with `chosen_mode` and `mode_rationale`. Acceptance: Contract section updated in all 10 files.
- [ ] 2.3 Implement mode-field validation and re-prompt logic in `internal/assets/*/sdd-orchestrator.md`. Acceptance: Result Processing section updated in all 10 files.

## Phase 3: Protocol and Documentation Integration

- [ ] 3.1 Update every phase protocol in `internal/assets/*/sdd-phase-protocols/*.md` with gate reference note. Acceptance: ~90 files contain the reference phrase in the first 500 chars.
- [ ] 3.2 Update shared SDD skills in `internal/assets/skills/sdd-*/SKILL.md` with gate reference note. Acceptance: 10 skill files updated.
- [ ] 3.3 Update `docs/adaptive-reasoning-v1.md` with "Mandatory Gate" section. Acceptance: Documentation reflects new gate behavior and scope.

## Phase 4: Testing and Verification

- [ ] 4.1 Update `internal/assets/assets_test.go` with `TestAdaptiveReasoningGateInjected`. Acceptance: Test asserts byte-identical gate injection across orchestrators.
- [ ] 4.2 Run Go tests: `go test ./internal/assets/...`. Acceptance: Tests pass.
