# Quick Odoo Patterns (Cheat Sheet)

Consolidated 80/20 reference for common Odoo tasks. Use for rapid prototyping.

> **Version-specific syntax** → `patterns-{version}/model-patterns.md`
> v17+: `invisible=` · v18+: `<tree>` → `<list>` · v19+: `SQL()`

---

## Model & Basic Fields

```python
class MyModel(models.Model):
    _name = 'my.module.model'
    _inherit = ['mail.thread']

    name = fields.Char(required=True)
    state = fields.Selection([('draft', 'Draft'), ('done', 'Done')], default='draft')
    partner_id = fields.Many2one('res.partner')
    line_ids = fields.One2many('my.line', 'parent_id')
```

## UI Elements (Views)

```xml
<!-- Form -->
<form>
    <header>
        <button name="action_do" type="object" invisible="state != 'draft'"/>
        <field name="state" widget="statusbar"/>
    </header>
    <sheet>
        <group><field name="name"/><field name="partner_id"/></group>
        <notebook><page string="Lines"><field name="line_ids"/></page></notebook>
    </sheet>
</form>

<!-- List (v18 renamed tree to list) -->
<list>
    <field name="name"/>
    <field name="state" widget="badge"/>
</list>
```

---

## Anti-Patterns

```python
# ❌ NEVER use attrs={'invisible': ...} in v17+. Use invisible="..." directly.

# ❌ NEVER use cr.execute() for simple queries. Use self.env['model'].search().

# ❌ NEVER forget to add 'mail' to depends if using mail.thread.
```

---

## Version Matrix (Quick Look)

| Syntax | v14-v16 | v17 | v18 | v19 |
|--------|---------|-----|-----|-----|
| Visibility | `attrs` | `invisible` | `invisible` | `invisible` |
| List View | `<tree>` | `<tree>` | `<list>` | `<list>` |
| Querying | `.search()` | `.search()` | `.search()` | `SQL()` |
| Logic | `api.multi` | `api` | `api` | `api` |
