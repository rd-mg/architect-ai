---
name: "Odoo Python Guidelines"
description: "Use when editing Odoo Python code (models, wizards, controllers). Covers ORM recordset safety, API conventions, and common pitfalls."
applyTo: "**/*.py"
---

# Odoo Python Guidelines

- Prefer recordset-safe code: support multi-record `self` unless method is clearly singleton.
- Use `super()` for overrides; preserve method contract.
- Avoid `sudo()` unless required; prefer correct ACLs or `with_company`/`with_context`.
- Careful with computed fields: declare dependencies, keep compute deterministic, avoid expensive loops.
- Keep business logic in models; keep wizards thin.
- **Odoo 17.0+ Display Name**: Override `_compute_display_name` instead of `name_get`.
- **Research Priority**: (1) `mcp-notebooklm-orchestrator` skill for high-level strategy, (2) `ripgrep` skill for local implementation patterns, (3) Context7 as last-resort fallback.
