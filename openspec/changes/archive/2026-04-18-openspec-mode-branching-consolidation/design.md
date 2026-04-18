# Design: OpenSpec Mode-Branching Consolidation

## Technical Approach

We will centralize the mode-branching logic into a single shared file and refactor all SDD phase skills to reference it. This eliminates the massive redundancy and prevents drift in how sub-agents handle persistence modes.

## Architecture Decisions

### Decision: Placement of Shared Protocol
**Choice**: `internal/assets/skills/_shared/mode-branching.md`
**Rationale**: This directory is already scanned by the orchestrator and known to agents. It avoids the need for a new "includes" mechanism in Go for now (Option B from TOPIC-11).

### Decision: Standardized Persistence Metadata
**Choice**: Each skill will explicitly declare its `Artifact Name`, `Topic Key`, and `Type` in its own persistence section, then reference the shared protocol for the logic.
**Rationale**: This keeps the "what" in the skill and the "how" in the shared protocol.

## Shared Protocol Content

\`\`\`markdown
# Mode-Branching Protocol

Follow these instructions based on the provided \`artifact_store\` mode:

### engram
- **Retrieval**: Use \`mem_get_observation(id)\` for previous artifacts (sdd-explore, sdd-propose, etc.). Use \`mem_search\` only to find IDs.
- **Persistence**: Save to Engram via \`mem_save\` or \`mem_update\`. Use the stable \`topic_key\` provided by the phase.
- **Filesystem**: Do NOT write or read any \`openspec/\` files.

### openspec
- **Retrieval**: Read from \`openspec/changes/{change-name}/\` or \`openspec/specs/\`.
- **Persistence**: Write to the host filesystem using \`write_to_file\` (or Bash).
- **State**: Update \`openspec/changes/{change-name}/state.yaml\` using the **Atomic Write Pattern** (tmp + rename) after every write.
- **Memory**: Do NOT call any \`mem_*\` tools.

### hybrid (Default Authority)
- **Retrieval**: Read from Engram (primary) with filesystem fallback. If state differs, filesystem is the authority.
- **Persistence**: Persist to BOTH Engram and filesystem. Follow the state maintenance rules from the \`openspec\` section.

### none
- **Action**: Return results inline only. NEVER modify the filesystem or Engram.

---

### Atomic Write Pattern (Filesystem)
To prevent data loss during session interrupts:
1. Write content to \`{filename}.tmp\`
2. Verify write success
3. Rename \`{filename}.tmp\` to \`{filename}\`
\`\`\`

## Refactoring Plan

| Skill | Artifact Name | Topic Key | Type |
|-------|---------------|-----------|------|
| sdd-propose | proposal.md | sdd/{change-name}/proposal | architecture |
| sdd-spec | spec.md | sdd/{change-name}/spec | architecture |
| sdd-design | design.md | sdd/{change-name}/design | architecture |
| sdd-tasks | tasks.md | sdd/{change-name}/tasks | architecture |
| sdd-apply | apply-progress.md | sdd/{change-name}/apply-progress | architecture |
| sdd-verify | verify.md | sdd/{change-name}/verify | architecture |
| sdd-archive | (N/A) | (N/A) | (N/A) |
| sdd-explore | explore.md | sdd/{change-name}/explore | architecture |
| sdd-onboard | (N/A) | (N/A) | (N/A) |
| sdd-init | (N/A) | (N/A) | (N/A) |
| skill-registry | skill-registry.md | skill-registry/{project} | config |

## Verification Plan

### Automated
- Run \`rg "mode-branching.md" internal/assets/skills/sdd-*/SKILL.md | wc -l\` and expect 10 (or 11 if including registry).

### Manual
- Inspect each skill to ensure the reference is correctly placed and per-skill nuances are preserved as notes.
