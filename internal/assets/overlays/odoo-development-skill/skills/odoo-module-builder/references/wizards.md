> **IMPORTANT (WORKSPACE RULE):** This reference explicitly targets older Odoo versions. Remember that for **Odoo 19.0**, you MUST use `<list>` (not `<tree>`), `<chatter/>`, `_compute_display_name`, and avoid `attrs`. For **Odoo 18.0**, use direct attributes. See `.github/copilot-instructions.md` for strict version variations, and consult Context7 documentation (https://context7.com/...) as your PREFERRED source of truth. Use local workspace assets only as a SECONDARY fallback.


# Odoo Wizards Reference

## Table of Contents
- [Wizard Structure](#wizard-structure)
- [Wizard Model](#wizard-model)
- [Wizard View](#wizard-view)
- [Wizard Action](#wizard-action)
- [Prepopulating Wizard Data](#prepopulating-wizard-data)
- [Multi-Record Wizards](#multi-record-wizards)

## Wizard Structure

Wizards use `TransientModel` — records are temporary and auto-purged. They open as modal dialogs.

Files to create:
```
wizards/
├── __init__.py
├── book_return_wizard.py
└── book_return_wizard_views.xml
```

Register in module `__init__.py`:
```python
from . import wizards
```

And `wizards/__init__.py`:
```python
from . import book_return_wizard
```

Register views in `__manifest__.py`:
```python
'data': [
    'wizards/book_return_wizard_views.xml',
],
```

## Wizard Model

```python
from odoo import models, fields, api

class BookReturnWizard(models.TransientModel):
    _name = 'library.book.return.wizard'
    _description = 'Return Book Wizard'

    book_id = fields.Many2one('library.book', string='Book', required=True)
    return_date = fields.Date(string='Return Date', default=fields.Date.today)
    condition = fields.Selection([
        ('good', 'Good'),
        ('damaged', 'Damaged'),
        ('lost', 'Lost'),
    ], string='Condition', default='good', required=True)
    notes = fields.Text(string='Notes')

    def action_confirm(self):
        """Process the book return."""
        self.ensure_one()
        self.book_id.write({
            'state': 'available' if self.condition == 'good' else 'lost',
        })
        return {'type': 'ir.actions.act_window_close'}
```

## Wizard View

```xml
<record id="view_book_return_wizard_form" model="ir.ui.view">
    <field name="name">library.book.return.wizard.form</field>
    <field name="model">library.book.return.wizard</field>
    <field name="arch" type="xml">
        <form string="Return Book">
            <group>
                <field name="book_id"/>
                <field name="return_date"/>
                <field name="condition"/>
                <field name="notes"/>
            </group>
            <footer>
                <button name="action_confirm" string="Confirm Return"
                        type="object" class="btn-primary"/>
                <button string="Cancel" class="btn-secondary" special="cancel"/>
            </footer>
        </form>
    </field>
</record>
```

## Wizard Action

```xml
<record id="action_book_return_wizard" model="ir.actions.act_window">
    <field name="name">Return Book</field>
    <field name="res_model">library.book.return.wizard</field>
    <field name="view_mode">form</field>
    <field name="target">new</field>  <!-- opens as modal dialog -->
</record>
```

### Launching from a Button in a Form View

```xml
<button name="%(action_book_return_wizard)d" string="Return"
        type="action" class="oe_highlight"/>
```

### Launching from Python (with context)

```python
def action_open_return_wizard(self):
    return {
        'name': 'Return Book',
        'type': 'ir.actions.act_window',
        'res_model': 'library.book.return.wizard',
        'view_mode': 'form',
        'target': 'new',
        'context': {
            'default_book_id': self.id,
        },
    }
```

## Prepopulating Wizard Data

Use `context` keys prefixed with `default_` to prepopulate fields:

```python
context = {
    'default_book_id': self.id,
    'default_return_date': fields.Date.today(),
}
```

Or use `@api.model` with `default_get`:

```python
@api.model
def default_get(self, fields_list):
    res = super().default_get(fields_list)
    active_id = self.env.context.get('active_id')
    if active_id:
        book = self.env['library.book'].browse(active_id)
        res['book_id'] = book.id
    return res
```

## Multi-Record Wizards

For wizards that operate on multiple selected records:

```python
@api.model
def default_get(self, fields_list):
    res = super().default_get(fields_list)
    active_ids = self.env.context.get('active_ids', [])
    res['book_ids'] = [(6, 0, active_ids)]
    return res

def action_confirm(self):
    for book in self.book_ids:
        book.write({'state': 'available'})
    return {'type': 'ir.actions.act_window_close'}
```

### Server Action to Launch Wizard on Selected Records

```xml
<record id="action_server_return_books" model="ir.actions.server">
    <field name="name">Return Selected Books</field>
    <field name="model_id" ref="model_library_book"/>
    <field name="binding_model_id" ref="model_library_book"/>
    <field name="binding_view_types">list</field>
    <field name="state">code</field>
    <field name="code">action = model.action_open_return_wizard()</field>
</record>
```
