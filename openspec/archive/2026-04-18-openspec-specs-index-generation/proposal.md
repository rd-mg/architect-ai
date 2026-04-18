# Proposal: OpenSpec INDEX Auto-Generation

## Problem Statement
The `openspec/specs/` directory is growing, making it difficult for agents to discover existing capabilities without expensive and exhaustive filesystem reads. There is no central index of "Source of Truth" specifications.

## Proposed Solution (Option B)
Implement a lightweight, shell-based indexing mechanism integrated directly into the SDD skills. This avoids Go-level complexities while delivering immediate value.

### Scope
- **`sdd-archive`**: Update to automatically regenerate `openspec/specs/INDEX.md` after a successful merge.
- **`sdd-explore` & `sdd-propose`**: Update instructions to consult `INDEX.md` as the first step in research.
- **Validation**: Ensure the index correctly lists domains, titles (extracted from first # heading), and paths.

### Tradeoffs
- **Pros**: Zero dependencies, instant implementation, highly portable (Unix-like).
- **Cons**: Bash dependency (not Windows native), basic error handling.

## DoD
- [ ] `sdd-archive/SKILL.md` includes the regeneration script.
- [ ] `sdd-explore/SKILL.md` references the index.
- [ ] `sdd-propose/SKILL.md` references the index.
- [ ] Manual verification: `openspec/specs/INDEX.md` exists and is populated correctly.
