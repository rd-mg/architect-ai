# Tasks: OpenSpec INDEX Auto-Generation

## Implementation

### Skill Patches
- [ ] **sdd-archive**: Insert index regeneration script after merge step.
  - File: `/home/rdmachadog/.gemini/antigravity/skills/sdd-archive/SKILL.md`
- [ ] **sdd-explore**: Add requirement to check `INDEX.md` at start.
  - File: `/home/rdmachadog/.gemini/antigravity/skills/sdd-explore/SKILL.md`
- [ ] **sdd-propose**: Add requirement to check `INDEX.md` at start.
  - File: `/home/rdmachadog/.gemini/antigravity/skills/sdd-propose/SKILL.md`

### Initial Indexing
- [ ] Manually run the indexing script once to bootstrap `openspec/specs/INDEX.md`.

## Verification

### Manual
- [ ] Verify `openspec/specs/INDEX.md` contains all current specs.
- [ ] Run a test archive (or dry-run) to confirm the index regenerates.
- [ ] Verify table formatting is correct in a markdown viewer.
