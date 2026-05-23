---
name: odoo-migration-15-16
description: >
  Migration guide for Odoo 15.0 → 16.0. Covers breaking changes in
  models, modules, OWL components, security rules, and version-specific
  behaviors. Bridged only when BOTH versions detected in the project
  (indicating a migration scenario).
---

# Odoo Migration: 15.0 → 16.0

Migration patterns for upgrading Odoo 15.0 to 16.0. Single-version projects use `patterns-15/` or `patterns-16/`.

## Files

- `model-migration.md` — Model and field changes
- `module-migration.md` — Manifest and module changes
- `owl-migration.md` — OWL component changes
- `security-migration.md` — Security rule changes
- `version-knowledge.md` — Behavioral differences

**Migration Sequence:**
1. Review ALL files
2. Identify changes for YOUR modules
3. Plan scripts (pre-migrate + post-migrate)
4. Test on prod data copy BEFORE staging
5. `migrations/{new-version}/pre-migrate.py` for schema/rename
6. `migrations/{new-version}/post-migrate.py` for data transforms

## Related Bundles

- `patterns-15/` — Source version
- `patterns-16/` — Target version
- `patterns-agnostic/` — Version-agnostic patterns
