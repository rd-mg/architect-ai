# Stock Inventory Patterns

Consolidated from the following source files:
- `stock-inventory-patterns.md`
- `lot-serial-patterns.md`
- `uom-patterns.md`

---


## Source: stock-inventory-patterns.md

# Stock and Inventory Patterns

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  STOCK & INVENTORY PATTERNS                                                  ║
║  Warehouse management, stock moves, and inventory operations                 ║
║  Use for inventory tracking, transfers, and logistics automation             ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## Module Setup

### Manifest Dependencies
```python
{
    'name': 'My Stock Module',
    'version': '18.0.1.0.0',
    'depends': ['stock'],
    'data': [
        'security/ir.model.access.csv',
        'views/stock_views.xml',
    ],
}
```

---

## Extending Stock Models

### Extend Product for Stock
```python
from odoo import api, fields, models


class ProductTemplate(models.Model):
    _inherit = 'product.template'

    x_min_stock_qty = fields.Float(
        string='Minimum Stock Quantity',
        default=0.0,
        help='Alert when stock falls below this level',
    )
    x_max_stock_qty = fields.Float(
        string='Maximum Stock Quantity',
        default=0.0,
    )
    x_reorder_qty = fields.Float(
        string='Reorder Quantity',
        default=1.0,
    )
    x_stock_status = fields.Selection(
        selection=[
            ('ok', 'In Stock'),
            ('low', 'Low Stock'),
            ('out', 'Out of Stock'),
        ],
        string='Stock Status',
        compute='_compute_stock_status',
        store=True,
    )

    @api.depends('qty_available', 'x_min_stock_qty')
    def _compute_stock_status(self):
        for product in self:
            if product.qty_available <= 0:
                product.x_stock_status = 'out'
            elif product.qty_available < product.x_min_stock_qty:
                product.x_stock_status = 'low'
            else:
                product.x_stock_status = 'ok'
```

### Extend Stock Picking
```python
class StockPicking(models.Model):
    _inherit = 'stock.picking'

    x_delivery_instructions = fields.Text(string='Delivery Instructions')
    x_priority_level = fields.Selection([
        ('normal', 'Normal'),
        ('urgent', 'Urgent'),
        ('critical', 'Critical'),
    ], string='Priority', default='normal')
    x_carrier_tracking = fields.Char(string='Carrier Tracking')

    def action_done(self):
        """Override to add custom logic after validation."""
        result = super().action_done()

        for picking in self:
            if picking.picking_type_code == 'outgoing':
                picking._send_shipping_notification()

        return result

    def _send_shipping_notification(self):
        """Send notification when shipment is done."""
        template = self.env.ref('my_module.email_template_shipment')
        if self.partner_id.email:
            template.send_mail(self.id)
```

### Extend Stock Move
```python
class StockMove(models.Model):
    _inherit = 'stock.move'

    x_custom_cost = fields.Float(string='Custom Cost')
    x_batch_number = fields.Char(string='Batch Number')

    def _action_done(self, cancel_backorder=False):
        """Override to add tracking after move completion."""
        result = super()._action_done(cancel_backorder=cancel_backorder)

        for move in self:
            if move.product_id.x_min_stock_qty:
                move._check_reorder_point()

        return result

    def _check_reorder_point(self):
        """Check if reorder is needed after stock change."""
        product = self.product_id
        if product.qty_available < product.x_min_stock_qty:
            self._create_reorder_notification()
```

---

## Stock Operations

### Create Stock Move
```python
def _create_stock_move(self, product, qty, src_location, dest_location):
    """Create a stock movement."""
    move = self.env['stock.move'].create({
        'name': f'Move: {product.name}',
        'product_id': product.id,
        'product_uom_qty': qty,
        'product_uom': product.uom_id.id,
        'location_id': src_location.id,
        'location_dest_id': dest_location.id,
        'picking_type_id': self._get_picking_type().id,
        'origin': self.name,
    })
    move._action_confirm()
    move._action_assign()
    return move
```

### Create Transfer (Picking)
```python
def _create_delivery(self, partner, lines):
    """Create outgoing delivery order."""
    warehouse = self.env['stock.warehouse'].search([
        ('company_id', '=', self.env.company.id),
    ], limit=1)

    picking = self.env['stock.picking'].create({
        'partner_id': partner.id,
        'picking_type_id': warehouse.out_type_id.id,
        'location_id': warehouse.lot_stock_id.id,
        'location_dest_id': partner.property_stock_customer.id,
        'origin': self.name,
    })

    for line in lines:
        self.env['stock.move'].create({
            'name': line['product'].name,
            'product_id': line['product'].id,
            'product_uom_qty': line['qty'],
            'product_uom': line['product'].uom_id.id,
            'picking_id': picking.id,
            'location_id': picking.location_id.id,
            'location_dest_id': picking.location_dest_id.id,
        })

    picking.action_confirm()
    picking.action_assign()

    return picking
```

### Internal Transfer
```python
def _create_internal_transfer(self, product, qty, src_location, dest_location):
    """Create internal transfer between locations."""
    warehouse = self.env['stock.warehouse'].search([
        ('company_id', '=', self.env.company.id),
    ], limit=1)

    picking = self.env['stock.picking'].create({
        'picking_type_id': warehouse.int_type_id.id,
        'location_id': src_location.id,
        'location_dest_id': dest_location.id,
        'origin': f'Internal Transfer: {self.name}',
    })

    self.env['stock.move'].create({
        'name': product.name,
        'product_id': product.id,
        'product_uom_qty': qty,
        'product_uom': product.uom_id.id,
        'picking_id': picking.id,
        'location_id': src_location.id,
        'location_dest_id': dest_location.id,
    })

    picking.action_confirm()
    picking.action_assign()

    return picking
```

---

## Inventory Adjustments

### Create Inventory Adjustment
```python
def _adjust_inventory(self, product, location, new_qty, reason):
    """Adjust inventory quantity for a product."""
    quant = self.env['stock.quant'].search([
        ('product_id', '=', product.id),
        ('location_id', '=', location.id),
    ], limit=1)

    current_qty = quant.quantity if quant else 0
    diff = new_qty - current_qty

    if diff == 0:
        return

    # Use inventory adjustment
    self.env['stock.quant'].with_context(inventory_mode=True).create({
        'product_id': product.id,
        'location_id': location.id,
        'inventory_quantity': new_qty,
    }).action_apply_inventory()
```

### Batch Inventory Update
```python
def _batch_inventory_adjustment(self, adjustments):
    """Process batch inventory adjustments.

    Args:
        adjustments: List of dicts with product_id, location_id, qty
    """
    for adj in adjustments:
        product = self.env['product.product'].browse(adj['product_id'])
        location = self.env['stock.location'].browse(adj['location_id'])

        quant = self.env['stock.quant'].search([
            ('product_id', '=', product.id),
            ('location_id', '=', location.id),
        ], limit=1)

        if quant:
            quant.with_context(inventory_mode=True).write({
                'inventory_quantity': adj['qty'],
            })
        else:
            self.env['stock.quant'].with_context(inventory_mode=True).create({
                'product_id': product.id,
                'location_id': location.id,
                'inventory_quantity': adj['qty'],
            })

    # Apply all adjustments
    quants_to_apply = self.env['stock.quant'].search([
        ('inventory_quantity_set', '=', True),
    ])
    quants_to_apply.action_apply_inventory()
```

---

## Stock Queries

### Get Available Quantity
```python
def _get_available_qty(self, product, location=None):
    """Get available quantity for a product."""
    if location:
        return product.with_context(location=location.id).qty_available
    return product.qty_available


def _get_free_qty(self, product, location=None):
    """Get free (unreserved) quantity."""
    if location:
        return product.with_context(location=location.id).free_qty
    return product.free_qty
```

### Get Stock by Location
```python
def _get_stock_by_location(self, product):
    """Get stock quantities per location."""
    quants = self.env['stock.quant'].search([
        ('product_id', '=', product.id),
        ('location_id.usage', '=', 'internal'),
    ])

    return {
        quant.location_id: {
            'quantity': quant.quantity,
            'reserved': quant.reserved_quantity,
            'available': quant.quantity - quant.reserved_quantity,
        }
        for quant in quants
    }
```

### Check Stock Availability
```python
def _check_stock_availability(self, lines):
    """Check if all lines can be fulfilled from stock.

    Args:
        lines: List of dicts with product_id, qty, location_id (optional)

    Returns:
        dict: {available: bool, missing: list of dicts}
    """
    missing = []

    for line in lines:
        product = self.env['product.product'].browse(line['product_id'])
        location_id = line.get('location_id')

        if location_id:
            available = product.with_context(location=location_id).free_qty
        else:
            available = product.free_qty

        if available < line['qty']:
            missing.append({
                'product': product,
                'requested': line['qty'],
                'available': available,
                'shortage': line['qty'] - available,
            })

    return {
        'available': len(missing) == 0,
        'missing': missing,
    }
```

---

## Lot and Serial Tracking

### Create with Lot
```python
def _create_move_with_lot(self, product, qty, lot_name, src_location, dest_location):
    """Create stock move with lot tracking."""
    # Ensure lot exists
    lot = self.env['stock.lot'].search([
        ('name', '=', lot_name),
        ('product_id', '=', product.id),
        ('company_id', '=', self.env.company.id),
    ], limit=1)

    if not lot:
        lot = self.env['stock.lot'].create({
            'name': lot_name,
            'product_id': product.id,
            'company_id': self.env.company.id,
        })

    move = self.env['stock.move'].create({
        'name': product.name,
        'product_id': product.id,
        'product_uom_qty': qty,
        'product_uom': product.uom_id.id,
        'location_id': src_location.id,
        'location_dest_id': dest_location.id,
    })

    move._action_confirm()
    move._action_assign()

    # Set lot on move line
    for move_line in move.move_line_ids:
        move_line.lot_id = lot

    move._action_done()
    return move
```

### Query by Lot
```python
def _get_stock_by_lot(self, product):
    """Get stock quantities per lot."""
    quants = self.env['stock.quant'].search([
        ('product_id', '=', product.id),
        ('location_id.usage', '=', 'internal'),
        ('lot_id', '!=', False),
    ])

    return {
        quant.lot_id: {
            'location': quant.location_id,
            'quantity': quant.quantity,
            'expiry_date': quant.lot_id.expiration_date,
        }
        for quant in quants
    }
```

---

## Warehouse Operations

### Get Default Warehouse
```python
def _get_warehouse(self):
    """Get warehouse for current company."""
    return self.env['stock.warehouse'].search([
        ('company_id', '=', self.env.company.id),
    ], limit=1)
```

### Get Picking Type
```python
def _get_picking_type(self, operation='outgoing'):
    """Get picking type by operation.

    Args:
        operation: 'incoming', 'outgoing', 'internal'
    """
    warehouse = self._get_warehouse()

    if operation == 'incoming':
        return warehouse.in_type_id
    elif operation == 'outgoing':
        return warehouse.out_type_id
    else:
        return warehouse.int_type_id
```

### Get Stock Locations
```python
def _get_stock_location(self):
    """Get main stock location."""
    warehouse = self._get_warehouse()
    return warehouse.lot_stock_id


def _get_customer_location(self):
    """Get customer (output) location."""
    return self.env.ref('stock.stock_location_customers')


def _get_supplier_location(self):
    """Get supplier (input) location."""
    return self.env.ref('stock.stock_location_suppliers')
```

---

## Scheduled Actions

### Low Stock Alert Cron
```python
@api.model
def _cron_check_low_stock(self):
    """Check for low stock products and send alerts."""
    low_stock_products = self.env['product.template'].search([
        ('x_min_stock_qty', '>', 0),
        ('qty_available', '<', 'x_min_stock_qty'),  # This needs SQL
    ])

    # Actually filter in Python due to comparison limitation
    products = self.env['product.template'].search([
        ('x_min_stock_qty', '>', 0),
        ('type', '=', 'product'),
    ])

    low_stock = products.filtered(
        lambda p: p.qty_available < p.x_min_stock_qty
    )

    if low_stock:
        self._send_low_stock_alert(low_stock)
```

### Auto Reorder Cron
```python
@api.model
def _cron_auto_reorder(self):
    """Automatically create purchase orders for low stock items."""
    products = self.env['product.template'].search([
        ('x_min_stock_qty', '>', 0),
        ('type', '=', 'product'),
    ])

    to_reorder = products.filtered(
        lambda p: p.qty_available < p.x_min_stock_qty
    )

    for product in to_reorder:
        self._create_reorder(product)
```

---

## Best Practices

1. **Use correct locations** - Don't hardcode location IDs
2. **Handle reservations** - Check `free_qty` not just `qty_available`
3. **Multi-company aware** - Always filter by company
4. **Lot tracking** - Enable when traceability is needed
5. **Use picking types** - Match the operation type
6. **Validate before confirm** - Check availability first
7. **Handle backorders** - Decide on backorder policy
8. **Performance** - Use `read_group` for aggregations
9. **Concurrency** - Use proper locking for inventory updates
10. **Test thoroughly** - Stock operations have many edge cases

---


## Source: lot-serial-patterns.md

# Lot and Serial Number Patterns

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  LOT & SERIAL NUMBER PATTERNS                                                ║
║  Product traceability, batch tracking, and serial number management          ║
║  Use for inventory tracking, recalls, and compliance requirements            ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## Tracking Types

| Tracking | Description | Use Case |
|----------|-------------|----------|
| `none` | No tracking | Basic products |
| `lot` | Batch/Lot tracking | Multiple items per lot |
| `serial` | Serial number | One item per serial |

---

## Product Tracking Configuration

### Enable Tracking on Product
```python
class ProductTemplate(models.Model):
    _inherit = 'product.template'

    tracking = fields.Selection([
        ('serial', 'By Unique Serial Number'),
        ('lot', 'By Lots'),
        ('none', 'No Tracking'),
    ], string='Tracking', default='none', required=True)

# Create product with lot tracking
product = self.env['product.template'].create({
    'name': 'Pharmaceutical Product',
    'type': 'product',
    'tracking': 'lot',  # Batch tracking
})

# Create product with serial tracking
product_serial = self.env['product.template'].create({
    'name': 'Electronic Device',
    'type': 'product',
    'tracking': 'serial',  # Unique serial per unit
})
```

---

## Creating Lots and Serials

### Create Lot/Serial
```python
# Create a lot (batch)
lot = self.env['stock.lot'].create({
    'name': 'LOT-2024-001',
    'product_id': product.id,
    'company_id': self.env.company.id,
})

# With expiration date
lot_with_expiry = self.env['stock.lot'].create({
    'name': 'LOT-2024-002',
    'product_id': product.id,
    'company_id': self.env.company.id,
    'expiration_date': fields.Datetime.now() + timedelta(days=365),
    'use_date': fields.Datetime.now() + timedelta(days=300),
    'removal_date': fields.Datetime.now() + timedelta(days=350),
    'alert_date': fields.Datetime.now() + timedelta(days=280),
})
```

### Auto-Generate Serial Numbers
```python
def generate_serial_numbers(self, product, quantity, prefix='SN'):
    """Generate unique serial numbers."""
    serials = []
    for i in range(int(quantity)):
        serial = self.env['stock.lot'].create({
            'name': f"{prefix}-{fields.Date.today().strftime('%Y%m%d')}-{i+1:04d}",
            'product_id': product.id,
            'company_id': self.env.company.id,
        })
        serials.append(serial)
    return serials
```

---

## Stock Moves with Lots

### Create Move with Lot
```python
def create_move_with_lot(self, product, lot, qty, location_src, location_dest):
    """Create stock move with lot tracking."""
    move = self.env['stock.move'].create({
        'name': f'Move {product.name}',
        'product_id': product.id,
        'product_uom_qty': qty,
        'product_uom': product.uom_id.id,
        'location_id': location_src.id,
        'location_dest_id': location_dest.id,
    })

    # Create move line with lot
    self.env['stock.move.line'].create({
        'move_id': move.id,
        'product_id': product.id,
        'lot_id': lot.id,
        'quantity': qty,
        'product_uom_id': product.uom_id.id,
        'location_id': location_src.id,
        'location_dest_id': location_dest.id,
    })

    move._action_confirm()
    move._action_assign()
    move._action_done()

    return move
```

### Receive with New Lot
```python
def receive_with_lot(self, picking, product, qty, lot_name):
    """Receive products creating new lot."""
    # Find or create lot
    lot = self.env['stock.lot'].search([
        ('name', '=', lot_name),
        ('product_id', '=', product.id),
    ]) or self.env['stock.lot'].create({
        'name': lot_name,
        'product_id': product.id,
        'company_id': picking.company_id.id,
    })

    # Find the move for this product
    move = picking.move_ids.filtered(lambda m: m.product_id == product)

    # Set lot on move line
    move.move_line_ids.write({
        'lot_id': lot.id,
        'quantity': qty,
    })

    return lot
```

---

## Querying Lots

### Find Lots for Product
```python
def get_product_lots(self, product):
    """Get all lots for a product."""
    return self.env['stock.lot'].search([
        ('product_id', '=', product.id),
    ])

def get_available_lots(self, product, location=None):
    """Get lots with available stock."""
    domain = [('product_id', '=', product.id)]

    lots = self.env['stock.lot'].search(domain)

    available_lots = []
    for lot in lots:
        # Get quantity for this lot
        quants = self.env['stock.quant'].search([
            ('lot_id', '=', lot.id),
            ('location_id.usage', '=', 'internal'),
        ])
        if location:
            quants = quants.filtered(lambda q: q.location_id == location)

        qty = sum(quants.mapped('quantity'))
        if qty > 0:
            available_lots.append({
                'lot': lot,
                'quantity': qty,
                'expiration_date': lot.expiration_date,
            })

    return available_lots
```

### Get Stock by Lot
```python
def get_lot_stock(self, lot):
    """Get stock quantity for a specific lot."""
    quants = self.env['stock.quant'].search([
        ('lot_id', '=', lot.id),
        ('location_id.usage', '=', 'internal'),
    ])
    return sum(quants.mapped('quantity'))
```

---

## Expiration Date Tracking

### Product Expiration Fields
```python
class ProductTemplate(models.Model):
    _inherit = 'product.template'

    use_expiration_date = fields.Boolean(
        string='Expiration Date',
        help='Track expiration dates on lots/serials',
    )

    # Default durations (in days)
    expiration_time = fields.Integer(string='Expiration Time')
    use_time = fields.Integer(string='Best Before Time')
    removal_time = fields.Integer(string='Removal Time')
    alert_time = fields.Integer(string='Alert Time')
```

### Auto-Calculate Expiration Dates
```python
class StockLot(models.Model):
    _inherit = 'stock.lot'

    @api.model_create_multi
    def create(self, vals_list):
        """Auto-set expiration dates from product."""
        for vals in vals_list:
            if 'product_id' in vals:
                product = self.env['product.product'].browse(vals['product_id'])
                if product.use_expiration_date:
                    now = fields.Datetime.now()
                    if not vals.get('expiration_date') and product.expiration_time:
                        vals['expiration_date'] = now + timedelta(days=product.expiration_time)
                    if not vals.get('use_date') and product.use_time:
                        vals['use_date'] = now + timedelta(days=product.use_time)
                    if not vals.get('removal_date') and product.removal_time:
                        vals['removal_date'] = now + timedelta(days=product.removal_time)
                    if not vals.get('alert_date') and product.alert_time:
                        vals['alert_date'] = now + timedelta(days=product.alert_time)

        return super().create(vals_list)
```

### Check Expiring Lots
```python
def get_expiring_lots(self, days=30):
    """Find lots expiring within specified days."""
    deadline = fields.Datetime.now() + timedelta(days=days)
    return self.env['stock.lot'].search([
        ('expiration_date', '!=', False),
        ('expiration_date', '<=', deadline),
        ('expiration_date', '>', fields.Datetime.now()),
    ])

def get_expired_lots(self):
    """Find expired lots with stock."""
    expired_lots = self.env['stock.lot'].search([
        ('expiration_date', '<', fields.Datetime.now()),
    ])

    # Filter to only those with stock
    return expired_lots.filtered(
        lambda l: sum(l.quant_ids.filtered(
            lambda q: q.location_id.usage == 'internal'
        ).mapped('quantity')) > 0
    )
```

---

## FIFO/FEFO Removal Strategy

### Configure Removal Strategy
```python
# On location or category
location = self.env['stock.location'].browse(location_id)
location.write({
    'removal_strategy_id': self.env.ref('stock.removal_fifo').id,
})

# FEFO (First Expiry, First Out)
location.write({
    'removal_strategy_id': self.env.ref('stock.removal_fefo').id,
})
```

### Manual Lot Selection
```python
def select_lot_for_delivery(self, product, quantity, strategy='fifo'):
    """Select lots for delivery based on strategy."""
    lots = self.get_available_lots(product)

    if strategy == 'fefo':
        # Sort by expiration date (earliest first)
        lots = sorted(lots, key=lambda l: l['expiration_date'] or datetime.max)
    elif strategy == 'fifo':
        # Sort by lot creation date (oldest first)
        lots = sorted(lots, key=lambda l: l['lot'].create_date)

    selected = []
    remaining = quantity
    for lot_data in lots:
        if remaining <= 0:
            break
        take = min(lot_data['quantity'], remaining)
        selected.append({
            'lot_id': lot_data['lot'].id,
            'quantity': take,
        })
        remaining -= take

    return selected
```

---

## Serial Number Validation

### Unique Serial Check
```python
class StockLot(models.Model):
    _inherit = 'stock.lot'

    @api.constrains('name', 'product_id', 'company_id')
    def _check_unique_serial(self):
        """Ensure serial numbers are unique."""
        for lot in self:
            if lot.product_id.tracking == 'serial':
                duplicates = self.search([
                    ('id', '!=', lot.id),
                    ('name', '=', lot.name),
                    ('product_id', '=', lot.product_id.id),
                    ('company_id', '=', lot.company_id.id),
                ])
                if duplicates:
                    raise ValidationError(
                        f"Serial number {lot.name} already exists for this product."
                    )
```

### Serial Format Validation
```python
import re

@api.constrains('name')
def _check_serial_format(self):
    """Validate serial number format."""
    pattern = r'^[A-Z]{2}-\d{4}-\d{6}$'  # e.g., SN-2024-000001
    for lot in self:
        if lot.product_id.tracking == 'serial':
            if not re.match(pattern, lot.name):
                raise ValidationError(
                    f"Invalid serial format: {lot.name}. "
                    f"Expected format: XX-YYYY-NNNNNN"
                )
```

---

## Traceability Reports

### Get Lot History
```python
def get_lot_traceability(self, lot):
    """Get complete history of a lot."""
    moves = self.env['stock.move.line'].search([
        ('lot_id', '=', lot.id),
        ('state', '=', 'done'),
    ])

    history = []
    for move in moves.sorted('date'):
        history.append({
            'date': move.date,
            'reference': move.reference,
            'from_location': move.location_id.complete_name,
            'to_location': move.location_dest_id.complete_name,
            'quantity': move.quantity,
            'picking': move.picking_id.name if move.picking_id else None,
        })

    return history
```

### Upstream/Downstream Traceability
```python
def get_upstream_lots(self, lot):
    """Find source lots (for manufacturing)."""
    # Find production orders that consumed this lot
    consumed = self.env['stock.move.line'].search([
        ('lot_id', '=', lot.id),
        ('location_dest_id.usage', '=', 'production'),
    ])

    # Find resulting products
    productions = consumed.mapped('move_id.raw_material_production_id')
    return productions.mapped('lot_producing_id')

def get_downstream_lots(self, lot):
    """Find destination lots (for recalls)."""
    # Find where this lot was consumed in production
    consumed = self.env['stock.move.line'].search([
        ('lot_id', '=', lot.id),
        ('move_id.raw_material_production_id', '!=', False),
    ])

    return consumed.mapped('move_id.raw_material_production_id.lot_producing_id')
```

---

## Custom Fields on Lots

### Extend Lot Model
```python
class StockLot(models.Model):
    _inherit = 'stock.lot'

    supplier_lot = fields.Char(string='Supplier Lot Number')
    production_date = fields.Date(string='Production Date')
    certificate_ids = fields.Many2many(
        'ir.attachment',
        string='Quality Certificates',
    )
    notes = fields.Text(string='Notes')

    # Quality control
    qc_status = fields.Selection([
        ('pending', 'Pending QC'),
        ('passed', 'Passed'),
        ('failed', 'Failed'),
        ('quarantine', 'Quarantine'),
    ], string='QC Status', default='pending')
```

---

## XML Data for Lots

### Pre-define Lots
```xml
<record id="lot_sample_001" model="stock.lot">
    <field name="name">SAMPLE-LOT-001</field>
    <field name="product_id" ref="product_sample"/>
    <field name="company_id" ref="base.main_company"/>
</record>
```

---

## Best Practices

1. **Choose tracking wisely** - Serial for unique items, lot for batches
2. **Use expiration dates** - For perishables and regulated products
3. **Validate formats** - Consistent naming conventions
4. **FEFO for perishables** - First expiry, first out
5. **Track QC status** - Before releasing to inventory
6. **Maintain traceability** - For recalls and compliance
7. **Supplier lot mapping** - Link to vendor's batch numbers
8. **Automate numbering** - Use sequences for consistency
9. **Regular audits** - Check expired/quarantined stock
10. **Document certificates** - Attach quality documents to lots

---


## Source: uom-patterns.md

# Unit of Measure Patterns

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  UNIT OF MEASURE (UoM) PATTERNS                                              ║
║  Product quantities, conversions, and multi-UoM handling                     ║
║  Use for inventory, sales, and purchasing with different units               ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## UoM Basics

### UoM Category and Units
```
UoM Category: Weight
├── kg (reference unit, factor = 1.0)
├── g (factor = 0.001)
├── lb (factor = 0.453592)
└── oz (factor = 0.0283495)

UoM Category: Unit
├── Unit(s) (reference unit)
├── Dozen (factor = 12)
└── Hundred (factor = 100)
```

---

## Creating UoM Categories and Units

### Create UoM Category
```python
# Create a custom UoM category
category = self.env['uom.category'].create({
    'name': 'Volume',
})
```

### Create Units of Measure
```python
# Reference unit (factor = 1)
liter = self.env['uom.uom'].create({
    'name': 'Liter',
    'category_id': category.id,
    'uom_type': 'reference',  # This is the base unit
    'rounding': 0.001,
})

# Smaller unit
milliliter = self.env['uom.uom'].create({
    'name': 'Milliliter',
    'category_id': category.id,
    'uom_type': 'smaller',
    'factor_inv': 1000,  # 1000 ml = 1 L
    'rounding': 1,
})

# Bigger unit
gallon = self.env['uom.uom'].create({
    'name': 'Gallon',
    'category_id': category.id,
    'uom_type': 'bigger',
    'factor': 3.78541,  # 1 gallon = 3.78541 L
    'rounding': 0.01,
})
```

---

## UoM on Products

### Product UoM Fields
```python
class ProductTemplate(models.Model):
    _inherit = 'product.template'

    # Default unit of measure (sales/inventory)
    uom_id = fields.Many2one(
        'uom.uom',
        string='Unit of Measure',
        required=True,
    )

    # Purchase unit of measure
    uom_po_id = fields.Many2one(
        'uom.uom',
        string='Purchase UoM',
        required=True,
    )

    # These must be in the same category for conversion
```

### Set Product UoM
```python
product = self.env['product.template'].create({
    'name': 'Cooking Oil',
    'type': 'product',
    'uom_id': liter.id,  # Sell in liters
    'uom_po_id': gallon.id,  # Buy in gallons
})
```

---

## UoM Conversion

### Convert Quantities
```python
def convert_uom(self, qty, from_uom, to_uom):
    """Convert quantity between units."""
    if from_uom.category_id != to_uom.category_id:
        raise UserError("Cannot convert between different UoM categories.")

    return from_uom._compute_quantity(qty, to_uom)

# Example: Convert 5 gallons to liters
liters = convert_uom(5, gallon, liter)  # = 18.927
```

### Compute Quantity Method
```python
# Built-in conversion
from_uom = self.env.ref('uom.product_uom_dozen')
to_uom = self.env.ref('uom.product_uom_unit')

# Convert 2 dozen to units
units = from_uom._compute_quantity(2, to_uom)  # = 24

# With rounding
units_rounded = from_uom._compute_quantity(
    2.5,
    to_uom,
    round=True,
    rounding_method='UP',
)
```

### Price Conversion
```python
def convert_price(self, price, from_uom, to_uom):
    """Convert unit price between UoMs."""
    return from_uom._compute_price(price, to_uom)

# If product costs $10 per gallon, what's the price per liter?
price_per_liter = convert_price(10, gallon, liter)  # ≈ $2.64
```

---

## UoM in Sales and Purchases

### Sale Order Line with UoM
```python
class SaleOrderLine(models.Model):
    _inherit = 'sale.order.line'

    product_uom = fields.Many2one(
        'uom.uom',
        string='Unit of Measure',
        domain="[('category_id', '=', product_uom_category_id)]",
    )

    product_uom_category_id = fields.Many2one(
        related='product_id.uom_id.category_id',
    )

    @api.onchange('product_id')
    def _onchange_product_id(self):
        """Set default UoM from product."""
        if self.product_id:
            self.product_uom = self.product_id.uom_id

    @api.onchange('product_uom')
    def _onchange_product_uom(self):
        """Recalculate price when UoM changes."""
        if self.product_id and self.product_uom:
            # Get price in product's UoM
            base_price = self.product_id.lst_price
            # Convert to selected UoM
            self.price_unit = self.product_id.uom_id._compute_price(
                base_price, self.product_uom
            )
```

### Purchase Order Line
```python
class PurchaseOrderLine(models.Model):
    _inherit = 'purchase.order.line'

    @api.onchange('product_id')
    def onchange_product_id(self):
        """Default to purchase UoM."""
        result = super().onchange_product_id()
        if self.product_id:
            self.product_uom = self.product_id.uom_po_id
        return result
```

---

## Stock Moves with UoM

### Create Stock Move
```python
def create_stock_move(self, product, qty, uom, location_src, location_dest):
    """Create stock move with UoM conversion."""
    # Quantity in product's UoM for stock
    product_qty = uom._compute_quantity(qty, product.uom_id)

    move = self.env['stock.move'].create({
        'name': product.name,
        'product_id': product.id,
        'product_uom_qty': qty,  # In move's UoM
        'product_uom': uom.id,
        'location_id': location_src.id,
        'location_dest_id': location_dest.id,
    })

    return move
```

### Check Stock in Different UoM
```python
def get_stock_in_uom(self, product, uom, location=None):
    """Get available stock converted to specific UoM."""
    if location:
        qty = product.with_context(location=location.id).qty_available
    else:
        qty = product.qty_available

    # Convert from product UoM to requested UoM
    return product.uom_id._compute_quantity(qty, uom)

# Get stock in dozens
stock_dozens = get_stock_in_uom(product, dozen_uom)
```

---

## UoM Rounding

### Rounding Options
```python
# Rounding precision on UoM
uom = self.env['uom.uom'].create({
    'name': 'Piece',
    'category_id': category.id,
    'uom_type': 'reference',
    'rounding': 1.0,  # Round to whole numbers
})

# Rounding methods
qty = from_uom._compute_quantity(
    10.3,
    to_uom,
    round=True,
    rounding_method='HALF-UP',  # Default, standard rounding
)

qty = from_uom._compute_quantity(
    10.3,
    to_uom,
    round=True,
    rounding_method='UP',  # Always round up
)
```

### Float Comparison
```python
from odoo.tools import float_compare, float_round

# Compare quantities with UoM precision
result = float_compare(
    qty1,
    qty2,
    precision_rounding=uom.rounding,
)
# Returns: -1 (less), 0 (equal), 1 (greater)

# Round to UoM precision
rounded_qty = float_round(qty, precision_rounding=uom.rounding)
```

---

## Custom Model with UoM

### Model with Quantity Field
```python
class InventoryAdjustment(models.Model):
    _name = 'inventory.adjustment'

    product_id = fields.Many2one('product.product', required=True)
    quantity = fields.Float(string='Quantity', digits='Product Unit of Measure')
    product_uom_id = fields.Many2one(
        'uom.uom',
        string='Unit of Measure',
        domain="[('category_id', '=', product_uom_category_id)]",
    )
    product_uom_category_id = fields.Many2one(
        related='product_id.uom_id.category_id',
        string='UoM Category',
    )

    # Quantity in product's base UoM
    product_qty = fields.Float(
        string='Quantity (Base UoM)',
        compute='_compute_product_qty',
        store=True,
    )

    @api.depends('quantity', 'product_uom_id', 'product_id')
    def _compute_product_qty(self):
        for record in self:
            if record.product_uom_id and record.product_id:
                record.product_qty = record.product_uom_id._compute_quantity(
                    record.quantity,
                    record.product_id.uom_id,
                )
            else:
                record.product_qty = record.quantity

    @api.onchange('product_id')
    def _onchange_product_id(self):
        if self.product_id:
            self.product_uom_id = self.product_id.uom_id
```

---

## UoM View Patterns

### Form View with UoM
```xml
<form>
    <group>
        <field name="product_id"/>
        <label for="quantity"/>
        <div class="o_row">
            <field name="quantity" class="oe_inline"/>
            <field name="product_uom_id" class="oe_inline"
                   options="{'no_create': True}"
                   groups="uom.group_uom"/>
        </div>
        <field name="product_uom_category_id" invisible="1"/>
    </group>
</form>
```

### List View
```xml
<tree>
    <field name="product_id"/>
    <field name="quantity"/>
    <field name="product_uom_id" groups="uom.group_uom"/>
    <field name="product_qty" string="Qty (Base)"/>
</tree>
```

---

## XML Data for UoM

### Define UoM in Data
```xml
<!-- UoM Category -->
<record id="uom_categ_length" model="uom.category">
    <field name="name">Length</field>
</record>

<!-- Reference unit -->
<record id="uom_meter" model="uom.uom">
    <field name="name">Meter</field>
    <field name="category_id" ref="uom_categ_length"/>
    <field name="uom_type">reference</field>
    <field name="rounding">0.01</field>
</record>

<!-- Smaller unit -->
<record id="uom_centimeter" model="uom.uom">
    <field name="name">Centimeter</field>
    <field name="category_id" ref="uom_categ_length"/>
    <field name="uom_type">smaller</field>
    <field name="factor_inv">100</field>
    <field name="rounding">1</field>
</record>

<!-- Bigger unit -->
<record id="uom_kilometer" model="uom.uom">
    <field name="name">Kilometer</field>
    <field name="category_id" ref="uom_categ_length"/>
    <field name="uom_type">bigger</field>
    <field name="factor">1000</field>
    <field name="rounding">0.001</field>
</record>
```

---

## Common UoM References

### Standard Odoo UoMs
```python
# Unit category
unit = self.env.ref('uom.product_uom_unit')
dozen = self.env.ref('uom.product_uom_dozen')

# Weight category
kg = self.env.ref('uom.product_uom_kgm')
gram = self.env.ref('uom.product_uom_gram')
lb = self.env.ref('uom.product_uom_lb')
oz = self.env.ref('uom.product_uom_oz')

# Time category
hour = self.env.ref('uom.product_uom_hour')
day = self.env.ref('uom.product_uom_day')

# Volume category (if installed)
litre = self.env.ref('uom.product_uom_litre')
```

---

## UoM Feature Group

### Enable Multi-UoM
```python
# Users need this group to see UoM fields
# Settings > Users > Technical Settings > Multiple Units of Measure

# In view, use groups attribute
# groups="uom.group_uom"
```

---

## Best Practices

1. **Same category** - Only convert within same UoM category
2. **Set rounding** - Appropriate precision for each unit
3. **Use product UoM** - For stock quantities, always convert to product's UoM
4. **Separate sale/purchase** - Different UoMs for buying vs selling
5. **Handle precision** - Use float_compare for equality checks
6. **Group visibility** - Hide UoM fields unless group_uom enabled
7. **Default UoM** - Always set sensible default
8. **Test conversions** - Verify conversion factors are correct
9. **Document units** - Clear names and descriptions
10. **Avoid mixing** - Don't mix UoMs from different categories

---

