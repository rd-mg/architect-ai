# Specification: Persistence Mode Branching Protocol

## Goal
Provide a single source of truth for how SDD sub-agents persist artifacts across different storage modes (engram, openspec, hybrid, none).

## Requirements

### Storage Modes
- **engram**: MUST use `mem_save` or `mem_update` with a stable `topic_key`.
- **openspec**: MUST use `write_to_file` (or Bash) to persist to the host filesystem.
- **hybrid**: MUST persist to BOTH engram and host filesystem. Host filesystem remains the authority for current state.
- **none**: MUST output artifacts inline only.

### Atomic Writes
- All host filesystem writes MUST be atomic to prevent partial data loss during session interrupts.

### Authoritative Merge (Hybrid)
- When resuming in Hybrid mode, the host filesystem state MUST be merged into Engram if they differ.

### Reference Pattern
- Individual skills MUST reference this protocol using:
  `Follow _shared/mode-branching.md for artifact-store branching.`
- Skills MUST specify their unique `artifact_name` and `topic_key`.

### Coverage
The following skills are verified to follow this protocol:
- `sdd-propose`
- `sdd-spec`
- `sdd-design`
- `sdd-tasks`
- `sdd-apply`
- `sdd-verify`
- `sdd-archive`
- `sdd-explore`
- `sdd-init`
- `sdd-onboard`
- `skill-registry`
