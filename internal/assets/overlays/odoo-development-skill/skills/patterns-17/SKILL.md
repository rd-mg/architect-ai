---
name: odoo-patterns-17
description: >
  Odoo 17.0-specific patterns for models, modules, OWL components, security,
  and version-specific behavior. Bridged only when Odoo 17.0 is detected in
  the project. Combine with patterns-agnostic/ for domain patterns.
---

# Odoo 17.0 Patterns

This bundle contains patterns specific to Odoo 17.0. Version-agnostic
domain patterns (accounting, stock, sale, etc.) are in `patterns-agnostic/`.

## Files in This Bundle

- `model-patterns.md` — Model definition, inheritance, fields, ORM methods
- `module-generator.md` — Module scaffolding, manifest, file layout
- `owl-components.md` — OWL component patterns
- `security-guide.md` — Security configuration, access rules, record rules
- `version-knowledge.md` — v17.0-specific behaviors and constraints

## Key Constraints for v17

- Python 3.10+
- OWL 2.x (continued)
- NO `attrs=` — use direct `invisible=`, `readonly=`, `required=` on fields
- Still uses `<tree>` view type
- `name_get` deprecated for display — use `_compute_display_name`
- Type hints recommended

## Before Writing Code for v17

1. Read the relevant file in this bundle (model-patterns.md, etc.)
2. Cross-reference `patterns-agnostic/` for domain-specific concerns
3. For migration work, see `migration-{prev}-17/` and `migration-17-{next}/`
