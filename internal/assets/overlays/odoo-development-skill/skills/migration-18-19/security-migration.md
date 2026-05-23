# Odoo Security Guide - Migration 18.0 → 19.0

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  MIGRATION GUIDE: ODOO 18.0 → 19.0 SECURITY                                  ║
║  Upgrading security code from v18 to v19.                                    ║
║  NOTE: v19 in development — patterns may change.                             ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## Security Changes Overview

| Component | v18 | v19 | Required |
|-----------|-----|-----|----------|
| Type hints | Recommended | **Mandatory** | **REQUIRED** |
| SQL builder | Recommended | **Mandatory** | **REQUIRED** |
| Python | 3.11+ | 3.12+ | Check |
| OWL | 2.x | 3.x | For OWL components |

## Breaking Changes

### 1. Type Hints Now Mandatory

**v18 (recommended):**
```python
class MyModel(models.Model):
    _name = 'my.model'
    name = fields.Char(required=True)
    partner_id = fields.Many2one('res.partner')
```

**v19 (mandatory):**
```python
from __future__ import annotations
class MyModel(models.Model):
    _name = 'my.model'
    name: str = fields.Char(required=True)
    partner_id: int = fields.Many2one('res.partner')
    line_ids: list[int] = fields.One2many('my.line', 'parent_id')
```

### 2. SQL Builder Now Mandatory

**v18 (allowed):**
```python
self.env.cr.execute("SELECT id FROM my_table WHERE company_id = %s", [self.env.company.id])
```

**v19 (mandatory):**
```python
from odoo.tools import SQL
query = SQL("SELECT id FROM %(table)s WHERE company_id = %(company_id)s",
    table=SQL.identifier('my_table'), company_id=self.env.company.id)
self.env.cr.execute(query)
```

## Migration Script
```python
import re

def add_type_hints(content):
    patterns = {
        r"(\w+)\s*=\s*fields\.Char\(": r"\1: str = fields.Char(",
        r"(\w+)\s*=\s*fields\.Boolean\(": r"\1: bool = fields.Boolean(",
        r"(\w+)\s*=\s*fields\.Integer\(": r"\1: int = fields.Integer(",
        r"(\w+)\s*=\s*fields\.Float\(": r"\1: float = fields.Float(",
        r"(\w+)\s*=\s*fields\.Many2one\(": r"\1: int = fields.Many2one(",
        r"(\w+)\s*=\s*fields\.One2many\(": r"\1: list[int] = fields.One2many(",
    }
    for pat, repl in patterns.items():
        content = re.sub(pat, repl, content)
    if "from __future__ import annotations" not in content:
        content = "from __future__ import annotations\n" + content
    return content
```

## Type Hint Examples

### Fields
```python
name: str = fields.Char(required=True)
active: bool = fields.Boolean(default=True)
sequence: int = fields.Integer(default=10)
amount: float = fields.Float()
state: str = fields.Selection([('draft','Draft'),('done','Done')], default='draft')
company_id: int = fields.Many2one('res.company')
partner_id: int = fields.Many2one('res.partner')
line_ids: list[int] = fields.One2many('secure.model.line', 'parent_id')
tag_ids: list[int] = fields.Many2many('secure.model.tag')
```

### Methods
```python
@api.model_create_multi
def create(self, vals_list: list[dict[str, Any]]) -> SecureModel: ...
def write(self, vals: dict[str, Any]) -> bool: ...
def unlink(self) -> bool: ...
def copy(self, default: dict[str, Any] | None = None) -> SecureModel: ...
def action_confirm(self) -> dict[str, Any] | bool: ...
```

## SQL Builder Migration

**Simple:**
```python
# v18
self.env.cr.execute("SELECT id FROM my_table WHERE active = %s", [True])
# v19
query = SQL("SELECT id FROM %(table)s WHERE active = %(active)s",
    table=SQL.identifier('my_table'), active=True)
self.env.cr.execute(query)
```

**Complex with joins:**
```python
query = SQL("""
    SELECT m.id, p.name AS partner_name, COALESCE(SUM(l.amount),0) AS total
    FROM %(main)s m LEFT JOIN %(partner)s p ON m.partner_id = p.id
    WHERE m.company_id IN %(cids)s AND m.state = %(state)s
    GROUP BY m.id, p.name
    ORDER BY %(order)s %(dir)s
""",
    main=SQL.identifier(self._table),
    partner=SQL.identifier('res_partner'),
    cids=tuple(self.env.companies.ids) or (0,),
    state='confirmed',
    order=SQL.identifier('total_amount'),
    dir=SQL('DESC'))
self.env.cr.execute(query)
```

## OWL 3.x Migration

**v18 (OWL 2.x):**
```javascript
import { Component, useState } from "@odoo/owl";
export class MyComponent extends Component {
    static template = "my_module.MyComponent";
    setup() { this.state = useState({ count: 0 }); }
}
```

**v19 (OWL 3.x):**
```javascript
import { Component, useState } from "@odoo/owl";
export class MyComponent extends Component {
    static template = "my_module.MyComponent";
    static props = {};
    setup() { this.state = useState({ count: 0 }); }
}
```

## Checklist

- [ ] **CRITICAL:** Type hints on ALL field definitions
- [ ] **CRITICAL:** Type hints on ALL method signatures
- [ ] **CRITICAL:** Convert ALL raw SQL to `SQL()` builder
- [ ] Add `from __future__ import annotations` to all Python files
- [ ] Update Python to 3.12+
- [ ] Migrate OWL 2.x → 3.x components
- [ ] Test all SQL queries
- [ ] Verify type hints don't cause runtime errors

## No Change Required

Security Groups, Access Rights, Record Rules, Field Groups, View visibility — unchanged.

## GitHub Reference

- `odoo/fields.py` — Type hint support
- `odoo/tools/sql.py` — SQL builder
- `odoo/models.py` — Model type annotations
- `addons/web/static/src/` — OWL 3.x
