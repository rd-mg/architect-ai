---
name: odoo-patterns-18
description: >
  Odoo 18.0-specific patterns for models, modules, OWL components, security,
  and version-specific behavior. Bridged only when Odoo 18.0 is detected in
  the project. Combine with patterns-agnostic/ for domain patterns.
---

# Odoo 18.0 Patterns

This bundle contains patterns specific to Odoo 18.0. Version-agnostic
domain patterns (accounting, stock, sale, etc.) are in `patterns-agnostic/`.

## Files in This Bundle

- `model-patterns.md` — Model definition, inheritance, fields, ORM methods
- `module-generator.md` — Module scaffolding, manifest, file layout
- `owl-components.md` — OWL component patterns
- `security-guide.md` — Security configuration, access rules, record rules
- `version-knowledge.md` — v18.0-specific behaviors and constraints

## Key Constraints for v18

- Python 3.11+
- OWL 2.x (enhanced)
- NO `attrs=` (removed in v17)
- Use `<list>` NOT `<tree>`
- Use `<chatter/>` shortcut element
- `tracking=True` mandatory for audited fields
- Type hints strongly recommended

## Before Writing Code for v18

1. Read the relevant file in this bundle (model-patterns.md, etc.)
2. Cross-reference `patterns-agnostic/` for domain-specific concerns
3. For migration work, see `migration-{prev}-18/` and `migration-18-{next}/`
