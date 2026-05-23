# Odoo Version Knowledge: 15 to 16 Migration

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  VERSION MIGRATION: 15.0 → 16.0                                              ║
║  Critical changes, breaking changes, and migration patterns                  ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## Breaking Changes Summary

| Category | Change | Impact |
|----------|--------|--------|
| ORM | `Command` class introduced | Medium - x2many |
| OWL | OWL 2.x replaces 1.x | High - Rewrite components |
| Views | `attrs=` deprecated | Medium - Start migrating |
| Assets | New asset bundling | Medium - Manifests |
| Batch | `@api.model_create_multi` preferred | Medium - Update create() |

## ORM Command Class

**v15:** `record.write({'line_ids': [(0,0,{'name':'Line 1'})]})`
**v16:** `record.write({'line_ids': [Command.create({'name':'Line 1'})]})`

| Tuple | Command | Purpose |
|-------|---------|---------|
| `(0,0,vals)` | `Command.create(vals)` | Create |
| `(1,id,vals)` | `Command.update(id,vals)` | Update |
| `(2,id,0)` | `Command.delete(id)` | Delete |
| `(3,id,0)` | `Command.unlink(id)` | Remove link |
| `(4,id,0)` | `Command.link(id)` | Add link |
| `(5,0,0)` | `Command.clear()` | Remove all |
| `(6,0,ids)` | `Command.set(ids)` | Replace |

## OWL 2.x Migration

**OWL 1.x (v15):** `const { Component } = owl;` — `owl.hooks.useState`
**OWL 2.x (v16):** `import { Component } from "@odoo/owl";` — `import { useState }`

## View Attrs Deprecation

**v15:** `attrs="{'invisible': [('state','=','draft')]}"`
**v16 starts:** `invisible="state == 'draft'"` — works in v16, removed in v17.

## Batch Create Pattern

**v15:** `@api.model` with `create(self, vals)` → still works
**v16 recommended:** `@api.model_create_multi` with `create(self, vals_list)`

## Asset System

**v15:** `'assets': {'web.assets_backend': ['path/to/js.js']}`
**v16:** `'assets': {'web.assets_backend': ['my_module/static/src/js/**/*']}` — glob patterns supported

## New Features in v16

- PDF preview: `<field name="document" widget="pdf_viewer"/>`
- Enhanced search views with domain/context
- Form view actions with bootstrap classes

## GitHub Verification

```
https://raw.githubusercontent.com/odoo/odoo/16.0/odoo/fields.py
https://raw.githubusercontent.com/odoo/odoo/16.0/odoo/tools/view_validation.py
```

## Checklist

- [ ] Replace tuple x2many with Command class
- [ ] Migrate OWL 1.x to 2.x
- [ ] Start replacing `attrs=` with inline expressions
- [ ] Update `@api.model` create → `@api.model_create_multi`
- [ ] Update asset declarations with glob patterns
- [ ] Test OWL components for 2.x compat
- [ ] Verify Command operations

## Common Errors

**`Command is not defined`** → Add `from odoo.fields import Command`
**OWL `useState` not found** → Update import: `import { useState } from "@odoo/owl"`
**`attrs attribute is deprecated`** → Replace with inline expressions
