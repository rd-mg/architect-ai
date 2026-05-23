# Odoo Module Migration: 14.0 → 15.0

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  MODULE MIGRATION: 14.0 → 15.0                                                ║
║  @api.multi removed, tracking=True standardized, OWL 1.x introduced          ║
║  VERIFY: https://github.com/odoo/odoo/tree/15.0                              ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## Overview

| Aspect | v14 | v15 |
|--------|-----|-----|
| @api.multi | Deprecated | REMOVED |
| track_visibility | Deprecated | Deprecated (warns) |
| tracking=True | Supported | Standard |
| OWL | N/A | OWL 1.x |
| Python | 3.6-3.8 | 3.7-3.9 |

## Breaking Changes

### 1. @api.multi REMOVED
```python
# v14:
from odoo import models, api
@api.multi
def action_confirm(self):
    for record in self:
        record.state = 'confirmed'

# v15:
from odoo import models
def action_confirm(self):
    for record in self:
        record.state = 'confirmed'
```

### 2. track_visibility → tracking
```python
# v14:
state = fields.Selection([('draft','Draft'),('done','Done')], track_visibility='onchange')

# v15:
state = fields.Selection([('draft','Draft'),('done','Done')], tracking=True)
```

## Manifest Changes

```python
# v14
{'version': '14.0.1.0.0', 'depends': ['base', 'mail'],
 'data': ['security/ir.model.access.csv', 'views/my_model_views.xml']}

# v15 — Add assets bundle
{'version': '15.0.1.0.0', 'depends': ['base', 'mail', 'web'],
 'data': ['security/ir.model.access.csv', 'views/my_model_views.xml'],
 'assets': {'web.assets_backend': ['my_module/static/src/js/**/*',
             'my_module/static/src/xml/**/*', 'my_module/static/src/scss/**/*']}}
```

## Model Migration

```python
# v14 Model:
class MyModel(models.Model):
    _name = 'my.model'
    _inherit = ['mail.thread']
    name = fields.Char(required=True, track_visibility='always')
    state = fields.Selection([('draft','Draft'),('confirmed','Confirmed'),('done','Done')],
        default='draft', track_visibility='onchange')

    @api.multi
    def action_confirm(self):
        for record in self:
            if not record.line_ids:
                raise UserError(_("Add at least one line."))
            record.state = 'confirmed'

    @api.model
    def create(self, vals):
        if not vals.get('code'):
            vals['code'] = self.env['ir.sequence'].next_by_code('my.model')
        return super(MyModel, self).create(vals)

# v15 Model:
class MyModel(models.Model):
    _name = 'my.model'
    _inherit = ['mail.thread', 'mail.activity.mixin']
    name = fields.Char(required=True, tracking=True)
    state = fields.Selection([('draft','Draft'),('confirmed','Confirmed'),('done','Done')],
        default='draft', tracking=True)

    def action_confirm(self):
        for record in self:
            if not record.line_ids:
                raise UserError(_("Add at least one line."))
            record.state = 'confirmed'

    @api.model
    def create(self, vals):
        if not vals.get('code'):
            vals['code'] = self.env['ir.sequence'].next_by_code('my.model')
        return super().create(vals)
```

## View Changes

Views remain compatible. No mandatory changes.

```xml
<form>
    <header><button name="action_confirm" type="object" string="Confirm"/></header>
    <sheet><group><field name="name"/></group></sheet>
    <div class="oe_chatter">
        <field name="message_follower_ids"/>
        <field name="activity_ids"/>
        <field name="message_ids"/>
    </div>
</form>
```

## Frontend Migration

### Legacy JS (v14) → OWL 1.x (v15)

```javascript
// v14 Legacy:
odoo.define('my_module.MyWidget', function (require) {
    var Widget = require('web.Widget');
    var MyWidget = Widget.extend({
        template: 'my_module.MyWidget',
        start: function () { return this._super.apply(this, arguments); },
    });
    core.action_registry.add('my_module.my_action', MyWidget);
    return MyWidget;
});

// v15 OWL 1.x:
/** @odoo-module **/
const { Component } = owl;
const { useService } = require('@web/core/utils/hooks');

class MyComponent extends Component {
    setup() {
        this.orm = useService('orm');
        onMounted(async () => { await this.loadData(); });
    }
}
MyComponent.template = 'my_module.MyComponent';
registry.category('actions').add('my_module.my_action', MyComponent);
```

### QWeb Template
```xml
<!-- v14 Legacy -->
<t t-name="my_module.MyWidget"><div class="my-widget">...</div></t>

<!-- v15 OWL -->
<t t-name="my_module.MyComponent" owl="1">
    <div class="my-component">
        <t t-if="state.loading"><div>Loading...</div></t>
        <t t-else="">
            <t t-foreach="state.data" t-as="item" t-key="item.id">
                <div t-esc="item.name"/>
            </t>
        </t>
    </div>
</t>
```

## Checklist

### Code
- [ ] Remove all `@api.multi`
- [ ] Replace `track_visibility='onchange'` → `tracking=True`
- [ ] Replace `track_visibility='always'` → `tracking=True`
- [ ] Update `super()` to Python 3 style
- [ ] Python 3.7+

### Manifest
- [ ] Version to `15.0.x.x.x`
- [ ] Add `assets` bundle for JS/CSS/XML
- [ ] Add `'web'` to depends if using OWL

### Optional
- [ ] Consider `@api.model_create_multi`
- [ ] Migrate legacy JS to OWL
- [ ] Add `mail.activity.mixin`

## Search Patterns
```bash
grep -r "@api.multi" --include="*.py"
grep -r "track_visibility" --include="*.py"
grep -r "super(.*self)" --include="*.py"
```

## Common Issues

**`AttributeError: 'api' object has no attribute 'multi'`** → Remove `@api.multi`
**`DeprecationWarning: track_visibility deprecated`** → Replace with `tracking=True`
**`ImportError: cannot import name 'Component' from 'owl'`** → Use `const { Component } = owl;`
