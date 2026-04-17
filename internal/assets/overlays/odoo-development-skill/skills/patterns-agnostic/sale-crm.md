# Sale Crm Patterns

Consolidated from the following source files:
- `sale-crm-patterns.md`
- `pricelist-pricing-patterns.md`
- `product-variant-patterns.md`

---


## Source: sale-crm-patterns.md

# Sale and CRM Integration Patterns

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  SALE & CRM INTEGRATION PATTERNS                                             ║
║  Sales orders, quotations, leads, opportunities, and pipelines               ║
║  Use for sales automation, CRM customization, and order workflows            ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## Module Setup

### Manifest for Sale Extension
```python
{
    'name': 'My Sale Extension',
    'version': '18.0.1.0.0',
    'depends': ['sale', 'sale_management'],
    'data': [
        'security/ir.model.access.csv',
        'views/sale_views.xml',
    ],
}
```

### Manifest for CRM Extension
```python
{
    'name': 'My CRM Extension',
    'version': '18.0.1.0.0',
    'depends': ['crm'],
    'data': [
        'security/ir.model.access.csv',
        'views/crm_views.xml',
    ],
}
```

---

## Extending Sale Orders

### Add Custom Fields
```python
from odoo import api, fields, models


class SaleOrder(models.Model):
    _inherit = 'sale.order'

    x_project_name = fields.Char(string='Project Name')
    x_delivery_priority = fields.Selection([
        ('normal', 'Normal'),
        ('urgent', 'Urgent'),
        ('critical', 'Critical'),
    ], string='Delivery Priority', default='normal')
    x_internal_notes = fields.Text(string='Internal Notes')
    x_approved_by = fields.Many2one(
        'res.users',
        string='Approved By',
        readonly=True,
    )
    x_approval_date = fields.Datetime(
        string='Approval Date',
        readonly=True,
    )
    x_margin_percent = fields.Float(
        string='Margin %',
        compute='_compute_margin_percent',
        store=True,
    )

    @api.depends('margin', 'amount_untaxed')
    def _compute_margin_percent(self):
        for order in self:
            if order.amount_untaxed:
                order.x_margin_percent = (order.margin / order.amount_untaxed) * 100
            else:
                order.x_margin_percent = 0.0
```

### Add Approval Workflow
```python
class SaleOrder(models.Model):
    _inherit = 'sale.order'

    state = fields.Selection(
        selection_add=[
            ('pending_approval', 'Pending Approval'),
        ],
        ondelete={'pending_approval': 'set default'},
    )
    x_requires_approval = fields.Boolean(
        string='Requires Approval',
        compute='_compute_requires_approval',
    )

    @api.depends('amount_total')
    def _compute_requires_approval(self):
        limit = float(self.env['ir.config_parameter'].sudo().get_param(
            'sale.approval_limit', default='10000'
        ))
        for order in self:
            order.x_requires_approval = order.amount_total > limit

    def action_confirm(self):
        """Override to add approval check."""
        for order in self:
            if order.x_requires_approval and order.state != 'pending_approval':
                order.write({'state': 'pending_approval'})
                order._send_approval_request()
                return True
        return super().action_confirm()

    def action_approve(self):
        """Approve order and continue confirmation."""
        self.write({
            'x_approved_by': self.env.uid,
            'x_approval_date': fields.Datetime.now(),
        })
        return super().action_confirm()

    def _send_approval_request(self):
        """Send approval request notification."""
        template = self.env.ref('my_module.email_template_approval_request')
        template.send_mail(self.id)
```

### Extend Sale Order Lines
```python
class SaleOrderLine(models.Model):
    _inherit = 'sale.order.line'

    x_delivery_date = fields.Date(string='Requested Delivery')
    x_is_custom = fields.Boolean(string='Custom Product')
    x_technical_notes = fields.Text(string='Technical Notes')
    x_cost_price = fields.Float(
        string='Cost Price',
        compute='_compute_cost_price',
        store=True,
    )
    x_line_margin = fields.Float(
        string='Line Margin',
        compute='_compute_line_margin',
        store=True,
    )

    @api.depends('product_id')
    def _compute_cost_price(self):
        for line in self:
            line.x_cost_price = line.product_id.standard_price

    @api.depends('price_subtotal', 'x_cost_price', 'product_uom_qty')
    def _compute_line_margin(self):
        for line in self:
            cost = line.x_cost_price * line.product_uom_qty
            line.x_line_margin = line.price_subtotal - cost
```

---

## Sale Order Automation

### Auto-Apply Discounts
```python
class SaleOrder(models.Model):
    _inherit = 'sale.order'

    @api.onchange('partner_id')
    def _onchange_partner_discount(self):
        """Apply partner-specific discount."""
        if self.partner_id and self.partner_id.x_discount_percent:
            for line in self.order_line:
                line.discount = self.partner_id.x_discount_percent


class ResPartner(models.Model):
    _inherit = 'res.partner'

    x_discount_percent = fields.Float(string='Default Discount %')
    x_credit_limit = fields.Float(string='Credit Limit')
    x_payment_terms_note = fields.Text(string='Payment Terms Note')
```

### Create Order from Template
```python
class SaleOrderTemplate(models.Model):
    _inherit = 'sale.order.template'

    x_auto_confirm = fields.Boolean(string='Auto Confirm Orders')


class SaleOrder(models.Model):
    _inherit = 'sale.order'

    @api.onchange('sale_order_template_id')
    def _onchange_sale_order_template_id(self):
        """Apply template and custom logic."""
        result = super()._onchange_sale_order_template_id()

        if self.sale_order_template_id.x_auto_confirm:
            # Schedule auto-confirmation
            self.x_auto_confirm_scheduled = True

        return result
```

---

## CRM Lead/Opportunity Extension

### Extend CRM Lead
```python
class CrmLead(models.Model):
    _inherit = 'crm.lead'

    x_industry_id = fields.Many2one('res.partner.industry', string='Industry')
    x_budget = fields.Monetary(string='Budget', currency_field='company_currency')
    x_timeline = fields.Selection([
        ('immediate', 'Immediate'),
        ('1_month', '1 Month'),
        ('3_months', '3 Months'),
        ('6_months', '6 Months'),
        ('1_year', '1 Year'),
    ], string='Purchase Timeline')
    x_competitor = fields.Char(string='Main Competitor')
    x_lead_score = fields.Integer(
        string='Lead Score',
        compute='_compute_lead_score',
        store=True,
    )
    x_next_action_date = fields.Date(string='Next Action Date')
    x_lost_reason_details = fields.Text(string='Lost Reason Details')

    @api.depends('expected_revenue', 'probability', 'x_budget', 'x_timeline')
    def _compute_lead_score(self):
        for lead in self:
            score = 0
            # Score based on revenue
            if lead.expected_revenue > 50000:
                score += 30
            elif lead.expected_revenue > 10000:
                score += 20
            elif lead.expected_revenue > 1000:
                score += 10

            # Score based on probability
            score += int(lead.probability * 0.5)

            # Score based on timeline
            timeline_scores = {
                'immediate': 20,
                '1_month': 15,
                '3_months': 10,
                '6_months': 5,
                '1_year': 2,
            }
            score += timeline_scores.get(lead.x_timeline, 0)

            lead.x_lead_score = score
```

### Lead Qualification
```python
class CrmLead(models.Model):
    _inherit = 'crm.lead'

    x_qualification_status = fields.Selection([
        ('unqualified', 'Unqualified'),
        ('mql', 'Marketing Qualified'),
        ('sql', 'Sales Qualified'),
        ('opportunity', 'Opportunity'),
    ], string='Qualification', default='unqualified')

    def action_qualify_mql(self):
        """Mark as Marketing Qualified Lead."""
        self.write({'x_qualification_status': 'mql'})
        self._schedule_qualification_followup()

    def action_qualify_sql(self):
        """Mark as Sales Qualified Lead."""
        self.write({'x_qualification_status': 'sql'})
        self.activity_schedule(
            'mail.mail_activity_data_call',
            summary='SQL Follow-up Call',
            user_id=self.user_id.id,
        )

    def action_convert_to_opportunity(self):
        """Convert to opportunity with custom logic."""
        self.write({
            'x_qualification_status': 'opportunity',
            'type': 'opportunity',
        })
        return self.action_opportunity_form()
```

### Pipeline Stage Automation
```python
class CrmStage(models.Model):
    _inherit = 'crm.stage'

    x_auto_activity = fields.Boolean(string='Auto Schedule Activity')
    x_activity_type_id = fields.Many2one(
        'mail.activity.type',
        string='Activity Type',
    )
    x_activity_days = fields.Integer(string='Days Until Due', default=3)
    x_email_template_id = fields.Many2one(
        'mail.template',
        string='Email Template',
    )


class CrmLead(models.Model):
    _inherit = 'crm.lead'

    def write(self, vals):
        """Auto-create activities on stage change."""
        result = super().write(vals)

        if 'stage_id' in vals:
            for lead in self:
                stage = lead.stage_id
                if stage.x_auto_activity and stage.x_activity_type_id:
                    lead.activity_schedule(
                        stage.x_activity_type_id.id,
                        date_deadline=fields.Date.today() + timedelta(
                            days=stage.x_activity_days
                        ),
                    )
                if stage.x_email_template_id:
                    stage.x_email_template_id.send_mail(lead.id)

        return result
```

---

## Quotation Templates

### Custom Quotation Sections
```python
class SaleOrderLine(models.Model):
    _inherit = 'sale.order.line'

    x_section_type = fields.Selection([
        ('product', 'Product'),
        ('service', 'Service'),
        ('option', 'Optional'),
    ], string='Section Type', default='product')
    x_is_optional = fields.Boolean(string='Optional Item')

    def _get_optional_lines(self):
        """Get optional lines for this order."""
        return self.order_id.order_line.filtered(lambda l: l.x_is_optional)
```

---

## Sales Reports

### Custom Sales Analysis
```python
class SaleReport(models.Model):
    _inherit = 'sale.report'

    x_margin_percent = fields.Float(string='Margin %', readonly=True)
    x_delivery_priority = fields.Selection([
        ('normal', 'Normal'),
        ('urgent', 'Urgent'),
        ('critical', 'Critical'),
    ], string='Priority', readonly=True)

    def _select_additional_fields(self):
        res = super()._select_additional_fields()
        res['x_margin_percent'] = """
            CASE WHEN s.amount_untaxed > 0
            THEN (s.margin / s.amount_untaxed) * 100
            ELSE 0 END
        """
        res['x_delivery_priority'] = "s.x_delivery_priority"
        return res

    def _group_by_sale(self):
        res = super()._group_by_sale()
        res += ", s.x_delivery_priority"
        return res
```

---

## Views

### Sale Order Form Extension
```xml
<?xml version="1.0" encoding="utf-8"?>
<odoo>
    <record id="view_order_form_inherit" model="ir.ui.view">
        <field name="name">sale.order.form.inherit</field>
        <field name="model">sale.order</field>
        <field name="inherit_id" ref="sale.view_order_form"/>
        <field name="arch" type="xml">
            <field name="payment_term_id" position="after">
                <field name="x_project_name"/>
                <field name="x_delivery_priority"/>
            </field>

            <xpath expr="//page[@name='other_information']" position="inside">
                <group string="Approval">
                    <field name="x_requires_approval"/>
                    <field name="x_approved_by"
                           invisible="not x_approved_by"/>
                    <field name="x_approval_date"
                           invisible="not x_approval_date"/>
                </group>
            </xpath>

            <xpath expr="//button[@name='action_confirm']" position="before">
                <button name="action_approve"
                        string="Approve"
                        type="object"
                        class="btn-primary"
                        invisible="state != 'pending_approval'"
                        groups="sales_team.group_sale_manager"/>
            </xpath>
        </field>
    </record>
</odoo>
```

### CRM Lead Form Extension
```xml
<record id="crm_lead_view_form_inherit" model="ir.ui.view">
    <field name="name">crm.lead.form.inherit</field>
    <field name="model">crm.lead</field>
    <field name="inherit_id" ref="crm.crm_lead_view_form"/>
    <field name="arch" type="xml">
        <field name="expected_revenue" position="after">
            <field name="x_budget"/>
            <field name="x_timeline"/>
            <field name="x_lead_score" widget="progressbar"/>
        </field>

        <div name="button_box" position="inside">
            <button class="oe_stat_button" type="object"
                    name="action_view_lead_score"
                    icon="fa-star">
                <field string="Score" name="x_lead_score"
                       widget="statinfo"/>
            </button>
        </div>
    </field>
</record>
```

---

## Best Practices

1. **Don't break standard flow** - Extend, don't replace core methods
2. **Use existing fields** - Check if field exists before adding
3. **Respect access rights** - Sales team vs manager permissions
4. **Performance** - Index frequently searched fields
5. **Multi-company** - Filter by company_id
6. **Currency handling** - Use Monetary fields properly
7. **Report integration** - Extend sale.report for analysis
8. **Email templates** - Use standard mail.template
9. **Activity types** - Use existing or create specific ones
10. **Testing** - Test quotation → order → invoice flow

---


## Source: pricelist-pricing-patterns.md

# Pricelist and Pricing Patterns

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  PRICELIST & PRICING PATTERNS                                                ║
║  Dynamic pricing, discounts, and multi-currency price management             ║
║  Use for sales pricing, promotions, and customer-specific pricing            ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## Pricelist Structure

### Pricelist Hierarchy
```
product.pricelist (Retail Pricelist)
├── product.pricelist.item (All Products: 10% discount)
├── product.pricelist.item (Category "Electronics": 15% discount)
├── product.pricelist.item (Product "Laptop": Fixed $999)
└── product.pricelist.item (Qty >= 10: 20% discount)
```

---

## Creating Pricelists

### Basic Pricelist
```python
pricelist = self.env['product.pricelist'].create({
    'name': 'Retail Pricelist',
    'currency_id': self.env.ref('base.USD').id,
    'company_id': self.env.company.id,
    'sequence': 10,
})
```

### Pricelist with Rules
```python
pricelist = self.env['product.pricelist'].create({
    'name': 'VIP Customers',
    'currency_id': self.env.ref('base.USD').id,
    'item_ids': [
        # 10% discount on all products
        (0, 0, {
            'applied_on': '3_global',  # All products
            'compute_price': 'percentage',
            'percent_price': 10,
        }),
        # 20% off electronics category
        (0, 0, {
            'applied_on': '2_product_category',
            'categ_id': electronics_categ.id,
            'compute_price': 'percentage',
            'percent_price': 20,
        }),
        # Fixed price for specific product
        (0, 0, {
            'applied_on': '1_product',
            'product_tmpl_id': laptop_template.id,
            'compute_price': 'fixed',
            'fixed_price': 899.00,
        }),
        # Quantity-based discount
        (0, 0, {
            'applied_on': '3_global',
            'min_quantity': 10,
            'compute_price': 'percentage',
            'percent_price': 15,
        }),
    ],
})
```

---

## Pricelist Item Types

### Applied On (Scope)
```python
# Global: All products
item = self.env['product.pricelist.item'].create({
    'pricelist_id': pricelist.id,
    'applied_on': '3_global',
    'compute_price': 'percentage',
    'percent_price': 5,
})

# Product Category
item = self.env['product.pricelist.item'].create({
    'pricelist_id': pricelist.id,
    'applied_on': '2_product_category',
    'categ_id': category.id,
    'compute_price': 'percentage',
    'percent_price': 10,
})

# Product Template (all variants)
item = self.env['product.pricelist.item'].create({
    'pricelist_id': pricelist.id,
    'applied_on': '1_product',
    'product_tmpl_id': template.id,
    'compute_price': 'fixed',
    'fixed_price': 50.00,
})

# Product Variant (specific)
item = self.env['product.pricelist.item'].create({
    'pricelist_id': pricelist.id,
    'applied_on': '0_product_variant',
    'product_id': variant.id,
    'compute_price': 'fixed',
    'fixed_price': 45.00,
})
```

### Compute Price Methods
```python
# Fixed price
item.write({
    'compute_price': 'fixed',
    'fixed_price': 99.99,
})

# Percentage discount
item.write({
    'compute_price': 'percentage',
    'percent_price': 15,  # 15% discount
})

# Formula based
item.write({
    'compute_price': 'formula',
    'base': 'list_price',  # or 'standard_price', 'pricelist'
    'price_discount': 10,  # 10% discount
    'price_surcharge': 5,  # Add $5
    'price_round': 0.99,   # Round to .99
    'price_min_margin': 10,  # Minimum 10% margin
    'price_max_margin': 50,  # Maximum 50% margin
})
```

---

## Getting Prices

### Get Product Price
```python
def get_product_price(self, product, pricelist, quantity=1.0, partner=None, date=None):
    """Get product price from pricelist."""
    if date is None:
        date = fields.Date.today()

    return pricelist._get_product_price(
        product,
        quantity,
        partner=partner,
        date=date,
    )

# Usage
price = get_product_price(product, customer_pricelist, quantity=5)
```

### Get Price Rule
```python
def get_price_with_rule(self, product, pricelist, quantity=1.0):
    """Get price and the rule that was applied."""
    price, rule_id = pricelist._get_product_price_rule(
        product,
        quantity,
    )
    rule = self.env['product.pricelist.item'].browse(rule_id)
    return {
        'price': price,
        'rule': rule,
        'discount_percent': rule.percent_price if rule else 0,
    }
```

### Batch Price Calculation
```python
def get_prices_for_products(self, products, pricelist, quantity=1.0):
    """Get prices for multiple products at once."""
    prices = {}
    for product in products:
        prices[product.id] = pricelist._get_product_price(
            product,
            quantity,
        )
    return prices
```

---

## Date-Based Pricing

### Time-Limited Promotions
```python
# Promotional pricing with date range
promo_item = self.env['product.pricelist.item'].create({
    'pricelist_id': pricelist.id,
    'applied_on': '3_global',
    'compute_price': 'percentage',
    'percent_price': 25,  # 25% off
    'date_start': fields.Date.today(),
    'date_end': fields.Date.today() + timedelta(days=7),  # 1 week promo
    'name': 'Summer Sale',
})
```

### Check Active Promotions
```python
def get_active_promotions(self, pricelist):
    """Get currently active promotional rules."""
    today = fields.Date.today()
    return self.env['product.pricelist.item'].search([
        ('pricelist_id', '=', pricelist.id),
        '|',
        ('date_start', '=', False),
        ('date_start', '<=', today),
        '|',
        ('date_end', '=', False),
        ('date_end', '>=', today),
    ])
```

---

## Multi-Currency Pricing

### Currency Conversion
```python
def convert_price(self, amount, from_currency, to_currency, company=None, date=None):
    """Convert price between currencies."""
    if company is None:
        company = self.env.company
    if date is None:
        date = fields.Date.today()

    return from_currency._convert(
        amount,
        to_currency,
        company,
        date,
    )

# Usage
usd_price = 100.00
eur_price = convert_price(
    usd_price,
    self.env.ref('base.USD'),
    self.env.ref('base.EUR'),
)
```

### Pricelist per Currency
```python
# Create pricelist for each currency
usd_pricelist = self.env['product.pricelist'].create({
    'name': 'USD Pricelist',
    'currency_id': self.env.ref('base.USD').id,
})

eur_pricelist = self.env['product.pricelist'].create({
    'name': 'EUR Pricelist',
    'currency_id': self.env.ref('base.EUR').id,
})
```

---

## Customer-Specific Pricing

### Assign Pricelist to Customer
```python
# Set customer's default pricelist
partner = self.env['res.partner'].browse(partner_id)
partner.write({
    'property_product_pricelist': vip_pricelist.id,
})
```

### Get Customer's Price
```python
def get_customer_price(self, partner, product, quantity=1.0):
    """Get price for specific customer."""
    pricelist = partner.property_product_pricelist
    if not pricelist:
        pricelist = self.env.ref('product.list0')  # Default pricelist

    return pricelist._get_product_price(product, quantity, partner=partner)
```

### Customer Price Tiers
```python
# Create customer-specific pricelist
customer_pricelist = self.env['product.pricelist'].create({
    'name': f'Pricelist for {partner.name}',
    'currency_id': partner.currency_id.id or self.env.company.currency_id.id,
    'item_ids': [
        (0, 0, {
            'applied_on': '3_global',
            'compute_price': 'percentage',
            'percent_price': partner.x_discount_rate or 0,
        }),
    ],
})
partner.property_product_pricelist = customer_pricelist
```

---

## Quantity Breaks

### Volume Discounts
```python
pricelist = self.env['product.pricelist'].create({
    'name': 'Volume Discount Pricelist',
    'item_ids': [
        # Base price (qty 1-9)
        (0, 0, {
            'applied_on': '1_product',
            'product_tmpl_id': product.id,
            'min_quantity': 1,
            'compute_price': 'fixed',
            'fixed_price': 100.00,
        }),
        # 10% off for 10-49
        (0, 0, {
            'applied_on': '1_product',
            'product_tmpl_id': product.id,
            'min_quantity': 10,
            'compute_price': 'fixed',
            'fixed_price': 90.00,
        }),
        # 20% off for 50-99
        (0, 0, {
            'applied_on': '1_product',
            'product_tmpl_id': product.id,
            'min_quantity': 50,
            'compute_price': 'fixed',
            'fixed_price': 80.00,
        }),
        # 30% off for 100+
        (0, 0, {
            'applied_on': '1_product',
            'product_tmpl_id': product.id,
            'min_quantity': 100,
            'compute_price': 'fixed',
            'fixed_price': 70.00,
        }),
    ],
})
```

---

## Sales Order Pricing

### Apply Pricelist in Sale Order
```python
class SaleOrder(models.Model):
    _inherit = 'sale.order'

    @api.onchange('partner_id')
    def onchange_partner_id(self):
        """Set pricelist from partner."""
        super().onchange_partner_id()
        if self.partner_id:
            self.pricelist_id = self.partner_id.property_product_pricelist


class SaleOrderLine(models.Model):
    _inherit = 'sale.order.line'

    @api.onchange('product_id', 'product_uom_qty')
    def _onchange_product_id_pricelist(self):
        """Update price from pricelist."""
        if self.product_id and self.order_id.pricelist_id:
            self.price_unit = self.order_id.pricelist_id._get_product_price(
                self.product_id,
                self.product_uom_qty or 1.0,
                partner=self.order_id.partner_id,
            )
```

---

## Custom Pricing Logic

### Override Price Calculation
```python
class ProductPricelist(models.Model):
    _inherit = 'product.pricelist'

    def _get_product_price(self, product, quantity, partner=None, date=False, uom_id=False):
        """Override to add custom pricing logic."""
        price = super()._get_product_price(
            product, quantity, partner=partner, date=date, uom_id=uom_id
        )

        # Custom logic: apply partner-specific discount
        if partner and partner.x_special_discount:
            price = price * (1 - partner.x_special_discount / 100)

        return price
```

### Dynamic Pricing
```python
class ProductProduct(models.Model):
    _inherit = 'product.product'

    def _get_dynamic_price(self):
        """Calculate price based on external factors."""
        base_price = self.list_price

        # Example: adjust based on stock level
        if self.qty_available < 10:
            # Low stock premium
            return base_price * 1.1
        elif self.qty_available > 100:
            # Overstock discount
            return base_price * 0.95

        return base_price
```

---

## XML Data for Pricelists

### Pricelist Definition
```xml
<record id="pricelist_wholesale" model="product.pricelist">
    <field name="name">Wholesale</field>
    <field name="currency_id" ref="base.USD"/>
    <field name="sequence">5</field>
</record>

<record id="pricelist_item_wholesale_global" model="product.pricelist.item">
    <field name="pricelist_id" ref="pricelist_wholesale"/>
    <field name="applied_on">3_global</field>
    <field name="compute_price">percentage</field>
    <field name="percent_price">20</field>
</record>

<record id="pricelist_item_wholesale_electronics" model="product.pricelist.item">
    <field name="pricelist_id" ref="pricelist_wholesale"/>
    <field name="applied_on">2_product_category</field>
    <field name="categ_id" ref="product.product_category_3"/>
    <field name="compute_price">percentage</field>
    <field name="percent_price">25</field>
</record>
```

---

## Price Display

### Show Prices with Taxes
```python
def get_display_price(self, product, pricelist, fiscal_position=None):
    """Get price as displayed to customer (with/without tax)."""
    price = pricelist._get_product_price(product, 1.0)

    if fiscal_position:
        taxes = product.taxes_id.filtered(
            lambda t: t.company_id == self.env.company
        )
        mapped_taxes = fiscal_position.map_tax(taxes)
        price = mapped_taxes.compute_all(
            price,
            currency=pricelist.currency_id,
            quantity=1.0,
            product=product,
        )['total_included']

    return price
```

### Format Price for Display
```python
def format_price(self, amount, currency):
    """Format price for display."""
    from odoo.tools import formatLang
    return formatLang(self.env, amount, currency_obj=currency)
```

---

## Best Practices

1. **Use rule priority** - Specific rules before global (sequence matters)
2. **Date ranges** - Use for promotions, not permanent pricing
3. **Test thoroughly** - Verify all quantity breaks work
4. **Currency consistency** - Match pricelist and partner currency
5. **Audit trail** - Log price changes
6. **Performance** - Cache frequent price lookups
7. **Multi-company** - Separate pricelists per company
8. **Clear naming** - Descriptive pricelist names
9. **Limit complexity** - Too many rules = slow calculation
10. **Document rules** - Explain business logic behind pricing

---


## Source: product-variant-patterns.md

# Product Variant Patterns

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  PRODUCT VARIANT PATTERNS                                                    ║
║  Product templates, variants, attributes, and configurators                  ║
║  Use for managing products with multiple variations (size, color, etc.)      ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## Product Structure Overview

```
product.template (T-Shirt)
├── product.product (T-Shirt, Red, S)
├── product.product (T-Shirt, Red, M)
├── product.product (T-Shirt, Red, L)
├── product.product (T-Shirt, Blue, S)
├── product.product (T-Shirt, Blue, M)
└── product.product (T-Shirt, Blue, L)
```

---

## Product Template vs Product

### Understanding the Difference
```python
# product.template = The "master" product
# product.product = Specific variant (SKU)

# Template fields (shared across variants)
template = self.env['product.template'].create({
    'name': 'T-Shirt',
    'type': 'product',
    'categ_id': category.id,
    'description': 'Cotton t-shirt',  # Same for all variants
})

# Variant fields (unique per variant)
variant = self.env['product.product'].create({
    'product_tmpl_id': template.id,
    'default_code': 'TSHIRT-RED-M',  # Unique SKU
    'barcode': '1234567890123',       # Unique barcode
})
```

### Accessing Template from Variant
```python
# From variant to template
variant = self.env['product.product'].browse(product_id)
template = variant.product_tmpl_id

# From template to variants
template = self.env['product.template'].browse(template_id)
variants = template.product_variant_ids
first_variant = template.product_variant_id  # Single variant
```

---

## Product Attributes

### Creating Attributes
```python
# Create attribute (e.g., Color, Size)
color_attr = self.env['product.attribute'].create({
    'name': 'Color',
    'display_type': 'color',  # radio, select, color, multi
    'create_variant': 'always',  # always, dynamic, no_variant
})

size_attr = self.env['product.attribute'].create({
    'name': 'Size',
    'display_type': 'radio',
    'create_variant': 'always',
})
```

### Creating Attribute Values
```python
# Create attribute values
colors = self.env['product.attribute.value'].create([
    {'name': 'Red', 'attribute_id': color_attr.id, 'html_color': '#FF0000'},
    {'name': 'Blue', 'attribute_id': color_attr.id, 'html_color': '#0000FF'},
    {'name': 'Green', 'attribute_id': color_attr.id, 'html_color': '#00FF00'},
])

sizes = self.env['product.attribute.value'].create([
    {'name': 'S', 'attribute_id': size_attr.id, 'sequence': 1},
    {'name': 'M', 'attribute_id': size_attr.id, 'sequence': 2},
    {'name': 'L', 'attribute_id': size_attr.id, 'sequence': 3},
    {'name': 'XL', 'attribute_id': size_attr.id, 'sequence': 4},
])
```

---

## Assigning Attributes to Products

### Add Attribute Lines to Template
```python
template = self.env['product.template'].create({
    'name': 'T-Shirt',
    'type': 'product',
    'attribute_line_ids': [
        (0, 0, {
            'attribute_id': color_attr.id,
            'value_ids': [(6, 0, colors.ids)],  # Red, Blue, Green
        }),
        (0, 0, {
            'attribute_id': size_attr.id,
            'value_ids': [(6, 0, sizes.ids)],  # S, M, L, XL
        }),
    ],
})

# This creates 3 colors × 4 sizes = 12 variants automatically
print(f"Created {len(template.product_variant_ids)} variants")
```

### Variant Creation Modes
```python
# 'always' - Variants created immediately for all combinations
# 'dynamic' - Variants created only when selected (sales/purchases)
# 'no_variant' - No variants, attribute stored on order line only
```

---

## Working with Variants

### Finding Specific Variant
```python
def get_variant(self, template, attribute_values):
    """Find variant matching specific attribute values."""
    domain = [('product_tmpl_id', '=', template.id)]

    for attr_value in attribute_values:
        domain.append(('product_template_attribute_value_ids.product_attribute_value_id', '=', attr_value.id))

    return self.env['product.product'].search(domain, limit=1)

# Usage
red = self.env['product.attribute.value'].search([('name', '=', 'Red')])
medium = self.env['product.attribute.value'].search([('name', '=', 'M')])
variant = get_variant(template, red | medium)
```

### Get Variant by Attribute Combination
```python
def _get_variant_for_combination(self, template, combination):
    """Get or create variant for attribute combination."""
    product = template._get_variant_for_combination(combination)
    if not product:
        # Create if dynamic variants enabled
        product = template._create_product_variant(combination)
    return product
```

---

## Variant Pricing

### Price Extra per Attribute Value
```python
# Add price extra to attribute value
template.attribute_line_ids.filtered(
    lambda l: l.attribute_id == size_attr
).product_template_value_ids.filtered(
    lambda v: v.name == 'XL'
).price_extra = 5.00  # +$5 for XL size

# Or via product.template.attribute.value
ptav = self.env['product.template.attribute.value'].search([
    ('product_tmpl_id', '=', template.id),
    ('product_attribute_value_id.name', '=', 'XL'),
])
ptav.price_extra = 5.00
```

### Get Variant Price
```python
def get_variant_price(self, variant, pricelist=None):
    """Get variant price including extras."""
    if pricelist:
        price = pricelist._get_product_price(
            variant,
            quantity=1.0,
            currency=pricelist.currency_id,
        )
    else:
        price = variant.lst_price  # Includes price_extra

    return price
```

---

## Variant-Specific Fields

### Extending Product Variant
```python
class ProductProduct(models.Model):
    _inherit = 'product.product'

    # Variant-specific fields
    variant_sku = fields.Char(string='Variant SKU')
    variant_weight = fields.Float(string='Variant Weight')

    # Override to make template field variant-specific
    weight = fields.Float(
        string='Weight',
        compute='_compute_weight',
        inverse='_set_weight',
        store=True,
    )

    @api.depends('product_tmpl_id.weight', 'variant_weight')
    def _compute_weight(self):
        for product in self:
            product.weight = product.variant_weight or product.product_tmpl_id.weight

    def _set_weight(self):
        for product in self:
            product.variant_weight = product.weight
```

---

## Configurable Products

### Product Configurator in Sales
```python
class SaleOrderLine(models.Model):
    _inherit = 'sale.order.line'

    # For no_variant attributes
    product_no_variant_attribute_value_ids = fields.Many2many(
        'product.template.attribute.value',
        string='Extra Values',
    )

    @api.onchange('product_id')
    def _onchange_product_id_variant_selector(self):
        """Open configurator for products with variants."""
        if self.product_id.product_tmpl_id.has_configurable_attributes:
            return {
                'type': 'ir.actions.act_window',
                'res_model': 'sale.product.configurator',
                'view_mode': 'form',
                'target': 'new',
                'context': {
                    'default_product_template_id': self.product_id.product_tmpl_id.id,
                },
            }
```

### Custom Configurator
```python
class ProductConfigurator(models.TransientModel):
    _name = 'product.configurator'
    _description = 'Product Configurator'

    product_template_id = fields.Many2one('product.template', required=True)
    attribute_line_ids = fields.One2many(
        'product.configurator.line',
        'configurator_id',
        string='Attributes',
    )

    @api.onchange('product_template_id')
    def _onchange_product_template(self):
        """Load attribute lines."""
        self.attribute_line_ids = [(5, 0, 0)]
        lines = []
        for attr_line in self.product_template_id.attribute_line_ids:
            lines.append((0, 0, {
                'attribute_id': attr_line.attribute_id.id,
                'value_ids': [(6, 0, attr_line.value_ids.ids)],
            }))
        self.attribute_line_ids = lines

    def action_configure(self):
        """Get configured variant."""
        combination = self.env['product.template.attribute.value']
        for line in self.attribute_line_ids:
            if line.selected_value_id:
                ptav = self.env['product.template.attribute.value'].search([
                    ('product_tmpl_id', '=', self.product_template_id.id),
                    ('product_attribute_value_id', '=', line.selected_value_id.id),
                ])
                combination |= ptav

        variant = self.product_template_id._get_variant_for_combination(combination)
        return variant


class ProductConfiguratorLine(models.TransientModel):
    _name = 'product.configurator.line'
    _description = 'Product Configurator Line'

    configurator_id = fields.Many2one('product.configurator')
    attribute_id = fields.Many2one('product.attribute')
    value_ids = fields.Many2many('product.attribute.value')
    selected_value_id = fields.Many2one(
        'product.attribute.value',
        domain="[('id', 'in', value_ids)]",
    )
```

---

## Variant Images

### Per-Variant Images
```python
class ProductProduct(models.Model):
    _inherit = 'product.product'

    # Variant has its own image or falls back to template
    image_variant_1920 = fields.Image(max_width=1920, max_height=1920)

    # The standard image field computes from variant or template
    image_1920 = fields.Image(compute='_compute_image_1920', store=True)

    @api.depends('image_variant_1920', 'product_tmpl_id.image_1920')
    def _compute_image_1920(self):
        for record in self:
            record.image_1920 = record.image_variant_1920 or record.product_tmpl_id.image_1920
```

### Image per Attribute Value
```python
# Assign image to attribute value (color swatch)
color_value = self.env['product.attribute.value'].browse(value_id)
color_value.write({
    'image': base64_encoded_image,
})
```

---

## Variant Stock

### Check Stock per Variant
```python
def check_variant_availability(self, variant, warehouse=None):
    """Check stock for specific variant."""
    if warehouse:
        qty = variant.with_context(warehouse=warehouse.id).qty_available
    else:
        qty = variant.qty_available

    return {
        'available': qty,
        'incoming': variant.incoming_qty,
        'outgoing': variant.outgoing_qty,
        'forecasted': variant.virtual_available,
    }
```

### Stock per Location
```python
def get_variant_stock_by_location(self, variant):
    """Get stock breakdown by location."""
    quants = self.env['stock.quant'].search([
        ('product_id', '=', variant.id),
        ('quantity', '>', 0),
    ])

    return [{
        'location': q.location_id.complete_name,
        'quantity': q.quantity,
        'reserved': q.reserved_quantity,
    } for q in quants]
```

---

## XML Data for Attributes

### Attribute Definition
```xml
<record id="product_attribute_color" model="product.attribute">
    <field name="name">Color</field>
    <field name="display_type">color</field>
    <field name="create_variant">always</field>
</record>

<record id="product_attribute_value_red" model="product.attribute.value">
    <field name="name">Red</field>
    <field name="attribute_id" ref="product_attribute_color"/>
    <field name="html_color">#FF0000</field>
    <field name="sequence">1</field>
</record>

<record id="product_attribute_value_blue" model="product.attribute.value">
    <field name="name">Blue</field>
    <field name="attribute_id" ref="product_attribute_color"/>
    <field name="html_color">#0000FF</field>
    <field name="sequence">2</field>
</record>
```

### Product with Attributes
```xml
<record id="product_template_tshirt" model="product.template">
    <field name="name">T-Shirt</field>
    <field name="type">product</field>
    <field name="categ_id" ref="product.product_category_all"/>
    <field name="list_price">29.99</field>
    <field name="attribute_line_ids" eval="[
        (0, 0, {
            'attribute_id': ref('product_attribute_color'),
            'value_ids': [(6, 0, [ref('product_attribute_value_red'), ref('product_attribute_value_blue')])],
        }),
    ]"/>
</record>
```

---

## Variant Search and Filtering

### Search by Attribute
```python
def search_by_attribute(self, attribute_name, value_name):
    """Find variants with specific attribute value."""
    return self.env['product.product'].search([
        ('product_template_attribute_value_ids.product_attribute_value_id.name', '=', value_name),
        ('product_template_attribute_value_ids.attribute_id.name', '=', attribute_name),
    ])

# Find all red products
red_products = search_by_attribute('Color', 'Red')
```

### Filter in View
```xml
<search>
    <field name="name"/>
    <field name="categ_id"/>
    <filter name="filter_red" string="Red Products"
            domain="[('product_template_attribute_value_ids.product_attribute_value_id.name', '=', 'Red')]"/>
    <group expand="0" string="Group By">
        <filter name="group_by_attribute" string="Color"
                context="{'group_by': 'product_template_attribute_value_ids'}"/>
    </group>
</search>
```

---

## Best Practices

1. **Use templates** - Define shared data on template, unique on variant
2. **Choose variant mode** - `always` for few combos, `dynamic` for many
3. **Price extras** - Use for simple attribute-based pricing
4. **Separate SKUs** - Each variant should have unique identifier
5. **Image strategy** - Template image as fallback, variant when different
6. **Stock tracking** - Always at variant level
7. **Limit combinations** - Too many variants = performance issues
8. **Archive unused** - Don't delete, archive discontinued variants
9. **Test configurator** - Ensure all valid combinations work
10. **Document attributes** - Clear naming for attributes and values

---

