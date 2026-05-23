# Odoo Module Migration: 18.0 → 19.0

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  MIGRATION GUIDE: Odoo 18.0 → 19.0                                           ║
║  ONLY changes between these specific versions.                               ║
║  Note: v19 in development — patterns may change.                             ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## Breaking Changes Summary

| Component | v18 | v19 | Action |
|-----------|-----|-----|--------|
| Type hints | Recommended | **Mandatory** | Must add |
| `SQL()` builder | Recommended | **Mandatory** | Must migrate |
| Raw SQL strings | Deprecated | **Removed** | Must migrate |
| OWL | 2.x | **3.x** | Must update |
| Python | 3.10+ | 3.12+ | Upgrade |
| `_check_company_auto` | Recommended | Standard | Already adopted |

## MANDATORY: Type Hints

**v18 (recommended):**
```python
def calculate_total(self, include_tax=True, discount=None):
    total = sum(self.mapped('amount'))
    return (total * 1.21 if include_tax else total) - (discount or 0)
```

**v19 (required):**
```python
def calculate_total(self, include_tax: bool = True, discount: Optional[float] = None) -> float:
    total = sum(self.mapped('amount'))
    return (total * 1.21 if include_tax else total) - (discount or 0.0)
```

**Complete:**
```python
class MyModel(models.Model):
    _name = 'my.model'

    @api.model_create_multi
    def create(self, vals_list: list[dict[str, Any]]) -> 'MyModel': ...
    def write(self, vals: dict[str, Any]) -> bool: ...
    def unlink(self) -> bool: ...
    def copy(self, default: Optional[dict[str, Any]] = None) -> 'MyModel': ...
    def action_confirm(self) -> None: ...
```

## MANDATORY: SQL Builder

**v18 (allowed):** `query = "SELECT id FROM %s WHERE company_id = %s" % (table, cid)`
**v19 (required):**
```python
from odoo.tools import SQL
query = SQL("SELECT id FROM %s WHERE company_id = %s", SQL.identifier(table), cid)
self.env.cr.execute(query)
```

**Complete ref:**
```python
SQL.identifier('my_table')                    # Safe table
SQL.identifier('table', 'column')             # Safe column
SQL('ORDER BY %s %s', SQL.identifier('date'), SQL('DESC'))  # Dynamic order
SQL("SELECT * FROM table WHERE id IN %s", (1,2,3))          # IN clause

# UNION
SQL("%s UNION ALL %s",
    SQL("SELECT id FROM t1 WHERE type = %s", 'a'),
    SQL("SELECT id FROM t2 WHERE type = %s", 'b'))
```

## OWL 3.x Migration

| Feature | OWL 2.x (v18) | OWL 3.x (v19) |
|---------|---------------|---------------|
| Reactivity | `useState` | Enhanced |
| Component | `Component` | Updated patterns |
| Lifecycle | Hooks-based | Refined hooks |
| Templates | QWeb | Enhanced QWeb |

**v19 Component:**
```javascript
import { Component, useState, useRef, onMounted } from "@odoo/owl";
import { useService } from "@web/core/utils/hooks";

export class MyComponent extends Component {
    static template = "my_module.MyComponent";
    static props = { recordId: { type: Number, optional: true } };
    setup() {
        this.orm = useService("orm");
        this.state = useState({ data: [], loading: true });
        onMounted(() => { this.loadData(); });
    }
}
registry.category("actions").add("my_module.my_component", MyComponent);
```

## Python 3.12+
```python
# Type param syntax (PEP 695)
def process_items[T](items: list[T]) -> list[T]: ...

# Enhanced match
match self.state:
    case 'draft': self.action_confirm()
    case _: raise UserError(_("Invalid"))
```

## Manifest Version

`'version': '18.0.1.0.0'` → `'version': '19.0.1.0.0'`

## Checklist

### Models (Python) — CRITICAL
- [ ] Type hints on ALL method signatures
- [ ] Return types on ALL methods
- [ ] Replace ALL raw SQL with `SQL()` builder
- [ ] Verify `from odoo.tools import SQL`
- [ ] Update to Python 3.12+

### OWL (JavaScript) — CRITICAL
- [ ] Update to OWL 3.x patterns
- [ ] Review lifecycle hooks
- [ ] Test all UI components

### Views (XML)
- [ ] Verify with v19
- [ ] Test dynamic visibility rules
- [ ] Verify template inheritance

### Security
- [ ] Verify record rules
- [ ] Test multi-company scenarios

### Manifest
- [ ] `18.0.x.x.x` → `19.0.x.x.x`
- [ ] Verify dependencies v19 compatible
- [ ] Update asset declarations for OWL 3.x

## Common Errors

**`Missing type annotation`** → Add type hints
**`Raw SQL not allowed`** → Use `SQL()` builder
**`Component lifecycle hook not found`** → Update to OWL 3.x patterns
**`SyntaxError requires Python 3.12+`** → Upgrade Python

## GitHub Reference
- https://github.com/odoo/odoo/tree/19.0 (when available)
- Odoo 19.0 release notes
