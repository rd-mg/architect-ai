# Accounting Patterns

Consolidated from the following source files:
- `accounting-patterns.md`
- `tax-fiscal-patterns.md`

---


## Source: accounting-patterns.md

# Accounting Integration Patterns

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  ACCOUNTING INTEGRATION PATTERNS                                             ║
║  Journal entries, invoicing, and financial operations                        ║
║  Use for ERP integrations, financial reporting, and accounting automation    ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## Module Setup

### Manifest Dependencies
```python
{
    'name': 'My Accounting Module',
    'version': '18.0.1.0.0',
    'depends': ['account'],
    'data': [
        'security/ir.model.access.csv',
        'data/account_data.xml',
        'views/account_views.xml',
    ],
}
```

---

## Journal Entries

### Create Journal Entry
```python
from odoo import api, fields, models
from odoo.exceptions import UserError


class AccountingMixin(models.AbstractModel):
    _name = 'accounting.mixin'
    _description = 'Accounting Mixin'

    def _create_journal_entry(self, lines, journal=None, ref=None, date=None):
        """Create a journal entry with multiple lines.

        Args:
            lines: List of dicts with account_id, debit, credit, partner_id
            journal: account.journal record (optional)
            ref: Reference string
            date: Entry date (defaults to today)

        Returns:
            account.move record
        """
        if not journal:
            journal = self.env['account.journal'].search([
                ('type', '=', 'general'),
                ('company_id', '=', self.env.company.id),
            ], limit=1)

        move_vals = {
            'journal_id': journal.id,
            'date': date or fields.Date.today(),
            'ref': ref or self.name,
            'line_ids': [(0, 0, {
                'account_id': line['account_id'],
                'partner_id': line.get('partner_id'),
                'name': line.get('name', ref or '/'),
                'debit': line.get('debit', 0.0),
                'credit': line.get('credit', 0.0),
            }) for line in lines],
        }

        move = self.env['account.move'].create(move_vals)
        return move

    def _post_journal_entry(self, lines, **kwargs):
        """Create and post journal entry."""
        move = self._create_journal_entry(lines, **kwargs)
        move.action_post()
        return move
```

### Balanced Entry Example
```python
def _create_expense_entry(self, amount, expense_account, description):
    """Create expense journal entry."""
    bank_account = self.env['account.account'].search([
        ('account_type', '=', 'asset_cash'),
        ('company_id', '=', self.env.company.id),
    ], limit=1)

    lines = [
        {
            'account_id': expense_account.id,
            'name': description,
            'debit': amount,
            'credit': 0.0,
        },
        {
            'account_id': bank_account.id,
            'name': description,
            'debit': 0.0,
            'credit': amount,
        },
    ]

    return self._post_journal_entry(lines, ref=description)
```

---

## Invoice Creation

### Customer Invoice
```python
def _create_customer_invoice(self, partner, lines, date=None):
    """Create customer invoice.

    Args:
        partner: res.partner record
        lines: List of dicts with product_id, quantity, price_unit
        date: Invoice date

    Returns:
        account.move record (invoice)
    """
    invoice_vals = {
        'move_type': 'out_invoice',
        'partner_id': partner.id,
        'invoice_date': date or fields.Date.today(),
        'invoice_line_ids': [(0, 0, {
            'product_id': line.get('product_id'),
            'name': line.get('name', line.get('product_id') and
                           self.env['product.product'].browse(line['product_id']).name),
            'quantity': line.get('quantity', 1),
            'price_unit': line['price_unit'],
            'tax_ids': line.get('tax_ids', [(6, 0, [])]),
        }) for line in lines],
    }

    invoice = self.env['account.move'].create(invoice_vals)
    return invoice


def _create_and_post_invoice(self, partner, lines, **kwargs):
    """Create and post customer invoice."""
    invoice = self._create_customer_invoice(partner, lines, **kwargs)
    invoice.action_post()
    return invoice
```

### Vendor Bill
```python
def _create_vendor_bill(self, partner, lines, date=None, ref=None):
    """Create vendor bill.

    Args:
        partner: res.partner (vendor)
        lines: List of dicts with product_id, quantity, price_unit
        date: Bill date
        ref: Vendor reference

    Returns:
        account.move record (bill)
    """
    bill_vals = {
        'move_type': 'in_invoice',
        'partner_id': partner.id,
        'invoice_date': date or fields.Date.today(),
        'ref': ref,
        'invoice_line_ids': [(0, 0, {
            'product_id': line.get('product_id'),
            'name': line.get('name', ''),
            'quantity': line.get('quantity', 1),
            'price_unit': line['price_unit'],
        }) for line in lines],
    }

    bill = self.env['account.move'].create(bill_vals)
    return bill
```

### Credit Note
```python
def _create_credit_note(self, invoice, reason=None):
    """Create credit note for an invoice.

    Args:
        invoice: Original account.move record
        reason: Reason for credit

    Returns:
        account.move record (credit note)
    """
    # Use the reversal wizard approach
    reversal_wizard = self.env['account.move.reversal'].with_context(
        active_model='account.move',
        active_ids=invoice.ids,
    ).create({
        'reason': reason or 'Credit Note',
        'refund_method': 'refund',  # 'refund', 'cancel', 'modify'
        'journal_id': invoice.journal_id.id,
    })

    result = reversal_wizard.reverse_moves()
    credit_note = self.env['account.move'].browse(result['res_id'])

    return credit_note
```

---

## Payment Processing

### Register Payment
```python
def _register_payment(self, invoice, amount=None, date=None, journal=None):
    """Register payment for an invoice.

    Args:
        invoice: account.move record
        amount: Payment amount (defaults to invoice amount)
        date: Payment date
        journal: Payment journal

    Returns:
        account.payment record
    """
    if not journal:
        journal = self.env['account.journal'].search([
            ('type', 'in', ['bank', 'cash']),
            ('company_id', '=', self.env.company.id),
        ], limit=1)

    payment_vals = {
        'payment_type': 'inbound' if invoice.move_type == 'out_invoice' else 'outbound',
        'partner_type': 'customer' if invoice.move_type in ['out_invoice', 'out_refund'] else 'supplier',
        'partner_id': invoice.partner_id.id,
        'amount': amount or invoice.amount_residual,
        'date': date or fields.Date.today(),
        'journal_id': journal.id,
        'ref': invoice.name,
    }

    payment = self.env['account.payment'].create(payment_vals)
    payment.action_post()

    # Reconcile with invoice
    lines_to_reconcile = (payment.move_id.line_ids + invoice.line_ids).filtered(
        lambda l: l.account_id.reconcile and not l.reconciled
    )
    lines_to_reconcile.reconcile()

    return payment
```

### Bulk Payment
```python
def _create_batch_payment(self, invoices, journal=None):
    """Create batch payment for multiple invoices.

    Args:
        invoices: account.move recordset

    Returns:
        account.payment record
    """
    if not invoices:
        raise UserError("No invoices to pay")

    # Group by partner
    partner = invoices[0].partner_id
    if any(inv.partner_id != partner for inv in invoices):
        raise UserError("All invoices must be for the same partner")

    total_amount = sum(invoices.mapped('amount_residual'))

    payment = self._register_payment(
        invoices[0],
        amount=total_amount,
        journal=journal,
    )

    # Reconcile all invoices
    for invoice in invoices[1:]:
        lines_to_reconcile = (payment.move_id.line_ids + invoice.line_ids).filtered(
            lambda l: l.account_id.reconcile and not l.reconciled
        )
        lines_to_reconcile.reconcile()

    return payment
```

---

## Account Queries

### Get Account by Type
```python
def _get_account(self, account_type, company=None):
    """Get account by type.

    Args:
        account_type: e.g., 'asset_receivable', 'liability_payable',
                     'expense', 'income', 'asset_cash'
    """
    company = company or self.env.company
    return self.env['account.account'].search([
        ('account_type', '=', account_type),
        ('company_id', '=', company.id),
    ], limit=1)


def _get_receivable_account(self):
    return self._get_account('asset_receivable')


def _get_payable_account(self):
    return self._get_account('liability_payable')


def _get_expense_account(self, product=None):
    if product and product.property_account_expense_id:
        return product.property_account_expense_id
    return self._get_account('expense')


def _get_income_account(self, product=None):
    if product and product.property_account_income_id:
        return product.property_account_income_id
    return self._get_account('income')
```

### Get Journal by Type
```python
def _get_journal(self, journal_type, company=None):
    """Get journal by type.

    Args:
        journal_type: 'sale', 'purchase', 'cash', 'bank', 'general'
    """
    company = company or self.env.company
    return self.env['account.journal'].search([
        ('type', '=', journal_type),
        ('company_id', '=', company.id),
    ], limit=1)
```

---

## Financial Reports

### Partner Balance
```python
def _get_partner_balance(self, partner, account_type='asset_receivable'):
    """Get partner balance for specific account type."""
    account = self._get_account(account_type)

    self.env.cr.execute("""
        SELECT COALESCE(SUM(debit - credit), 0)
        FROM account_move_line
        WHERE partner_id = %s
        AND account_id = %s
        AND parent_state = 'posted'
    """, (partner.id, account.id))

    return self.env.cr.fetchone()[0]


def _get_customer_receivable(self, partner):
    """Get customer receivable balance."""
    return self._get_partner_balance(partner, 'asset_receivable')


def _get_vendor_payable(self, partner):
    """Get vendor payable balance."""
    return self._get_partner_balance(partner, 'liability_payable')
```

### Account Balance
```python
def _get_account_balance(self, account, date_from=None, date_to=None):
    """Get account balance for date range."""
    domain = [
        ('account_id', '=', account.id),
        ('parent_state', '=', 'posted'),
    ]

    if date_from:
        domain.append(('date', '>=', date_from))
    if date_to:
        domain.append(('date', '<=', date_to))

    lines = self.env['account.move.line'].search(domain)
    return sum(lines.mapped('balance'))
```

### Aged Receivables
```python
def _get_aged_receivables(self, partner=None):
    """Get aged receivables report data."""
    today = fields.Date.today()
    periods = [
        ('0-30', 0, 30),
        ('31-60', 31, 60),
        ('61-90', 61, 90),
        ('90+', 91, 9999),
    ]

    domain = [
        ('account_id.account_type', '=', 'asset_receivable'),
        ('parent_state', '=', 'posted'),
        ('reconciled', '=', False),
    ]

    if partner:
        domain.append(('partner_id', '=', partner.id))

    lines = self.env['account.move.line'].search(domain)

    result = {period[0]: 0.0 for period in periods}

    for line in lines:
        days = (today - line.date_maturity).days if line.date_maturity else 0
        for period_name, min_days, max_days in periods:
            if min_days <= days <= max_days:
                result[period_name] += line.amount_residual
                break

    return result
```

---

## Tax Handling

### Get Taxes
```python
def _get_sale_taxes(self, product=None):
    """Get applicable sale taxes."""
    if product:
        return product.taxes_id
    return self.env['account.tax'].search([
        ('type_tax_use', '=', 'sale'),
        ('company_id', '=', self.env.company.id),
    ])


def _get_purchase_taxes(self, product=None):
    """Get applicable purchase taxes."""
    if product:
        return product.supplier_taxes_id
    return self.env['account.tax'].search([
        ('type_tax_use', '=', 'purchase'),
        ('company_id', '=', self.env.company.id),
    ])
```

### Calculate Tax
```python
def _compute_tax_amount(self, amount, taxes, price_include=False):
    """Compute tax amount for given amount and taxes.

    Args:
        amount: Base amount
        taxes: account.tax recordset
        price_include: Whether amount includes tax

    Returns:
        dict with total, taxes breakdown
    """
    tax_results = taxes.compute_all(
        amount,
        currency=self.env.company.currency_id,
        quantity=1.0,
        product=None,
        partner=None,
        is_refund=False,
    )

    return {
        'total_included': tax_results['total_included'],
        'total_excluded': tax_results['total_excluded'],
        'taxes': tax_results['taxes'],
    }
```

---

## Reconciliation

### Auto Reconcile
```python
def _auto_reconcile_partner(self, partner):
    """Auto-reconcile partner's open items."""
    receivable_account = self._get_receivable_account()

    lines = self.env['account.move.line'].search([
        ('partner_id', '=', partner.id),
        ('account_id', '=', receivable_account.id),
        ('reconciled', '=', False),
        ('parent_state', '=', 'posted'),
    ])

    # Group by exact amount match
    by_amount = {}
    for line in lines:
        amount = abs(line.balance)
        if amount not in by_amount:
            by_amount[amount] = {'debit': [], 'credit': []}

        if line.balance > 0:
            by_amount[amount]['debit'].append(line)
        else:
            by_amount[amount]['credit'].append(line)

    # Reconcile matching amounts
    for amount, grouped in by_amount.items():
        if grouped['debit'] and grouped['credit']:
            to_reconcile = grouped['debit'][0] + grouped['credit'][0]
            to_reconcile.reconcile()
```

---

## Best Practices

1. **Always balance entries** - Debits must equal credits
2. **Use correct account types** - Receivable, payable, income, expense
3. **Post entries** - Draft entries don't affect financials
4. **Handle multi-currency** - Use currency conversion methods
5. **Respect fiscal year** - Check date restrictions
6. **Use proper journals** - Sales, purchase, bank, cash, general
7. **Reconcile regularly** - Match payments to invoices
8. **Multi-company aware** - Always filter by company
9. **Tax compliance** - Use correct tax accounts
10. **Audit trail** - Don't delete, use reversals

---


## Source: tax-fiscal-patterns.md

# Tax and Fiscal Position Patterns

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  TAX & FISCAL POSITION PATTERNS                                              ║
║  Tax configuration, fiscal positions, and tax calculations                   ║
║  Use for multi-tax scenarios, international sales, and tax compliance        ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## Tax Types Overview

| Type | Description | Use Case |
|------|-------------|----------|
| Sale | Applied on sales | Customer invoices |
| Purchase | Applied on purchases | Vendor bills |
| None | No tax impact | Internal transfers |

---

## Creating Taxes

### Basic Tax Configuration
```python
# Create a sales tax
sales_tax = self.env['account.tax'].create({
    'name': 'Sales Tax 10%',
    'type_tax_use': 'sale',
    'amount_type': 'percent',
    'amount': 10.0,
    'company_id': self.env.company.id,
    'price_include': False,  # Tax excluded from price
})

# Create a purchase tax
purchase_tax = self.env['account.tax'].create({
    'name': 'Purchase Tax 10%',
    'type_tax_use': 'purchase',
    'amount_type': 'percent',
    'amount': 10.0,
    'company_id': self.env.company.id,
})
```

### Tax Amount Types
```python
# Percentage tax
percent_tax = self.env['account.tax'].create({
    'name': 'VAT 21%',
    'amount_type': 'percent',
    'amount': 21.0,
    'type_tax_use': 'sale',
})

# Fixed amount tax
fixed_tax = self.env['account.tax'].create({
    'name': 'Eco Tax',
    'amount_type': 'fixed',
    'amount': 5.0,  # $5 per unit
    'type_tax_use': 'sale',
})

# Division tax (price includes tax)
division_tax = self.env['account.tax'].create({
    'name': 'VAT Included 21%',
    'amount_type': 'division',
    'amount': 21.0,
    'type_tax_use': 'sale',
    'price_include': True,
})

# Group of taxes
group_tax = self.env['account.tax'].create({
    'name': 'Combined Tax',
    'amount_type': 'group',
    'type_tax_use': 'sale',
    'children_tax_ids': [(6, 0, [tax1.id, tax2.id])],
})
```

---

## Tax on Products

### Set Default Taxes
```python
class ProductTemplate(models.Model):
    _inherit = 'product.template'

    # Customer taxes (for sales)
    taxes_id = fields.Many2many(
        'account.tax',
        'product_taxes_rel',
        'prod_id', 'tax_id',
        string='Customer Taxes',
        domain=[('type_tax_use', '=', 'sale')],
    )

    # Supplier taxes (for purchases)
    supplier_taxes_id = fields.Many2many(
        'account.tax',
        'product_supplier_taxes_rel',
        'prod_id', 'tax_id',
        string='Vendor Taxes',
        domain=[('type_tax_use', '=', 'purchase')],
    )

# Create product with taxes
product = self.env['product.template'].create({
    'name': 'Taxable Product',
    'list_price': 100.00,
    'taxes_id': [(6, 0, [sales_tax.id])],
    'supplier_taxes_id': [(6, 0, [purchase_tax.id])],
})
```

---

## Tax Calculations

### Compute Tax Amounts
```python
def compute_taxes(self, price_unit, quantity, taxes, currency=None, partner=None):
    """Compute tax amounts for a line."""
    if currency is None:
        currency = self.env.company.currency_id

    result = taxes.compute_all(
        price_unit,
        currency=currency,
        quantity=quantity,
        product=None,
        partner=partner,
    )

    return {
        'total_excluded': result['total_excluded'],
        'total_included': result['total_included'],
        'taxes': result['taxes'],
        'base': result['base'],
    }

# Example
result = compute_taxes(100.0, 2, sales_tax)
# result['total_excluded'] = 200.0
# result['total_included'] = 220.0
# result['taxes'] = [{'amount': 20.0, 'name': 'Sales Tax 10%', ...}]
```

### Tax Computation Details
```python
def get_tax_breakdown(self, order):
    """Get tax breakdown for an order."""
    tax_totals = {}

    for line in order.order_line:
        taxes = line.tax_id.compute_all(
            line.price_unit,
            order.currency_id,
            line.product_uom_qty,
            line.product_id,
            order.partner_id,
        )

        for tax_data in taxes['taxes']:
            tax_id = tax_data['id']
            if tax_id not in tax_totals:
                tax_totals[tax_id] = {
                    'name': tax_data['name'],
                    'base': 0,
                    'amount': 0,
                }
            tax_totals[tax_id]['base'] += tax_data['base']
            tax_totals[tax_id]['amount'] += tax_data['amount']

    return list(tax_totals.values())
```

---

## Fiscal Positions

### What is a Fiscal Position?
```
Fiscal Position = Tax Mapping Rules
- Maps taxes to different taxes (or no tax)
- Maps accounts to different accounts
- Based on customer location, type, or other criteria
```

### Create Fiscal Position
```python
# Create fiscal position for EU customers
eu_fiscal = self.env['account.fiscal.position'].create({
    'name': 'EU Customers',
    'auto_apply': True,
    'country_group_id': self.env.ref('base.europe').id,
    'tax_ids': [
        # Map domestic VAT to EU VAT
        (0, 0, {
            'tax_src_id': domestic_vat.id,
            'tax_dest_id': eu_vat.id,
        }),
    ],
    'account_ids': [
        # Map domestic account to EU account
        (0, 0, {
            'account_src_id': domestic_revenue.id,
            'account_dest_id': eu_revenue.id,
        }),
    ],
})

# Create fiscal position for exports (no tax)
export_fiscal = self.env['account.fiscal.position'].create({
    'name': 'Export - No Tax',
    'auto_apply': True,
    'tax_ids': [
        # Map VAT to no tax
        (0, 0, {
            'tax_src_id': domestic_vat.id,
            'tax_dest_id': False,  # No tax
        }),
    ],
})
```

### Map Taxes with Fiscal Position
```python
def get_taxes_for_partner(self, product, partner):
    """Get applicable taxes for a partner."""
    # Get product's default taxes
    taxes = product.taxes_id

    # Get partner's fiscal position
    fiscal_position = partner.property_account_position_id

    if fiscal_position:
        # Map taxes through fiscal position
        taxes = fiscal_position.map_tax(taxes)

    return taxes

# Usage
taxes = get_taxes_for_partner(product, customer)
```

---

## Auto-Apply Fiscal Positions

### Configuration for Auto-Apply
```python
fiscal_position = self.env['account.fiscal.position'].create({
    'name': 'B2B Foreign',
    'auto_apply': True,

    # Match by country
    'country_id': self.env.ref('base.de').id,  # Germany

    # Or by country group
    'country_group_id': eu_group.id,

    # Or by VAT requirement
    'vat_required': True,  # Partner must have VAT number

    # Sequence for priority
    'sequence': 10,
})
```

### Get Fiscal Position for Partner
```python
def get_fiscal_position(self, partner, delivery_partner=None):
    """Determine fiscal position for a partner."""
    # Use delivery address if provided
    partner_to_check = delivery_partner or partner

    # Find applicable fiscal position
    fiscal_position = self.env['account.fiscal.position']._get_fiscal_position(
        partner_to_check,
        delivery=delivery_partner,
    )

    return fiscal_position

# Or use partner's default
fiscal_position = partner.property_account_position_id
```

---

## Tax Included in Price

### Price Include Configuration
```python
# Tax where price includes tax
vat_included = self.env['account.tax'].create({
    'name': 'VAT 20% (Included)',
    'amount': 20.0,
    'amount_type': 'percent',
    'type_tax_use': 'sale',
    'price_include': True,
    'include_base_amount': False,
})

# Calculate base from tax-included price
def get_base_from_included(self, price_total, tax):
    """Extract base amount from price including tax."""
    result = tax.compute_all(
        price_total,
        quantity=1,
    )
    return result['total_excluded']

# If product price is $120 including 20% VAT
base = get_base_from_included(120, vat_included)  # = $100
```

---

## Tax Repartition

### Define Tax Accounts
```python
# Tax with repartition lines (where tax goes)
tax = self.env['account.tax'].create({
    'name': 'VAT 21%',
    'amount': 21.0,
    'type_tax_use': 'sale',
    'invoice_repartition_line_ids': [
        # Base line
        (0, 0, {
            'repartition_type': 'base',
        }),
        # Tax line
        (0, 0, {
            'repartition_type': 'tax',
            'account_id': vat_payable_account.id,
            'tag_ids': [(6, 0, [vat_tag.id])],
        }),
    ],
    'refund_repartition_line_ids': [
        # Base line for refunds
        (0, 0, {
            'repartition_type': 'base',
        }),
        # Tax line for refunds
        (0, 0, {
            'repartition_type': 'tax',
            'account_id': vat_payable_account.id,
            'tag_ids': [(6, 0, [vat_tag.id])],
        }),
    ],
})
```

---

## Tax Groups

### Configure Tax Groups
```python
# Tax group for display
tax_group = self.env['account.tax.group'].create({
    'name': 'VAT',
    'sequence': 10,
})

tax = self.env['account.tax'].create({
    'name': 'VAT 21%',
    'amount': 21.0,
    'tax_group_id': tax_group.id,
    # ...
})
```

---

## Sales Order with Taxes

### Tax Handling in Sale Order
```python
class SaleOrderLine(models.Model):
    _inherit = 'sale.order.line'

    @api.depends('product_uom_qty', 'price_unit', 'tax_id')
    def _compute_amount(self):
        for line in self:
            taxes = line.tax_id.compute_all(
                line.price_unit,
                line.order_id.currency_id,
                line.product_uom_qty,
                product=line.product_id,
                partner=line.order_id.partner_shipping_id,
            )
            line.update({
                'price_tax': taxes['total_included'] - taxes['total_excluded'],
                'price_total': taxes['total_included'],
                'price_subtotal': taxes['total_excluded'],
            })

    @api.onchange('product_id')
    def _onchange_product_id_taxes(self):
        """Apply taxes with fiscal position mapping."""
        if self.product_id:
            taxes = self.product_id.taxes_id
            fiscal = self.order_id.fiscal_position_id
            if fiscal:
                taxes = fiscal.map_tax(taxes)
            self.tax_id = taxes
```

---

## Tax Reporting

### Get Tax Report Data
```python
def get_tax_report(self, date_from, date_to, company=None):
    """Generate tax report data."""
    company = company or self.env.company

    # Find all posted invoices in period
    invoices = self.env['account.move'].search([
        ('company_id', '=', company.id),
        ('state', '=', 'posted'),
        ('date', '>=', date_from),
        ('date', '<=', date_to),
        ('move_type', 'in', ['out_invoice', 'out_refund', 'in_invoice', 'in_refund']),
    ])

    # Aggregate by tax
    tax_data = {}
    for invoice in invoices:
        for line in invoice.line_ids.filtered(lambda l: l.tax_line_id):
            tax = line.tax_line_id
            if tax.id not in tax_data:
                tax_data[tax.id] = {
                    'tax': tax,
                    'base_amount': 0,
                    'tax_amount': 0,
                }
            # Get base and tax amounts
            tax_data[tax.id]['tax_amount'] += line.balance

    return list(tax_data.values())
```

---

## XML Data for Taxes

### Define Tax in Data
```xml
<record id="tax_sale_21" model="account.tax">
    <field name="name">VAT 21%</field>
    <field name="type_tax_use">sale</field>
    <field name="amount_type">percent</field>
    <field name="amount">21</field>
    <field name="tax_group_id" ref="account.tax_group_taxes"/>
    <field name="company_id" ref="base.main_company"/>
</record>

<record id="fiscal_position_export" model="account.fiscal.position">
    <field name="name">Export</field>
    <field name="auto_apply" eval="True"/>
    <field name="company_id" ref="base.main_company"/>
</record>

<record id="fiscal_position_tax_map_1" model="account.fiscal.position.tax">
    <field name="position_id" ref="fiscal_position_export"/>
    <field name="tax_src_id" ref="tax_sale_21"/>
    <!-- tax_dest_id empty = no tax -->
</record>
```

---

## Best Practices

1. **Separate by type** - Sale vs Purchase taxes
2. **Use fiscal positions** - For tax mapping, not manual overrides
3. **Auto-apply rules** - Set up automatic fiscal position assignment
4. **Test calculations** - Verify tax computations thoroughly
5. **Handle refunds** - Configure refund repartition lines
6. **Group taxes** - Use tax groups for reporting
7. **Price include** - Be consistent within product category
8. **Document rules** - Explain fiscal position logic
9. **Multi-company** - Taxes are company-specific
10. **Compliance** - Follow local tax regulations

---

