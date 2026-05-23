# Odoo Version Knowledge: 17 to 18 Migration

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  VERSION MIGRATION: 17.0 → 18.0                                              ║
║  Critical changes, breaking changes, and migration patterns                  ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## Breaking Changes Summary

| Category | Change | Impact |
|----------|--------|--------|
| Multi-Company | `_check_company_auto` recommended | High - Review all models |
| Relations | `check_company=True` recommended | High - Update Many2one |
| SQL | `SQL()` builder recommended | Medium - Update raw queries |
| Type Hints | Recommended for methods | Low - Add gradually |
| Security | Enhanced company isolation | Medium - Review rules |

## Multi-Company Framework

v18 introduces stricter multi-company validation with automatic checking.

**v17 (manual):**
```python
class MyModel(models.Model):
    _name = 'my.model'
    company_id = fields.Many2one('res.company')
    partner_id = fields.Many2one('res.partner')
    @api.constrains('partner_id', 'company_id')
    def _check_company(self):
        for record in self:
            if record.partner_id.company_id and record.partner_id.company_id != record.company_id:
                raise ValidationError("Mismatch!")
```

**v18 (automatic):**
```python
class MyModel(models.Model):
    _name = 'my.model'
    _check_company_auto = True
    company_id = fields.Many2one('res.company', required=True)
    partner_id = fields.Many2one('res.partner', check_company=True)
```

### _check_company_auto Behavior
Validates on write(). All fields with `check_company=True` checked against `company_id`.

## Record Rules: allowed_company_ids

```xml
<record id="rule_my_model_company" model="ir.rule">
    <field name="domain_force">[('company_id', 'in', allowed_company_ids)]</field>
</record>
```

## SQL() Builder Introduction

**v17 (discouraged):**
```python
self.env.cr.execute("SELECT id FROM my_model WHERE state = %s", ('confirmed',))
```

**v18 (recommended):**
```python
from odoo.tools import SQL
self.env.cr.execute(SQL("SELECT id FROM my_model WHERE state = %s", 'confirmed'))
self.env.cr.execute(SQL("UPDATE my_model SET state = %(state)s WHERE id IN %(ids)s",
    state='done', ids=tuple(self.ids)))
```

**Composable queries:**
```python
base = SQL("SELECT * FROM my_model WHERE active = %s", True)
filtered = SQL("%s AND company_id = %s", base, self.env.company.id)
table = SQL.identifier('my_model')
column = SQL.identifier('state')
query = SQL("SELECT %s FROM %s", column, table)
```

## Type Hints (Recommended)

```python
class MyModel(models.Model):
    _name = 'my.model'
    _check_company_auto = True
    name: str = fields.Char(required=True)
    active: bool = fields.Boolean(default=True)
    amount: float = fields.Float()

    def action_confirm(self) -> bool:
        for record in self:
            record.state = 'confirmed'
        return True

    @api.model_create_multi
    def create(self, vals_list: list[dict]) -> 'MyModel':
        return super().create(vals_list)
```

## Enhanced Indexing

```python
name = fields.Char(index='trigram')       # ILIKE
code = fields.Char(index='btree_not_null') # Exclude NULLs
date = fields.Date(index=True)             # Standard
```

## OWL 2.x Improvements

```javascript
setup() {
    this.orm = useService("orm");
    this.notification = useService("notification");
    this.state = useState({ records: [], loading: true });
    onWillStart(async () => { await this.loadRecords(); });
}
```

## GitHub Verification

```
https://raw.githubusercontent.com/odoo/odoo/18.0/odoo/models.py
https://raw.githubusercontent.com/odoo/odoo/18.0/odoo/tools/sql.py
```

## Checklist

- [ ] Add `_check_company_auto = True` to multi-company models
- [ ] Add `check_company=True` to relational fields
- [ ] Update record rules to use `allowed_company_ids`
- [ ] Start migrating raw SQL to `SQL()` builder
- [ ] Add type hints to new/modified methods
- [ ] Review and test multi-company scenarios
- [ ] Update index definitions for search fields

## Common Errors

**`ValidationError: check_company failed`** → Ensure related records same company
**`String SQL queries discouraged`** → Use `SQL()` builder
**Company isolation violations** → Add `_check_company_auto` and update rules

## Multi-Company Testing
```python
def test_multi_company_isolation(self):
    company2 = self.env['res.company'].create({'name': 'C2'})
    user_c2 = self.env['res.users'].create({
        'name':'U2', 'login':'u2', 'company_id': company2.id, 'company_ids': [(6,0,[company2.id])]})
    record = self.env['my.model'].create({'name':'T', 'company_id': self.env.company.id})
    visible = self.env['my.model'].with_user(user_c2).search([])
    self.assertNotIn(record.id, visible.ids)
```

## v17 vs v18 Model Comparison

```python
# v17:
class M17(models.Model):
    _name = 'my.model'
    company_id = fields.Many2one('res.company')
    partner_id = fields.Many2one('res.partner')
    @api.constrains('partner_id')
    def _check(self):
        for r in self:
            if r.partner_id.company_id and r.partner_id.company_id != r.company_id:
                raise ValidationError("Mismatch")

# v18:
class M18(models.Model):
    _name = 'my.model'
    _check_company_auto = True
    company_id = fields.Many2one('res.company', required=True)
    partner_id = fields.Many2one('res.partner', check_company=True)
```
