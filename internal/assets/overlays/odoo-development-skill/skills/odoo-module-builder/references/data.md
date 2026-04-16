> **IMPORTANT (WORKSPACE RULE):** This reference explicitly targets older Odoo versions. Remember that for **Odoo 19.0**, you MUST use `<list>` (not `<tree>`), `<chatter/>`, `_compute_display_name`, and avoid `attrs`. For **Odoo 18.0**, use direct attributes. See `.github/copilot-instructions.md` for strict version variations, and consult Context7 documentation (https://context7.com/...) as your PREFERRED source of truth. Use local workspace assets only as a SECONDARY fallback.


# Odoo Data Files Reference

## Table of Contents
- [Data File Types](#data-file-types)
- [Sequences](#sequences)
- [Cron Jobs](#cron-jobs)
- [Mail Templates](#mail-templates)
- [Server Actions](#server-actions)
- [Demo Data](#demo-data)
- [Seed / Default Data](#seed--default-data)

## Data File Types

Files in `data/` are loaded during module install/upgrade. List them in `__manifest__.py` `data` key.

```
data/
├── sequence_data.xml
├── cron_data.xml
├── mail_template_data.xml
├── server_action_data.xml
└── demo_data.xml          # or list in 'demo' key instead
```

## Sequences

Auto-incrementing reference numbers (e.g., `LIB/00001`):

```xml
<!-- data/sequence_data.xml -->
<odoo noupdate="1">
    <record id="seq_library_book" model="ir.sequence">
        <field name="name">Library Book Sequence</field>
        <field name="code">library.book</field>
        <field name="prefix">LIB/</field>
        <field name="padding">5</field>
        <field name="number_increment">1</field>
        <field name="number_next">1</field>
    </record>
</odoo>
```

Usage in Python:

```python
@api.model_create_multi
def create(self, vals_list):
    for vals in vals_list:
        if vals.get('reference', 'New') == 'New':
            vals['reference'] = self.env['ir.sequence'].next_by_code('library.book') or 'New'
    return super().create(vals_list)
```

### `noupdate="1"`

Data inside `<odoo noupdate="1">` is only loaded on initial install, not on module upgrade. Use for:
- Sequences (users may customize prefix/padding)
- Default records users might edit
- Mail templates

Omit `noupdate` (or set `"0"`) for data that should always be refreshed on upgrade.

## Cron Jobs

Scheduled actions running periodically:

```xml
<!-- data/cron_data.xml -->
<odoo noupdate="1">
    <record id="cron_check_overdue_books" model="ir.cron">
        <field name="name">Library: Check Overdue Books</field>
        <field name="model_id" ref="model_library_book"/>
        <field name="state">code</field>
        <field name="code">model._cron_check_overdue()</field>
        <field name="interval_number">1</field>
        <field name="interval_type">days</field>
        <field name="numbercall">-1</field>  <!-- -1 = unlimited -->
        <field name="active" eval="True"/>
    </record>
</odoo>
```

Corresponding model method:

```python
def _cron_check_overdue(self):
    overdue = self.search([
        ('state', '=', 'borrowed'),
        ('return_date', '<', fields.Date.today()),
    ])
    for book in overdue:
        book.message_post(body="This book is overdue!")
```

### Interval Types

`minutes`, `hours`, `days`, `weeks`, `months`

## Mail Templates

Email templates with QWeb:

```xml
<!-- data/mail_template_data.xml -->
<odoo noupdate="1">
    <record id="mail_template_book_borrowed" model="mail.template">
        <field name="name">Library: Book Borrowed</field>
        <field name="model_id" ref="model_library_book"/>
        <field name="subject">Book Borrowed: {{ object.name }}</field>
        <field name="email_from">{{ (object.company_id.email or user.email) }}</field>
        <field name="email_to">{{ object.borrower_id.email }}</field>
        <field name="body_html" type="html">
            <div style="margin: 0; padding: 0;">
                <p>Dear <t t-out="object.borrower_id.name"/>,</p>
                <p>You have borrowed: <strong><t t-out="object.name"/></strong></p>
                <p>Please return by: <t t-out="object.return_date"/></p>
            </div>
        </field>
        <field name="auto_delete" eval="True"/>
    </record>
</odoo>
```

Sending from Python:

```python
def action_borrow(self):
    template = self.env.ref('library_management.mail_template_book_borrowed')
    template.send_mail(self.id, force_send=True)
```

## Server Actions

Custom actions triggered from UI or automation:

```xml
<!-- data/server_action_data.xml -->
<odoo>
    <record id="action_mark_books_available" model="ir.actions.server">
        <field name="name">Mark as Available</field>
        <field name="model_id" ref="model_library_book"/>
        <field name="binding_model_id" ref="model_library_book"/>
        <field name="binding_view_types">list,form</field>
        <field name="state">code</field>
        <field name="code">
            for record in records:
                record.write({'state': 'available'})
        </field>
    </record>
</odoo>
```

`binding_model_id` makes the action appear in the Action dropdown menu for that model.

## Demo Data

Demo/sample records loaded only when "Load demo data" is enabled:

```xml
<!-- data/demo_data.xml -->
<odoo>
    <record id="demo_book_1" model="library.book">
        <field name="name">The Odoo Book</field>
        <field name="isbn">978-0-123456-78-9</field>
        <field name="state">available</field>
        <field name="category_id" ref="demo_category_programming"/>
    </record>

    <record id="demo_book_2" model="library.book">
        <field name="name">Python Cookbook</field>
        <field name="isbn">978-0-987654-32-1</field>
        <field name="state">available</field>
    </record>
</odoo>
```

List in `__manifest__.py`:
```python
'demo': [
    'data/demo_data.xml',
],
```

## Seed / Default Data

Non-demo data loaded on every install (e.g., default categories):

```xml
<!-- data/default_data.xml -->
<odoo noupdate="1">
    <record id="category_fiction" model="library.book.category">
        <field name="name">Fiction</field>
    </record>
    <record id="category_nonfiction" model="library.book.category">
        <field name="name">Non-Fiction</field>
    </record>
</odoo>
```

List in `__manifest__.py` `data` key (not `demo`).
