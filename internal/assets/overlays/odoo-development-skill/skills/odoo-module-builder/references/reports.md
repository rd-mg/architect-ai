> **IMPORTANT (WORKSPACE RULE):** This reference explicitly targets older Odoo versions. Remember that for **Odoo 19.0**, you MUST use `<list>` (not `<tree>`), `<chatter/>`, `_compute_display_name`, and avoid `attrs`. For **Odoo 18.0**, use direct attributes. See `.github/copilot-instructions.md` for strict version variations, and consult Context7 documentation (https://context7.com/...) as your PREFERRED source of truth. Use local workspace assets only as a SECONDARY fallback.


# Odoo Reports Reference

## Table of Contents
- [Report Types](#report-types)
- [QWeb Report Definition](#qweb-report-definition)
- [QWeb Template](#qweb-template)
- [QWeb Directives](#qweb-directives)
- [Paper Format](#paper-format)
- [Report Styling](#report-styling)
- [Custom Report Model](#custom-report-model)

## Report Types

| Type | Output | Use Case |
|------|--------|----------|
| `qweb-pdf` | PDF via wkhtmltopdf | Invoices, receipts, formal documents |
| `qweb-html` | HTML in browser | Previews, dashboards |

## QWeb Report Definition

```xml
<!-- reports/library_book_report.xml -->
<odoo>
    <record id="action_report_library_book" model="ir.actions.report">
        <field name="name">Book Card</field>
        <field name="model">library.book</field>
        <field name="report_type">qweb-pdf</field>
        <field name="report_name">library_management.report_book_card</field>
        <field name="report_file">library_management.report_book_card</field>
        <field name="print_report_name">'Book Card - %s' % object.name</field>
        <field name="binding_model_id" ref="model_library_book"/>
        <field name="binding_type">report</field>
    </record>
</odoo>
```

- `report_name`: Template XML ID (module.template_id)
- `binding_model_id`: Makes report available from the model's Print menu
- `print_report_name`: Python expression for the PDF filename

## QWeb Template

```xml
<!-- reports/library_book_report_template.xml -->
<odoo>
    <template id="report_book_card">
        <t t-call="web.html_container">
            <t t-foreach="docs" t-as="doc">
                <t t-call="web.external_layout">
                    <div class="page">
                        <h2><t t-esc="doc.name"/></h2>
                        <div class="row mt-3">
                            <div class="col-6">
                                <strong>ISBN:</strong>
                                <span t-field="doc.isbn"/>
                            </div>
                            <div class="col-6">
                                <strong>Category:</strong>
                                <span t-field="doc.category_id.name"/>
                            </div>
                        </div>
                        <table class="table table-sm mt-4">
                            <thead>
                                <tr>
                                    <th>Author</th>
                                    <th>Email</th>
                                </tr>
                            </thead>
                            <tbody>
                                <tr t-foreach="doc.author_ids" t-as="author">
                                    <td><t t-esc="author.name"/></td>
                                    <td><t t-esc="author.email"/></td>
                                </tr>
                            </tbody>
                        </table>
                        <div class="mt-3">
                            <strong>Description:</strong>
                            <p t-field="doc.description"/>
                        </div>
                    </div>
                </t>
            </t>
        </t>
    </template>
</odoo>
```

### Key Template Elements

- `web.html_container`: Wraps multiple documents
- `web.external_layout`: Adds company header/footer
- `docs`: Recordset passed automatically (the selected records)
- `doc`: Current record in the loop

## QWeb Directives

| Directive | Purpose | Example |
|-----------|---------|---------|
| `t-esc` | Output escaped text | `<t t-esc="doc.name"/>` |
| `t-raw` | Output raw HTML (unsafe) | `<t t-raw="doc.description"/>` |
| `t-field` | Formatted field output | `<span t-field="doc.date" t-options='{"widget": "date"}'/>`|
| `t-foreach` | Loop | `<t t-foreach="doc.line_ids" t-as="line">` |
| `t-if` | Conditional | `<t t-if="doc.state == 'available'">` |
| `t-elif` / `t-else` | Else branches | `<t t-elif="doc.state == 'lost'">` |
| `t-set` | Variable assignment | `<t t-set="total" t-value="sum(...)"/>` |
| `t-att-*` | Dynamic attribute | `<div t-att-class="'text-danger' if x else ''"/>` |
| `t-call` | Include sub-template | `<t t-call="module.sub_template"/>` |

### `t-field` Format Options

```xml
<!-- Date formatting -->
<span t-field="doc.date_published" t-options='{"widget": "date", "format": "dd/MM/yyyy"}'/>

<!-- Monetary -->
<span t-field="doc.price" t-options='{"widget": "monetary", "display_currency": doc.currency_id}'/>

<!-- Duration -->
<span t-field="doc.duration" t-options='{"widget": "duration", "unit": "hour"}'/>
```

## Paper Format

```xml
<record id="paperformat_book_card" model="report.paperformat">
    <field name="name">Book Card Format</field>
    <field name="default" eval="False"/>
    <field name="format">A4</field>
    <field name="orientation">Portrait</field>
    <field name="margin_top">40</field>
    <field name="margin_bottom">20</field>
    <field name="margin_left">7</field>
    <field name="margin_right">7</field>
    <field name="header_line" eval="False"/>
    <field name="header_spacing">35</field>
    <field name="dpi">90</field>
</record>

<!-- Link paper format to report -->
<record id="action_report_library_book" model="ir.actions.report">
    <field name="paperformat_id" ref="paperformat_book_card"/>
</record>
```

## Report Styling

Add CSS within the template:

```xml
<template id="report_book_card">
    <t t-call="web.html_container">
        <style>
            .book-header { font-size: 1.5em; color: #333; }
            .book-table th { background-color: #f5f5f5; }
        </style>
        <!-- report content -->
    </t>
</template>
```

## Custom Report Model

For reports needing computed data not on the source model:

```python
class BookReportMixin(models.AbstractModel):
    _name = 'report.library_management.report_book_card'
    _description = 'Book Card Report'

    @api.model
    def _get_report_values(self, docids, data=None):
        docs = self.env['library.book'].browse(docids)
        return {
            'doc_ids': docids,
            'doc_model': 'library.book',
            'docs': docs,
            'data': data,
            'total_books': len(docs),
        }
```

The model name must be `report.<report_name>` matching the `report_name` field in the report action.
