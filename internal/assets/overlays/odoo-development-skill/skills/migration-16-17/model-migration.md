# Odoo Model Patterns Migration: 16.0 → 17.0

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  MODEL MIGRATION GUIDE: Odoo 16.0 → 17.0                                     ║
║  Focus: Python models, decorators, CRUD methods                              ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## Breaking Changes Summary

| Pattern | v16 | v17 | Action |
|---------|-----|-----|--------|
| `@api.model_create_multi` | Recommended | **Mandatory** | Must add |
| `attrs` in views | Deprecated | Removed | Must migrate |
| `states` in views | Deprecated | Removed | Must migrate |

## MANDATORY: @api.model_create_multi

**v16:**
```python
@api.model
def create(self, vals):
    if not vals.get('name'):
        vals['name'] = self.env['ir.sequence'].next_by_code('my.model')
    return super().create(vals)
```

**v17:**
```python
@api.model_create_multi
def create(self, vals_list):
    for vals in vals_list:
        if not vals.get('name'):
            vals['name'] = self.env['ir.sequence'].next_by_code('my.model')
    return super().create(vals_list)
```

## View Visibility Changes

**v16:**
```python
def get_view_attrs(self):
    return {'invisible': [('state','!=','draft')], 'readonly': [('locked','=',True)]}
```

**v17:**
```python
def get_view_visibility(self):
    return {'invisible': "state != 'draft'", 'readonly': "locked"}
```

## Command Class (Already Required in v16)

All x2many must use Command class:
```python
from odoo import Command
self.write({'line_ids': [Command.create({'name':'New'}), Command.delete(1), Command.clear()]})
```

## Python Version

v16: 3.8+ → v17: 3.10+

**New Python features:**
```python
match self.state:
    case 'draft': self.action_confirm()
    case 'confirmed': self.action_done()
    case _: pass

def process(self, data: list[dict]) -> bool: ...
```

## Checklist

### Per Model
- [ ] Update `create()` to `@api.model_create_multi`
- [ ] Change signature `create(vals)` → `create(vals_list)`
- [ ] Iterate `vals_list` in create logic
- [ ] Verify all x2many use Command class
- [ ] Update view visibility logic for Python expressions

### Testing
- [ ] Test bulk create operations
- [ ] Verify single record creation still works
- [ ] Test all form views for visibility
- [ ] Test all buttons for state visibility

## Common Errors

**`create() got an unexpected keyword argument 'vals'`** → Update signature for `vals_list`
**`create() should return recordset`** → Ensure super().create() called with `vals_list`

## Automated Migration

```python
# Before:
@api.model
def create(self, vals):
    # ... logic with vals ...
    return super().create(vals)

# After:
@api.model_create_multi
def create(self, vals_list):
    for vals in vals_list:
        # ... logic with vals ...
    return super().create(vals_list)
```
