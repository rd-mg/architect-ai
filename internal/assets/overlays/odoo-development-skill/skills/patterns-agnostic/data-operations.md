# Data Operations Patterns

Consolidated from the following source files:
- `import-export-patterns.md` (architect-ai)
- `data-migration-patterns.md` (architect-ai)
- `sequence-numbering-patterns.md` (architect-ai)
- `context-environment-patterns.md` (architect-ai)
- `external-api-patterns.md` (architect-ai)

> **Version-specific syntax** → `patterns-{version}/model-patterns.md`
> `SQL()` mandatory from v19+ · `with_company` in v13+ · `create_multi` in v16+

---

## Environment & Context

### Managing Context & Sudo
```python
# 1. Modify context (temporary)
records = self.with_context(active_test=False).search([])

# 2. Sudo for bypass access rules (dangerous - use sparingly)
self.sudo().write({'state': 'done'})

# 3. Switching Company
company_records = self.with_company(company_id).search([])
```

### Batch Operations
```python
@api.model_create_multi
def create(self, vals_list):
    # ✅ ALWAYS use create_multi for performance
    return super().create(vals_list)
```

---

## Migration Scripts

### Pre-Migration (Schema Changes)
```python
# migrations/18.0.1.1/pre-migrate.py
def migrate(cr, version):
    # Rename column before ORM initializes to avoid data loss
    cr.execute("ALTER TABLE my_table RENAME COLUMN old_name TO new_name")
```

### Post-Migration (Data Transformation)
```python
# migrations/18.0.1.1/post-migrate.py
from odoo import api, SUPERUSER_ID

def migrate(cr, version):
    env = api.Environment(cr, SUPERUSER_ID, {})
    # Use ORM to update records safely
    env['my.model'].search([])._compute_total()
```

---

## External API (XML-RPC)

```python
import xmlrpc.client
common = xmlrpc.client.ServerProxy(f'{url}/xmlrpc/2/common')
uid = common.authenticate(db, username, password, {})
models = xmlrpc.client.ServerProxy(f'{url}/xmlrpc/2/object')
ids = models.execute_kw(db, uid, password, 'res.partner', 'search', [[['is_company', '=', True]]])
```

---

## Anti-Patterns

```python
# ❌ NEVER use cr.commit() inside Odoo methods (use for migrations/crons only).

# ❌ NEVER use with_context in a loop: cache the environment instead.
# ✅ CORRECT: env = self.with_context(key=val).env

# ❌ NEVER use cr.execute() with string formatting: SQL injection risk.
# ✅ CORRECT: cr.execute("SELECT * FROM table WHERE id = %s", (id,))
```

---

## Version Matrix

| Feature | v14-v15 | v16 | v17 | v18-v19 |
|---------|---------|-----|-----|---------|
| Create | `create()` | `@api.model_create_multi` | `create_multi` | `create_multi` |
| SQL Queries | `cr.execute` | `cr.execute` | `cr.execute` | `SQL()` |
| Environment | `Environment` | `Environment` | `Environment` | `Environment` |
| Migrations | `pre/post` | `pre/post` | `pre/post` | `pre/post` |
