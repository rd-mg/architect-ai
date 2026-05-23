---
name: "odoo-change-checklist"
description: "Generate a safe checklist for changing an Odoo addon (models/views/manifest/tests) in this repo."
argument-hint: "Describe the change (addon, feature, files)"
agent: "agent"
model: ['GPT-5.2 (copilot)', 'GPT-5.3-codex (copilot)', 'GPT-5.3-codex (copilot)', 'Gemini 3.1 Pro (copilot)']
tools: ['file_search', 'read_file', 'grep_search', 'run_in_terminal', 'read_file', 'edit_file', 'code-mode', 'browser_run_code', 'set_config_value', 'github/issue_write', 'github/update_pull_request', 'github/push_files', 'github/sub_issue_write', 'github/list_tags', 'github/fork_repository', 'github/list_branches', 'container-tools/get-config', 'google_notebo/ask_question']
---

Generate concise checklist for requested change in this Odoo addons workspace.

Include:
- Impacted addon(s)
- Files likely to change (`models/`, `views/`, `data/`, `security/`, `tests/`, `__manifest__.py`)
- Compatibility risks (external IDs, view inheritance, access rights)
- Validation steps (Odoo tests if available; else `python -m compileall .`)

Output as Markdown checklist.

Verify version-specific syntax (e.g., `<list>` not `<tree>` for Odoo 18/19, `hasclass()` for XPath).
Prefer Context7 for official docs; fall back to local resources.
