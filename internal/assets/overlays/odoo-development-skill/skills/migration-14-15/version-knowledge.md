# Odoo Version Knowledge: 14 to 15 Migration

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  VERSION MIGRATION: 14.0 → 15.0                                              ║
║  Critical changes, breaking changes, and migration patterns                  ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## Breaking Changes Summary

| Category | Change | Impact |
|----------|--------|--------|
| API | `@api.multi` REMOVED | High - Update all methods |
| Fields | `track_visibility` deprecated | Medium - Use `tracking` |
| OWL | OWL 1.x introduced | High - New frontend |
| Assets | `assets_backend` pattern change | Medium - Update manifests |
| Python | 3.8+ required | Low - Compatibility |

## Critical: @api.multi Removal

**Most significant breaking change in v15**

**v14:**
```python
@api.multi
def action_confirm(self):
    for order in self:
        order.state = 'sale'

@api.multi
def action_cancel(self):
    return self.write({'state': 'cancel'})
```

**v15:**
```python
def action_confirm(self):
    for order in self:
        order.state = 'sale'

def action_cancel(self):
    return self.write({'state': 'cancel'})
```

**Migration:**
```python
import re
def migrate_api_multi(content):
    return re.sub(r'\n\s*@api\.multi\n', '\n', content)
```

## Field Tracking Migration

**v14:** `track_visibility='onchange'` / `track_visibility='always'`
**v15:** `tracking=True`

## OWL Introduction (v15)

**Legacy (v14):**
```javascript
odoo.define('my_module.MyWidget', function (require) {
    var Widget = require('web.Widget');
    var MyWidget = Widget.extend({ template: 'my_module.MyTemplate' });
    return MyWidget;
});
```

**OWL 1.x (v15):**
```javascript
/** @odoo-module **/
const { Component } = owl;
class MyComponent extends Component {
    static template = xml`<div>...</div>`;
}
```

## Asset Bundle Changes

**v14:** `'qweb': ['static/src/xml/my_templates.xml']`
**v15:** `'assets': {'web.assets_backend': [...], 'web.assets_qweb': [...]}`

## Removed/Deprecated Features

| Feature | Status | Replacement |
|---------|--------|-------------|
| `@api.multi` | REMOVED | No decorator |
| `@api.one` | REMOVED | Loop in method |
| `track_visibility` | Deprecated | `tracking=True` |
| `qweb` in manifest | Deprecated | Use `assets` |
| `website_published` | Deprecated | `is_published` |

## GitHub Verification

```
https://raw.githubusercontent.com/odoo/odoo/14.0/odoo/api.py
https://raw.githubusercontent.com/odoo/odoo/15.0/odoo/api.py
https://raw.githubusercontent.com/odoo/odoo/14.0/addons/sale/models/sale_order.py
https://raw.githubusercontent.com/odoo/odoo/15.0/addons/sale/models/sale_order.py
```

## Checklist

- [ ] Remove all `@api.multi`
- [ ] Remove all `@api.one` (rewrite logic)
- [ ] Replace `track_visibility` → `tracking`
- [ ] Update manifest `qweb` → `assets`
- [ ] Test button actions without decorators
- [ ] Verify field tracking works
- [ ] Update Python (3.8+)
- [ ] Consider OWL for new frontend

## Common Errors

**`AttributeError: module 'odoo.api' has no attribute 'multi'`** → Remove `@api.multi`
**`Unknown field 'track_visibility'`** → Replace with `tracking=True`
**`qweb key not supported in manifest`** → Move to `assets.web.assets_qweb`
