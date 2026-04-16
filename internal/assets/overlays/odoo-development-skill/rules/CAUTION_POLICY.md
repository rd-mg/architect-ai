# Architect-AI Project Policy: Caution in Code Modifications

**MANDATE: CAUTION IN CODE MODIFICATIONS**

All AI agents in this project MUST adhere to the following principles when adding, deleting, or refactoring code:

1. **Preserve Legacy Context**: NEVER delete sections from skill templates, documentation, or configuration files unless they are explicitly marked as "DEPRECATED" or "REMOVED" by the official Odoo 19 documentation (verified via NotebookLM/Github).
2. **Expansion over Erasure**: When updating a skill to a newer version (e.g., v18 -> v19), prefer **extending** the file with version-specific sections (e.g., `## Version 19.0 Patterns`) rather than deleting previous version's information.
3. **Surgical Precision**: Edits must be strictly limited to the necessary blocks. If a block of legacy code needs to be replaced, ensure the legacy context is documented or archived before removal.
4. **Validation Required**: Before executing any `replace` or `write_file` operation that removes content, the agent MUST summarize to the user exactly what is being removed and WHY.
5. **Human Approval**: The user must explicitly confirm any deletion of content that exceeds 5 lines of documentation or code.

This policy takes precedence over the "Concise/Minimal" mandates when dealing with existing codebase documentation.
