# Proposal: OpenSpec Delta-Merge Conflict Detection (TOPIC-10)

## Intent
Prevent silent data loss during concurrent SDD changes by enforcing a SHA-based base-version check before merging deltas into the main specifications.

## Scope
- **SHA-256 base-stamping**: Every delta spec authored by `sdd-spec` will include front-matter declaring the exact version of the main spec it was based on.
- **Conflict Hard Gate**: `sdd-archive` will refuse to merge if the main spec has drifted since the delta was written.
- **Recovery Workflow**: Provide a clear conflict report and runbook for manual resolution (rebasing).

## Approach
- Add a new `openspec` Go component for SHA computation and front-matter parsing.
- Introduce `architect-ai sdd-archive-preflight` for dry-run validation.
- Update agent skills to automatically manage metadata and verify state before merging.

## Affected Areas
- `internal/components/openspec/`
- `internal/cli/`
- `internal/assets/skills/sdd-spec/SKILL.md`
- `internal/assets/skills/sdd-archive/SKILL.md`

## Risks
- **Operational Friction**: Users must re-run `sdd-spec` (rebase) if conflicts occur. This is a tradeoff for correctness.

## Success Criteria
- [ ] Concurrent changes touching the same spec are detected at archive time.
- [ ] `merge-conflict.md` is generated with actionable recovery steps.
- [ ] Delta front-matter is correctly stripped during merge to keep main specs clean.
