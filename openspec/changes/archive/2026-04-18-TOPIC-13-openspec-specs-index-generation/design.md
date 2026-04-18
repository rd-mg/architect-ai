# Design: OpenSpec INDEX Auto-Generation

## Technical Approach

The implementation will utilize shell-based logic embedded directly within the SDD skill instructions. This approach ensures that any agent executing these phases will automatically follow the indexing requirements without needing additional system binaries (other than standard POSIX tools).

## Architecture Decisions

### Decision: Embedded Shell Logic vs. Separate Script

**Choice**: Embedded Shell Logic in `SKILL.md`.
**Alternatives considered**: Python/Go script, separate Bash script in `bin/`.
**Rationale**: Embedding logic ensures that the "Source of Truth" for the workflow (the Skill) also contains the automation code. This minimizes deployment overhead and keeps the project structure clean for a Pareto win.

### Decision: First Heading Parsing

**Choice**: `head -1` + `sed`.
**Alternatives considered**: Full Markdown parser, regex in Go.
**Rationale**: Most OpenSpec files follow a strict `# Heading` format at line 1. A simple one-liner is sufficient for 99% of cases and handles empty/missing files gracefully.

## Data Flow

1. **sdd-archive** merges delta specs.
2. **Archive Step 3d** triggers INDEX regeneration.
3. Script iterates over `openspec/specs/*/`.
4. Script extracts domain name and spec title.
5. Script rebuilds `openspec/specs/INDEX.md` as a Markdown table.

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `/home/rdmachadog/.gemini/antigravity/skills/sdd-archive/SKILL.md` | Modify | Add index regeneration shell block. |
| `/home/rdmachadog/.gemini/antigravity/skills/sdd-explore/SKILL.md` | Modify | Add requirement to check INDEX.md. |
| `/home/rdmachadog/.gemini/antigravity/skills/sdd-propose/SKILL.md` | Modify | Add requirement to check INDEX.md. |

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Manual | INDEX generation | Archive a change and verify `openspec/specs/INDEX.md` content. |
| Manual | Agent adherence | Launch `sdd-propose` and verify it reads `INDEX.md` in logs. |

## Migration / Rollout

No migration required. The first `sdd-archive` execution will bootstrap the `INDEX.md`.
