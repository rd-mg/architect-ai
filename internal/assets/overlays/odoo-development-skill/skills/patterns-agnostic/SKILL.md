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

| Intent / Keywords | File |
|-------------------|------|
| account, invoice, journal, payment | `accounting.md` |
| tax, fiscal, vat, withholding | `accounting.md` |
| stock, inventory, warehouse, picking | `stock-inventory.md` |
| lot, serial, batch, expiry | `stock-inventory.md` |
| uom, unit, measure | `stock-inventory.md` |
| sale, order, quotation, crm | `sale-crm.md` |
| pricelist, price, discount | `sale-crm.md` |
| product, variant, attribute | `sale-crm.md` |
| hr, employee, contract, timesheet | `hr-employee.md` |
| project, task, subtask | `hr-employee.md` |
| purchase, vendor, procurement, RFQ | `purchase-procurement.md` |
| website, portal, public, landing | `website-portal.md` |
| token, access, portal user | `website-portal.md` |
| view, form, tree, list, kanban, search | `views-widgets.md` |
| widget, statusbar, badge, image | `views-widgets.md` |
| qweb, template, t-if, t-foreach | `views-widgets.md` |
| action, window, server action | `views-widgets.md` |
| menu, navigation, menuitem | `views-widgets.md` |
| dashboard, kpi, analytics | `views-widgets.md` |
| field, char, many2one, selection | `models-fields.md` |
| computed, depends, inverse | `models-fields.md` |
| constraint, validation, check | `models-fields.md` |
| onchange, dynamic, domain | `models-fields.md` |
| inherit, extend, override | `models-fields.md` |
| workflow, state, statusbar | `models-fields.md` |
| wizard, transient, dialog | `models-fields.md` |
| controller, http, api, rest | `infrastructure.md` |
| cron, scheduled, automation | `infrastructure.md` |
| mail, email, chatter, activity | `infrastructure.md` |
| assets, js, css, scss, bundle | `infrastructure.md` |
| logging, debug, error, exception | `infrastructure.md` |
| settings, config, parameter | `infrastructure.md` |
| attachment, binary, file, image | `infrastructure.md` |
| report, pdf, print | `infrastructure.md` |
| multi-company, company_id | `infrastructure.md` |
| import, export, csv, excel | `data-operations.md` |
| migration, upgrade, version | `data-operations.md` |
| sequence, numbering | `data-operations.md` |
| context, env, sudo | `data-operations.md` |
| external, api, webhook, integration | `data-operations.md` |
| quick, snippet, cheatsheet, 80/20 | `quick-patterns.md` |

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
