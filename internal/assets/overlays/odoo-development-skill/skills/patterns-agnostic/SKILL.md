---
name: odoo-patterns-agnostic
description: >
  Version-agnostic Odoo development patterns covering domain areas (accounting,
  stock, sale, HR, etc.) and infrastructure (views, models, controllers, etc.).
  Bridged for all Odoo projects regardless of version. Version-specific syntax
  is in patterns-{version}/ bundles.
---

# Odoo Patterns — Version-Agnostic

These patterns apply across Odoo versions 14-19. Syntax differences between
versions live in `patterns-{version}/` bundles. When in doubt, consult BOTH
this bundle (for the concept) AND the version-specific bundle (for the exact
syntax).

## Pattern Discovery Index

When you need a specific pattern, consult the matching file:
See `discovery-index.md` for the full mapping of intents to files.

## Loading Protocol

This index is always bridged. Individual domain files are NOT pre-loaded.

When you need a specific domain pattern:
1. Check this index for the matching file
2. Read that file directly: cat .agent/skills/patterns-agnostic/{file}
3. Apply patterns. Do NOT load files you don't need.

Exception: sdd-verify always needs models-fields.md and views-widgets.md.
Pre-load those for verify phases only.

## Don't Reinvent the Wheel — Key OCA Repos

When searching for existing modules, start here:

### By Domain
- **Accounting**: https://github.com/OCA/account-financial-reporting, https://github.com/OCA/account-financial-tools
- **Stock/Warehouse**: https://github.com/OCA/stock-logistics-workflow, https://github.com/OCA/stock-logistics-warehouse
- **Sale**: https://github.com/OCA/sale-workflow
- **Purchase**: https://github.com/OCA/purchase-workflow
- **HR**: https://github.com/OCA/hr
- **POS**: https://github.com/OCA/pos
- **Website**: https://github.com/OCA/website
- **Server Tools**: https://github.com/OCA/server-tools, https://github.com/OCA/server-ux
- **Reporting**: https://github.com/OCA/reporting-engine
- **Connector**: https://github.com/OCA/connector
- **Localizations**: https://github.com/OCA/l10n-{country_code} (e.g., `l10n-spain`, `l10n-brazil`)

## Version-Agnostic Principles

These apply regardless of version:

1. **ORM first**: use Odoo ORM methods (`search`, `create`, `write`, `unlink`) — NEVER raw SQL unless absolutely necessary
2. **Inheritance, don't replace**: use `_inherit` to extend; use `replace` position in XML sparingly
3. **Security by default**: every new model gets `ir.model.access.csv`
4. **Respect Multi-Company**: every query on shareable records respects `company_ids`
5. **Don't bypass user_id checks**: use `sudo()` only with documented justification
6. **Prefer computed over stored**: store only what's needed for search/sort; compute the rest
7. **Test in TransactionCase**: rollback between tests ensures isolation
8. **Version-aware syntax**: ALWAYS check the version-specific pattern file before writing

## Resources

See also:
- `patterns-{version}/` bundles for version-specific syntax
- `migration-{from}-{to}/` bundles for migration between versions
- `rules/coding-style.md` — universal coding style
- `rules/security.md` — security hardening
- `sdd-supplements/domain-map.md` — DDD bounded contexts
