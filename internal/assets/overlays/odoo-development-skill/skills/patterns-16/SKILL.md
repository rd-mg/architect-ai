---
name: odoo-patterns-16
description: >
  Odoo 16.0-specific patterns for models, modules, OWL components, security,
  and version-specific behavior. Bridged only when Odoo 16.0 is detected in
  the project. Combine with patterns-agnostic/ for domain patterns.
---

# Odoo 16.0 Patterns

This bundle contains patterns specific to Odoo 16.0. Version-agnostic
domain patterns (accounting, stock, sale, etc.) are in `patterns-agnostic/`.

## Files in This Bundle

- `model-patterns.md` — Model definition, inheritance, fields, ORM methods
- `module-generator.md` — Module scaffolding, manifest, file layout
- `owl-components.md` — OWL component patterns
- `security-guide.md` — Security configuration, access rules, record rules
- `version-knowledge.md` — v16.0-specific behaviors and constraints

## Key Constraints for v16

- Python 3.8+
- OWL 2.x (new)
- ES6 imports instead of AMD/require
- Uses `attrs=` for field conditions in XML
- Uses `<tree>` view type
- `@api.model_create_multi` recommended

## Before Writing Code for v16

1. Read the relevant file in this bundle (model-patterns.md, etc.)
2. Cross-reference `patterns-agnostic/` for domain-specific concerns
3. For migration work, see `migration-{prev}-16/` and `migration-16-{next}/`
