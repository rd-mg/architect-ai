# Odoo Module Migration: 16.0 → 17.0

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  MIGRATION GUIDE: Odoo 16.0 → 17.0                                           ║
║  ONLY changes between these specific versions.                               ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## Breaking Changes Summary

| Component | v16 | v17 | Action |
|-----------|-----|-----|--------|
| `attrs` | Deprecated | **REMOVED** | Must migrate |
| `states` | Deprecated | **REMOVED** | Must migrate |
| `@api.model_create_multi` | Recommended | **Mandatory** | Must add |
| Direct attributes | Supported | Required | Use exclusively |
| OWL | 2.x | 2.x (enhanced) | Minor |

## CRITICAL: attrs Removal

**v16 (deprecated):**
```xml
<button name="action_confirm" attrs="{'invisible': [('state','!=','draft')]}"/>
<field name="partner_id" attrs="{'readonly': [('state','!=','draft')]}"/>
```

**v17 (required):**
```xml
<button name="action_confirm" invisible="state != 'draft'"/>
<field name="partner_id" readonly="state != 'draft'"/>
```

### Migration Pattern

| v16 attrs | v17 Replacement |
|-----------|-----------------|
| `[('f','=',v)]` | `f == v` |
| `[('f','!=',v)]` | `f != v` |
| `[('f','in',[a,b])]` | `f in [a, b]` |
| `[('f','not in',[a,b])]` | `f not in [a, b]` |
| `[('f','=',True)]` | `f` |
| `[('f','=',False)]` | `not f` |
| Multiple AND | `cond1 and cond2` |
| Multiple OR | `cond1 or cond2` |

**Complex:**
```xml
<!-- v16 -->
<field name="amount" attrs="{'invisible': [('state','=','draft'),('amount','=',0)],
    'readonly': ['|',('state','!=','draft'),('locked','=',True)]}"/>

<!-- v17 -->
<field name="amount" invisible="state == 'draft' and amount == 0"
    readonly="state != 'draft' or locked"/>
```

## CRITICAL: states Removal

**v16:** `<field name="partner_id" states="draft,sent"/>`
**v17:** `<field name="partner_id" invisible="state not in ('draft','sent')"/>`

## MANDATORY: @api.model_create_multi

**v16 (optional):**
```python
@api.model
def create(self, vals):
    if not vals.get('name'):
        vals['name'] = self.env['ir.sequence'].next_by_code('my.model')
    return super().create(vals)
```

**v17 (required):**
```python
@api.model_create_multi
def create(self, vals_list):
    for vals in vals_list:
        if not vals.get('name'):
            vals['name'] = self.env['ir.sequence'].next_by_code('my.model')
    return super().create(vals_list)
```

## Manifest Version

`'version': '16.0.1.0.0'` → `'version': '17.0.1.0.0'`

## Python Changes

v16: 3.8+ → v17: 3.10+. Type hints encouraged:
```python
def process_data(self, partner_id: int, options: dict = None) -> bool: ...
```

## OWL 2.x Updates

Minor. No breaking changes. Registration unchanged:
```javascript
registry.category("actions").add("my_module.component", MyComponent);
```

## Checklist

### Views (XML)
- [ ] Replace ALL `attrs="{'invisible': ...}"` → `invisible="..."`
- [ ] Replace ALL `attrs="{'readonly': ...}"` → `readonly="..."`
- [ ] Replace ALL `attrs="{'required': ...}"` → `required="..."`
- [ ] Replace ALL `states="..."` → `invisible="state not in (...)"`
- [ ] Convert domain syntax to Python expressions

### Models (Python)
- [ ] Add `@api.model_create_multi` to ALL `create()` methods
- [ ] Update `create(vals)` → `create(vals_list)` signature
- [ ] Update Python version compat to 3.10+
- [ ] Add type hints where appropriate

### Manifest
- [ ] Version `16.0.x.x.x` → `17.0.x.x.x`
- [ ] Verify dependencies v17 compatible

### Testing
- [ ] Run all tests with v17
- [ ] Test all form views for visibility rules
- [ ] Test all buttons for state-based visibility

## Common Errors

**`attrs="..." not supported`** → Convert to Python expressions
**`create() got unexpected keyword argument 'vals'`** → Use `vals_list`
**Invalid expression** → Use `==` not `=` for comparison

## Automated Migration Script

```python
def convert_domain_to_expr(domain_str):
    patterns = [
        (r"\('(\w+)',\s*'=',\s*'([^']+)'\)", r"\1 == '\2'"),
        (r"\('(\w+)',\s*'=',\s*True\)", r"\1"),
        (r"\('(\w+)',\s*'=',\s*False\)", r"not \1"),
        (r"\('(\w+)',\s*'!=',\s*'([^']+)'\)", r"\1 != '\2'"),
    ]
    for p, r in patterns:
        domain_str = re.sub(p, r, domain_str)
    return domain_str
```

## GitHub Reference

- https://github.com/odoo/odoo/tree/17.0
- Odoo 17.0 release notes
