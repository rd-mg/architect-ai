# Purchase & Procurement Patterns

Consolidated from the following source files:
- `purchase-order-patterns.md` (architect-fix)
- `vendor-bill-patterns.md` (architect-fix)
- `procurement-rule-patterns.md` (architect-fix)

> **Version-specific syntax** → `patterns-{version}/model-patterns.md`
> `_check_company_auto` in v18+ · SQL() mandatory in v19+

---

## Purchase & Vendor Flows

### Custom Approval Flow
```python
class PurchaseOrder(models.Model):
    _inherit = 'purchase.order'

    @api.depends('amount_total')
    def _compute_approval_required(self):
        for order in self:
            order.approval_required = order.amount_total > 5000.0

    def button_confirm(self):
        if self.approval_required and not self.approved_by_id:
            raise UserError("Approval required before confirmation.")
        return super().button_confirm()
```

### Programmatic PO Creation
```python
def create_po(self, vendor, products_data):
    order = self.env['purchase.order'].create({'partner_id': vendor.id})
    for p in products_data:
        self.env['purchase.order.line'].create({
            'order_id': order.id,
            'product_id': p['id'],
            'product_qty': p['qty'],
            'price_unit': p['price'],
            'date_planned': fields.Date.today(),
        })
    return order
```

---

## Procurement & Rules

### Trigger Replenishment
```python
def trigger_mto(self, product, qty, location):
    # Triggers Odoo's procurement engine (MTO, Reorder Points)
    self.env['procurement.group'].run([
        self.env['procurement.group'].Procurement(
            product, qty, product.uom_id, location,
            "Manual trigger", "origin_ref", self.env.company, {}
        )
    ])
```

---

## Anti-Patterns

```python
# ❌ NEVER use product.standard_price as vendor price.
# ✅ CORRECT: Use seller = product._select_seller(partner_id=vendor).

# ❌ NEVER create account.move for bills without move_type='in_invoice'.

# ❌ NEVER confirm a PO without checking custom approval states in overrides.
```

---

## Version Matrix

| Feature | v14-v17 | v18 | v19 |
|---------|---------|-----|-----|
| Multi-company | `company_id` | `check_company=True` | `check_company=True` |
| Constraints | `@api.constrains` | `@api.constrains` | `SQL()` |
| Invoice Creation | `action_view_invoice` | `action_create_invoice` | `action_create_invoice` |
