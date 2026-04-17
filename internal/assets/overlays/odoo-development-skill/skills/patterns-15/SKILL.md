---
name: odoo-patterns-15
description: >
  Odoo 15.0-specific patterns for models, modules, OWL components, security,
  and version-specific behavior. Bridged only when Odoo 15.0 is detected in
  the project. Combine with patterns-agnostic/ for domain patterns.
---

# Odoo 15.0 Patterns

This bundle contains patterns specific to Odoo 15.0. Version-agnostic
domain patterns (accounting, stock, sale, etc.) are in `patterns-agnostic/`.

## Files in This Bundle

- `model-patterns.md` — Model definition, inheritance, fields, ORM methods
- `module-generator.md` — Module scaffolding, manifest, file layout
- `owl-components.md` — OWL component patterns
- `security-guide.md` — Security configuration, access rules, record rules
- `version-knowledge.md` — v15.0-specific behaviors and constraints

## Key Constraints for v15

- Python 3.7+
- OWL 1.x (legacy)
- Mixed widget/OWL component usage
- Uses `attrs=` for field conditions in XML
- Uses `<tree>` view type

## Before Writing Code for v15

1. Read the relevant file in this bundle (model-patterns.md, etc.)
2. Cross-reference `patterns-agnostic/` for domain-specific concerns
3. For migration work, see `migration-{prev}-15/` and `migration-15-{next}/`
