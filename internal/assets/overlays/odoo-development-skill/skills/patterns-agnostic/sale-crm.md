# Sale & CRM Patterns

Consolidated from the following source files:
- `sale-crm-patterns.md` (architect-ai)
- `pricelist-pricing-patterns.md` (architect-ai)
- `product-variant-patterns.md` (architect-ai)

> **Version-specific syntax** → `patterns-{version}/model-patterns.md`
> `attrs=` removed in v17+ · `<tree>` renamed `<list>` in v18+ · SQL() mandatory in v19+

---

## Sales Automation

### Custom Approval Workflow
```python
class SaleOrder(models.Model):
    _inherit = 'sale.order'

    approval_state = fields.Selection([
        ('draft', 'Draft'),
        ('pending', 'Pending Approval'),
        ('approved', 'Approved'),
    ], default='draft')

    def action_confirm(self):
        if self.amount_total > 5000 and self.approval_state != 'approved':
            self.approval_state = 'pending'
            return False
        return super().action_confirm()
```

### Margin Calculation
```python
class SaleOrderLine(models.Model):
    _inherit = 'sale.order.line'

    margin_percent = fields.Float(compute='_compute_margin_percent', store=True)

    @api.depends('price_subtotal', 'purchase_price')
    def _compute_margin_percent(self):
        for line in self:
            if line.price_subtotal:
                line.margin_percent = (line.price_subtotal - line.purchase_price) / line.price_subtotal
```

---

## CRM & Pipelines

### Lead Qualification Logic
```python
class CrmLead(models.Model):
    _inherit = 'crm.lead'

    lead_score = fields.Integer(compute='_compute_lead_score', store=True)

    @api.depends('expected_revenue', 'probability')
    def _compute_lead_score(self):
        for lead in self:
            lead.lead_score = int(lead.expected_revenue * lead.probability / 100)
```

---

## Pricing & Pricelists

### Get Price with Rule
```python
def get_price(self, product, pricelist, qty=1):
    # Returns (price, rule_id)
    return pricelist._get_product_price_rule(product, qty)
```

---

## Anti-Patterns

```python
# ❌ NEVER hardcode currency conversion.
# ✅ CORRECT: Use from_currency._convert(amount, to_currency, company, date).

# ❌ NEVER use attrs= in v17+ views.
# ✅ CORRECT: Use invisible="state != 'draft'" directly on the field/button.

# ❌ NEVER bypass pricelists for customer pricing.
# ✅ CORRECT: Always use partner.property_product_pricelist.
```

---

## Version Matrix

| Feature | v14-v16 | v17 | v18 | v19 |
|---------|---------|-----|-----|-----|
| `attrs=` | ✅ | ❌ | ❌ | ❌ |
| `<tree>` | ✅ | ✅ | `<list>` | `<list>` |
| Tracking | `track_visibility`| `tracking=True` | `tracking=True` | `tracking=True` |
| Multi-company| `company_id` | `company_id` | `check_company` | `check_company` |
