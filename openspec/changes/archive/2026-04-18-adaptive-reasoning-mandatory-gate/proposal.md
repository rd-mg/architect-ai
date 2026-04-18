# Proposal: Adaptive Reasoning Mandatory Gate

## Intent

Make the `adaptive-reasoning` classifier a structural mandatory gate in every sub-agent launch. Currently, it is an optional skill that agents may ignore. Moving it to the prompt prelude ensures every sub-agent scores the task (scope, ambiguity, risk, verification) and selects a reasoning mode (Mode 1/2/3) before executing its phase protocol, optimizing token spend and reasoning depth.

## Scope

### In Scope
- Create `internal/assets/skills/_shared/adaptive-reasoning-gate.md` (single source of truth).
- Inject gate markers into 10 SDD orchestrators (claude, antigravity, opencode, codex, kiro, cursor, gemini, generic, windsurf, qwen).
- Update ~90 phase protocols and shared SDD skills with gate reference notes.
- Extend sub-agent result contracts with `chosen_mode` and `mode_rationale` fields.
- Implement re-prompt logic for missing mode declarations in all orchestrators.
- Add Go assets test to verify byte-identical gate injection.
- Update `docs/adaptive-reasoning-v1.md` with gate documentation.

### Out of Scope
- Runtime Go code changes for the classifier (prompt-only enforcement).
- Enforcement for platform-managed built-in sub-agents (e.g., Antigravity Browser/Terminal).

## Capabilities

### New Capabilities
- `adaptive-reasoning-gate`: Structural gate for mandatory reasoning depth classification.

### Modified Capabilities
- `sdd-orchestrator`: Extended result contract and validation logic.
- `sdd-phase-protocols`: Integrated reference to the adaptive reasoning gate.

## Approach

1. Create the shared gate fragment.
2. Update the orchestrator templates to include markers and inject the fragment.
3. Use a script to bulk-update phase protocols with the reference note.
4. Manually update orchestrator result contract descriptions.
5. Implement the Go test to guard against injection drift.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/assets/skills/_shared/adaptive-reasoning-gate.md` | New | Shared gate text |
| `internal/assets/*/sdd-orchestrator.md` | Modified | Injection markers and result contract |
| `internal/assets/*/sdd-phase-protocols/*.md` | Modified | Gate reference notes |
| `internal/assets/skills/sdd-*/SKILL.md` | Modified | Gate reference notes |
| `internal/assets/assets_test.go` | Modified | New injection and presence tests |
| `docs/adaptive-reasoning-v1.md` | Modified | Added "Mandatory Gate" section |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Injection drift across orchestrators | Med | Hardened Go test with byte-identical check |
| Sub-agent refusal to output Mode line | Low | Re-prompt logic + fallback to Mode 1 |
| Fragile regex parsing | Low | Tolerant regex scanning first 5 lines |

## Rollback Plan

Revert the merge commit. Orchestrators will lose the markers and validation. Old session sub-agents that don't output the mode line will trigger the new re-prompt once (if the session is in-flight during deployment), then proceed normally.

## Success Criteria

- [ ] Every SDD orchestrator contains the injected gate block.
- [ ] Every phase protocol references the gate in its opening.
- [ ] `go test ./internal/assets/...` passes with new assertions.
- [ ] Documentation explains the gate and its scope.
