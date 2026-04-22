# Models & Fields Patterns

Consolidated from the following source files:
- `field-type-patterns.md` (architect-fix)
- `compute-depends-patterns.md` (architect-fix)
- `constraint-validation-patterns.md` (architect-fix)
- `inheritance-override-patterns.md` (architect-fix)
- `workflow-state-patterns.md` (architect-fix)
- `wizard-transient-patterns.md` (architect-fix)

> **Version-specific syntax** → `patterns-{version}/model-patterns.md`
> `@api.model_create_multi` mandatory from v16+ · `attrs=` removed in v17+ · SQL() mandatory in v19+

---

## Core Field Definitions

### Scalar and Relational Fields
```python
from odoo import api, fields, models

class MyModel(models.Model):
    _name = 'my_module.my_model'
    _description = 'My Model'
    _order = 'sequence, name'

    name = fields.Char(string='Name', required=True, index=True, tracking=True, translate=True)
    active = fields.Boolean(default=True)
    state = fields.Selection(selection=[('draft', 'Draft'), ('done', 'Done')], default='draft', tracking=True)
    
    # Monetary requires a currency_id field
    currency_id = fields.Many2one('res.currency', default=lambda self: self.env.company.currency_id)
    amount = fields.Monetary(string='Price', currency_field='currency_id')

    # Relational
    partner_id = fields.Many2one('res.partner', ondelete='restrict', index=True)
    line_ids = fields.One2many('my_module.line', 'parent_id', copy=True)
    tag_ids = fields.Many2many('my_module.tag', relation='my_model_tag_rel', column1='model_id', column2='tag_id')
```

---

## Compute & Logic Triggers

### Computed Fields (Stored & Inverse)
```python
    total = fields.Monetary(compute='_compute_total', store=True)
    price_with_tax = fields.Float(compute='_compute_pw_tax', inverse='_inverse_pw_tax')

    @api.depends('line_ids.price_subtotal')
    def _compute_total(self):
        for rec in self:
            rec.total = sum(rec.line_ids.mapped('price_subtotal'))

    def _inverse_pw_tax(self):
        for rec in self:
            rec.price_unit = rec.price_with_tax / 1.21 # Example inverse logic
```

### Constraints and Validations
```python
    @api.constrains('date_start', 'date_end')
    def _check_dates(self):
        for rec in self:
            if rec.date_start > rec.date_end:
                raise ValidationError("Start date must be before end date.")

    _sql_constraints = [
        ('name_uniq', 'UNIQUE(name)', 'Name must be unique!'),
        ('pos_qty', 'CHECK(quantity >= 0)', 'Quantity must be positive!'),
    ]
```

---

## Inheritance & State Machine

### Inheritance Types
```python
# 1. Classical (Extension)
class ResPartner(models.Model):
    _inherit = 'res.partner'
    custom_field = fields.Char()

# 2. Prototype (Copy)
class MyTask(models.Model):
    _name = 'my.task'
    _inherit = 'project.task'

# 3. Delegation (Shared Table)
class UserExtension(models.Model):
    _name = 'user.ext'
    _inherits = {'res.users': 'user_id'}
    user_id = fields.Many2one('res.users', required=True, ondelete='cascade')
```

### Wizard (TransientModel)
```python
class MyWizard(models.TransientModel):
    _name = 'my.wizard'
    
    @api.model
    def default_get(self, fields):
        res = super().default_get(fields)
        res['record_ids'] = self.env.context.get('active_ids')
        return res
```

---

## Anti-Patterns

```python
# ❌ NEVER use mutable defaults: tag_ids = fields.Many2many(default=[])
# ✅ CORRECT: default=lambda self: self.env['my.tag']

# ❌ NEVER compute without @api.depends: Odoo won't know when to recompute.

# ❌ NEVER use sudo() in a loop: Use self.sudo().write() instead of for r in self: r.sudo().write().
```

---

## Version Matrix

| Feature | v14-v15 | v16 | v17 | v18-v19 |
|---------|---------|-----|-----|---------|
| Create | `create()` | `@api.model_create_multi` | `create_multi` | `create_multi` |
| Multi-company| `company_id` | `company_id` | `company_id` | `check_company=True`|
| SQL | `_sql_constraints`| `_sql_constraints`| `_sql_constraints`| `SQL()` |
| Tracking | `track_visibility`| `tracking=True` | `tracking=True` | `tracking=True` |
