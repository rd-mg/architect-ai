---
name: "odoo-change-checklist"
description: "Generate a safe checklist for changing an Odoo addon (models/views/manifest/tests) in this repo."
argument-hint: "Describe the change (addon, feature, files)"
agent: "agent"
model: ['GPT-5.2 (copilot)', 'GPT-5.3-codex (copilot)', 'GPT-5.3-codex (copilot)', 'Gemini 3.1 Pro (copilot)']
tools: ['file_search', 'read_file', 'grep_search', 'run_in_terminal', 'read_file', 'edit_file', 'code-mode', 'browser_run_code', 'set_config_value', 'github/issue_write', 'github/update_pull_request', 'github/push_files', 'github/sub_issue_write', 'github/list_tags', 'github/fork_repository', 'github/list_branches', 'container-tools/get-config', 'google_notebo/ask_question']
---

Create a concise checklist to implement the requested change in this Odoo addons workspace.

Include:
- Which addon(s) are impacted
- Which files are likely to change (`models/`, `views/`, `data/`, `security/`, `tests/`, `__manifest__.py`)
- Compatibility risks (external IDs, view inheritance, access rights)
- Suggested validation steps (Odoo tests if available; otherwise `python -m compileall .`)

Output as a Markdown checklist.

- Verify version-specific syntax considerations (e.g., `<list>` instead of `<tree>` for Odoo 18/19, `hasclass()` for XPath).
- Note that official documentation lookups should prefer Context7, else use local resources.
