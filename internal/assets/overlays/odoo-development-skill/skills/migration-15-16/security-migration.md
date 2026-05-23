# Odoo Security Guide - Migration 15.0 → 16.0

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  MIGRATION GUIDE: ODOO 15.0 → 16.0 SECURITY                                  ║
║  Guide for upgrading security code from v15 to v16.                          ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## Security Changes Overview

| Component | v15 | v16 | Required |
|-----------|-----|-----|----------|
| x2many | Tuple syntax | Command class | Recommended |
| View visibility | `attrs` | `attrs` (deprecated) | Prepare v17 |
| OWL | 1.x | 2.x | For OWL components |
| @api.model_create_multi | Optional | Recommended | Recommended |

## Key Changes

### 1. Command Class for x2many

**v15:**
```python
def action_update_lines(self):
    self.write({'line_ids': [(0,0,{'name':'New'}), (1,id,{'name':'Updated'}), (2,id,0)]})
```

**v16:**
```python
from odoo import Command
def action_update_lines(self):
    self.write({'line_ids': [Command.create({'name':'New'}), Command.update(id,{'name':'Updated'}), Command.delete(id)]})
```

### 2. Prepare for attrs Deprecation

**v15/v16 (deprecated):**
```xml
<field name="notes" attrs="{'invisible': [('state','=','draft')]}"/>
```

**v16 (recommended, ready for v17):**
```xml
<field name="notes" invisible="state == 'draft'"/>
```

### 3. @api.model_create_multi

**v15:**
```python
@api.model
def create(self, vals):
    return super().create(vals)
```

**v16 (recommended):**
```python
@api.model_create_multi
def create(self, vals_list):
    return super().create(vals_list)
```

## Command Class Reference

```python
from odoo import Command
Command.create(values)       # (0, 0, values)
Command.update(id, values)   # (1, id, values)
Command.delete(id)           # (2, id, 0)
Command.unlink(id)           # (3, id, 0)
Command.link(id)             # (4, id, 0)
Command.clear()              # (5, 0, 0)
Command.set(ids)             # (6, 0, ids)
```

## Security Implications

Command class improves security by: more explicit operations, less tuple index errors, better type checking.

**Secure Line Creation:**
```python
def action_create_secure_lines(self):
    self.ensure_one()
    if not self.env.user.has_group('my_module.group_manager'):
        raise AccessError(_("Only managers can create lines."))
    self.write({'line_ids': [Command.create({'name':'Line 1', 'user_id': self.env.user.id})]})
```

## OWL 2.x Security Changes

**v15 (OWL 1.x):**
```javascript
odoo.define('my_module.MyComponent', function (require) {
    const { Component } = owl;
    registry.category('actions').add('my_action', MyComponent);
});
```

**v16 (OWL 2.x):**
```javascript
/** @odoo-module **/
import { Component } from "@odoo/owl";
import { registry } from "@web/core/registry";
export class MyComponent extends Component {}
registry.category('actions').add('my_action', MyComponent);
```

## No Change Required

Security Groups, Access Rights, Record Rules, Field Groups — same syntax.

## Checklist

- [ ] Import `Command` for x2many operations
- [ ] Start migrating `attrs` → direct attributes (prep for v17)
- [ ] Update to `@api.model_create_multi`
- [ ] Migrate OWL 1.x → 2.x
- [ ] Test x2many with new Command syntax
- [ ] Update JS module syntax

## Dual Compatibility (v15/v16)

```python
try:
    from odoo import Command
except ImportError:
    class Command:
        @staticmethod
        def create(values): return (0, 0, values)
        @staticmethod
        def update(id, values): return (1, id, values)
        # ... etc
```

## GitHub Reference

- `odoo/fields.py` — x2many field handling
- `addons/web/static/src/` — OWL 2.x components
