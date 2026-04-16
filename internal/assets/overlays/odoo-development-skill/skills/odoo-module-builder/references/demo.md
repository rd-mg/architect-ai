> **IMPORTANT (WORKSPACE RULE):** This reference explicitly targets older Odoo versions. Remember that for **Odoo 19.0**, you MUST use `<list>` (not `<tree>`), `<chatter/>`, `_compute_display_name`, and avoid `attrs`. For **Odoo 18.0**, use direct attributes. See `.github/copilot-instructions.md` for strict version variations, and consult Context7 documentation (https://context7.com/...) as your PREFERRED source of truth. Use local workspace assets only as a SECONDARY fallback.


# Odoo Demo Data Reference

## Table of Contents
- [Overview](#overview)
- [Directory Structure](#directory-structure)
- [Demo Data Files](#demo-data-files)
- [Relational Records](#relational-records)
- [Demo vs Regular Data](#demo-vs-regular-data)
- [Best Practices](#best-practices)

## Overview

Demo data provides sample records loaded only when the database is created with "Load demonstration data" enabled. It helps users explore module features and is used in automated tests.

## Directory Structure

Two common patterns:

```
# Pattern 1: Separate demo/ folder
demo/
├── demo_categories.xml
├── demo_books.xml
└── demo_loans.xml

# Pattern 2: Inside data/ folder (also valid)
data/
├── demo_data.xml
```

Both work — just reference files correctly in `__manifest__.py`.

## Demo Data Files

Register in the `demo` key (not `data`):

```python
# __manifest__.py
{
    'demo': [
        'demo/demo_categories.xml',
        'demo/demo_books.xml',
        'demo/demo_loans.xml',
    ],
}
```

### XML Demo Records

```xml
<!-- demo/demo_categories.xml -->
<odoo>
    <record id="demo_category_fiction" model="library.book.category">
        <field name="name">Fiction</field>
    </record>
    <record id="demo_category_science" model="library.book.category">
        <field name="name">Science</field>
    </record>
    <record id="demo_category_history" model="library.book.category">
        <field name="name">History</field>
    </record>
</odoo>
```

```xml
<!-- demo/demo_books.xml -->
<odoo>
    <record id="demo_book_1" model="library.book">
        <field name="name">The Great Gatsby</field>
        <field name="isbn">978-0-7432-7356-5</field>
        <field name="category_id" ref="demo_category_fiction"/>
        <field name="state">available</field>
        <field name="date_published">1925-04-10</field>
        <field name="price">12.99</field>
        <field name="author_ids" eval="[(4, ref('base.res_partner_2'))]"/>
    </record>

    <record id="demo_book_2" model="library.book">
        <field name="name">A Brief History of Time</field>
        <field name="isbn">978-0-553-10953-5</field>
        <field name="category_id" ref="demo_category_science"/>
        <field name="state">available</field>
        <field name="price">15.50</field>
    </record>
</odoo>
```

### CSV Demo Data

```csv
id,name,isbn,category_id:id,state,price
demo_book_csv_1,Don Quixote,978-0-060-93434-7,demo_category_fiction,available,9.99
demo_book_csv_2,The Origin of Species,978-0-451-52965-9,demo_category_science,available,11.50
```

## Relational Records

### Many2one

```xml
<field name="category_id" ref="demo_category_fiction"/>
```

### Many2many

```xml
<!-- Link to existing records -->
<field name="tag_ids" eval="[(4, ref('demo_tag_classic')), (4, ref('demo_tag_bestseller'))]"/>

<!-- Replace all links -->
<field name="tag_ids" eval="[(6, 0, [ref('demo_tag_classic'), ref('demo_tag_bestseller')])]"/>
```

### One2many (inline creation)

```xml
<record id="demo_order_1" model="library.order">
    <field name="partner_id" ref="base.res_partner_1"/>
    <field name="line_ids" eval="[
        (0, 0, {'book_id': ref('demo_book_1'), 'quantity': 1, 'price': 12.99}),
        (0, 0, {'book_id': ref('demo_book_2'), 'quantity': 2, 'price': 15.50}),
    ]"/>
</record>
```

### Referencing Base Module Demo Data

Odoo's `base` module provides demo partners, users, and companies:

```xml
<field name="partner_id" ref="base.res_partner_1"/>
<field name="user_id" ref="base.user_demo"/>
<field name="company_id" ref="base.main_company"/>
```

## Demo vs Regular Data

| Aspect | `data` | `demo` |
|--------|--------|--------|
| Manifest key | `'data': [...]` | `'demo': [...]` |
| When loaded | Always (install + upgrade) | Only with demo flag |
| Purpose | Required config, security, views | Sample records for exploration |
| In production | Yes | No (never enable demo in prod) |
| Used in tests | Yes | Yes (tests run with demo by default) |

## Best Practices

- Keep demo data realistic — use plausible names, dates, and amounts
- Reference existing base demo records (partners, users) instead of creating duplicates
- Load order matters: create parent records before children (categories before books)
- Use `noupdate="0"` or omit it — demo data should be refreshable
- Include enough records to demonstrate all states and features (draft, confirmed, done)
- Add demo data for all models the user will interact with
- Test that demo data loads without errors: `./odoo-bin -d testdb --test-enable -i module_name`
