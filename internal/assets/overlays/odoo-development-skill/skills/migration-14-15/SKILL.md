---
name: odoo-migration-14-15
description: >
  Migration guide for Odoo 14.0 → 15.0. Covers breaking changes in
  models, modules, OWL components, security rules, and version-specific
  behaviors. Bridged only when BOTH versions detected in the project
  (indicating a migration scenario).
---

# Odoo Migration: 14.0 → 15.0

Migration patterns for upgrading Odoo 14.0 to 15.0. Single-version projects use `patterns-14/` or `patterns-15/`.

## Files

- `model-migration.md` — Model and field changes
- `module-migration.md` — Manifest and module structure changes
- `owl-migration.md` — OWL component changes
- `security-migration.md` — Security rule changes
- `version-knowledge.md` — Behavioral differences

**Migration Sequence:**
1. Review ALL files
2. Identify changes for YOUR modules
3. Plan scripts (pre-migrate + post-migrate) per module
4. Test on prod data copy BEFORE staging
5. `migrations/{new-version}/pre-migrate.py` for schema/rename
6. `migrations/{new-version}/post-migrate.py` for data transforms

## Related Bundles

- `patterns-14/` — Source version patterns
- `patterns-15/` — Target version patterns
- `patterns-agnostic/` — Version-agnostic domain patterns
