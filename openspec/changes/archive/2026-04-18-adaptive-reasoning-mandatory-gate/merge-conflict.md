# Merge Conflict Report

Change: `adaptive-reasoning-mandatory-gate`
Generated: 2026-04-18T05:10:30Z

`sdd-archive` refused to merge the delta specs below because the main
spec changed after this delta was authored. Resolve manually; see
`docs/openspec-merge-conflict.md`.

## Domain `sdd-orchestration`

- Delta file: `openspec/changes/adaptive-reasoning-mandatory-gate/specs/sdd-orchestration/spec.md`
- Main file:  `openspec/specs/sdd-orchestration/spec.md`
- Expected base SHA: `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855`
- Actual  base SHA: `4f1f561c45279de1a5d2a1cd068453cf60b510d5952707a147ad175587b57b72`
- Delta captured at: `2026-04-18T05:01:47Z`

**Main spec has changed since this delta was authored. Another change archived in the meantime.**

### Recovery

1. `git log openspec/specs/sdd-orchestration/spec.md` — find what changed.
2. Re-run `sdd-spec` for this delta to rebase onto current main.
3. Or edit the delta manually and update `openspec_delta.base_sha` to the new SHA.
4. Re-run `architect-ai sdd-archive-preflight adaptive-reasoning-mandatory-gate`.

