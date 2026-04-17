---
name: "debug-copilot-customizations"
description: "Troubleshoot why Copilot agents/prompts/instructions are (not) applied in VS Code. Uses #debugEventsSnapshot and checks common settings from recent VS Code Updates (agent permissions, Autopilot)."
argument-hint: "Paste symptoms + what you expected"
agent: "ask"
model: ['GPT-5.2 (copilot)', 'GPT-5.3-codex (copilot)', 'GPT-5.3-codex (copilot)', 'Gemini 3.1 Pro (copilot)']
tools: ['file_search', 'read_file', 'grep_search', 'run_in_terminal', 'read_file', 'edit_file', 'code-mode', 'browser_run_code', 'set_config_value', 'github/issue_write', 'github/update_pull_request', 'github/push_files', 'github/sub_issue_write', 'github/list_tags', 'github/fork_repository', 'github/list_branches', 'container-tools/get-config', 'google_notebo/ask_question']
---

Help me debug Copilot customization behavior in VS Code.

Ask me to attach `#debugEventsSnapshot` and then:
- Identify which workspace instructions, `.instructions.md`, prompts, and custom agents were loaded
- Check for conflicts (multiple instruction sources, overly broad `applyTo: "**"`, YAML frontmatter parse issues)
- Remind about agent permission levels (Default Approvals vs Bypass vs Autopilot)
- Provide a minimal set of steps to fix discovery/loading issues
