# Odoo Version Knowledge: 16 to 17 Migration

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  VERSION MIGRATION: 16.0 → 17.0                                              ║
║  Critical changes, breaking changes, and migration patterns                  ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## Breaking Changes Summary

| Category | Change | Impact |
|----------|--------|--------|
| Views | `attrs=` REMOVED | **CRITICAL** - Must migrate |
| ORM | `@api.model_create_multi` required | High - All create() |
| Domain | New expression syntax | Medium |
| Security | Enhanced record rule validation | Medium |
| JS | ES6 modules fully enforced | Medium |

## CRITICAL: attrs Removal

Most significant breaking change in v17.

**v16:**
```xml
<field name="partner_id" attrs="{'invisible': [('state','=','draft')], 'readonly': [('state','!=','draft')]}"/>
<group attrs="{'invisible': [('show_details','=',False)]}"/>
```

**v17:**
```xml
<field name="partner_id" invisible="state == 'draft'" readonly="state != 'draft'"/>
<group invisible="not show_details"/>
```

### Expression Syntax

| attrs Domain | v17 Expression |
|--------------|----------------|
| `[('f','=',v)]` | `f == v` |
| `[('f','!=',v)]` | `f != v` |
| `[('f','>',v)]` | `f > v` |
| `[('f','in',[a,b])]` | `f in [a, b]` |
| `[('f','=',True)]` | `f` |
| `[('f','=',False)]` | `not f` |
| `['&', A, B]` | `A and B` |
| `['\|', A, B]` | `A or B` |
| `['!', A]` | `not A` |

**Complex:**
```xml
<!-- Old: attrs="{'invisible': ['|', ('state','=','draft'), '&', ('type','=','service'), ('qty','=',0)]}" -->
<field name="x" invisible="state == 'draft' or (type == 'service' and qty == 0)"/>
```

## Create Method

**v16 (allowed):**
```python
@api.model
def create(self, vals):
    if 'name' not in vals: vals['name'] = 'Default'
    return super().create(vals)
```

**v17 (required):**
```python
@api.model_create_multi
def create(self, vals_list):
    for vals in vals_list:
        if 'name' not in vals: vals['name'] = 'Default'
    return super().create(vals_list)
```

## New: Tree Column Visibility

```xml
<field name="internal_code" column_invisible="True"/>
<field name="cost" column_invisible="not context.get('show_cost')"/>
```

## Domain Expression Improvements

**Parent access:** `invisible="parent.state == 'done'"`
**Context:** `invisible="context.get('hide_price')"`

## Enhanced Record Rules

v17 validates domain expressions more strictly.

## ORM Improvements

Better prefetching. Explicit control:
```python
records = self.with_prefetch(('partner_id', 'line_ids'))
```

## Strict ES6 Enforcement

```javascript
/** @odoo-module **/
import { Component } from "@odoo/owl";
// v17 requires proper ES6 imports
```

## GitHub Verification

```
https://raw.githubusercontent.com/odoo/odoo/17.0/odoo/tools/view_validation.py
https://raw.githubusercontent.com/odoo/odoo/17.0/odoo/osv/expression.py
```

## Checklist

- [ ] **CRITICAL:** Replace ALL `attrs=` with inline expressions
- [ ] Update all `@api.model` create → `@api.model_create_multi`
- [ ] Review record rules for stricter validation
- [ ] Test all view visibility conditions
- [ ] Verify JS modules use ES6 imports
- [ ] Test form/tree/kanban views thoroughly

## Common Errors

**`attrs attribute not supported`** → Replace with inline `invisible=`/`readonly=`/`required=`
**`create() takes 2 positional but 3 given`** → Change to `@api.model_create_multi create(self, vals_list)`
**`Invalid domain expression`** → Review syntax — v17 stricter about field refs
