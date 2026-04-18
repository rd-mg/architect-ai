# Proposal: TOPIC-14 — Apply-Progress Continuity

## Intent
Enable symmetric task-level progress resumption across all Architect-AI persistence modes (Engram, OpenSpec, Hybrid, None).

## Scope
- Update all agent orchestrator prompts (9+ files).
- Update `sdd-apply` and `sdd-verify` skills.
- Enforce "Filesystem Wins" authority in Hybrid mode.
- Block verification when implementation is incomplete.

## Reasoning
Currently, the `sdd-apply` continuation check only supports Engram. In `openspec` or `hybrid` modes, sub-agents may re-execute already completed tasks, leading to data overwrites and inefficiency.

## Affected Areas
- `internal/assets/**/sdd-orchestrator.md`
- `internal/assets/skills/sdd-apply/SKILL.md`
- `internal/assets/skills/sdd-verify/SKILL.md`
