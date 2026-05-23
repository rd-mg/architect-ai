# Odoo Module Migration: 15.0 → 16.0

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  MODULE MIGRATION: 15.0 → 16.0                                                ║
║  Command class introduced, attrs deprecated, OWL 2.x                          ║
║  VERIFY: https://github.com/odoo/odoo/tree/16.0                              ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## Overview

| Aspect | v15 | v16 |
|--------|-----|-----|
| X2many | Tuple syntax | Command class (recommended) |
| attrs | Full support | DEPRECATED (warns) |
| OWL | 1.x | 2.x |
| Python | 3.7-3.9 | 3.8-3.10 |
| create | @api.model | @api.model_create_multi recommended |

## Key Changes

### 1. Command Class for X2many

```python
# v15 (still works):
line_ids = [(0, 0, {'name': 'Line 1'}), (1, id, {'name': 'Updated'}), (2, id, 0)]

# v16 (recommended):
from odoo.fields import Command
line_ids = [Command.create({'name': 'Line 1'}), Command.update(id, {'name': 'Updated'}), Command.delete(id)]
```

### 2. attrs DEPRECATED

```xml
<!-- v15: -->
<field name="partner_id" attrs="{'invisible': [('state','=','draft')], 'required': [('state','=','confirmed')]}"/>

<!-- v16 (recommended): -->
<field name="partner_id" invisible="state == 'draft'" required="state == 'confirmed'"/>
```

### 3. OWL 1.x → 2.x

```javascript
// v15 OWL 1.x:
const { Component, useState } = owl;
const { useService } = require('@web/core/utils/hooks');

// v16 OWL 2.x:
import { Component, useState } from "@odoo/owl";
import { useService } from "@web/core/utils/hooks";
```

## Manifest Changes

```python
# v15
{'version': '15.0.1.0.0', 'assets': {'web.assets_backend': ['my_module/static/src/js/**/*']}}

# v16 — Same structure, update version
{'version': '16.0.1.0.0', 'assets': {'web.assets_backend': ['my_module/static/src/**/*.js']}}
```

## Complete Model Example

```python
# v15:
class MyModel(models.Model):
    _name = 'my.model'
    name = fields.Char(required=True, tracking=True)
    state = fields.Selection([('draft','Draft'),('done','Done')], default='draft', tracking=True)

    @api.model
    def create(self, vals):
        if not vals.get('code'):
            vals['code'] = self.env['ir.sequence'].next_by_code('my.model')
        return super().create(vals)

    def create_with_lines(self):
        return self.env['my.model'].create({
            'line_ids': [(0, 0, {'name': 'Line 1'})],
        })

# v16:
from odoo.fields import Command

class MyModel(models.Model):
    _name = 'my.model'
    name = fields.Char(required=True, tracking=True)
    state = fields.Selection([('draft','Draft'),('done','Done')], default='draft', tracking=True, index=True)

    @api.model_create_multi
    def create(self, vals_list):
        for vals in vals_list:
            if not vals.get('code'):
                vals['code'] = self.env['ir.sequence'].next_by_code('my.model')
        return super().create(vals_list)

    def create_with_lines(self):
        return self.env['my.model'].create({
            'line_ids': [Command.create({'name': 'Line 1'}), Command.create({'name': 'Line 2'})],
        })
```

## View Migration

### Form View

```xml
<!-- v15 (with attrs): -->
<button name="action_confirm" attrs="{'invisible': [('state','!=','draft')]}"/>
<field name="partner_id" attrs="{'required': [('state','=','confirmed')]}"/>

<!-- v16 (recommended): -->
<button name="action_confirm" invisible="state != 'draft'"/>
<field name="partner_id" required="state == 'confirmed'"/>
```

### Enhanced Tree Decorations
```xml
<tree decoration-danger="state == 'cancelled'" decoration-success="state == 'done'"
      decoration-warning="is_overdue">
    <field name="name"/>
    <field name="state" widget="badge"/>
</tree>
```

## OWL Migration (1.x → 2.x)

```javascript
// v15:
const { Component, useState, onWillStart } = owl;
const { useService } = require('@web/core/utils/hooks');
class MyComponent extends Component { setup() { ... } }
MyComponent.template = 'my_module.MyComponent';
registry.category('actions').add('my_action', MyComponent);

// v16:
import { Component, useState, onWillStart } from "@odoo/owl";
import { useService } from "@web/core/utils/hooks";
export class MyComponent extends Component {
    static template = "my_module.MyComponent";
    static props = { recordId: { type: Number, optional: true } };
    setup() { ... }
}
registry.category('actions').add('my_action', MyComponent);
```

## attrs Conversion Reference

| attrs Domain | Python Expression |
|--------------|-------------------|
| `[('state','=','draft')]` | `state == 'draft'` |
| `[('state','!=','draft')]` | `state != 'draft'` |
| `[('state','in',['a','b'])]` | `state in ('a','b')` |
| `['\|', A, B]` | `A or B` |
| `['&', A, B]` | `A and B` |

## Checklist

### Code
- [ ] Add `from odoo.fields import Command`
- [ ] Replace tuple x2many with Command class
- [ ] Consider `@api.model_create_multi` for creates
- [ ] Add `index=True` to frequently searched fields

### View Changes
- [ ] Replace `attrs=` with Python expression attributes
- [ ] Convert invisible/required/readonly domains

### OWL
- [ ] Change `require()` → ES `import`
- [ ] Update OWL imports from `@odoo/owl`
- [ ] Add `static template` and `static props`
- [ ] Use arrow functions in t-on-click

### Manifest
- [ ] Update version to `16.0.x.x.x`
- [ ] Update asset glob patterns

## Common Issues

**`DeprecationWarning: attrs is deprecated`** → Migrate to Python expressions
**`ImportError: cannot import 'Command'`** → Ensure v16; use tuple for v15
**OWL Import Errors** → Use ES `import` statements
