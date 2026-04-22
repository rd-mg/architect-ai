# Accounting Patterns

Consolidated from the following source files:
- `invoice-bill-patterns.md` (architect-ai)
- `payment-reconciliation-patterns.md` (architect-ai)
- `journal-entry-patterns.md` (architect-ai)
- `tax-fiscal-patterns.md` (architect-ai)

> **Version-specific syntax** → `patterns-{version}/model-patterns.md`
> `move_type` introduced in v13+ · `line_ids` vs `invoice_line_ids` logic

---

## Invoices & Moves

### Creating an Invoice
```python
class AccountMove(models.Model):
    _inherit = 'account.move'

    def create_custom_invoice(self, partner, lines):
        return self.create({
            'move_type': 'out_invoice',
            'partner_id': partner.id,
            'invoice_line_ids': [(0, 0, {
                'product_id': l['product_id'],
                'quantity': l['qty'],
                'price_unit': l['price'],
            }) for l in lines]
        })
```

### Reconciling Payments
```python
def reconcile_payment(self, payment, invoice):
    # Links a payment to an invoice
    (payment.move_id.line_ids + invoice.line_ids).filtered(
        lambda l: l.account_id == payment.destination_account_id
    ).reconcile()
```

---

## Anti-Patterns

```python
# ❌ NEVER update account.move.line directly for posted moves.
# ✅ CORRECT: Reverse/Cancel the move, or use credit notes.

# ❌ NEVER use cr.execute to change accounting balances.
# ✅ CORRECT: Always use the ORM to ensure journal consistency.

# ❌ NEVER skip tax calculation on manual lines.
# ✅ CORRECT: Call _compute_tax_totals() or ensure taxes_id is set.
```

---

## Version Matrix

| Feature | v14-v16 | v17 | v18 | v19 |
|---------|---------|-----|-----|-----|
| Move Type | `move_type` | `move_type` | `move_type` | `move_type` |
| Lines | `invoice_line_ids`| `invoice_line_ids`| `invoice_line_ids`| `invoice_line_ids`|
| Reconcile | `reconcile()` | `reconcile()` | `reconcile()` | `reconcile()` |
| Tax Totals | `tax_totals_json`| `tax_totals` | `tax_totals` | `tax_totals` |
