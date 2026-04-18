# Tasks: Engram Research Persistence

## Phase 1: Foundation / Documentation

- [ ] 1.1 Update `skills/_shared/engram-convention.md` with research-class topic keys and 168h TTL rules.
- [ ] 1.2 Update `internal/assets/claude/sdd-orchestrator.md` to include `research_cache_hits/misses` in the global result contract.

## Phase 2: Orchestrator Protocol Updates

- [ ] 2.1 Update `internal/assets/claude/sdd-phase-protocols/sdd-explore.md` with Cache-Lookup-Before-Call step.
- [ ] 2.2 Update `internal/assets/claude/sdd-phase-protocols/sdd-propose.md` with research metric reporting.
- [ ] 2.3 Update `internal/assets/claude/sdd-phase-protocols/sdd-design.md` with research metric reporting.
- [ ] 2.4 Update `internal/assets/claude/sdd-phase-protocols/sdd-verify.md` with research metric reporting.

## Phase 3: Research-Touching Protocols

- [ ] 3.1 Update `internal/assets/claude/sdd-phase-protocols/sdd-tasks.md` with cache lookup logic.
- [ ] 3.2 Update `internal/assets/claude/sdd-phase-protocols/sdd-apply.md` with cache lookup logic (if research is needed).

## Phase 4: Verification / Testing

- [ ] 4.1 Create a mock session to verify `topic_key` generation (prefix + len).
- [ ] 4.2 Verify 168h TTL enforcement in a simulated orchestrator call.
- [ ] 4.3 Verify sub-agent result metrics propagation.
