# Odoo Model Patterns Migration: 18.0 → 19.0

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  MODEL MIGRATION GUIDE: Odoo 18.0 → 19.0                                     ║
║  Focus: Mandatory type hints, mandatory SQL builder                          ║
║  Note: v19 in development — patterns may change.                             ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## Breaking Changes Summary

| Feature | v18 | v19 | Action |
|---------|-----|-----|--------|
| Type hints | Recommended | **Mandatory** | Must add |
| `SQL()` builder | Recommended | **Mandatory** | Must use |
| Raw SQL strings | Deprecated | **Removed** | Must migrate |

## MANDATORY: Type Hints

**v18 (optional):**
```python
def calculate_totals(self, options=None):
    options = options or {}
    results = []
    for record in self:
        total = sum(record.line_ids.mapped('amount'))
        if options.get('include_tax'): total *= 1.21
        results.append({'id': record.id, 'total': total})
    return results

@api.model_create_multi
def create(self, vals_list):
    for vals in vals_list:
        if not vals.get('name'): vals['name'] = 'New'
    return super().create(vals_list)
```

**v19 (mandatory):**
```python
from __future__ import annotations
from typing import Any, Optional
from collections.abc import Sequence

def calculate_totals(self, options: Optional[dict[str, Any]] = None) -> list[dict[str, Any]]:
    options = options or {}
    for record in self:
        total: float = sum(record.line_ids.mapped('amount'))
        if options.get('include_tax'): total *= 1.21
        results.append({'id': record.id, 'total': total})
    return results

@api.model_create_multi
def create(self, vals_list: list[dict[str, Any]]) -> 'MyModel':
    for vals in vals_list:
        if not vals.get('name'): vals['name'] = 'New'
    return super().create(vals_list)
```

## MANDATORY: SQL Builder

**v18 (allowed, deprecated):**
```python
def _get_report_data(self):
    query = "SELECT id, name FROM %s WHERE company_id = %%s" % self._table
    self.env.cr.execute(query, (self.env.company.id,))
```

**v19 (required):**
```python
from odoo.tools import SQL

def _get_report_data(self) -> list[dict[str, Any]]:
    query = SQL("SELECT id, name FROM %s WHERE company_id = %s ORDER BY %s",
        SQL.identifier(self._table), self.env.company.id, SQL('create_date DESC'))
    self.env.cr.execute(query)
    return self.env.cr.dictfetchall()
```

## Type Hint Reference

```python
from __future__ import annotations
from typing import Any, Optional, Union, Literal
from collections.abc import Sequence, Mapping, Iterable

class MyModel(models.Model):
    _name = 'my.model'

    @api.model_create_multi
    def create(self, vals_list: list[dict[str, Any]]) -> 'MyModel': ...
    def write(self, vals: dict[str, Any]) -> bool: ...
    def unlink(self) -> bool: ...
    def copy(self, default: Optional[dict[str, Any]] = None) -> 'MyModel': ...

    @api.depends('line_ids.amount')
    def _compute_total(self) -> None: ...

    @api.constrains('amount')
    def _check_amount(self) -> None: ...

    def action_confirm(self) -> None: ...
    def action_view_records(self) -> dict[str, Any]: ...

    def process_data(self, partner_ids: list[int],
        options: Optional[dict[str, Any]] = None,
        mode: Literal['create','update','delete'] = 'create') -> tuple[int, list[str]]: ...
```

## SQL Builder Complete Reference

```python
from odoo.tools import SQL

query = SQL("SELECT * FROM %s WHERE id = %s", SQL.identifier('my_table'), record_id)

# Complex with joins
query = SQL("""
    SELECT m.id, p.name as partner, COALESCE(SUM(l.amount),0) as total
    FROM %s m LEFT JOIN %s p ON p.id = m.partner_id
    WHERE m.company_id IN %s AND m.state = %s
    GROUP BY m.id, p.name
    ORDER BY %s LIMIT %s OFFSET %s
""", SQL.identifier('my_model'), SQL.identifier('res_partner'),
    tuple(company_ids), 'confirmed', SQL('total DESC'), limit, offset)

# Dynamic conditions
conditions = [SQL("company_id = %s", cid)]
if state: conditions.append(SQL("state = %s", state))
where = SQL(" AND ").join(conditions)
query = SQL("SELECT * FROM %s WHERE %s", SQL.identifier(self._table), where)
```

## Checklist

### All Models
- [ ] Add `from __future__ import annotations` at file top
- [ ] Type hints on ALL method parameters
- [ ] Return type annotations on ALL methods

### SQL
- [ ] Replace ALL raw SQL with `SQL()` builder
- [ ] Verify all queries work

### Python
- [ ] Ensure Python 3.12+

### Testing
- [ ] Run type checker (mypy, pyright)
- [ ] Test all SQL queries
- [ ] Verify all methods work

## Common Errors

**`Missing type annotation for parameter`** → Add type hints
**`Raw SQL strings deprecated`** → Use `SQL()` builder
**`SyntaxError (type hint)`** → Ensure Python 3.12+
**`Circular import with type hints`** → Use `from __future__ import annotations`
