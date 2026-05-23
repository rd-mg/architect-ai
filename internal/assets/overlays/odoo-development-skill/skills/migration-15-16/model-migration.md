# Odoo Model Patterns Migration: 15.0 → 16.0

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  MODEL MIGRATION: 15.0 → 16.0                                                ║
║  Command class introduced, attrs deprecated (still works)                    ║
║  VERIFY: https://github.com/odoo/odoo/tree/16.0/odoo/models.py               ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## Overview

v15→v16 is **non-breaking** for model code. Main changes:
- `Command` class for x2many (recommended)
- `attrs` deprecation in views (start migrating)
- `@api.model_create_multi` recommended

## New: Command Class

Cleaner API for x2many field operations.

**Import:** `from odoo.fields import Command`

| Command | Tuple Equivalent | Description |
|---------|------------------|-------------|
| `Command.create(vals)` | `(0, 0, vals)` | Create new |
| `Command.update(id, vals)` | `(1, id, vals)` | Update existing |
| `Command.delete(id)` | `(2, id, 0)` | Delete |
| `Command.unlink(id)` | `(3, id, 0)` | Unlink (M2M) |
| `Command.link(id)` | `(4, id, 0)` | Link existing |
| `Command.clear()` | `(5, 0, 0)` | Clear all |
| `Command.set(ids)` | `(6, 0, ids)` | Replace all |

### Examples

```python
# v15 (still works):
self.env['sale.order'].create({
    'order_line': [(0, 0, {'product_id': p.id})],
})
self.write({'tag_ids': [(4, tag.id, 0)]})

# v16 Command class:
from odoo.fields import Command

self.env['sale.order'].create({
    'order_line': [Command.create({'product_id': p.id})],
})
self.write({'tag_ids': [Command.link(tag.id)]})
```

## Create Method: model_create_multi

**Still works in v16 but `@api.model_create_multi` strongly recommended.** Mandatory in v17.

```python
# v15 (still works):
@api.model
def create(self, vals):
    if not vals.get('code'):
        vals['code'] = self.env['ir.sequence'].next_by_code('my.model')
    return super().create(vals)

# v16 (recommended):
@api.model_create_multi
def create(self, vals_list):
    for vals in vals_list:
        if not vals.get('code'):
            vals['code'] = self.env['ir.sequence'].next_by_code('my.model')
    return super().create(vals_list)
```

## Field Indexing

v16 encourages more explicit indexing:

```python
name = fields.Char(required=True, index='trigram')   # For LIKE searches
code = fields.Char(index=True)                        # Standard B-tree
state = fields.Selection([...], index=True)           # For filtering
```

| Type | Use Case |
|------|----------|
| `index=True` | Standard B-tree |
| `index='trigram'` | ILIKE/pattern |
| `index='btree_not_null'` | B-tree excluding NULL |

## Complete Migration Example

```python
# v15 Model:
class SaleOrderLine(models.Model):
    _name = 'sale.order.line'
    order_id = fields.Many2one('sale.order', required=True)
    product_id = fields.Many2one('product.product', required=True)

    @api.model
    def create(self, vals):
        if not vals.get('price_unit') and vals.get('product_id'):
            product = self.env['product.product'].browse(vals['product_id'])
            vals['price_unit'] = product.list_price
        return super().create(vals)

# v16 Model (recommended):
from odoo.fields import Command

class SaleOrderLine(models.Model):
    _name = 'sale.order.line'
    order_id = fields.Many2one('sale.order', required=True, index=True)
    product_id = fields.Many2one('product.product', required=True)

    @api.model_create_multi
    def create(self, vals_list):
        for vals in vals_list:
            if not vals.get('price_unit') and vals.get('product_id'):
                product = self.env['product.product'].browse(vals['product_id'])
                vals['price_unit'] = product.list_price
        return super().create(vals_list)
```

## View Changes (Start Migrating)

`attrs` still works in v16 but deprecated. Start converting:

```xml
<!-- v15 (works, deprecated): -->
<field name="partner_id" attrs="{'invisible': [('state', '=', 'draft')]}"/>

<!-- v16 (recommended): -->
<field name="partner_id" invisible="state == 'draft'"/>
```

## Checklist

### Recommended (Non-Breaking)
- [ ] Add `from odoo.fields import Command` to files using x2many
- [ ] Replace tuple syntax with Command class
- [ ] Update `@api.model` create → `@api.model_create_multi`
- [ ] Add `index=True` to frequently filtered fields
- [ ] Consider `index='trigram'` for searchable text

### Views (Start for v17)
- [ ] Identify all `attrs=` usage in XML
- [ ] Start converting to `invisible=`, `readonly=`, `required=`
- [ ] Test conditional visibility with new syntax

## Search and Replace

| Pattern | Replacement |
|---------|-------------|
| `(0, 0, vals)` | `Command.create(vals)` |
| `(1, id, vals)` | `Command.update(id, vals)` |
| `(2, id, 0)` | `Command.delete(id)` |
| `(3, id, 0)` | `Command.unlink(id)` |
| `(4, id, 0)` | `Command.link(id)` |
| `(5, 0, 0)` | `Command.clear()` |
| `(6, 0, ids)` | `Command.set(ids)` |

## Testing

1. Test x2many operations with Command class
2. Test batch creates with model_create_multi
3. If migrating attrs, verify visibility logic
4. Verify indexing improvements
