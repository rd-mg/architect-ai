# Odoo Module Migration: 17.0 → 18.0

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  MIGRATION GUIDE: Odoo 17.0 → 18.0                                           ║
║  ONLY changes between these specific versions.                               ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## Changes Summary

| Component | v17 | v18 | Action |
|-----------|-----|-----|--------|
| `_check_company_auto` | N/A | Recommended | Add to models |
| `check_company` on fields | N/A | Recommended | Add to relations |
| `SQL()` builder | New | Recommended | Use for raw SQL |
| Type hints | Optional | Recommended | Add where possible |
| Raw SQL | Allowed | Deprecated | Migrate to SQL() |
| Python | 3.10+ | 3.10+, 3.12 recommended | Verify |

## NEW: Company Check Automation

**v17 (manual):**
```python
class MyModel(models.Model):
    _name = 'my.model'
    company_id = fields.Many2one('res.company')
    partner_id = fields.Many2one('res.partner')

    def write(self, vals):
        if 'partner_id' in vals:
            partner = self.env['res.partner'].browse(vals['partner_id'])
            if partner.company_id and partner.company_id != self.company_id:
                raise UserError(_("Company mismatch."))
        return super().write(vals)
```

**v18 (automatic):**
```python
class MyModel(models.Model):
    _name = 'my.model'
    _check_company_auto = True
    company_id = fields.Many2one('res.company', required=True)
    partner_id = fields.Many2one('res.partner', check_company=True)
```

## NEW: SQL Builder

**v17 (raw):** String formatting vulnerable to injection
**v18 (SQL builder):**
```python
from odoo.tools import SQL

query = SQL("SELECT * FROM %s WHERE id = %s", SQL.identifier('my_table'), 123)
self.env.cr.execute(query)
```

**SQL Builder Reference:**
```python
SQL.identifier('my_table')                    # Safe table/column
SQL.identifier('my_table', 'column_name')     # Safe column ref
SQL('ORDER BY create_date DESC')              # Raw fragment (careful)
ids = (1, 2, 3)
SQL("SELECT * FROM table WHERE id IN %s", ids)  # IN clause
```

## RECOMMENDED: Type Hints

**v17:** `def process_partner(self, partner_id, options=None):`
**v18:** `def process_partner(self, partner_id: int, options: Optional[dict] = None) -> str:`

## Multi-Company Updates

```python
class MyModel(models.Model):
    _name = 'my.model'
    _check_company_auto = True
    company_id = fields.Many2one('res.company', required=True, default=lambda self: self.env.company)
    partner_id = fields.Many2one('res.partner', check_company=True)
```

**Security Rule:**
```xml
<field name="domain_force">['|', ('company_id', '=', False), ('company_id', 'in', allowed_company_ids)]</field>
```

## OWL Updates (2.x Enhanced)

```javascript
// v18: Consistent service usage
setup() {
    this.company = useService("company");
    this.orm = useService("orm");
    this.notification = useService("notification");
}
```

## Manifest Version

`'version': '17.0.1.0.0'` → `'version': '18.0.1.0.0'`

## Checklist

### Models
- [ ] Add `_check_company_auto = True` to multi-company models
- [ ] Add `check_company=True` to relational fields
- [ ] Replace raw SQL with `SQL()` builder
- [ ] Add type hints to method signatures
- [ ] Update Python compat to 3.10+ (3.12 recommended)

### Security
- [ ] Update record rules: `company_ids` → `allowed_company_ids`
- [ ] Test multi-company access scenarios

### OWL
- [ ] Verify service usage patterns
- [ ] Test client actions

### Manifest
- [ ] `17.0.x.x.x` → `18.0.x.x.x`
- [ ] Verify dependencies v18 compatible

### Testing
- [ ] Test multi-company scenarios
- [ ] Test company switching
- [ ] Verify SQL queries

## Common Issues

**Company validation errors** → Ensure relational fields have proper company or use `check_company=False`
**SQL syntax errors** → Ensure `SQL()` properly imported from `odoo.tools`
**Type hint import errors** → Use Python 3.10+

## Performance

`_check_company_auto` adds overhead:
- Enable only on models that need it
- Use `with_context(check_company=False)` for batch operations

## Backward Compatibility

```python
try:
    from odoo.tools import SQL
    HAS_SQL_BUILDER = True
except ImportError:
    HAS_SQL_BUILDER = False
```

## GitHub Reference

- https://github.com/odoo/odoo/tree/18.0
- Odoo 18.0 release notes
