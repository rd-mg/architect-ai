# Odoo Security Guide - Migration 16.0 → 17.0

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  MIGRATION GUIDE: ODOO 16.0 → 17.0 SECURITY                                  ║
║  Upgrading security code from v16 to v17.                                    ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## Security Changes Overview

| Component | v16 | v17 | Required |
|-----------|-----|-----|----------|
| View visibility | `attrs` (deprecated) | Direct attributes | **REQUIRED** |
| Create method | `@api.model_create_multi` optional | Mandatory | **REQUIRED** |
| Python | 3.8+ | 3.10+ | Check |
| Record rules | `company_ids` | `company_ids` | No change |

## Breaking Changes

### 1. attrs Attribute REMOVED

**v16 (breaks in v17):**
```xml
<field name="secret_field" attrs="{'invisible': [('state','=','draft')]}"/>
<field name="amount" attrs="{'readonly': [('state','!=','draft')], 'required': [('type','=','invoice')]}"/>
```

**v17:**
```xml
<field name="secret_field" invisible="state == 'draft'"/>
<field name="amount" readonly="state != 'draft'" required="type == 'invoice'"/>
```

### 2. Domain Syntax Changed

**v16:** `attrs="{'invisible': ['|', ('state','=','draft'), ('type','=','invoice')]}"`
**v17:** `invisible="state == 'draft' or type == 'invoice'"`

### 3. Group Checks

**v16:** No direct way in attrs
**v17:** `invisible="not user_has_groups('base.group_system')"`

## Migration Script for attrs
```python
def convert_domain_to_expression(domain_str):
    # Simplified converter
    domain_str = domain_str.replace("('", '"').replace("', '", ' == ').replace("')", '"')
    return domain_str
```

## Manual Migration Examples

**Simple:** `attrs="{'invisible': [('state','=','draft')]}"` → `invisible="state == 'draft'"`

**AND:** `[('state','=','draft'), ('type','=','draft')]` → `state == 'draft' and type == 'draft'`

**OR:** `['|', ('state','=','draft'), ('state','=','cancelled')]` → `state == 'draft' or state == 'cancelled'`

**Complex:** `['|', '&', ('state','=','draft'), ('type','=','a'), ('active','=',False)]` → `(state == 'draft' and type == 'a') or not active`

**Multiple attrs:** `attrs="{'readonly': [('state','!=','draft')], 'required': [('type','=','invoice')]}"` → `readonly="state != 'draft'" required="type == 'invoice'"`

## Create Method

**v16 (optional):**
```python
@api.model
def create(self, vals): return super().create(vals)
# or
@api.model_create_multi
def create(self, vals_list): return super().create(vals_list)
```

**v17 (mandatory):**
```python
@api.model_create_multi
def create(self, vals_list): return super().create(vals_list)
```

## No Change Required

Security Groups, Access Rights, Record Rules, Field Groups — same syntax.

## Checklist

- [ ] **CRITICAL:** Replace ALL `attrs` with direct attributes
- [ ] Convert domain syntax to Python expressions
- [ ] Update create methods to `@api.model_create_multi`
- [ ] Update Python to 3.10+
- [ ] Test all view visibility conditions
- [ ] Test button visibility with different states
- [ ] Verify group-based visibility with `user_has_groups()`

## Common Mistakes

**Wrong:** `invisible="('state', '=', 'draft')"` → **Correct:** `invisible="state == 'draft'"`
**Wrong:** `invisible="active = False"` → **Correct:** `invisible="not active"`
**Wrong:** `invisible="state == draft"` → **Correct:** `invisible="state == 'draft'"`

## Testing After Migration
```python
def test_view_visibility(self):
    record = self.env['my.model'].create({'name':'Test','state':'draft'})
    view = self.env.ref('my_module.view_form')
    # Verify invisible fields not shown for draft
```

## Rollback

Cannot support v16+v17 with same view files. Options:
1. Version-specific view files
2. Conditional manifest loading (not recommended)
3. Separate branches

## GitHub Reference

- `odoo/tools/view_validation.py`
- `addons/web/static/src/views/`
