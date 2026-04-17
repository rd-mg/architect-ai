---
name: odoo-migration-14-15
description: >
  Migration guide for Odoo 14.0 → 15.0. Covers breaking changes in
  models, modules, OWL components, security rules, and version-specific
  behaviors. Bridged only when BOTH versions are detected in the project
  (indicating a migration scenario).
---

# Odoo Migration: 14.0 → 15.0

This bundle contains migration patterns specific to upgrading from Odoo
14.0 to 15.0. If your project only targets one version, consult
`patterns-14/` or `patterns-15/` instead.

## Files in This Bundle

- `model-migration.md` — Model and field changes
- `module-migration.md` — Manifest and module structure changes
- `owl-migration.md` — OWL component version changes
- `security-migration.md` — Security rule changes
- `version-knowledge.md` — General behavioral differences

## Migration Sequence

When migrating a real project:
1. Review ALL files in this bundle
2. Identify which changes apply to YOUR modules
3. Plan migration scripts (pre-migrate + post-migrate) per module
4. Test migrations on a copy of production data BEFORE running on staging
5. Use `migrations/{new-version}/pre-migrate.py` for schema/rename changes
6. Use `migrations/{new-version}/post-migrate.py` for data transformations

## Related Bundles

- `patterns-14/` — Source version patterns (what you have now)
- `patterns-15/` — Target version patterns (what you're migrating to)
- `patterns-agnostic/` — Version-agnostic domain patterns
