# Specification: 04-ripgrep-odoo-local-skill

## Capability: odoo-discovery (local)

### Purpose
Provide high-fidelity codebase discovery for Odoo monorepos using local ripgrep execution.

### Preconditions
- `rg` binary is installed and in the system path.
- Odoo monorepo exists at `~/gitproj/odoo/` (or as configured).

### Behavior
- The skill executes `rg` with pre-defined flags optimized for Odoo domains (backend, frontend, spreadsheet).
- It filters out non-relevant files (tests, migrations, manifests) by default unless overridden.
- It returns line content with context (-C 3) and file paths.

### Postconditions
- Results are presented to the agent for analysis.
- No files are modified by the discovery process.

### Error Handling
- If `rg` fails (e.g., path not found), return a clear error message to the agent.
- Handle empty results gracefully.

### Invariants
- The tool must never execute destructive shell commands.
- Searches must stay within the specified Odoo directories.

### Test Hooks
- Dry-run mode to print the command that would be executed.
- Verification of `rg` output parsing.

## Capability: skill-registry (update)

### Purpose
Register the `ripgrep-odoo` skill so it can be automatically loaded by agents.

### Preconditions
- `.atl/skill-registry.md` exists and is writable.

### Behavior
- Append or insert the `ripgrep-odoo` entry into the registry.
- Include trigger keywords: "odoo", "o-spreadsheet", "owl", "enterprise".

### Postconditions
- The skill registry contains a valid link to the new skill's `SKILL.md`.

### Error Handling
- Prevent duplicate registrations.

### Invariants
- Registry structure must remain a valid Markdown table.

### Test Hooks
- Check if `architect-ai skill-registry verify` (if exists) or manual grep finds the entry.
