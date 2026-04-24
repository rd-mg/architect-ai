---
name: "Odoo XML Guidelines"
description: "Use when editing Odoo XML (views, actions, menus, reports, data). Covers inheritance, XPath robustness, and ID stability."
applyTo: "**/*.xml"
---

# Odoo XML Guidelines

- Prefer robust XPath selectors (avoid relying on exact positions).
- Keep `id` values stable; removing/renaming IDs is a breaking change.
- For view inheritance, keep changes minimal and scoped to the target view.
- When adding new records, ensure model + required fields are correct and consider `noupdate` where appropriate.

- **Odoo 18.0 & 19.0 Syntax**: Use `<list>` instead of `<tree>`. Use `<chatter/>` instead of the old chatter div structures. Do not use `attrs` dictionary syntax; use direct attributes like `invisible="expr"`.
- ALWAYS USE `hasclass()` rather than relying on exact classes in XPath.
- **Research Priority**: (1) Apply `mcp-notebooklm-orchestrator` skill for code-based and high-level strategy, (2) Apply `ripgrep` skill for local implementation patterns, (3) Use Context7 only as a last-resort fallback.
