# Tasks: OpenSpec INDEX Auto-Generation

## Phase 1: Skill Updates (Core Logic)

- [x] 1.1 Update `/home/rdmachadog/.gemini/antigravity/skills/sdd-archive/SKILL.md`: Add Step 3d (INDEX Regeneration) shell block.
- [x] 1.2 Update `/home/rdmachadog/.gemini/antigravity/skills/sdd-explore/SKILL.md`: Add requirement to check `openspec/specs/INDEX.md` before exploration.
- [x] 1.3 Update `/home/rdmachadog/.gemini/antigravity/skills/sdd-propose/SKILL.md`: Add requirement to check `openspec/specs/INDEX.md` before proposal.

## Phase 2: Manual Verification

- [x] 2.1 Trigger a manual `sdd-archive` simulation (run the shell script) to verify `INDEX.md` creation.
- [x] 2.2 Verify `INDEX.md` content: check Domain, Title, and Path accuracy.
- [x] 2.3 Verify `sdd-propose` adherence: ensure the sub-agent identifies the index.
