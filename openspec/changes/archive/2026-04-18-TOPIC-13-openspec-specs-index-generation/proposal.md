# Proposal: OpenSpec INDEX Auto-Generation

## Intent

Automate the generation of `openspec/specs/INDEX.md` to provide a centralized view of project capabilities. This reduces token consumption for agents during exploration and proposal phases by avoiding bulk reading of all spec files.

## Scope

### In Scope
- Shell-based INDEX generation logic added to `sdd-archive/SKILL.md`.
- Integration of INDEX-checking logic in `sdd-explore/SKILL.md` and `sdd-propose/SKILL.md`.
- Automated regeneration of the index after every successful change archive.

### Out of Scope
- Full Go-based CLI implementation (deferred).
- Windows-specific shell support (handled via POSIX/Bash assumption).

## Capabilities

### New Capabilities
- None (This is a workflow/infrastructure improvement).

### Modified Capabilities
- None (Internal SDD process update).

## Approach

Implement a Pareto fix using a shell one-liner in the `sdd-archive` skill. After the merge step, the skill will iterate through `openspec/specs/*/`, extract the first heading from `spec.md`, and rebuild `INDEX.md`. 

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `sdd-archive/SKILL.md` | Modified | Add INDEX regeneration step. |
| `sdd-explore/SKILL.md` | Modified | Add INDEX check pre-exploration. |
| `sdd-propose/SKILL.md` | Modified | Add INDEX check pre-proposal. |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Fragile parsing of spec titles | Low | Use simple `head -1` + `sed` as a start; refine if needed. |
| Missing `spec.md` files | Low | Use `2>/dev/null` to prevent script failure; log issues. |

## Rollback Plan

Remove the added shell blocks from the `SKILL.md` files and delete `openspec/specs/INDEX.md`.

## Dependencies

- None.

## Success Criteria

- [ ] `sdd-archive` successfully regenerates `openspec/specs/INDEX.md`.
- [ ] `INDEX.md` contains correct domain, title, and path mapping.
- [ ] `sdd-explore` and `sdd-propose` reference the index in their prompts.
