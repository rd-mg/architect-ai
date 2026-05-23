---
name: "Odoo XML Guidelines"
description: "Use when editing Odoo XML (views, actions, menus, reports, data). Covers inheritance, XPath robustness, and ID stability."
applyTo: "**/*.xml"
---

# Odoo XML Guidelines

- Prefer robust XPath selectors; avoid relying on exact positions.
- Keep `id` values stable; removing/renaming IDs is a breaking change.
- For view inheritance, keep changes minimal and scoped to target view.
- When adding new records, ensure model + required fields are correct; consider `noupdate` where appropriate.
- **Odoo 18.0 & 19.0**: `<list>` instead of `<tree>`. `<chatter/>` instead of old chatter div structures. No `attrs` dict syntax; use direct `invisible="expr"`.
- ALWAYS USE `hasclass()` rather than exact classes in XPath.
- **Research Priority**: (1) `mcp-notebooklm-orchestrator` skill for high-level strategy, (2) `ripgrep` skill for local implementation patterns, (3) Context7 as last-resort fallback.
