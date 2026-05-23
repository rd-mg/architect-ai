# Architect-AI Project Policy: Caution in Code Modifications

**MANDATE: CAUTION IN CODE MODIFICATIONS**

All AI agents MUST adhere when adding, deleting, or refactoring code:

1. **Preserve Legacy Context**: NEVER delete sections from skill templates, docs, or config files unless marked "DEPRECATED" or "REMOVED" by official Odoo 19 docs (verified via NotebookLM/Github).
2. **Expansion over Erasure**: When updating skill version (e.g., v18 → v19), prefer **extending** with version-specific sections (e.g., `## Version 19.0 Patterns`) over deleting prior versions.
3. **Surgical Precision**: Edits limited to necessary blocks. If legacy code replaced, document or archive context before removal.
4. **Validation Required**: Before any `replace` or `write_file` that removes content, agent MUST summarize to user exactly what is removed and WHY.
5. **Human Approval**: User must explicitly confirm any deletion exceeding 5 lines of documentation or code.

This policy overrides "Concise/Minimal" mandates for existing codebase documentation.
