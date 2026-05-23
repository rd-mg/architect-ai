---
name: odoo-migration-18-19
description: >
  Migration guide for Odoo 18.0 → 19.0. Covers breaking changes in
  models, modules, OWL components, security rules, and version-specific
  behaviors. Bridged only when BOTH versions detected in the project
  (indicating a migration scenario).
---

# Odoo Migration: 18.0 → 19.0

Migration patterns for upgrading Odoo 18.0 to 19.0. Single-version projects use `patterns-18/` or `patterns-19/`.

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

- `patterns-18/` — Source version
- `patterns-19/` — Target version
- `patterns-agnostic/` — Version-agnostic patterns
