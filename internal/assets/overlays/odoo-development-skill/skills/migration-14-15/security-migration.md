# Odoo Security Guide - Migration 14.0 → 15.0

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  MIGRATION GUIDE: ODOO 14.0 → 15.0 SECURITY                                  ║
║  Guide for upgrading security code from v14 to v15.                          ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## Security Changes Overview

| Component | v14 | v15 | Required |
|-----------|-----|-----|----------|
| @api.multi | Deprecated | **Removed** | **REQUIRED** |
| Field tracking | `track_visibility` | `tracking` | **REQUIRED** |
| Python | 3.6+ | 3.8+ | Check |
| Chatter | Legacy | Simplified | Recommended |

## Breaking Changes

### 1. @api.multi REMOVED

**v14:**
```python
@api.multi
def action_confirm(self):
    for record in self:
        record.state = 'confirmed'
```

**v15:**
```python
def action_confirm(self):
    for record in self:
        record.state = 'confirmed'
```

### 2. track_visibility → tracking

**v14:**
```python
name = fields.Char(string='Name', track_visibility='onchange')
state = fields.Selection([...], track_visibility='always')
```

**v15:**
```python
name = fields.Char(string='Name', tracking=True)
state = fields.Selection([...], tracking=True)
```

**Note:** `tracking=True` replaces both `onchange` and `always`. Distinction no longer needed.

## Migration Script
```python
import re
def migrate_track_visibility(content):
    content = re.sub(r"track_visibility=['\"]onchange['\"]", "tracking=True", content)
    content = re.sub(r"track_visibility=['\"]always['\"]", "tracking=True", content)
    return content

def remove_api_multi(content):
    content = re.sub(r"^\s*@api\.multi\s*\n", "", content, flags=re.MULTILINE)
    return content
```

## Detailed Migration

### Model with Tracking

**v14:**
```python
class MyModel(models.Model):
    _name = 'my.model'
    _inherit = ['mail.thread', 'mail.activity.mixin']
    name = fields.Char(track_visibility='onchange')
    state = fields.Selection([('draft','Draft'),('confirmed','Confirmed')],
        default='draft', track_visibility='always')

    @api.multi
    def action_confirm(self):
        for record in self:
            record.state = 'confirmed'
```

**v15:**
```python
class MyModel(models.Model):
    _name = 'my.model'
    _inherit = ['mail.thread', 'mail.activity.mixin']
    name = fields.Char(tracking=True)
    state = fields.Selection([('draft','Draft'),('confirmed','Confirmed')],
        default='draft', tracking=True)

    def action_confirm(self):
        for record in self:
            record.state = 'confirmed'
```

### Chatter Widget

**v14:**
```xml
<div class="oe_chatter">
    <field name="message_follower_ids" widget="mail_followers"/>
    <field name="activity_ids" widget="mail_activity"/>
    <field name="message_ids" widget="mail_thread"/>
</div>
```

**v15:** widget attributes optional — Odoo auto-detects.

```xml
<div class="oe_chatter">
    <field name="message_follower_ids"/>
    <field name="activity_ids"/>
    <field name="message_ids"/>
</div>
```

## No Change Required

**Security Groups, Access Rights, Record Rules, View attrs, Field groups** — same syntax.

## Checklist

- [ ] **CRITICAL:** Remove ALL `@api.multi`
- [ ] **CRITICAL:** Replace `track_visibility` → `tracking=True`
- [ ] Update Python to 3.8+
- [ ] Update chatter widgets (optional)
- [ ] Test all methods that had `@api.multi`
- [ ] Verify tracking works on mail.thread models

## Common Mistakes

**Wrong (v15):** `@api.multi` still in code → remove decorator, keep method.
**Wrong:** `track_visibility='onchange'` → use `tracking=True`.

## Testing After Migration
```python
def test_tracking(self):
    record = self.env['my.model'].create({'name': 'Test'})
    record.write({'name': 'Updated'})
    messages = record.message_ids.filtered(lambda m: m.tracking_value_ids)
    self.assertTrue(messages, "Tracking message should be created")
```

## GitHub Reference
- `odoo/api.py` — Decorator changes
- `odoo/models.py` — Field tracking implementation
