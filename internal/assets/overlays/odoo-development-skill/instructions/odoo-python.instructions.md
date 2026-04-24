---
name: "Odoo Python Guidelines"
description: "Use when editing Odoo Python code (models, wizards, controllers). Covers ORM recordset safety, API conventions, and common pitfalls."
applyTo: "**/*.py"
---

# Odoo Python Guidelines

- Prefer recordset-safe code (support multi-record `self` unless the method is clearly singleton).
- Use `super()` for overrides and preserve the method contract.
- Avoid `sudo()` unless required; prefer correct ACLs or `with_company`/`with_context`.
- Be careful with computed fields: declare dependencies, keep compute deterministic, avoid expensive loops.
- Keep business logic in models; keep wizards thin.

- **Odoo 17.0+ Display Name**: Do not use `name_get`. Instead, override `_compute_display_name`.
- **Research Priority**: (1) Apply `mcp-notebooklm-orchestrator` skill for code-based and high-level strategy, (2) Apply `ripgrep` skill for local implementation patterns, (3) Use Context7 only as a last-resort fallback.
