# Odoo Model Patterns Migration: 14.0 → 15.0

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  MODEL MIGRATION: 14.0 → 15.0                                                ║
║  @api.multi REMOVED, tracking=True standardized                              ║
║  VERIFY: https://github.com/odoo/odoo/tree/15.0/odoo/models.py               ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## Breaking Changes

### 1. @api.multi REMOVED

**Most critical.** All `@api.multi` methods fail in v15.

```python
# v14 (BREAKS in v15):
@api.multi
def action_confirm(self):
    for record in self:
        record.state = 'confirmed'

# v15 (remove @api.multi):
def action_confirm(self):
    for record in self:
        record.state = 'confirmed'
```

### 2. track_visibility → tracking

```python
# v14 (deprecated):
name = fields.Char(track_visibility='always')
state = fields.Selection([...], track_visibility='onchange')

# v15:
name = fields.Char(tracking=True)
state = fields.Selection([...], tracking=True)
```

### 3. super() Syntax

```python
# v14 (old style):
return super(MyModel, self).create(vals)

# v15 (Python 3):
return super().create(vals)
```

## Field Changes

| v14 Syntax | v15 Syntax |
|------------|------------|
| `track_visibility='always'` | `tracking=True` |
| `track_visibility='onchange'` | `tracking=True` |
| `track_visibility=True` | `tracking=True` |

```python
# v14:
class MyModel(models.Model):
    _inherit = 'mail.thread'
    name = fields.Char(track_visibility='always')
    state = fields.Selection([('draft','Draft'),('done','Done')], track_visibility='onchange')

# v15:
class MyModel(models.Model):
    _inherit = ['mail.thread', 'mail.activity.mixin']
    name = fields.Char(tracking=True)
    state = fields.Selection([('draft','Draft'),('done','Done')], tracking=True)
```

## CRUD Methods Migration

### Create
```python
# v14:
@api.model
def create(self, vals):
    if not vals.get('code'):
        vals['code'] = self.env['ir.sequence'].next_by_code('my.model')
    return super(MyModel, self).create(vals)

# v15 (single):
@api.model
def create(self, vals):
    if not vals.get('code'):
        vals['code'] = self.env['ir.sequence'].next_by_code('my.model')
    return super().create(vals)

# v15 (batch — recommended):
@api.model_create_multi
def create(self, vals_list):
    for vals in vals_list:
        if not vals.get('code'):
            vals['code'] = self.env['ir.sequence'].next_by_code('my.model')
    return super().create(vals_list)
```

### Write
```python
# v14:
@api.multi
def write(self, vals):
    if 'state' in vals and vals['state'] == 'done':
        for record in self:
            if not record.line_ids:
                raise UserError(_("Add at least one line."))
    return super(MyModel, self).write(vals)

# v15:
def write(self, vals):
    if 'state' in vals and vals['state'] == 'done':
        for record in self:
            if not record.line_ids:
                raise UserError(_("Add at least one line."))
    return super().write(vals)
```

### Unlink
```python
# v14:
@api.multi
def unlink(self):
    for record in self:
        if record.state == 'done':
            raise UserError(_("Cannot delete done records."))
    return super(MyModel, self).unlink()

# v15:
def unlink(self):
    for record in self:
        if record.state == 'done':
            raise UserError(_("Cannot delete done records."))
    return super().unlink()
```

### Copy
```python
# v14:
@api.multi
def copy(self, default=None):
    self.ensure_one()
    default = dict(default or {})
    default['name'] = _("%s (copy)") % self.name
    return super(MyModel, self).copy(default)

# v15:
def copy(self, default=None):
    self.ensure_one()
    default = dict(default or {})
    default['name'] = _("%s (copy)") % self.name
    return super().copy(default)
```

## Action Methods Migration

```python
# v14:
@api.multi
def action_confirm(self):
    for record in self:
        if record.state != 'draft':
            raise UserError(_("Only draft can be confirmed."))
        record.state = 'confirmed'
    return True

@api.multi
def action_view_partner(self):
    self.ensure_one()
    return {'type': 'ir.actions.act_window', 'res_model': 'res.partner',
            'res_id': self.partner_id.id, 'view_mode': 'form'}

# v15:
def action_confirm(self):
    for record in self:
        if record.state != 'draft':
            raise UserError(_("Only draft can be confirmed."))
        record.state = 'confirmed'
    return True

def action_view_partner(self):
    self.ensure_one()
    return {'type': 'ir.actions.act_window', 'res_model': 'res.partner',
            'res_id': self.partner_id.id, 'view_mode': 'form'}
```

## Computed Fields (No Change)
```python
total = fields.Float(compute='_compute_total', store=True)

@api.depends('line_ids.amount')
def _compute_total(self):
    for record in self:
        record.total = sum(record.line_ids.mapped('amount'))
```

## Constraints (No Change)
```python
@api.constrains('date_start', 'date_end')
def _check_dates(self):
    for record in self:
        if record.date_start and record.date_end and record.date_start > record.date_end:
            raise ValidationError(_("End after start."))
```

## Migration Script
```bash
grep -rn "@api.multi" --include="*.py"
grep -rn "track_visibility" --include="*.py"
grep -rn "super(.*self)" --include="*.py"
```

## Search and Replace

| Find | Replace |
|------|---------|
| `@api.multi\n    def` | `def` |
| `track_visibility='always'` | `tracking=True` |
| `track_visibility='onchange'` | `tracking=True` |
| `track_visibility=True` | `tracking=True` |
| `super(ClassName, self)` | `super()` |

## Checklist

- [ ] Remove ALL `@api.multi`
- [ ] Replace ALL `track_visibility` → `tracking=True`
- [ ] Update `super()` to Python 3 style
- [ ] Add `mail.activity.mixin` where appropriate
- [ ] Consider `@api.model_create_multi` for batch creates
- [ ] Test all action methods
- [ ] Test all CRUD operations
- [ ] Verify mail tracking

## Common Errors

**`AttributeError: 'api' object has no attribute 'multi'`** → Remove `@api.multi`
**`DeprecationWarning: track_visibility deprecated`** → Replace with `tracking=True`
**`TypeError: create() got multiple values for argument 'vals'`** → Use consistent pattern
