# Odoo Version Knowledge: 18 to 19 Migration

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  VERSION MIGRATION: 18.0 → 19.0                                              ║
║  Critical changes, breaking changes, and migration patterns                  ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## Breaking Changes Summary

| Category | Change | Impact |
|----------|--------|--------|
| SQL | `SQL()` **REQUIRED** | **CRITICAL** - All raw SQL |
| Type Hints | **REQUIRED** | High - All methods |
| SQL Constraints | `models.Constraint()` class **REQUIRED** | High - All SQL constraints |
| res.users | `groups_id` cannot set in create() | High - User creation |
| OWL | OWL 3.x replaces 2.x | High - Component rewrite |
| Multi-Company | `_check_company_auto` required | High - All multi-company |
| Python | 3.12+ required | Medium |

## CRITICAL: SQL() Builder Required

**v18 (worked, discouraged):** `self.env.cr.execute("SELECT id FROM m WHERE state = %s", ('draft',))`
**v19 (required):** `self.env.cr.execute(SQL("SELECT id FROM m WHERE state = %s", 'draft'))`

```python
from odoo.tools import SQL

def _get_statistics(self) -> dict:
    self.env.cr.execute(SQL("""
        SELECT state, COUNT(*) as c, SUM(amount) as total
        FROM my_model WHERE company_id = %s AND create_date >= %s
        GROUP BY state
    """, self.env.company.id, fields.Date.today() - timedelta(days=30)))
    return {r['state']: r for r in self.env.cr.dictfetchall()}
```

## SQL Constraints: models.Constraint() Required

**v18 (worked):**
```python
_sql_constraints = [
    ('check_pct', 'CHECK(percentage >= 0 AND percentage <= 100)',
     'Percentage must be 0-100.'),
]
```

**v19 (required):**
```python
_check_percentage = models.Constraint(
    'CHECK(percentage >= 0 AND percentage <= 100)',
    'Percentage must be 0-100.',
)
```

**Migration:**
```python
# v18
_sql_constraints = [('code_unique', 'UNIQUE(code)', 'Must be unique.')]

# v19
_code_unique = models.Constraint('UNIQUE(code)', 'Must be unique.')
_amount_positive = models.Constraint('CHECK(amount >= 0)', 'Amount positive.')
```

## res.users: groups_id Cannot Set in create()

**v18 (worked):**
```python
user = self.env['res.users'].create({
    'name': 'Portal', 'login': 'p@e.com',
    'groups_id': [(6,0,[ref('base.group_portal').id])],
})
```

**v19 (must add separately):**
```python
user = self.env['res.users'].create({'name':'Portal', 'login':'p@e.com'})
portal = self.env.ref('base.group_portal')
portal.write({'users': [(4, user.id)]})
```

## Type Hints Required

```python
from typing import Optional, Any
from odoo import api, fields, models

class MyModel(models.Model):
    _name = 'my.model'
    _check_company_auto = True
    name: str = fields.Char(required=True, index=True)
    state: str = fields.Selection([('draft','Draft'),('done','Done')], default='draft')
    amount: float = fields.Monetary(currency_field='currency_id')

    def action_confirm(self) -> bool:
        for record in self:
            if record.state == 'draft':
                record.state = 'confirmed'
        return True

    @api.model_create_multi
    def create(self, vals_list: list[dict[str, Any]]) -> 'MyModel':
        for vals in vals_list:
            if 'code' not in vals:
                vals['code'] = self.env['ir.sequence'].next_by_code('my.model')
        return super().create(vals_list)
```

## OWL 3.x Migration

**OWL 2.x (v18):**
```javascript
static props = { recordId: Number, onSave: Function };
```

**OWL 3.x (v19):**
```javascript
static props = {
    recordId: { type: Number, required: true },
    onSave: { type: Function, optional: true },
    config: { type: Object, optional: true },
};
static defaultProps = { config: {} };
```

## Multi-Company Enforcement

```python
class MyModel(models.Model):
    _name = 'my.model'
    _check_company_auto = True      # REQUIRED
    company_id = fields.Many2one('res.company', required=True, readonly=True,
        default=lambda self: self.env.company, index=True)
    partner_id = fields.Many2one('res.partner', check_company=True)  # REQUIRED
```

## Python 3.12+

```python
type RecordList = list['MyModel']
from typing import Self
class MyModel(models.Model):
    def copy(self) -> Self: return super().copy()
```

## GitHub Verification

```
https://raw.githubusercontent.com/odoo/odoo/master/odoo/tools/sql.py
https://raw.githubusercontent.com/odoo/odoo/master/addons/web/static/src/core/
https://raw.githubusercontent.com/odoo/odoo/master/odoo/models.py
```

## Checklist

- [ ] **CRITICAL:** Migrate ALL raw SQL to `SQL()` builder
- [ ] Add type hints to ALL public methods
- [ ] Update OWL 2.x → 3.x components
- [ ] `_check_company_auto = True` on all models
- [ ] `check_company=True` on all relational fields
- [ ] Verify Python 3.12+ compatibility
- [ ] Update OWL prop definitions
- [ ] Test all SQL queries
- [ ] Review and update all tests

## Common Errors

**`SQL string queries not supported`** → Wrap with `SQL()`: `self.env.cr.execute(SQL("...", p1, p2))`
**`Type hints required for public method`** → Add return+param types
**`Props validation failed`** → Update OWL props: `{ type: Number, required: true }`
