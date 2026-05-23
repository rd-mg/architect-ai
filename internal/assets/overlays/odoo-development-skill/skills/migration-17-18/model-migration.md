# Odoo Model Patterns Migration: 17.0 → 18.0

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  MODEL MIGRATION GUIDE: Odoo 17.0 → 18.0                                     ║
║  Focus: Company checks, SQL builder, type hints                              ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## New Features Summary

| Feature | v17 | v18 | Recommendation |
|---------|-----|-----|----------------|
| `_check_company_auto` | N/A | Available | Add to multi-company models |
| `check_company` on fields | N/A | Available | Add to cross-company fields |
| `SQL()` builder | N/A | Recommended | Use for raw SQL |
| Type hints | Optional | Recommended | Add where possible |

## NEW: Automatic Company Validation

**v17 (manual):**
```python
class MyModel(models.Model):
    _name = 'my.model'
    company_id = fields.Many2one('res.company')
    partner_id = fields.Many2one('res.partner')

    @api.constrains('partner_id', 'company_id')
    def _check_partner_company(self):
        for record in self:
            if record.partner_id.company_id and record.partner_id.company_id != record.company_id:
                raise ValidationError(_("Partner company must match."))
```

**v18 (automatic):**
```python
class MyModel(models.Model):
    _name = 'my.model'
    _check_company_auto = True  # Enable automatic checking

    company_id = fields.Many2one('res.company', required=True)
    partner_id = fields.Many2one('res.partner', check_company=True)
    # No manual constraint needed!
```

## NEW: SQL Builder

**v17 (raw SQL):**
```python
def _get_statistics(self):
    self.env.cr.execute("""
        SELECT partner_id, SUM(amount) as total
        FROM %s WHERE company_id = %%s AND state = %%s
        GROUP BY partner_id
    """ % self._table, (self.env.company.id, 'confirmed'))
```

**v18 (SQL Builder):**
```python
from odoo.tools import SQL

def _get_statistics(self):
    query = SQL("""
        SELECT partner_id, SUM(amount) as total
        FROM %s WHERE company_id = %s AND state = %s
        GROUP BY partner_id
    """, SQL.identifier(self._table), self.env.company.id, 'confirmed')
    self.env.cr.execute(query)
```

**Benefits:** SQL injection prevention, type-safe identifiers, clear param binding.

## RECOMMENDED: Type Hints

**v17:**
```python
def calculate_total(self, include_tax=True, discount=None):
    total = sum(self.mapped('amount'))
    return (total * 1.21 if include_tax else total) - (discount or 0)
```

**v18:**
```python
from typing import Optional

def calculate_total(self, include_tax: bool = True, discount: Optional[float] = None) -> float:
    total = sum(self.mapped('amount'))
    return (total * 1.21 if include_tax else total) - (discount or 0.0)
```

## Security Rules

**v17:** `[('company_id', 'in', company_ids)]`
**v18:** `[('company_id', 'in', allowed_company_ids)]`

## Checklist

### Multi-Company Models
- [ ] Add `_check_company_auto = True`
- [ ] Add `check_company=True` to relevant Many2one fields
- [ ] Remove manual company validation constraints
- [ ] Update record rules → `allowed_company_ids`

### SQL Queries
- [ ] Import `from odoo.tools import SQL`
- [ ] Replace raw SQL with `SQL()` builder
- [ ] Use `SQL.identifier()` for table/column names

### Methods (Recommended)
- [ ] Add type hints to params and return types

### Testing
- [ ] Test multi-company scenarios
- [ ] Test company switching
- [ ] Verify SQL queries

## Backward Compatibility (v17+v18)

```python
try:
    from odoo.tools import SQL
    HAS_SQL = True
except ImportError:
    HAS_SQL = False
```

## Common Issues

**`company_id required when _check_company_auto is True`** → Add `required=True`
**`'str' object has no attribute 'as_string'`** → Use `SQL.identifier()` for table names
**`ImportError: cannot import Optional`** → Use Python 3.10+ or `from typing import Optional`
