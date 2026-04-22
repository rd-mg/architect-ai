# Stock & Inventory Patterns

Consolidated from the following source files:
- `stock-picking-patterns.md` (architect-ai)
- `inventory-adjustment-patterns.md` (architect-ai)
- `warehouse-location-patterns.md` (architect-ai)
- `stock-move-patterns.md` (architect-ai)

> **Version-specific syntax** → `patterns-{version}/model-patterns.md`
> `qty_done` renamed `quantity` in v17+ · `product_qty` vs `reserved_qty` logic

---

## Picking & Moves

### Validating a Picking Programmatically
```python
def validate_picking(self, picking):
    for move in picking.move_ids:
        # v17+ uses 'quantity', older use 'qty_done'
        move.quantity = move.product_uom_qty 
    picking.button_validate()
```

### Reservation Logic
```python
def reserve_stock(self, picking):
    # Triggers Odoo's reservation engine
    picking.action_assign()
    if picking.state != 'assigned':
        raise UserError("Not enough stock to reserve.")
```

---

## Warehouse & Locations

### Finding Stock in Location
```python
def get_stock_at_location(self, product, location):
    # qty_available is computed based on location in context
    return product.with_context(location=location.id).qty_available
```

---

## Anti-Patterns

```python
# ❌ NEVER manually update qty_available: it is a computed field.
# ✅ CORRECT: Use inventory adjustments or stock.move to change stock.

# ❌ NEVER bypass the picking state machine (draft -> waiting -> assigned -> done).

# ❌ NEVER use product.qty_available without location context if you need specific site stock.
```

---

## Version Matrix

| Feature | v14-v16 | v17 | v18 | v19 |
|---------|---------|-----|-----|-----|
| Done Qty | `qty_done` | `quantity` | `quantity` | `quantity` |
| Reservation| `action_assign` | `action_assign` | `action_assign` | `action_assign` |
| Adjustments| `stock.inventory`| `stock.quant` | `stock.quant` | `stock.quant` |
| UI/Views | `<tree>` | `<tree>` | `<list>` | `<list>` |
