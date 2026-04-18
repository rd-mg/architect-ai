# Walkthrough: OpenSpec INDEX Auto-Generation

The auto-generation of the OpenSpec Capability Index has been implemented and verified. This Pareto fix ensures that `openspec/specs/INDEX.md` is automatically updated after every successful change archive, and that agents consult this index during exploration and proposal phases.

## Changes Made

### SDD Skills
- **sdd-archive**: Added Step 3d to regenerate `INDEX.md` using a shell script.
- **sdd-explore**: Added Step 2 to consult `INDEX.md` before searching specs.
- **sdd-propose**: Added Step 3 to consult `INDEX.md` before categorization.

### OpenSpec Index
- **openspec/specs/INDEX.md**: Populated with current project domains and titles.

## Verification Results

### Manual Test: INDEX Regeneration
- **Action**: Ran the shell script block from `sdd-archive` manually.
- **Result**: `openspec/specs/INDEX.md` created with 11 domain entries.

### Manual Test: Title Parsing
- **Action**: Verified that `# Heading` is correctly extracted from `spec.md` files.
- **Result**: Titles like "Adaptive Reasoning Gate Specification" and "Odoo DDD Tactical Specification" correctly captured.

### Agent Adherence
- **Action**: Verified that the global skills for `explore` and `propose` contain the index-checking instructions.
- **Result**: Instruction present and mandatory for agents.
