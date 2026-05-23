# Odoo Security Guide - Migration 17.0 → 18.0

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  MIGRATION GUIDE: ODOO 17.0 → 18.0 SECURITY                                  ║
║  Upgrading security code from v17 to v18.                                    ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## Security Changes Overview

| Component | v17 | v18 | Required |
|-----------|-----|-----|----------|
| Multi-company | `company_ids` | `allowed_company_ids` | Recommended |
| Company check | Manual domain | `_check_company_auto` | Recommended |
| Field company | Manual validation | `check_company=True` | Recommended |
| Type hints | Optional | Recommended | Optional |
| SQL | Parameterized | `SQL()` builder | Recommended |
| View syntax | Direct attributes | Direct attributes | No change |

## Breaking Changes

**None.** v17→v18 is mostly additive.

## Recommended Migrations

### 1. Multi-Company Record Rules

**v17:** `('company_id', 'in', company_ids)`
**v18:** `('company_id', 'in', allowed_company_ids)`

### 2. Model Company Validation

**v17 (manual):**
```python
class MyModel(models.Model):
    _name = 'my.model'
    company_id = fields.Many2one('res.company', default=lambda self: self.env.company)
    partner_id = fields.Many2one('res.partner', domain="[('company_id','in',[company_id,False])]")

    @api.constrains('partner_id', 'company_id')
    def _check_company(self):
        for record in self:
            if record.partner_id.company_id and record.partner_id.company_id != record.company_id:
                raise ValidationError(_("Mismatch"))
```

**v18 (automatic):**
```python
class MyModel(models.Model):
    _name = 'my.model'
    _check_company_auto = True
    company_id = fields.Many2one('res.company', default=lambda self: self.env.company)
    partner_id = fields.Many2one('res.partner', check_company=True)
    # No manual constraint needed
```

**Migration:** Add `_check_company_auto = True`, add `check_company=True` to relational fields, remove manual `_check_company` constraint.

### 3. Type Hints on Fields

**v17:** `name = fields.Char(required=True)`
**v18:** `name: str = fields.Char(required=True)`

### 4. SQL Builder

**v17:**
```python
self.env.cr.execute("SELECT id, name FROM my_table WHERE company_id = %s", [id])
```

**v18:**
```python
from odoo.tools import SQL
query = SQL("SELECT id, name FROM %(table)s WHERE company_id = %(company_id)s",
    table=SQL.identifier('my_table'), company_id=id)
self.env.cr.execute(query)
```

## No Change Required

Security Groups, Access Rights, Record Rules, View Security, Field Groups — unchanged.

## Checklist

- [ ] Update record rules: `company_ids` → `allowed_company_ids`
- [ ] Add `_check_company_auto = True` to multi-company models
- [ ] Add `check_company=True` to relational fields
- [ ] Remove manual company validation constraints
- [ ] Add type hints to relational fields (recommended)
- [ ] Convert raw SQL to `SQL()` builder (recommended)
- [ ] Test security rules with different user roles
- [ ] Verify multi-company access

## Testing
```python
def test_company_check(self):
    company_a = self.env['res.company'].create({'name': 'A'})
    company_b = self.env['res.company'].create({'name': 'B'})
    partner_a = self.env['res.partner'].create({'name':'PA', 'company_id': company_a.id})
    with self.assertRaises(ValidationError):
        self.env['my.model'].create({'name':'Test', 'company_id': company_b.id, 'partner_id': partner_a.id})
```

## Dual v17+v18 Support
```python
class MyModel(models.Model):
    _name = 'my.model'
    _check_company_auto = True  # Ignored in v17, works in v18
    partner_id = fields.Many2one('res.partner', check_company=True)  # Ignored in v17
    # Keep manual constraint for v17 compat
    @api.constrains('partner_id', 'company_id')
    def _check_company_compat(self):
        for record in self:
            if record.partner_id.company_id and record.partner_id.company_id != record.company_id:
                raise ValidationError(_("Mismatch"))
```

## GitHub Reference

- `odoo/models.py` — `_check_company_auto` implementation
- `odoo/fields.py` — `check_company` parameter
- `odoo/tools/sql.py` — `SQL` builder
