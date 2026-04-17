---
name: odoo-patterns-19
description: >
  Odoo 19.0-specific patterns for models, modules, OWL components, security,
  and version-specific behavior. Bridged only when Odoo 19.0 is detected in
  the project. Combine with patterns-agnostic/ for domain patterns.
---

# Odoo 19.0 Patterns

This bundle contains patterns specific to Odoo 19.0. Version-agnostic
domain patterns (accounting, stock, sale, etc.) are in `patterns-agnostic/`.

## Files in This Bundle

- `model-patterns.md` — Model definition, inheritance, fields, ORM methods
- `module-generator.md` — Module scaffolding, manifest, file layout
- `owl-components.md` — OWL component patterns
- `security-guide.md` — Security configuration, access rules, record rules
- `version-knowledge.md` — v19.0-specific behaviors and constraints
- `v19-features.md` — v19-exclusive features (AI server actions, passkeys, Hoot testing, etc.)

## Key Constraints for v19

- Python 3.12+
- OWL 3.x (new)
- Type hints REQUIRED on all public methods
- `SQL()` builder MANDATORY for any raw SQL
- Use `<list>` NOT `<tree>`
- Uses `hoot` testing framework (new)
- Supports AI server actions, passkeys, WebRTC IoT

## Before Writing Code for v19

1. Read the relevant file in this bundle (model-patterns.md, etc.)
2. Cross-reference `patterns-agnostic/` for domain-specific concerns
3. For migration work, see `migration-{prev}-19/` and `migration-19-{next}/`
