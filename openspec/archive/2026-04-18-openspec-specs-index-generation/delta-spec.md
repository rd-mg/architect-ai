# Spec: OpenSpec INDEX Auto-Generation

## Requirements

### R1: Index Format
The generated `openspec/specs/INDEX.md` MUST follow this structure:
- Heading: `# OpenSpec Capability Index`
- Table columns: `| Domain | Title | Path |`
- Row content:
  - **Domain**: Directory name under `openspec/specs/`.
  - **Title**: Content of the first H1 heading in `spec.md`.
  - **Path**: Relative path to the spec file (e.g., `specs/auth/spec.md`).

### R2: Indexing Script
The script MUST:
1. Initialize the header and table structure.
2. Iterate through all subdirectories in `openspec/specs/`.
3. Check for existence of `spec.md` in each directory.
4. Extract the title using a robust pattern (handling potential whitespace).
5. Append a row for each valid spec.

### R3: Error Handling
If a `spec.md` is missing or the directory is empty, the script SHOULD log a warning to stderr but NOT fail the archive process.

### R4: Skill Integration
- **sdd-archive**: The script must run at the end of the merge lifecycle.
- **sdd-explore/propose**: Agents MUST use `cat openspec/specs/INDEX.md` as their first action when researching existing capabilities.

## Scenarios

### Scenario: Initial Generation
- **Given** no `INDEX.md` exists.
- **When** the script runs.
- **Then** a new `INDEX.md` is created with all existing specs.

### Scenario: Update on Archive
- **Given** an existing `INDEX.md`.
- **When** a new change is archived.
- **Then** the `INDEX.md` is overwritten with the updated list including the new spec.
