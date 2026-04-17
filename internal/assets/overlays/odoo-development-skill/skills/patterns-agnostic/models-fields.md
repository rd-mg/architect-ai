# Models Fields Patterns

Consolidated from the following source files:
- `field-type-reference.md`
- `computed-field-patterns.md`
- `constraint-patterns.md`
- `onchange-dynamic-patterns.md`
- `inheritance-patterns.md`
- `domain-filter-patterns.md`
- `workflow-state-patterns.md`
- `wizard-patterns.md`

---


## Source: field-type-reference.md

# Odoo Field Type Reference

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  FIELD TYPE REFERENCE                                                        ║
║  Complete reference for all Odoo field types with version-specific notes     ║
║  Use when defining model fields                                              ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## Field Types Overview

| Type | Python Class | Storage | Use Case |
|------|-------------|---------|----------|
| Char | `fields.Char` | VARCHAR | Short text, names |
| Text | `fields.Text` | TEXT | Long text, descriptions |
| Html | `fields.Html` | TEXT | Rich formatted text |
| Integer | `fields.Integer` | INTEGER | Whole numbers |
| Float | `fields.Float` | FLOAT | Decimal numbers |
| Monetary | `fields.Monetary` | NUMERIC | Currency amounts |
| Boolean | `fields.Boolean` | BOOLEAN | True/False |
| Date | `fields.Date` | DATE | Dates without time |
| Datetime | `fields.Datetime` | TIMESTAMP | Dates with time |
| Selection | `fields.Selection` | VARCHAR | Choice from list |
| Binary | `fields.Binary` | BYTEA | Files, images |
| Many2one | `fields.Many2one` | INTEGER (FK) | Single relation |
| One2many | `fields.One2many` | Virtual | Reverse of Many2one |
| Many2many | `fields.Many2many` | Junction table | Multiple relations |

---

## String Fields

### Char
```python
# Basic
name = fields.Char(string='Name')

# With constraints
name = fields.Char(
    string='Name',
    required=True,
    size=64,                    # Max length (rarely used)
    trim=True,                  # Strip whitespace (default True)
    translate=True,             # Enable translations
)

# With tracking (v15+)
name = fields.Char(
    string='Name',
    required=True,
    tracking=True,              # Track in chatter
    index=True,                 # Create index
)

# With index types (v16+)
code = fields.Char(
    string='Code',
    index='btree_not_null',     # Exclude NULL values from index
)
search_name = fields.Char(
    string='Search Name',
    index='trigram',            # For ILIKE searches
)
```

### Text
```python
description = fields.Text(
    string='Description',
    translate=True,
    tracking=True,
)
```

### Html
```python
content = fields.Html(
    string='Content',
    sanitize=True,              # Clean HTML (default True)
    sanitize_tags=True,         # Remove unsafe tags
    sanitize_attributes=True,   # Remove unsafe attributes
    sanitize_style=True,        # Clean style attributes
    strip_style=False,          # Remove all styles
    strip_classes=False,        # Remove all classes
)
```

---

## Numeric Fields

### Integer
```python
sequence = fields.Integer(
    string='Sequence',
    default=10,
    index=True,
)

count = fields.Integer(
    string='Count',
    compute='_compute_count',
    store=True,
)
```

### Float
```python
# Basic
quantity = fields.Float(string='Quantity')

# With precision
quantity = fields.Float(
    string='Quantity',
    digits='Product Unit of Measure',  # Named precision
)

# With explicit precision
amount = fields.Float(
    string='Amount',
    digits=(16, 2),                    # (total, decimal)
)

# Computed with aggregation
total = fields.Float(
    string='Total',
    compute='_compute_total',
    store=True,
    group_operator='sum',              # For group_by aggregation
)
```

### Monetary
```python
# Requires currency field
currency_id = fields.Many2one(
    comodel_name='res.currency',
    string='Currency',
    default=lambda self: self.env.company.currency_id,
)

amount = fields.Monetary(
    string='Amount',
    currency_field='currency_id',      # REQUIRED: link to currency
)

# Related currency (common pattern)
currency_id = fields.Many2one(
    comodel_name='res.currency',
    related='company_id.currency_id',
    store=True,
)
```

---

## Boolean Fields

```python
active = fields.Boolean(
    string='Active',
    default=True,
)

is_done = fields.Boolean(
    string='Done',
    compute='_compute_is_done',
    store=True,
)

# Copy behavior
copy_this = fields.Boolean(default=True)              # Copied by default
dont_copy = fields.Boolean(default=True, copy=False)  # Not copied
```

---

## Date/Time Fields

### Date
```python
date = fields.Date(
    string='Date',
    default=fields.Date.today,         # Today as default
)

date = fields.Date(
    string='Date',
    default=fields.Date.context_today, # Today in user's timezone
)

# Computed date
deadline = fields.Date(
    string='Deadline',
    compute='_compute_deadline',
    store=True,
    index=True,
)
```

### Datetime
```python
datetime = fields.Datetime(
    string='Date Time',
    default=fields.Datetime.now,       # Current datetime
)

# With copy behavior
create_datetime = fields.Datetime(
    string='Created',
    default=fields.Datetime.now,
    copy=False,
)
```

---

## Selection Fields

### Basic Selection
```python
state = fields.Selection(
    selection=[
        ('draft', 'Draft'),
        ('confirmed', 'Confirmed'),
        ('done', 'Done'),
        ('cancelled', 'Cancelled'),
    ],
    string='Status',
    default='draft',
    required=True,
    tracking=True,
)
```

### Extending Selection (inheritance)
```python
# In inherited model
state = fields.Selection(
    selection_add=[
        ('approved', 'Approved'),       # Add new option
        ('rejected', 'Rejected'),
    ],
    ondelete={                          # Handle deletion (v14+)
        'approved': 'set default',
        'rejected': 'cascade',
    },
)
```

### Dynamic Selection
```python
type = fields.Selection(
    selection='_get_type_selection',
    string='Type',
)

@api.model
def _get_type_selection(self):
    return [
        ('type1', 'Type 1'),
        ('type2', 'Type 2'),
    ]
```

---

## Binary Fields

```python
# File attachment
document = fields.Binary(
    string='Document',
    attachment=True,                   # Store as attachment (recommended)
)
document_name = fields.Char(string='File Name')

# Image with auto-resize
image = fields.Image(
    string='Image',
    max_width=1920,
    max_height=1920,
)

# Image variants (automatic)
image_128 = fields.Image(
    string='Image 128',
    related='image',
    max_width=128,
    max_height=128,
    store=True,
)
```

---

## Relational Fields

### Many2one
```python
# Basic
partner_id = fields.Many2one(
    comodel_name='res.partner',
    string='Partner',
)

# With constraints
partner_id = fields.Many2one(
    comodel_name='res.partner',
    string='Partner',
    required=True,
    ondelete='cascade',                # cascade, set null, restrict
    index=True,
    tracking=True,
)

# With domain
partner_id = fields.Many2one(
    comodel_name='res.partner',
    string='Customer',
    domain=[('customer_rank', '>', 0)],
)

# Dynamic domain
partner_id = fields.Many2one(
    comodel_name='res.partner',
    string='Partner',
    domain="[('company_id', '=', company_id)]",
)

# v18+ Multi-company
partner_id = fields.Many2one(
    comodel_name='res.partner',
    string='Partner',
    check_company=True,                # REQUIRED in v18+
)
```

### One2many
```python
# Basic (REQUIRES inverse_name)
line_ids = fields.One2many(
    comodel_name='my.model.line',
    inverse_name='model_id',           # REQUIRED
    string='Lines',
)

# With copy behavior
line_ids = fields.One2many(
    comodel_name='my.model.line',
    inverse_name='model_id',
    string='Lines',
    copy=True,                         # Copy lines when record copied
)

# With domain
active_line_ids = fields.One2many(
    comodel_name='my.model.line',
    inverse_name='model_id',
    string='Active Lines',
    domain=[('active', '=', True)],
)
```

### Many2many
```python
# Basic (auto table name)
tag_ids = fields.Many2many(
    comodel_name='my.model.tag',
    string='Tags',
)

# With explicit relation table
tag_ids = fields.Many2many(
    comodel_name='my.model.tag',
    relation='my_model_tag_rel',       # Junction table name
    column1='model_id',                # This model's column
    column2='tag_id',                  # Related model's column
    string='Tags',
)

# With domain
active_tag_ids = fields.Many2many(
    comodel_name='my.model.tag',
    string='Active Tags',
    domain=[('active', '=', True)],
)
```

---

## Computed Fields

### Basic Computed
```python
full_name = fields.Char(
    string='Full Name',
    compute='_compute_full_name',
)

@api.depends('first_name', 'last_name')
def _compute_full_name(self):
    for record in self:
        record.full_name = f"{record.first_name or ''} {record.last_name or ''}".strip()
```

### Stored Computed
```python
total = fields.Float(
    string='Total',
    compute='_compute_total',
    store=True,                        # Save to database
)

@api.depends('line_ids.amount')
def _compute_total(self):
    for record in self:
        record.total = sum(record.line_ids.mapped('amount'))
```

### Inverse (Editable Computed)
```python
full_name = fields.Char(
    string='Full Name',
    compute='_compute_full_name',
    inverse='_inverse_full_name',      # Makes field editable
    store=True,
)

def _inverse_full_name(self):
    for record in self:
        if record.full_name:
            parts = record.full_name.split(' ', 1)
            record.first_name = parts[0]
            record.last_name = parts[1] if len(parts) > 1 else ''
```

### Search on Computed
```python
total = fields.Float(
    string='Total',
    compute='_compute_total',
    search='_search_total',            # Enable searching
)

def _search_total(self, operator, value):
    # Return domain that filters records
    if operator == '>':
        ids = self.search([]).filtered(lambda r: r.total > value).ids
        return [('id', 'in', ids)]
    return []
```

---

## Related Fields

```python
# Simple related
partner_name = fields.Char(
    string='Partner Name',
    related='partner_id.name',
)

# Stored related (for performance/search)
partner_name = fields.Char(
    string='Partner Name',
    related='partner_id.name',
    store=True,
    index=True,
)

# Readonly related (can't edit through this field)
partner_email = fields.Char(
    string='Partner Email',
    related='partner_id.email',
    readonly=True,
)
```

---

## Common Field Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `string` | str | Field label |
| `help` | str | Tooltip text |
| `required` | bool | Cannot be empty |
| `readonly` | bool | Cannot be edited in UI |
| `index` | bool/str | Create database index |
| `default` | value/callable | Default value |
| `copy` | bool | Copy when duplicating |
| `groups` | str | Access groups (comma separated) |
| `tracking` | bool | Track in chatter (v15+) |
| `store` | bool | Store computed field |
| `compute` | str | Compute method name |
| `depends` | str | Dependency fields |
| `inverse` | str | Inverse method name |

---

## Version-Specific Notes

### v14
```python
# Use track_visibility (deprecated in v15)
name = fields.Char(track_visibility='onchange')
```

### v15+
```python
# Use tracking
name = fields.Char(tracking=True)
```

### v16+
```python
# Index types
code = fields.Char(index='btree_not_null')
name = fields.Char(index='trigram')
```

### v18+
```python
# Multi-company fields
partner_id = fields.Many2one('res.partner', check_company=True)

# Type hints on model
class MyModel(models.Model):
    name: str = fields.Char(required=True)
    amount: float = fields.Float()
```

### v19+
```python
# Type hints required
class MyModel(models.Model):
    _name = 'my.model'

    name: str = fields.Char(required=True)
    active: bool = fields.Boolean(default=True)
    amount: float = fields.Monetary(currency_field='currency_id')
```

---

## Field Naming Conventions

| Suffix | Field Type | Example |
|--------|-----------|---------|
| `_id` | Many2one | `partner_id`, `company_id` |
| `_ids` | One2many, Many2many | `line_ids`, `tag_ids` |
| `_count` | Integer (computed) | `order_count`, `task_count` |
| `_date` | Date | `create_date`, `due_date` |
| `_datetime` | Datetime | `start_datetime` |
| `is_` | Boolean | `is_done`, `is_locked` |
| `has_` | Boolean | `has_children`, `has_invoice` |
| `x_` | Custom field | `x_custom_field` (for studio/inheritance) |

---

## Security-Sensitive Fields

```python
# Hide from non-admin
salary = fields.Monetary(
    string='Salary',
    groups='hr.group_hr_manager',      # Only HR managers
)

# Multiple groups
secret_field = fields.Char(
    string='Secret',
    groups='base.group_system,hr.group_hr_manager',
)
```

---


## Source: computed-field-patterns.md

# Computed Field Patterns

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  COMPUTED FIELD PATTERNS                                                     ║
║  @api.depends, compute methods, inverse, and search                          ║
║  Use for derived values, aggregations, and dynamic data                      ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## Basic Computed Fields

### Simple Computation
```python
from odoo import api, fields, models


class MyModel(models.Model):
    _name = 'my.model'

    first_name = fields.Char()
    last_name = fields.Char()

    # Basic computed field
    full_name = fields.Char(
        string='Full Name',
        compute='_compute_full_name',
    )

    @api.depends('first_name', 'last_name')
    def _compute_full_name(self):
        for record in self:
            parts = filter(None, [record.first_name, record.last_name])
            record.full_name = ' '.join(parts)
```

### Stored Computed Field
```python
class MyModel(models.Model):
    _name = 'my.model'

    quantity = fields.Float()
    price = fields.Float()

    # Stored - saved to database, recomputed on dependency change
    subtotal = fields.Float(
        string='Subtotal',
        compute='_compute_subtotal',
        store=True,
    )

    @api.depends('quantity', 'price')
    def _compute_subtotal(self):
        for record in self:
            record.subtotal = record.quantity * record.price
```

### Readonly vs Editable
```python
class MyModel(models.Model):
    _name = 'my.model'

    # Non-stored are always readonly
    calculated_value = fields.Float(compute='_compute_value')

    # Stored computed can be readonly (default) or editable
    total = fields.Float(
        compute='_compute_total',
        store=True,
        readonly=True,  # Default
    )

    # Editable stored computed (rare)
    adjustable_total = fields.Float(
        compute='_compute_adjustable_total',
        store=True,
        readonly=False,
    )
```

---

## Dependency Patterns

### Field Dependencies
```python
@api.depends('field1', 'field2', 'field3')
def _compute_value(self):
    for record in self:
        record.value = record.field1 + record.field2 + record.field3
```

### Related Field Dependencies
```python
@api.depends('partner_id.name', 'partner_id.email')
def _compute_partner_info(self):
    for record in self:
        record.partner_info = f"{record.partner_id.name} <{record.partner_id.email}>"
```

### One2many/Many2many Dependencies
```python
@api.depends('line_ids.amount', 'line_ids.quantity')
def _compute_total(self):
    for record in self:
        record.total = sum(
            line.amount * line.quantity
            for line in record.line_ids
        )
```

### Deep Dependencies
```python
@api.depends('order_id.partner_id.country_id.code')
def _compute_country_code(self):
    for record in self:
        record.country_code = record.order_id.partner_id.country_id.code or ''
```

### No Dependencies (Always Recompute)
```python
# Use for values that depend on external factors
@api.depends()
def _compute_current_date(self):
    for record in self:
        record.current_date = fields.Date.today()
```

---

## Inverse Methods

### Editable Computed Field
```python
class MyModel(models.Model):
    _name = 'my.model'

    unit_price = fields.Float()
    quantity = fields.Float(default=1.0)

    # Editable computed field with inverse
    total_price = fields.Float(
        string='Total Price',
        compute='_compute_total_price',
        inverse='_inverse_total_price',
    )

    @api.depends('unit_price', 'quantity')
    def _compute_total_price(self):
        for record in self:
            record.total_price = record.unit_price * record.quantity

    def _inverse_total_price(self):
        """When user edits total, recalculate unit price."""
        for record in self:
            if record.quantity:
                record.unit_price = record.total_price / record.quantity
```

### Inverse with Related
```python
class MyModel(models.Model):
    _name = 'my.model'

    partner_id = fields.Many2one('res.partner')

    # Editable related-like field
    partner_email = fields.Char(
        string='Partner Email',
        compute='_compute_partner_email',
        inverse='_inverse_partner_email',
    )

    @api.depends('partner_id.email')
    def _compute_partner_email(self):
        for record in self:
            record.partner_email = record.partner_id.email

    def _inverse_partner_email(self):
        for record in self:
            if record.partner_id:
                record.partner_id.email = record.partner_email
```

---

## Search Methods

### Custom Search Implementation
```python
class MyModel(models.Model):
    _name = 'my.model'

    amount = fields.Float()
    tax_rate = fields.Float(default=0.21)

    total_with_tax = fields.Float(
        string='Total with Tax',
        compute='_compute_total_with_tax',
        search='_search_total_with_tax',
    )

    @api.depends('amount', 'tax_rate')
    def _compute_total_with_tax(self):
        for record in self:
            record.total_with_tax = record.amount * (1 + record.tax_rate)

    def _search_total_with_tax(self, operator, value):
        """Enable searching on computed field."""
        # Convert the search to base fields
        if operator in ('=', '!=', '<', '<=', '>', '>='):
            # Search for records where amount * (1 + tax_rate) matches
            # Simplified: assume default tax_rate
            adjusted_value = value / 1.21
            return [('amount', operator, adjusted_value)]
        return []
```

### Search with SQL
```python
def _search_full_name(self, operator, value):
    """Search on concatenated fields."""
    if operator == 'ilike':
        return [
            '|',
            ('first_name', 'ilike', value),
            ('last_name', 'ilike', value),
        ]
    elif operator == '=':
        # Exact match on full name
        self.env.cr.execute("""
            SELECT id FROM my_model
            WHERE CONCAT(first_name, ' ', last_name) = %s
        """, (value,))
        ids = [r[0] for r in self.env.cr.fetchall()]
        return [('id', 'in', ids)]
    return []
```

---

## Common Computation Patterns

### Aggregations
```python
# Sum
@api.depends('line_ids.amount')
def _compute_total_amount(self):
    for record in self:
        record.total_amount = sum(record.line_ids.mapped('amount'))

# Count
@api.depends('line_ids')
def _compute_line_count(self):
    for record in self:
        record.line_count = len(record.line_ids)

# Average
@api.depends('line_ids.score')
def _compute_average_score(self):
    for record in self:
        scores = record.line_ids.mapped('score')
        record.average_score = sum(scores) / len(scores) if scores else 0
```

### Status/State Computation
```python
@api.depends('line_ids.state')
def _compute_state(self):
    for record in self:
        if not record.line_ids:
            record.state = 'draft'
        elif all(line.state == 'done' for line in record.line_ids):
            record.state = 'done'
        elif any(line.state == 'in_progress' for line in record.line_ids):
            record.state = 'in_progress'
        else:
            record.state = 'pending'
```

### Boolean Checks
```python
@api.depends('amount_total', 'amount_paid')
def _compute_is_paid(self):
    for record in self:
        record.is_paid = record.amount_paid >= record.amount_total

@api.depends('date_deadline')
def _compute_is_overdue(self):
    today = fields.Date.today()
    for record in self:
        record.is_overdue = (
            record.date_deadline and
            record.date_deadline < today and
            record.state != 'done'
        )
```

### Date Computations
```python
from datetime import timedelta
from dateutil.relativedelta import relativedelta

@api.depends('start_date', 'duration_days')
def _compute_end_date(self):
    for record in self:
        if record.start_date and record.duration_days:
            record.end_date = record.start_date + timedelta(days=record.duration_days)
        else:
            record.end_date = False

@api.depends('birth_date')
def _compute_age(self):
    today = fields.Date.today()
    for record in self:
        if record.birth_date:
            delta = relativedelta(today, record.birth_date)
            record.age = delta.years
        else:
            record.age = 0
```

### Related Record Properties
```python
@api.depends('partner_id')
def _compute_partner_details(self):
    for record in self:
        partner = record.partner_id
        record.partner_phone = partner.phone or ''
        record.partner_email = partner.email or ''
        record.partner_country = partner.country_id.name or ''
```

### Currency Conversion
```python
@api.depends('amount', 'currency_id', 'company_id')
def _compute_amount_company_currency(self):
    for record in self:
        if record.currency_id != record.company_id.currency_id:
            record.amount_company = record.currency_id._convert(
                record.amount,
                record.company_id.currency_id,
                record.company_id,
                record.date or fields.Date.today(),
            )
        else:
            record.amount_company = record.amount
```

---

## Performance Optimization

### Batch Computation
```python
@api.depends('partner_id')
def _compute_partner_order_count(self):
    """Optimized: single query for all records."""
    if not self:
        return

    # Single query for all partners
    data = self.env['sale.order'].read_group(
        [('partner_id', 'in', self.mapped('partner_id').ids)],
        ['partner_id'],
        ['partner_id'],
    )
    counts = {d['partner_id'][0]: d['partner_id_count'] for d in data}

    for record in self:
        record.partner_order_count = counts.get(record.partner_id.id, 0)
```

### Prefetching
```python
@api.depends('line_ids.product_id.categ_id')
def _compute_categories(self):
    # Prefetch all products at once
    self.mapped('line_ids.product_id')

    for record in self:
        categories = record.line_ids.mapped('product_id.categ_id')
        record.category_names = ', '.join(categories.mapped('name'))
```

### Conditional Computation
```python
@api.depends('state')
def _compute_expensive_value(self):
    """Only compute for records that need it."""
    for record in self:
        if record.state in ('draft', 'cancel'):
            record.expensive_value = 0
        else:
            record.expensive_value = record._calculate_expensive()
```

---

## Related Fields (Special Computed)

### Using Related
```python
class MyModel(models.Model):
    _name = 'my.model'

    partner_id = fields.Many2one('res.partner')

    # Shorthand for computed field following relation
    partner_name = fields.Char(
        related='partner_id.name',
        string='Partner Name',
    )

    partner_country_id = fields.Many2one(
        related='partner_id.country_id',
        string='Partner Country',
    )

    # Stored related (denormalization)
    partner_email = fields.Char(
        related='partner_id.email',
        store=True,
    )
```

### Related vs Computed
```python
# Use related for simple field access
partner_phone = fields.Char(related='partner_id.phone')

# Use computed for transformations
partner_phone_formatted = fields.Char(compute='_compute_phone_formatted')

@api.depends('partner_id.phone')
def _compute_phone_formatted(self):
    for record in self:
        phone = record.partner_id.phone or ''
        record.partner_phone_formatted = f"+{phone}" if phone else ''
```

---

## Best Practices

1. **Always iterate over self** - Even for single records
2. **List all dependencies** - Include all fields that affect the result
3. **Use store=True wisely** - Only when needed for search/sort
4. **Optimize batch operations** - Use read_group, mapped()
5. **Handle empty values** - Check for False/None
6. **Avoid circular dependencies** - Field A depends on B, B depends on A
7. **Use related for simple cases** - Cleaner than custom compute
8. **Add search for non-stored** - If users need to filter
9. **Test with multiple records** - Ensure batch processing works
10. **Document complex logic** - Explain business rules

---

## Common Mistakes

```python
# Bad - Not iterating
@api.depends('amount')
def _compute_total(self):
    self.total = self.amount * 2  # Wrong!

# Good
@api.depends('amount')
def _compute_total(self):
    for record in self:
        record.total = record.amount * 2

# Bad - Missing dependency
@api.depends('quantity')  # Missing 'price'!
def _compute_subtotal(self):
    for record in self:
        record.subtotal = record.quantity * record.price

# Good
@api.depends('quantity', 'price')
def _compute_subtotal(self):
    for record in self:
        record.subtotal = record.quantity * record.price
```

---


## Source: constraint-patterns.md

# Constraint Patterns

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  CONSTRAINT PATTERNS                                                         ║
║  SQL constraints, Python constraints, and data validation                    ║
║  Use for data integrity, validation rules, and business logic enforcement    ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## Constraint Types

| Type | When Checked | Use Case |
|------|--------------|----------|
| SQL | Database level | Uniqueness, basic checks |
| Python | ORM level | Complex business rules |

---

## SQL Constraints

### Odoo 19: models.Constraint() Class (REQUIRED)
```python
from odoo import models, fields


class MyModel(models.Model):
    _name = 'my.model'
    _description = 'My Model'

    name = fields.Char(required=True)
    code = fields.Char()
    amount = fields.Float()
    percentage = fields.Float()
    date_start = fields.Date()
    date_end = fields.Date()
    company_id = fields.Many2one('res.company')

    # Unique constraint
    _code_unique = models.Constraint(
        'UNIQUE(code)',
        'Code must be unique.',
    )

    # Unique per company
    _code_company_unique = models.Constraint(
        'UNIQUE(code, company_id)',
        'Code must be unique per company.',
    )

    # Check constraint
    _amount_positive = models.Constraint(
        'CHECK(amount >= 0)',
        'Amount must be positive.',
    )

    # Range constraint
    _percentage_range = models.Constraint(
        'CHECK(percentage >= 0 AND percentage <= 100)',
        'Percentage must be between 0 and 100.',
    )

    # Date constraint
    _dates_check = models.Constraint(
        'CHECK(date_end >= date_start)',
        'End date must be after start date.',
    )

    # Not null with condition
    _code_required_if_active = models.Constraint(
        'CHECK(active = false OR code IS NOT NULL)',
        'Code is required for active records.',
    )
```

### Odoo 18 and Earlier: _sql_constraints List (DEPRECATED in v19)
```python
from odoo import models, fields


class MyModel(models.Model):
    _name = 'my.model'
    _description = 'My Model'

    name = fields.Char(required=True)
    code = fields.Char()
    amount = fields.Float()
    percentage = fields.Float()
    date_start = fields.Date()
    date_end = fields.Date()
    company_id = fields.Many2one('res.company')

    _sql_constraints = [
        # Unique constraint
        ('code_unique', 'UNIQUE(code)',
         'Code must be unique.'),

        # Unique per company
        ('code_company_unique', 'UNIQUE(code, company_id)',
         'Code must be unique per company.'),

        # Check constraint
        ('amount_positive', 'CHECK(amount >= 0)',
         'Amount must be positive.'),

        # Range constraint
        ('percentage_range', 'CHECK(percentage >= 0 AND percentage <= 100)',
         'Percentage must be between 0 and 100.'),

        # Date constraint
        ('dates_check', 'CHECK(date_end >= date_start)',
         'End date must be after start date.'),

        # Not null with condition
        ('code_required_if_active',
         'CHECK(active = false OR code IS NOT NULL)',
         'Code is required for active records.'),
    ]
```

### Common SQL Constraint Patterns (Odoo 19 Syntax)

#### Uniqueness
```python
# Simple unique
_name_unique = models.Constraint(
    'UNIQUE(name)',
    'Name must be unique.',
)

# Unique combination
_name_date_unique = models.Constraint(
    'UNIQUE(name, date)',
    'Name must be unique per date.',
)

# Unique per parent
_name_parent_unique = models.Constraint(
    'UNIQUE(name, parent_id)',
    'Name must be unique within parent.',
)

# Unique per company (common pattern)
_reference_company_unique = models.Constraint(
    'UNIQUE(reference, company_id)',
    'Reference must be unique per company.',
)
```

#### Value Checks
```python
# Positive value
_quantity_positive = models.Constraint(
    'CHECK(quantity > 0)',
    'Quantity must be greater than zero.',
)

# Non-negative
_balance_non_negative = models.Constraint(
    'CHECK(balance >= 0)',
    'Balance cannot be negative.',
)

# Range
_discount_range = models.Constraint(
    'CHECK(discount >= 0 AND discount <= 100)',
    'Discount must be between 0% and 100%.',
)

# Not equal
_not_self_parent = models.Constraint(
    'CHECK(id != parent_id)',
    'Record cannot be its own parent.',
)
```

#### Conditional Checks
```python
# Required if condition
_email_required_for_customers = models.Constraint(
    'CHECK(is_customer = false OR email IS NOT NULL)',
    'Email is required for customers.',
)

# Either/or
_product_or_description = models.Constraint(
    'CHECK(product_id IS NOT NULL OR description IS NOT NULL)',
    'Either product or description is required.',
)

# Mutually exclusive
_exclusive_type = models.Constraint(
    'CHECK((type_a = true AND type_b = false) OR (type_a = false AND type_b = true) OR (type_a = false AND type_b = false))',
    'Cannot be both type A and type B.',
)
```

---

## Python Constraints

### Basic Python Constraint
```python
from odoo import api, models, fields
from odoo.exceptions import ValidationError


class MyModel(models.Model):
    _name = 'my.model'

    @api.constrains('amount')
    def _check_amount(self):
        for record in self:
            if record.amount < 0:
                raise ValidationError("Amount cannot be negative.")
```

### Multiple Fields
```python
@api.constrains('date_start', 'date_end')
def _check_dates(self):
    for record in self:
        if record.date_start and record.date_end:
            if record.date_start > record.date_end:
                raise ValidationError(
                    "End date must be after start date."
                )
```

### Complex Validation
```python
@api.constrains('line_ids')
def _check_lines(self):
    for record in self:
        if not record.line_ids:
            raise ValidationError("At least one line is required.")

        total = sum(record.line_ids.mapped('amount'))
        if total != record.amount_total:
            raise ValidationError(
                f"Line amounts ({total}) must equal total ({record.amount_total})."
            )
```

### Cross-Record Validation
```python
@api.constrains('code', 'company_id')
def _check_code_unique(self):
    """Check uniqueness with more control than SQL constraint."""
    for record in self:
        if not record.code:
            continue

        duplicate = self.search([
            ('id', '!=', record.id),
            ('code', '=ilike', record.code),  # Case-insensitive
            ('company_id', '=', record.company_id.id),
        ], limit=1)

        if duplicate:
            raise ValidationError(
                f"Code '{record.code}' already exists in this company."
            )
```

### Validation with External Data
```python
@api.constrains('email')
def _check_email_format(self):
    import re
    email_pattern = r'^[\w\.-]+@[\w\.-]+\.\w+$'

    for record in self:
        if record.email and not re.match(email_pattern, record.email):
            raise ValidationError(
                f"Invalid email format: {record.email}"
            )

@api.constrains('phone')
def _check_phone_format(self):
    import re
    phone_pattern = r'^\+?[\d\s-]{8,}$'

    for record in self:
        if record.phone and not re.match(phone_pattern, record.phone):
            raise ValidationError(
                f"Invalid phone format: {record.phone}"
            )
```

### Validation with Context
```python
@api.constrains('quantity')
def _check_quantity_available(self):
    """Skip validation during import."""
    if self.env.context.get('import_mode'):
        return

    for record in self:
        available = record.product_id.qty_available
        if record.quantity > available:
            raise ValidationError(
                f"Requested quantity ({record.quantity}) exceeds "
                f"available stock ({available})."
            )
```

---

## Advanced Patterns

### Hierarchical Validation
```python
@api.constrains('parent_id')
def _check_hierarchy(self):
    """Prevent circular references."""
    if not self._check_recursion():
        raise ValidationError(
            "Error! You cannot create recursive categories."
        )
```

### State-Dependent Validation
```python
@api.constrains('state', 'partner_id', 'line_ids')
def _check_state_requirements(self):
    for record in self:
        if record.state == 'confirmed':
            if not record.partner_id:
                raise ValidationError(
                    "Partner is required for confirmed records."
                )
            if not record.line_ids:
                raise ValidationError(
                    "Lines are required for confirmed records."
                )
```

### Aggregate Validation
```python
@api.constrains('percentage')
def _check_total_percentage(self):
    """Ensure percentages sum to 100%."""
    for record in self:
        siblings = self.search([
            ('parent_id', '=', record.parent_id.id),
        ])
        total = sum(siblings.mapped('percentage'))

        if abs(total - 100) > 0.01:  # Allow small rounding errors
            raise ValidationError(
                f"Percentages must sum to 100%. Current total: {total}%"
            )
```

### Business Period Validation
```python
@api.constrains('date', 'company_id')
def _check_fiscal_period(self):
    """Ensure date is in open fiscal period."""
    for record in self:
        period = self.env['account.fiscal.year'].search([
            ('company_id', '=', record.company_id.id),
            ('date_from', '<=', record.date),
            ('date_to', '>=', record.date),
        ], limit=1)

        if not period:
            raise ValidationError(
                f"No fiscal year defined for date {record.date}."
            )

        if period.state == 'closed':
            raise ValidationError(
                f"Cannot post to closed fiscal period: {period.name}"
            )
```

---

## Constraint with Detailed Messages

### Multiple Error Collection
```python
@api.constrains('name', 'code', 'amount', 'date_start', 'date_end')
def _check_all_fields(self):
    """Validate all fields and collect errors."""
    for record in self:
        errors = []

        if not record.name or len(record.name) < 3:
            errors.append("Name must be at least 3 characters.")

        if record.code and not record.code.isalnum():
            errors.append("Code must be alphanumeric only.")

        if record.amount <= 0:
            errors.append("Amount must be greater than zero.")

        if record.date_start and record.date_end:
            if record.date_start > record.date_end:
                errors.append("Start date must be before end date.")

        if errors:
            raise ValidationError("\n".join(errors))
```

### Field-Specific Error Messages
```python
@api.constrains('quantity', 'product_id')
def _check_quantity(self):
    for record in self:
        if record.quantity <= 0:
            raise ValidationError(
                f"Invalid quantity for product '{record.product_id.name}': "
                f"must be greater than zero."
            )

        min_qty = record.product_id.x_min_order_qty
        if min_qty and record.quantity < min_qty:
            raise ValidationError(
                f"Minimum order quantity for '{record.product_id.name}' "
                f"is {min_qty}. Requested: {record.quantity}"
            )
```

---

## When to Use Which

### Use SQL Constraints For:
- Simple uniqueness checks
- Basic numeric checks (positive, range)
- Database-level integrity
- Performance-critical validations

### Use Python Constraints For:
- Complex business logic
- Cross-record validation
- External data validation
- Conditional validation
- Custom error messages
- Validation that needs ORM features

---

## Best Practices

1. **Prefer SQL for simple checks** - More efficient, database-level
2. **Use Python for complex logic** - More flexibility
3. **Clear error messages** - Tell user what's wrong and how to fix
4. **Validate early** - Catch errors before processing
5. **Consider context** - Skip validation during imports if appropriate
6. **Test constraints** - Write tests for validation logic
7. **Don't over-constrain** - Balance integrity vs usability
8. **Document constraints** - Explain business rules
9. **Handle upgrades** - New constraints may fail on existing data
10. **Performance** - Avoid heavy queries in constraints

---

## Handling Existing Data

### Adding Constraints to Existing Tables
```python
# In migration script
def migrate(cr, version):
    """Fix data before adding constraint."""
    # Fix invalid data first
    cr.execute("""
        UPDATE my_model
        SET amount = 0
        WHERE amount < 0
    """)

    # Remove duplicates
    cr.execute("""
        DELETE FROM my_model a
        USING my_model b
        WHERE a.id > b.id
        AND a.code = b.code
        AND a.company_id = b.company_id
    """)
```

---


## Source: onchange-dynamic-patterns.md

# Onchange and Dynamic Form Patterns

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  ONCHANGE & DYNAMIC FORM PATTERNS                                            ║
║  @api.onchange, dynamic domains, and form field updates                      ║
║  Use for interactive form behavior and field auto-population                 ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## Basic Onchange

### Simple Field Update
```python
from odoo import api, fields, models


class SaleOrderLine(models.Model):
    _name = 'sale.order.line'

    product_id = fields.Many2one('product.product')
    name = fields.Char(string='Description')
    price_unit = fields.Float()
    quantity = fields.Float(default=1.0)
    discount = fields.Float()

    @api.onchange('product_id')
    def _onchange_product_id(self):
        """Update fields when product changes."""
        if self.product_id:
            self.name = self.product_id.name
            self.price_unit = self.product_id.list_price
        else:
            self.name = ''
            self.price_unit = 0.0
```

### Onchange with Warning
```python
@api.onchange('quantity')
def _onchange_quantity(self):
    """Warn if quantity exceeds stock."""
    if self.product_id and self.quantity:
        available = self.product_id.qty_available
        if self.quantity > available:
            return {
                'warning': {
                    'title': 'Insufficient Stock',
                    'message': (
                        f"Requested quantity ({self.quantity}) exceeds "
                        f"available stock ({available})."
                    ),
                }
            }
```

### Onchange with Domain
```python
@api.onchange('partner_id')
def _onchange_partner_id(self):
    """Filter products based on partner."""
    if self.partner_id:
        # Clear current selection if doesn't match new domain
        if self.product_id and self.product_id.partner_id != self.partner_id:
            self.product_id = False

        return {
            'domain': {
                'product_id': [('partner_id', '=', self.partner_id.id)],
            }
        }
    return {
        'domain': {
            'product_id': [],
        }
    }
```

---

## Onchange vs Computed

### When to Use Onchange
```python
# Onchange: user can modify the auto-filled value
@api.onchange('partner_id')
def _onchange_partner_id(self):
    """Fill address from partner (user can edit)."""
    if self.partner_id:
        self.street = self.partner_id.street
        self.city = self.partner_id.city
```

### When to Use Computed
```python
# Computed: value always derived from other fields
partner_id = fields.Many2one('res.partner')
partner_name = fields.Char(
    compute='_compute_partner_name',
    store=True,
)

@api.depends('partner_id')
def _compute_partner_name(self):
    for record in self:
        record.partner_name = record.partner_id.name or ''
```

### Combined Pattern
```python
class Order(models.Model):
    _name = 'my.order'

    partner_id = fields.Many2one('res.partner')

    # Computed (always calculated)
    partner_email = fields.Char(
        compute='_compute_partner_email',
        store=True,
    )

    # Editable (filled by onchange, user can modify)
    delivery_address = fields.Text()

    @api.depends('partner_id')
    def _compute_partner_email(self):
        for record in self:
            record.partner_email = record.partner_id.email or ''

    @api.onchange('partner_id')
    def _onchange_partner_id(self):
        """Pre-fill delivery address."""
        if self.partner_id:
            self.delivery_address = self._format_address(self.partner_id)
```

---

## Complex Onchange Patterns

### Cascading Onchange
```python
class Order(models.Model):
    _name = 'my.order'

    country_id = fields.Many2one('res.country')
    state_id = fields.Many2one('res.country.state')
    city_id = fields.Many2one('res.city')

    @api.onchange('country_id')
    def _onchange_country_id(self):
        """Clear state and city when country changes."""
        self.state_id = False
        self.city_id = False
        if self.country_id:
            return {
                'domain': {
                    'state_id': [('country_id', '=', self.country_id.id)],
                }
            }
        return {'domain': {'state_id': []}}

    @api.onchange('state_id')
    def _onchange_state_id(self):
        """Clear city when state changes."""
        self.city_id = False
        if self.state_id:
            return {
                'domain': {
                    'city_id': [('state_id', '=', self.state_id.id)],
                }
            }
        return {'domain': {'city_id': []}}
```

### Onchange on One2many
```python
class Order(models.Model):
    _name = 'my.order'

    line_ids = fields.One2many('my.order.line', 'order_id')
    amount_total = fields.Float(compute='_compute_amount_total')

    @api.onchange('line_ids')
    def _onchange_line_ids(self):
        """Recalculate total when lines change."""
        # This triggers when lines are added/removed/modified
        self.amount_total = sum(
            line.quantity * line.price_unit
            for line in self.line_ids
        )
```

### Onchange with External Lookup
```python
@api.onchange('vat')
def _onchange_vat(self):
    """Look up company info from VAT number."""
    if self.vat:
        try:
            # External API call
            info = self._lookup_vat(self.vat)
            if info:
                self.name = info.get('name', '')
                self.street = info.get('address', '')
                return {
                    'warning': {
                        'title': 'Company Found',
                        'message': f"Company info loaded for VAT {self.vat}",
                    }
                }
        except Exception as e:
            return {
                'warning': {
                    'title': 'VAT Lookup Failed',
                    'message': str(e),
                }
            }
```

---

## Dynamic Domains

### Domain in Field Definition
```python
class Task(models.Model):
    _name = 'my.task'

    project_id = fields.Many2one('project.project')

    # Static domain
    user_id = fields.Many2one(
        'res.users',
        domain=[('share', '=', False)],  # Internal users only
    )

    # Dynamic domain (string evaluated at runtime)
    assignee_id = fields.Many2one(
        'res.users',
        domain="[('id', 'in', allowed_user_ids)]",
    )

    allowed_user_ids = fields.Many2many(
        'res.users',
        compute='_compute_allowed_users',
    )

    @api.depends('project_id')
    def _compute_allowed_users(self):
        for record in self:
            if record.project_id:
                record.allowed_user_ids = record.project_id.member_ids
            else:
                record.allowed_user_ids = self.env['res.users'].search([])
```

### Domain in View
```xml
<form>
    <group>
        <field name="project_id"/>
        <!-- Domain evaluated in context -->
        <field name="task_id"
               domain="[('project_id', '=', project_id)]"/>

        <!-- Domain with parent reference -->
        <field name="line_ids">
            <tree editable="bottom">
                <field name="product_id"
                       domain="[('categ_id', '=', parent.category_id)]"/>
            </tree>
        </field>
    </group>
</form>
```

### Context-Based Domain
```python
@api.onchange('type')
def _onchange_type(self):
    """Set domain based on type selection."""
    domains = {
        'product': [('type', '=', 'product')],
        'service': [('type', '=', 'service')],
        'consumable': [('type', '=', 'consu')],
    }
    return {
        'domain': {
            'product_id': domains.get(self.type, []),
        }
    }
```

---

## Form Field Visibility

### Dynamic Visibility with Onchange
```python
class Document(models.Model):
    _name = 'my.document'

    type = fields.Selection([
        ('internal', 'Internal'),
        ('external', 'External'),
    ])

    # Fields that appear based on type
    internal_ref = fields.Char()
    external_partner_id = fields.Many2one('res.partner')

    @api.onchange('type')
    def _onchange_type(self):
        """Clear irrelevant fields when type changes."""
        if self.type == 'internal':
            self.external_partner_id = False
        elif self.type == 'external':
            self.internal_ref = False
```

### View with Conditional Visibility
```xml
<form>
    <group>
        <field name="type"/>
        <field name="internal_ref" invisible="type != 'internal'"/>
        <field name="external_partner_id" invisible="type != 'external'"/>
    </group>
</form>
```

---

## Onchange Return Values

### Complete Return Structure
```python
@api.onchange('field_name')
def _onchange_field(self):
    """Full onchange return options."""
    # Set field values directly
    self.other_field = 'value'
    self.computed_value = self.field_name * 2

    return {
        # Filter options for other fields
        'domain': {
            'related_field': [('condition', '=', True)],
        },
        # Show warning to user
        'warning': {
            'title': 'Warning Title',
            'message': 'Warning message to display',
            'type': 'notification',  # or 'dialog'
        },
    }
```

### Warning Types
```python
# Notification (non-blocking)
return {
    'warning': {
        'title': 'Info',
        'message': 'This is informational',
        'type': 'notification',
    }
}

# Dialog (blocking)
return {
    'warning': {
        'title': 'Attention',
        'message': 'Please confirm this action',
        'type': 'dialog',
    }
}
```

---

## Multiple Field Onchange

### Single Decorator, Multiple Fields
```python
@api.onchange('quantity', 'price_unit', 'discount')
def _onchange_amount(self):
    """Recalculate when any pricing field changes."""
    subtotal = self.quantity * self.price_unit
    discount_amount = subtotal * (self.discount / 100)
    self.amount = subtotal - discount_amount
```

### Order of Execution
```python
# Onchanges execute in definition order
@api.onchange('partner_id')
def _onchange_partner_id(self):
    """First: set pricelist from partner."""
    if self.partner_id:
        self.pricelist_id = self.partner_id.property_product_pricelist

@api.onchange('pricelist_id')
def _onchange_pricelist_id(self):
    """Second: update prices based on pricelist."""
    if self.pricelist_id:
        self._update_line_prices()
```

---

## Onchange in Wizards

### Wizard Onchange Pattern
```python
class ConfigWizard(models.TransientModel):
    _name = 'config.wizard'

    template_id = fields.Many2one('config.template')
    name = fields.Char()
    settings = fields.Text()

    @api.onchange('template_id')
    def _onchange_template_id(self):
        """Load template settings."""
        if self.template_id:
            self.name = self.template_id.name
            self.settings = self.template_id.default_settings
```

---

## Best Practices

### Do's
```python
# Good: Clear dependent fields
@api.onchange('parent_id')
def _onchange_parent_id(self):
    self.child_id = False  # Clear when parent changes

# Good: Handle empty values
@api.onchange('partner_id')
def _onchange_partner_id(self):
    if self.partner_id:
        self.phone = self.partner_id.phone
    else:
        self.phone = ''

# Good: Return domains for filtering
@api.onchange('category_id')
def _onchange_category_id(self):
    return {
        'domain': {
            'product_id': [('categ_id', '=', self.category_id.id)]
                          if self.category_id else [],
        }
    }
```

### Don'ts
```python
# Bad: Heavy computation in onchange
@api.onchange('product_id')
def _onchange_product_id(self):
    # Don't do expensive operations
    all_orders = self.env['sale.order'].search([])  # Bad!

# Bad: Database writes in onchange
@api.onchange('quantity')
def _onchange_quantity(self):
    self.env['stock.move'].create({})  # Bad! Record doesn't exist yet

# Bad: Modifying records outside self
@api.onchange('partner_id')
def _onchange_partner_id(self):
    self.partner_id.last_accessed = fields.Date.today()  # Bad!
```

---

## Debugging Onchange

### Logging Onchange Values
```python
import logging
_logger = logging.getLogger(__name__)

@api.onchange('field_name')
def _onchange_field_name(self):
    _logger.info(
        "Onchange triggered: field=%s, value=%s, self.id=%s",
        'field_name', self.field_name, self.id
    )
    # Note: self.id is NewId for unsaved records
```

### Testing Onchange
```python
def test_onchange_product(self):
    """Test product onchange fills name and price."""
    line = self.env['sale.order.line'].new({
        'order_id': self.order.id,
    })

    line.product_id = self.product
    line._onchange_product_id()

    self.assertEqual(line.name, self.product.name)
    self.assertEqual(line.price_unit, self.product.list_price)
```

---

## Summary Table

| Feature | Onchange | Computed |
|---------|----------|----------|
| Trigger | User interaction | Dependency change |
| User editable | Yes | No (unless inverse) |
| Works before save | Yes | Yes |
| Works after save | No | Yes |
| Database query | Avoid | OK with store=True |
| Return domains | Yes | No |
| Return warnings | Yes | No |
| Use case | Auto-fill suggestions | Derived values |

---


## Source: inheritance-patterns.md

# Model and View Inheritance Patterns

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  INHERITANCE PATTERNS                                                        ║
║  Extending models, views, and controllers without modifying core code        ║
║  Use for customizations, extensions, and module integrations                 ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## Inheritance Types Overview

| Type | `_name` | `_inherit` | Use Case |
|------|---------|-----------|----------|
| Extension | None | `'model'` | Add fields/methods to existing model |
| Delegation | `'new'` | `'model'` | Link new model to existing |
| Prototype | `'new'` | `['model']` | Copy structure from existing |

---

## Model Extension (Most Common)

### Add Fields to Existing Model
```python
from odoo import api, fields, models


class ResPartner(models.Model):
    _inherit = 'res.partner'

    # New fields
    x_loyalty_points = fields.Integer(
        string='Loyalty Points',
        default=0,
    )
    x_customer_tier = fields.Selection(
        selection=[
            ('bronze', 'Bronze'),
            ('silver', 'Silver'),
            ('gold', 'Gold'),
        ],
        string='Customer Tier',
        compute='_compute_customer_tier',
        store=True,
    )
    x_account_manager_id = fields.Many2one(
        comodel_name='res.users',
        string='Account Manager',
    )

    @api.depends('x_loyalty_points')
    def _compute_customer_tier(self):
        for partner in self:
            if partner.x_loyalty_points >= 1000:
                partner.x_customer_tier = 'gold'
            elif partner.x_loyalty_points >= 500:
                partner.x_customer_tier = 'silver'
            else:
                partner.x_customer_tier = 'bronze'
```

### Override Methods
```python
class SaleOrder(models.Model):
    _inherit = 'sale.order'

    def action_confirm(self):
        """Override confirm to add custom logic."""
        # Pre-processing
        for order in self:
            order._check_credit_limit()

        # Call original method
        result = super().action_confirm()

        # Post-processing
        for order in self:
            order._send_confirmation_notification()

        return result

    def _check_credit_limit(self):
        """Custom credit check before confirmation."""
        if self.partner_id.credit_limit and \
           self.amount_total > self.partner_id.credit_limit:
            raise UserError("Order exceeds credit limit.")

    def _send_confirmation_notification(self):
        """Send notification after confirmation."""
        template = self.env.ref('my_module.email_template_order_confirm')
        template.send_mail(self.id)
```

### Extend Computed Fields
```python
class SaleOrderLine(models.Model):
    _inherit = 'sale.order.line'

    @api.depends('product_uom_qty', 'discount', 'price_unit', 'tax_id')
    def _compute_amount(self):
        """Extend to add loyalty discount."""
        super()._compute_amount()

        for line in self:
            if line.order_id.partner_id.x_customer_tier == 'gold':
                # Apply additional 5% discount for gold customers
                line.price_subtotal *= 0.95
```

### Add to Selection Field
```python
class SaleOrder(models.Model):
    _inherit = 'sale.order'

    # Extend existing selection
    state = fields.Selection(
        selection_add=[
            ('pending_approval', 'Pending Approval'),
            ('approved', 'Approved'),
        ],
        ondelete={
            'pending_approval': 'set default',
            'approved': 'set default',
        },
    )
```

---

## Delegation Inheritance

### Link to Existing Model
```python
class Employee(models.Model):
    _name = 'hr.employee'
    _inherits = {'res.partner': 'partner_id'}

    partner_id = fields.Many2one(
        comodel_name='res.partner',
        string='Related Partner',
        required=True,
        ondelete='cascade',
    )

    # Employee-specific fields
    department_id = fields.Many2one('hr.department')
    job_id = fields.Many2one('hr.job')

    # Inherited fields from res.partner are accessible directly
    # employee.name -> partner.name
    # employee.email -> partner.email
```

### Custom Delegation
```python
class ProductVariant(models.Model):
    _name = 'product.product'
    _inherits = {'product.template': 'product_tmpl_id'}

    product_tmpl_id = fields.Many2one(
        comodel_name='product.template',
        string='Product Template',
        required=True,
        ondelete='cascade',
    )

    # Variant-specific fields
    barcode = fields.Char(string='Barcode')
    default_code = fields.Char(string='Internal Reference')
```

---

## Abstract Models (Mixins)

### Create Reusable Mixin
```python
class TimestampMixin(models.AbstractModel):
    _name = 'timestamp.mixin'
    _description = 'Timestamp Mixin'

    created_at = fields.Datetime(
        string='Created At',
        default=fields.Datetime.now,
        readonly=True,
    )
    updated_at = fields.Datetime(
        string='Updated At',
        readonly=True,
    )

    def write(self, vals):
        vals['updated_at'] = fields.Datetime.now()
        return super().write(vals)


class ApprovalMixin(models.AbstractModel):
    _name = 'approval.mixin'
    _description = 'Approval Mixin'

    approval_state = fields.Selection(
        selection=[
            ('draft', 'Draft'),
            ('pending', 'Pending Approval'),
            ('approved', 'Approved'),
            ('rejected', 'Rejected'),
        ],
        string='Approval Status',
        default='draft',
        tracking=True,
    )
    approved_by = fields.Many2one(
        comodel_name='res.users',
        string='Approved By',
        readonly=True,
    )
    approved_date = fields.Datetime(
        string='Approval Date',
        readonly=True,
    )

    def action_submit_for_approval(self):
        self.write({'approval_state': 'pending'})

    def action_approve(self):
        self.write({
            'approval_state': 'approved',
            'approved_by': self.env.uid,
            'approved_date': fields.Datetime.now(),
        })

    def action_reject(self):
        self.write({'approval_state': 'rejected'})
```

### Use Mixins
```python
class PurchaseRequest(models.Model):
    _name = 'purchase.request'
    _description = 'Purchase Request'
    _inherit = ['mail.thread', 'timestamp.mixin', 'approval.mixin']

    name = fields.Char(string='Reference', required=True)
    amount = fields.Monetary(string='Amount')
    # Gets all fields and methods from mixins
```

---

## View Inheritance

### Extend Form View
```xml
<?xml version="1.0" encoding="utf-8"?>
<odoo>
    <record id="view_partner_form_inherit" model="ir.ui.view">
        <field name="name">res.partner.form.inherit.my_module</field>
        <field name="model">res.partner</field>
        <field name="inherit_id" ref="base.view_partner_form"/>
        <field name="arch" type="xml">
            <!-- Add after existing field -->
            <field name="email" position="after">
                <field name="x_loyalty_points"/>
                <field name="x_customer_tier"/>
            </field>

            <!-- Add new page to notebook -->
            <xpath expr="//notebook" position="inside">
                <page string="Loyalty" name="loyalty">
                    <group>
                        <field name="x_loyalty_points"/>
                        <field name="x_customer_tier"/>
                        <field name="x_account_manager_id"/>
                    </group>
                </page>
            </xpath>

            <!-- Replace existing element -->
            <field name="title" position="replace">
                <field name="title" placeholder="Select title..."/>
            </field>

            <!-- Add attributes -->
            <field name="phone" position="attributes">
                <attribute name="required">1</attribute>
            </field>

            <!-- Hide field (v17+) -->
            <field name="fax" position="attributes">
                <attribute name="invisible">1</attribute>
            </field>
        </field>
    </record>
</odoo>
```

### XPath Expressions

```xml
<!-- By field name -->
<field name="partner_id" position="after">

<!-- By xpath -->
<xpath expr="//field[@name='partner_id']" position="after">

<!-- First field in group -->
<xpath expr="//group[1]/field[1]" position="before">

<!-- Field inside specific group -->
<xpath expr="//group[@name='sale_info']/field[@name='date_order']" position="after">

<!-- Page by name -->
<xpath expr="//page[@name='other_info']" position="inside">

<!-- Button by name -->
<xpath expr="//button[@name='action_confirm']" position="before">

<!-- Div by class -->
<xpath expr="//div[hasclass('oe_title')]" position="inside">

<!-- Last element -->
<xpath expr="//sheet/*[last()]" position="after">
```

### Position Types
| Position | Effect |
|----------|--------|
| `inside` | Add as child (at end) |
| `after` | Add as sibling after |
| `before` | Add as sibling before |
| `replace` | Replace entire element |
| `attributes` | Modify attributes |

### Extend Tree View
```xml
<record id="view_order_tree_inherit" model="ir.ui.view">
    <field name="name">sale.order.tree.inherit</field>
    <field name="model">sale.order</field>
    <field name="inherit_id" ref="sale.view_order_tree"/>
    <field name="arch" type="xml">
        <field name="amount_total" position="after">
            <field name="x_margin" optional="show"/>
            <field name="x_priority" decoration-danger="x_priority == 'high'"/>
        </field>
    </field>
</record>
```

### Extend Search View
```xml
<record id="view_order_search_inherit" model="ir.ui.view">
    <field name="name">sale.order.search.inherit</field>
    <field name="model">sale.order</field>
    <field name="inherit_id" ref="sale.view_sales_order_filter"/>
    <field name="arch" type="xml">
        <filter name="my_quotations" position="after">
            <filter string="High Priority" name="high_priority"
                    domain="[('x_priority', '=', 'high')]"/>
        </filter>

        <xpath expr="//group" position="inside">
            <filter string="Priority" name="group_priority"
                    context="{'group_by': 'x_priority'}"/>
        </xpath>
    </field>
</record>
```

---

## Controller Inheritance

### Extend HTTP Controller
```python
from odoo import http
from odoo.http import request
from odoo.addons.website_sale.controllers.main import WebsiteSale


class WebsiteSaleExtend(WebsiteSale):

    @http.route()
    def cart(self, **post):
        """Extend cart to add custom data."""
        response = super().cart(**post)

        # Add custom values to qcontext
        if hasattr(response, 'qcontext'):
            response.qcontext['x_loyalty_points'] = \
                request.env.user.partner_id.x_loyalty_points

        return response

    @http.route('/shop/cart/update_loyalty', type='json', auth='user')
    def update_loyalty(self, points_to_use):
        """New endpoint for loyalty point redemption."""
        order = request.website.sale_get_order()
        if order:
            order.x_loyalty_points_used = points_to_use
        return {'success': True}
```

---

## Report Inheritance

### Extend Report Template
```xml
<template id="report_invoice_document_inherit"
          inherit_id="account.report_invoice_document">
    <!-- Add custom section -->
    <xpath expr="//div[@id='informations']" position="after">
        <div class="row mt-3">
            <div class="col-6">
                <strong>Customer Tier:</strong>
                <span t-field="o.partner_id.x_customer_tier"/>
            </div>
            <div class="col-6">
                <strong>Loyalty Points Earned:</strong>
                <span t-esc="int(o.amount_total / 10)"/>
            </div>
        </div>
    </xpath>

    <!-- Modify existing content -->
    <xpath expr="//span[@t-field='o.name']" position="attributes">
        <attribute name="class">h2 text-primary</attribute>
    </xpath>
</template>
```

---

## Security Inheritance

### Extend Access Rights
```csv
id,name,model_id:id,group_id:id,perm_read,perm_write,perm_create,perm_unlink
# Extend existing model access
access_partner_loyalty_user,res.partner.loyalty.user,base.model_res_partner,base.group_user,1,1,0,0
access_partner_loyalty_manager,res.partner.loyalty.manager,base.model_res_partner,my_module.group_loyalty_manager,1,1,1,1
```

### Add Record Rules
```xml
<record id="rule_partner_loyalty_user" model="ir.rule">
    <field name="name">Partner: Loyalty User See Own</field>
    <field name="model_id" ref="base.model_res_partner"/>
    <field name="domain_force">[
        '|',
        ('x_account_manager_id', '=', user.id),
        ('x_account_manager_id', '=', False)
    ]</field>
    <field name="groups" eval="[(4, ref('my_module.group_loyalty_user'))]"/>
</record>
```

---

## Best Practices

### 1. Use Proper Naming
```python
# Fields: x_ prefix for custom fields
x_custom_field = fields.Char()

# Views: include module name
inherit_id="base.view_partner_form"
name="res.partner.form.inherit.my_module"
```

### 2. Call Super Properly
```python
# Good - always call super
def action_confirm(self):
    result = super().action_confirm()
    self._custom_logic()
    return result

# Bad - skipping super breaks inheritance chain
def action_confirm(self):
    self._custom_logic()
    # Missing super() call!
```

### 3. Use Specific XPath
```xml
<!-- Good - specific path -->
<xpath expr="//field[@name='partner_id']" position="after">

<!-- Bad - fragile, may break -->
<xpath expr="//field[3]" position="after">
```

### 4. Handle Dependencies
```python
# Manifest
{
    'depends': ['sale', 'account'],  # Declare all inherited modules
}
```

### 5. Preserve Original Behavior
```python
# Good - extend, don't replace
def _compute_amount(self):
    super()._compute_amount()
    # Add to computed value
    for line in self:
        line.price_subtotal += line.x_extra_fee

# Bad - completely replaces original
def _compute_amount(self):
    for line in self:
        line.price_subtotal = line.quantity * line.price_unit
```

### 6. Never Use 'string' Attribute as Selector
```xml
<!-- Good - use 'name' attribute (stable identifier) -->
<xpath expr="//page[@name='other_info']" position="inside">
<xpath expr="//field[@name='partner_id']" position="after">
<xpath expr="//button[@name='action_confirm']" position="before">

<!-- Bad - 'string' attribute is translated and may change -->
<xpath expr="//page[@string='Other Information']" position="inside">
<xpath expr="//button[@string='Confirm']" position="before">
```

The `string` attribute should not be used as a selector in view inheritance because:
- It contains translatable text that varies by language
- Labels may be changed in future Odoo versions
- Other modules may override the same string differently

Always prefer `name` attributes which are stable technical identifiers.

### 7. Always Verify XML IDs and Views Before Extension

**⚠️ CRITICAL RULE:** Do not trust your memory or make assumptions about XML IDs, views, records, or any other Odoo identifiers. Your memory is flawed by design. Always use the Odoo indexer or search tools to look them up. **Always read the actual Odoo source code** when in doubt.

```python
# Bad - trusting memory or assumptions about XML IDs
view_id = self.env.ref('base.view_partner_form')  # May not exist!

# Good - verify existence before using
try:
    view_id = self.env.ref('base.view_partner_form', raise_if_not_found=False)
    if not view_id:
        raise ValueError("View not found")
except ValueError:
    # Handle missing view
    pass
```

```xml
<!-- Bad - assuming XML ID exists without verification -->
<record id="view_partner_form_inherit" model="ir.ui.view">
    <field name="name">res.partner.form.inherit.my_module</field>
    <field name="model">res.partner</field>
    <field name="inherit_id" ref="base.view_partner_form"/>
    <!-- This will fail if base.view_partner_form doesn't exist -->
</record>

<!-- Good - verify XML ID exists first using Odoo code/indexer -->
<!-- Before writing this inheritance, verify the XML ID exists:
     1. Use grep/search to find the view definition in Odoo source
     2. Check ir.ui.view records in database
     3. Use Odoo indexer/IDE tools to look up the XML ID
     4. READ THE ACTUAL ODOO SOURCE CODE - never rely on memory
-->
<record id="view_partner_form_inherit" model="ir.ui.view">
    <field name="name">res.partner.form.inherit.my_module</field>
    <field name="model">res.partner</field>
    <field name="inherit_id" ref="base.view_partner_form"/>
    <!-- Verified that base.view_partner_form exists by reading Odoo source -->
</record>
```

Why this matters:
- XML IDs can change between Odoo versions
- Views may be renamed, removed, or restructured
- Some models may only have certain view types (e.g., only kanban, no form/tree)
- Not all standard models have all view types defined
- Runtime errors occur when referencing non-existent XML IDs
- Memory of Odoo's structure is unreliable and prone to errors

**Real-world example:**
In Odoo 19, the `res.users.api.keys` model only has a kanban view - no form or tree view exists. Attempting to extend a non-existent form view will cause runtime errors. Always verify what views actually exist before attempting to extend them.

```python
# Example: Check what views exist for a model
def check_available_views(self, model_name):
    """Check what view types exist for a model."""
    views = self.env['ir.ui.view'].search([
        ('model', '=', model_name),
        ('type', '!=', False)
    ])
    return {view.type for view in views}

# Before extending: check if the view type exists
# available_views = check_available_views('res.users.api.keys')
# Result might be: {'kanban'}  # Only kanban, no form or tree!
```

**Always verify XML IDs and view existence before use:**
1. **READ THE ACTUAL ODOO SOURCE CODE** - this is the primary method, never rely on memory
2. Use grep/ripgrep to search Odoo source code for the exact XML ID
3. Check the database `ir.model.data` table for the record
4. Use Odoo indexer or IDE integration tools to look up identifiers
5. Verify which view types exist for a model before extending them
6. Check the specific Odoo version's codebase (views differ between versions)
7. Never assume an XML ID exists - always look it up using the tools above

```python
# Example: Verifying a view exists before inheritance
def _check_view_exists(self, xml_id):
    """Check if XML ID exists before using it."""
    try:
        return self.env.ref(xml_id, raise_if_not_found=False)
    except ValueError:
        return False

# Check if specific view type exists for a model
def _check_view_type_exists(self, model_name, view_type):
    """Check if a specific view type exists for a model."""
    return self.env['ir.ui.view'].search([
        ('model', '=', model_name),
        ('type', '=', view_type)
    ], limit=1)

# In data files, you can use noupdate="1" with error handling
# Or use module dependencies to ensure base modules are loaded
```

**Workflow when extending views:**
1. Identify the model you want to extend
2. Search Odoo source code to find available views for that model
3. Verify the view type exists (form, tree, kanban, etc.)
4. Look up the exact XML ID using search/indexer
5. Read the actual view structure to understand the elements
6. Test your XPath expressions against the actual view structure
7. Only then write your view inheritance
8. Never assume or guess - always verify

### 8. Always Test XPath Expressions Before Use

**⚠️ CRITICAL RULE:** After looking up the XML ID and before writing view inheritance, **read the actual view structure** and verify that your XPath expressions will match the correct elements. XPath errors are a common source of view inheritance failures.

```xml
<!-- Bad - assuming structure without verification -->
<record id="view_apikeys_kanban_inherit" model="ir.ui.view">
    <field name="name">res.users.apikeys.kanban.inherit</field>
    <field name="model">res.users.apikeys</field>
    <field name="inherit_id" ref="base.res_users_apikeys_view_kanban"/>
    <field name="arch" type="xml">
        <!-- Wrong XPath - this doesn't match actual structure! -->
        <xpath expr="//div[hasclass('flex-row')]" position="inside">
            <span>My content</span>
        </xpath>
    </field>
</record>
```

**Correct workflow:**
1. **Read the actual view** to understand its structure
2. **Verify the XPath** matches the actual elements
3. Write the correct inheritance

```xml
<!-- Good - verified XPath matches actual structure -->
<record id="view_apikeys_kanban_inherit" model="ir.ui.view">
    <field name="name">res.users.apikeys.kanban.inherit</field>
    <field name="model">res.users.apikeys</field>
    <field name="inherit_id" ref="base.res_users_apikeys_view_kanban"/>
    <field name="arch" type="xml">
        <!-- Correct XPath after reading the view structure:
             The view has <t t-name="card"> not a div with flex-row class -->
        <xpath expr="//t[@t-name='card']/div/div" position="inside">
            <span>My content</span>
        </xpath>
    </field>
</record>
```

**Real-world example from Odoo 19:**
When extending `res.users.apikeys` kanban view, developers might assume there's a `div` with `flex-row` class. However, reading the actual view reveals it uses `<t t-name="card">` structure instead. Using the wrong XPath will cause runtime errors.

**How to test XPath expressions:**
1. Read the source XML view file from Odoo codebase
2. Identify the exact element structure and attributes
3. Write your XPath to match the actual structure
4. Test by upgrading your module and checking for errors
5. If errors occur, re-read the view and adjust XPath

```python
# Example: Read a view to understand its structure
def read_view_structure(self, xml_id):
    """Read view arch to understand structure before inheritance."""
    view = self.env.ref(xml_id, raise_if_not_found=False)
    if view:
        # Print or log the arch to see actual structure
        print(view.arch)
        return view.arch
    return None

# Before writing inheritance:
# read_view_structure('base.res_users_apikeys_view_kanban')
# Examine output to understand structure and write correct XPath
```

**Common XPath mistakes:**
- Assuming class names that don't exist
- Using wrong element names (div vs t vs field)
- Not checking for QWeb template structures (t-name, t-if, etc.)
- Copying XPath from different Odoo versions
- Not accounting for nested structures

**Best practices for XPath:**
1. Always read the actual view first
2. Use specific, precise selectors (element name + attributes)
3. Prefer `name` attributes over classes or string values
4. Test XPath against the actual structure
5. Document why you chose that specific XPath
6. Never copy-paste XPath without verification

### 8. Always Test XPath Expressions Before Use

When inheriting views, incorrect XPath expressions are a common source of errors. Always verify your XPath expressions match the actual view structure by reading the base view first.

```xml
<!-- Bad - guessing the XPath without reading the view -->
<record id="view_apikeys_kanban_inherit" model="ir.ui.view">
    <field name="name">res.users.apikeys.kanban.inherit</field>
    <field name="model">res.users.apikeys</field>
    <field name="inherit_id" ref="base.res_users_apikeys_view_kanban"/>
    <field name="arch" type="xml">
        <!-- This XPath is incorrect - assuming structure without verification -->
        <xpath expr="//div[hasclass('flex-row')]" position="inside">
            <span t-if="record.is_readonly.raw_value"
                  class="badge text-bg-warning ms-2">
                Read-Only
            </span>
        </xpath>
    </field>
</record>

<!-- Good - verified XPath by reading the actual view structure -->
<record id="view_apikeys_kanban_inherit" model="ir.ui.view">
    <field name="name">res.users.apikeys.kanban.inherit</field>
    <field name="model">res.users.apikeys</field>
    <field name="inherit_id" ref="base.res_users_apikeys_view_kanban"/>
    <field name="arch" type="xml">
        <!-- Correct XPath after reading the view and finding <t t-name="card"> -->
        <xpath expr="//t[@t-name='card']/div/div" position="inside">
            <span t-if="record.is_readonly.raw_value"
                  class="badge text-bg-warning ms-2">
                Read-Only
            </span>
        </xpath>
    </field>
</record>
```

**Best Practice Workflow for XPath Expressions:**

1. **Read the base view first** - Never write XPath expressions from memory
2. **Identify the exact element** - Find the precise structure in the view
3. **Write the XPath expression** - Use the actual element names and attributes
4. **Test the inheritance** - Verify the view renders correctly
5. **Debug if needed** - If it fails, re-read the view and correct the XPath

**Common XPath mistakes:**
- Using `hasclass()` with incorrect class names
- Assuming div structure without checking actual elements (could be `<t>`, `<span>`, etc.)
- Not accounting for QWeb-specific elements like `<t t-name="...">`
- Guessing element hierarchy instead of reading the actual view

**Real-world example from Odoo 19:**

The `res.users.apikeys` kanban view uses `<t t-name="card">` for the card template, not a simple `<div>`. An incorrect XPath like `//div[hasclass('flex-row')]` will fail, while the correct XPath `//t[@t-name='card']/div/div` works because it matches the actual structure.

**How to verify XPath expressions:**
```python
# Read the view to check structure
view = self.env.ref('base.res_users_apikeys_view_kanban')
print(view.arch)  # Examine the XML structure

# Or use grep/search in Odoo source code
# grep -r "res_users_apikeys_view_kanban" odoo/addons/base/
```

Always read first, then write. Never trust your memory about view structures - they change between Odoo versions and even within the same version as views are refactored.

---

## Common Inheritance Patterns

### Add Workflow State
```python
class SaleOrder(models.Model):
    _inherit = 'sale.order'

    state = fields.Selection(
        selection_add=[
            ('waiting_approval', 'Waiting Approval'),
        ],
        ondelete={'waiting_approval': 'set default'},
    )

    def action_submit_approval(self):
        self.write({'state': 'waiting_approval'})

    def action_approve(self):
        self.action_confirm()
```

### Add Smart Button
```xml
<xpath expr="//div[@name='button_box']" position="inside">
    <button class="oe_stat_button" type="object"
            name="action_view_loyalty_history"
            icon="fa-star">
        <field string="Points" name="x_loyalty_points"
               widget="statinfo"/>
    </button>
</xpath>
```

### Conditional Field Display
```xml
<!-- v17+ syntax -->
<field name="x_approval_notes"
       invisible="state not in ['waiting_approval', 'approved']"/>

<!-- Pre-v17 syntax -->
<field name="x_approval_notes"
       attrs="{'invisible': [('state', 'not in', ['waiting_approval', 'approved'])]}"/>
```

---


## Source: domain-filter-patterns.md

# Domain and Filter Patterns

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  DOMAIN & FILTER PATTERNS                                                    ║
║  Search domains, record filtering, and query optimization                    ║
║  Use for search views, record rules, and programmatic filtering              ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## Domain Syntax

### Basic Operators
| Operator | Description | Example |
|----------|-------------|---------|
| `=` | Equal | `('state', '=', 'draft')` |
| `!=` | Not equal | `('state', '!=', 'cancel')` |
| `>` | Greater than | `('amount', '>', 100)` |
| `>=` | Greater or equal | `('date', '>=', '2024-01-01')` |
| `<` | Less than | `('quantity', '<', 10)` |
| `<=` | Less or equal | `('date', '<=', today)` |
| `like` | Pattern match (case sensitive) | `('name', 'like', 'Test%')` |
| `ilike` | Pattern match (case insensitive) | `('name', 'ilike', '%test%')` |
| `=like` | SQL LIKE | `('code', '=like', 'ABC%')` |
| `=ilike` | SQL ILIKE | `('code', '=ilike', 'abc%')` |
| `in` | In list | `('state', 'in', ['draft', 'sent'])` |
| `not in` | Not in list | `('state', 'not in', ['cancel'])` |
| `child_of` | Hierarchical child | `('category_id', 'child_of', parent_id)` |
| `parent_of` | Hierarchical parent | `('category_id', 'parent_of', child_id)` |

### Logical Operators
```python
# AND (implicit between tuples)
domain = [
    ('state', '=', 'confirmed'),
    ('date', '>=', '2024-01-01'),
]

# OR (prefix notation)
domain = [
    '|',
    ('state', '=', 'draft'),
    ('state', '=', 'sent'),
]

# NOT
domain = [
    '!',
    ('state', '=', 'cancel'),
]

# Complex: (A AND B) OR (C AND D)
domain = [
    '|',
    '&', ('state', '=', 'draft'), ('user_id', '=', uid),
    '&', ('state', '=', 'confirmed'), ('amount', '>', 1000),
]
```

---

## Common Domain Patterns

### Date Ranges
```python
from datetime import date, datetime, timedelta
from odoo import fields

# Today
today = fields.Date.today()
domain = [('date', '=', today)]

# This week
week_start = today - timedelta(days=today.weekday())
week_end = week_start + timedelta(days=6)
domain = [
    ('date', '>=', week_start),
    ('date', '<=', week_end),
]

# This month
month_start = today.replace(day=1)
next_month = (month_start + timedelta(days=32)).replace(day=1)
domain = [
    ('date', '>=', month_start),
    ('date', '<', next_month),
]

# Last 30 days
domain = [('date', '>=', today - timedelta(days=30))]

# Between dates
domain = [
    ('date', '>=', date_from),
    ('date', '<=', date_to),
]
```

### Company Filtering
```python
# Current company only
domain = [('company_id', '=', self.env.company.id)]

# User's companies
domain = [('company_id', 'in', self.env.user.company_ids.ids)]

# Or no company (shared)
domain = [
    '|',
    ('company_id', '=', False),
    ('company_id', '=', self.env.company.id),
]
```

### User/Partner Filtering
```python
# Current user
domain = [('user_id', '=', self.env.uid)]

# Current user's partner
domain = [('partner_id', '=', self.env.user.partner_id.id)]

# Team members
domain = [('user_id', 'in', self.env.user.team_id.member_ids.ids)]

# No assigned user
domain = [('user_id', '=', False)]
```

### Related Field Filtering
```python
# Filter by related field
domain = [('partner_id.country_id', '=', country_id)]

# Multiple levels
domain = [('order_id.partner_id.is_company', '=', True)]

# Related Many2many
domain = [('tag_ids.name', 'ilike', 'important')]
```

### Null/Empty Checks
```python
# Field is empty
domain = [('name', '=', False)]

# Field is not empty
domain = [('name', '!=', False)]

# Empty string vs False
domain = [('name', 'in', [False, ''])]

# Has related records
domain = [('line_ids', '!=', False)]
```

---

## Dynamic Domains

### In Python Methods
```python
def _get_records_domain(self):
    """Build domain dynamically."""
    domain = [('active', '=', True)]

    if self.partner_id:
        domain.append(('partner_id', '=', self.partner_id.id))

    if self.date_from:
        domain.append(('date', '>=', self.date_from))

    if self.date_to:
        domain.append(('date', '<=', self.date_to))

    if self.state_filter:
        domain.append(('state', '=', self.state_filter))

    return domain

def action_search(self):
    domain = self._get_records_domain()
    records = self.env['my.model'].search(domain)
    return records
```

### In Field Definitions
```python
# Static domain
partner_id = fields.Many2one(
    'res.partner',
    domain=[('is_company', '=', True)],
)

# Dynamic domain (string)
partner_id = fields.Many2one(
    'res.partner',
    domain="[('company_id', '=', company_id)]",
)

# Complex dynamic domain
def _get_partner_domain(self):
    return [
        ('is_company', '=', True),
        ('country_id', '=', self.env.company.country_id.id),
    ]

partner_id = fields.Many2one(
    'res.partner',
    domain=lambda self: self._get_partner_domain(),
)
```

### In Views (XML)
```xml
<!-- Static domain -->
<field name="partner_id" domain="[('is_company', '=', True)]"/>

<!-- Dynamic using other fields -->
<field name="product_id"
       domain="[('categ_id', '=', category_id)]"/>

<!-- With context -->
<field name="user_id"
       domain="[('company_id', '=', company_id)]"
       context="{'default_company_id': company_id}"/>

<!-- Complex domain -->
<field name="location_id"
       domain="[
           ('usage', '=', 'internal'),
           '|',
           ('company_id', '=', company_id),
           ('company_id', '=', False)
       ]"/>
```

---

## Search Views

### Basic Search View
```xml
<record id="my_model_view_search" model="ir.ui.view">
    <field name="name">my.model.search</field>
    <field name="model">my.model</field>
    <field name="arch" type="xml">
        <search string="Search">
            <!-- Search fields -->
            <field name="name"/>
            <field name="partner_id"/>
            <field name="reference" filter_domain="[
                '|',
                ('name', 'ilike', self),
                ('reference', 'ilike', self)
            ]"/>

            <!-- Filters -->
            <filter string="My Records" name="my_records"
                    domain="[('user_id', '=', uid)]"/>
            <filter string="Today" name="today"
                    domain="[('date', '=', context_today().strftime('%Y-%m-%d'))]"/>
            <filter string="This Month" name="this_month"
                    domain="[
                        ('date', '>=', (context_today() + relativedelta(day=1)).strftime('%Y-%m-%d')),
                        ('date', '&lt;', (context_today() + relativedelta(months=1, day=1)).strftime('%Y-%m-%d'))
                    ]"/>
            <separator/>
            <filter string="Draft" name="draft"
                    domain="[('state', '=', 'draft')]"/>
            <filter string="Confirmed" name="confirmed"
                    domain="[('state', '=', 'confirmed')]"/>
            <separator/>
            <filter string="Archived" name="archived"
                    domain="[('active', '=', False)]"/>

            <!-- Group By -->
            <group expand="0" string="Group By">
                <filter string="Partner" name="group_partner"
                        context="{'group_by': 'partner_id'}"/>
                <filter string="State" name="group_state"
                        context="{'group_by': 'state'}"/>
                <filter string="Date" name="group_date"
                        context="{'group_by': 'date:month'}"/>
            </group>
        </search>
    </field>
</record>
```

### Advanced Search Features
```xml
<search>
    <!-- Multi-field search -->
    <field name="name" string="Name/Reference"
           filter_domain="[
               '|', '|',
               ('name', 'ilike', self),
               ('reference', 'ilike', self),
               ('partner_id.name', 'ilike', self)
           ]"/>

    <!-- Search related fields -->
    <field name="partner_id" operator="child_of"/>

    <!-- Date range filters -->
    <filter string="Last 7 Days" name="last_7_days"
            domain="[('create_date', '>=', (context_today() - relativedelta(days=7)).strftime('%Y-%m-%d'))]"/>

    <!-- Negative filter -->
    <filter string="Without Partner" name="no_partner"
            domain="[('partner_id', '=', False)]"/>

    <!-- Combined AND filter -->
    <filter string="Urgent Draft" name="urgent_draft"
            domain="[('state', '=', 'draft'), ('priority', '=', 'high')]"/>

    <!-- Dynamic context filter -->
    <filter string="My Team" name="my_team"
            domain="[('user_id.team_id', '=', %(sales_team.team_id)d)]"/>
</search>
```

---

## Record Rules

### Basic Record Rule
```xml
<!-- Users see only their own records -->
<record id="rule_my_model_user" model="ir.rule">
    <field name="name">My Model: User Rule</field>
    <field name="model_id" ref="model_my_model"/>
    <field name="domain_force">[('user_id', '=', user.id)]</field>
    <field name="groups" eval="[(4, ref('base.group_user'))]"/>
    <field name="perm_read" eval="True"/>
    <field name="perm_write" eval="True"/>
    <field name="perm_create" eval="True"/>
    <field name="perm_unlink" eval="True"/>
</record>
```

### Manager Rule (Override)
```xml
<!-- Managers see all records -->
<record id="rule_my_model_manager" model="ir.rule">
    <field name="name">My Model: Manager Rule</field>
    <field name="model_id" ref="model_my_model"/>
    <field name="domain_force">[(1, '=', 1)]</field>
    <field name="groups" eval="[(4, ref('my_module.group_manager'))]"/>
</record>
```

### Global Rule (No Groups)
```xml
<!-- Global rule applying to everyone -->
<record id="rule_my_model_company" model="ir.rule">
    <field name="name">My Model: Company Rule</field>
    <field name="model_id" ref="model_my_model"/>
    <field name="domain_force">[
        '|',
        ('company_id', '=', False),
        ('company_id', 'in', company_ids)
    ]</field>
    <field name="global" eval="True"/>
</record>
```

---

## Domain Helper Functions

### Domain Utilities
```python
from odoo.osv import expression


class DomainHelper(models.AbstractModel):
    _name = 'domain.helper'

    def combine_domains(self, *domains):
        """Combine multiple domains with AND."""
        return expression.AND(list(domains))

    def combine_domains_or(self, *domains):
        """Combine multiple domains with OR."""
        return expression.OR(list(domains))

    def normalize_domain(self, domain):
        """Normalize domain to standard form."""
        return expression.normalize_domain(domain)

    def is_false_domain(self, domain):
        """Check if domain is always false."""
        return expression.is_false(self, domain)

    def distribute_not(self, domain):
        """Push NOT operators down in domain."""
        return expression.distribute_not(domain)


# Usage
domain1 = [('state', '=', 'draft')]
domain2 = [('user_id', '=', self.env.uid)]

# AND combination
combined = expression.AND([domain1, domain2])
# Result: [('state', '=', 'draft'), ('user_id', '=', uid)]

# OR combination
combined = expression.OR([domain1, domain2])
# Result: ['|', ('state', '=', 'draft'), ('user_id', '=', uid)]
```

### Domain Parsing
```python
from odoo.osv.expression import DOMAIN_OPERATORS

def parse_domain(domain):
    """Parse and analyze domain."""
    result = {
        'fields': set(),
        'operators': [],
    }

    for element in domain:
        if isinstance(element, tuple):
            result['fields'].add(element[0])
            result['operators'].append(element[1])
        elif element in DOMAIN_OPERATORS:
            pass  # Logical operator

    return result
```

---

## Performance Tips

### Indexed Fields
```python
# Add index for frequently filtered fields
name = fields.Char(string='Name', index=True)
date = fields.Date(string='Date', index=True)
state = fields.Selection([...], index=True)
partner_id = fields.Many2one('res.partner', index=True)
```

### Efficient Domains
```python
# Good - Uses index
domain = [('state', '=', 'confirmed')]

# Bad - Function prevents index usage
domain = [('state', '=like', 'conf%')]

# Good - Specific IDs
domain = [('id', 'in', record_ids)]

# Bad - Too many OR conditions
domain = ['|'] * 99 + [('field', '=', v) for v in range(100)]
# Better - Use 'in'
domain = [('field', 'in', list(range(100)))]
```

### Limit Results
```python
# Always limit when possible
records = self.env['my.model'].search(domain, limit=100)

# Use search_count for counts
count = self.env['my.model'].search_count(domain)
```

---

## Best Practices

1. **Use `in` operator** for multiple values instead of multiple OR
2. **Index filtered fields** for better performance
3. **Limit results** when full result set not needed
4. **Use expression module** for combining domains
5. **Test complex domains** with actual data
6. **Document complex domains** with comments
7. **Use `child_of`/`parent_of`** for hierarchical data
8. **Prefer `ilike`** over `like` for user searches
9. **Handle empty values** explicitly
10. **Use record rules** for security, not business logic

---


## Source: workflow-state-patterns.md

# Workflow and State Machine Patterns

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  WORKFLOW & STATE MACHINE PATTERNS                                           ║
║  State transitions, approvals, and business process flows                    ║
║  Use for modeling business processes with defined states and transitions     ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## Basic State Machine

### Simple State Field
```python
from odoo import api, fields, models
from odoo.exceptions import UserError


class MyDocument(models.Model):
    _name = 'my.document'
    _description = 'My Document'

    name = fields.Char(required=True)

    state = fields.Selection([
        ('draft', 'Draft'),
        ('confirmed', 'Confirmed'),
        ('done', 'Done'),
        ('cancel', 'Cancelled'),
    ], string='Status', default='draft', required=True, tracking=True)

    # State transition methods
    def action_confirm(self):
        """Transition to confirmed state."""
        for record in self:
            if record.state != 'draft':
                raise UserError("Only draft documents can be confirmed.")
            record.state = 'confirmed'

    def action_done(self):
        """Transition to done state."""
        for record in self:
            if record.state != 'confirmed':
                raise UserError("Only confirmed documents can be marked done.")
            record.state = 'done'

    def action_cancel(self):
        """Cancel the document."""
        for record in self:
            if record.state == 'done':
                raise UserError("Cannot cancel completed documents.")
            record.state = 'cancel'

    def action_draft(self):
        """Reset to draft state."""
        for record in self:
            if record.state != 'cancel':
                raise UserError("Only cancelled documents can be reset to draft.")
            record.state = 'draft'
```

### View with Statusbar
```xml
<form>
    <header>
        <button name="action_confirm" string="Confirm"
                type="object" invisible="state != 'draft'"
                class="oe_highlight"/>
        <button name="action_done" string="Mark Done"
                type="object" invisible="state != 'confirmed'"
                class="oe_highlight"/>
        <button name="action_cancel" string="Cancel"
                type="object" invisible="state in ['done', 'cancel']"/>
        <button name="action_draft" string="Reset to Draft"
                type="object" invisible="state != 'cancel'"/>
        <field name="state" widget="statusbar"
               statusbar_visible="draft,confirmed,done"/>
    </header>
    <sheet>
        <group>
            <field name="name" readonly="state != 'draft'"/>
        </group>
    </sheet>
</form>
```

---

## State-Dependent Field Access

### Readonly in Certain States
```python
class MyDocument(models.Model):
    _name = 'my.document'

    state = fields.Selection([
        ('draft', 'Draft'),
        ('confirmed', 'Confirmed'),
        ('done', 'Done'),
    ], default='draft')

    # These fields readonly after draft
    partner_id = fields.Many2one('res.partner')
    date = fields.Date()
    line_ids = fields.One2many('my.document.line', 'document_id')

    # Computed field for lock status
    is_locked = fields.Boolean(compute='_compute_is_locked')

    @api.depends('state')
    def _compute_is_locked(self):
        for record in self:
            record.is_locked = record.state != 'draft'
```

### View with State-Based Readonly
```xml
<form>
    <header>
        <field name="state" widget="statusbar"/>
    </header>
    <sheet>
        <group>
            <!-- Odoo 17+ syntax -->
            <field name="partner_id" readonly="state != 'draft'"/>
            <field name="date" readonly="state != 'draft'"/>

            <!-- Or use computed field -->
            <field name="partner_id" readonly="is_locked"/>
        </group>

        <notebook>
            <page string="Lines">
                <field name="line_ids" readonly="state != 'draft'">
                    <tree editable="bottom">
                        <field name="product_id"/>
                        <field name="quantity"/>
                    </tree>
                </field>
            </page>
        </notebook>
    </sheet>
</form>
```

---

## Approval Workflow

### Multi-Level Approval
```python
class ApprovalDocument(models.Model):
    _name = 'approval.document'
    _inherit = ['mail.thread', 'mail.activity.mixin']

    name = fields.Char(required=True)
    amount = fields.Float()

    state = fields.Selection([
        ('draft', 'Draft'),
        ('submitted', 'Submitted'),
        ('first_approval', 'First Approval'),
        ('second_approval', 'Second Approval'),
        ('approved', 'Approved'),
        ('rejected', 'Rejected'),
    ], default='draft', tracking=True)

    # Approval tracking
    submitted_by = fields.Many2one('res.users', readonly=True)
    submitted_date = fields.Datetime(readonly=True)
    first_approver_id = fields.Many2one('res.users', readonly=True)
    first_approval_date = fields.Datetime(readonly=True)
    second_approver_id = fields.Many2one('res.users', readonly=True)
    second_approval_date = fields.Datetime(readonly=True)
    rejection_reason = fields.Text()

    def action_submit(self):
        """Submit for approval."""
        self.ensure_one()
        if self.state != 'draft':
            raise UserError("Document must be in draft state.")

        self.write({
            'state': 'submitted',
            'submitted_by': self.env.uid,
            'submitted_date': fields.Datetime.now(),
        })

        # Notify approvers
        self._notify_approvers()

    def action_first_approve(self):
        """First level approval."""
        self.ensure_one()
        self._check_approval_rights('first')

        self.write({
            'state': 'first_approval',
            'first_approver_id': self.env.uid,
            'first_approval_date': fields.Datetime.now(),
        })

        # Check if second approval needed
        if self.amount > 10000:
            self._notify_second_approvers()
        else:
            self.action_final_approve()

    def action_second_approve(self):
        """Second level approval."""
        self.ensure_one()
        self._check_approval_rights('second')

        self.write({
            'state': 'second_approval',
            'second_approver_id': self.env.uid,
            'second_approval_date': fields.Datetime.now(),
        })
        self.action_final_approve()

    def action_final_approve(self):
        """Mark as fully approved."""
        self.write({'state': 'approved'})
        self._execute_approved_actions()

    def action_reject(self):
        """Reject the document."""
        self.ensure_one()
        if not self.rejection_reason:
            raise UserError("Please provide a rejection reason.")

        self.write({'state': 'rejected'})
        self._notify_rejection()

    def _check_approval_rights(self, level):
        """Verify user can approve at this level."""
        if level == 'first':
            group = 'my_module.group_first_approver'
        else:
            group = 'my_module.group_second_approver'

        if not self.env.user.has_group(group):
            raise UserError("You don't have approval rights.")

    def _notify_approvers(self):
        """Send notification to approvers."""
        # Implementation depends on notification method
        pass
```

### Approval View
```xml
<form>
    <header>
        <button name="action_submit" string="Submit for Approval"
                type="object" invisible="state != 'draft'"
                class="oe_highlight"/>
        <button name="action_first_approve" string="Approve"
                type="object" invisible="state != 'submitted'"
                class="oe_highlight"
                groups="my_module.group_first_approver"/>
        <button name="action_second_approve" string="Final Approve"
                type="object" invisible="state != 'first_approval'"
                class="oe_highlight"
                groups="my_module.group_second_approver"/>
        <button name="%(action_reject_wizard)d" string="Reject"
                type="action" invisible="state not in ['submitted', 'first_approval']"/>
        <field name="state" widget="statusbar"
               statusbar_visible="draft,submitted,first_approval,approved"/>
    </header>
    <sheet>
        <div class="oe_title">
            <h1><field name="name"/></h1>
        </div>
        <group>
            <group>
                <field name="amount"/>
            </group>
            <group>
                <field name="submitted_by" invisible="not submitted_by"/>
                <field name="submitted_date" invisible="not submitted_date"/>
                <field name="first_approver_id" invisible="not first_approver_id"/>
                <field name="second_approver_id" invisible="not second_approver_id"/>
            </group>
        </group>
        <group string="Rejection" invisible="state != 'rejected'">
            <field name="rejection_reason"/>
        </group>
    </sheet>
    <div class="oe_chatter">
        <field name="message_follower_ids"/>
        <field name="activity_ids"/>
        <field name="message_ids"/>
    </div>
</form>
```

---

## Transition Validation

### Allowed Transitions Matrix
```python
class StateMachine(models.Model):
    _name = 'state.machine'

    # Define allowed transitions
    TRANSITIONS = {
        'draft': ['submitted', 'cancel'],
        'submitted': ['approved', 'rejected', 'cancel'],
        'approved': ['done', 'cancel'],
        'rejected': ['draft'],
        'done': [],
        'cancel': ['draft'],
    }

    state = fields.Selection([
        ('draft', 'Draft'),
        ('submitted', 'Submitted'),
        ('approved', 'Approved'),
        ('rejected', 'Rejected'),
        ('done', 'Done'),
        ('cancel', 'Cancelled'),
    ], default='draft')

    def _transition_to(self, new_state):
        """Safely transition to new state."""
        for record in self:
            allowed = self.TRANSITIONS.get(record.state, [])
            if new_state not in allowed:
                raise UserError(
                    f"Cannot transition from '{record.state}' to '{new_state}'. "
                    f"Allowed: {', '.join(allowed) or 'none'}"
                )
            record.state = new_state

    def action_submit(self):
        self._transition_to('submitted')

    def action_approve(self):
        self._transition_to('approved')

    def action_reject(self):
        self._transition_to('rejected')

    def action_done(self):
        self._transition_to('done')

    def action_cancel(self):
        self._transition_to('cancel')

    def action_draft(self):
        self._transition_to('draft')
```

### Transition with Pre/Post Hooks
```python
class DocumentWorkflow(models.Model):
    _name = 'document.workflow'

    state = fields.Selection([
        ('draft', 'Draft'),
        ('in_progress', 'In Progress'),
        ('review', 'Under Review'),
        ('done', 'Done'),
    ], default='draft')

    def _before_transition(self, from_state, to_state):
        """Hook called before state change."""
        # Validate transition is allowed
        # Check prerequisites
        pass

    def _after_transition(self, from_state, to_state):
        """Hook called after state change."""
        # Send notifications
        # Create related records
        # Update dependent fields
        pass

    def _do_transition(self, new_state):
        """Execute state transition with hooks."""
        for record in self:
            old_state = record.state
            record._before_transition(old_state, new_state)
            record.state = new_state
            record._after_transition(old_state, new_state)

    def action_start(self):
        """Start work on document."""
        for record in self:
            record._do_transition('in_progress')

    def action_submit_review(self):
        """Submit for review."""
        for record in self:
            record._do_transition('review')

    def action_complete(self):
        """Mark as complete."""
        for record in self:
            record._do_transition('done')
```

---

## Conditional State Transitions

### Amount-Based Approval
```python
class PurchaseRequest(models.Model):
    _name = 'purchase.request'

    amount_total = fields.Float()
    state = fields.Selection([
        ('draft', 'Draft'),
        ('pending', 'Pending Approval'),
        ('manager_approved', 'Manager Approved'),
        ('director_approved', 'Director Approved'),
        ('approved', 'Approved'),
        ('rejected', 'Rejected'),
    ], default='draft')

    # Approval thresholds
    MANAGER_LIMIT = 5000
    DIRECTOR_LIMIT = 20000

    def action_submit(self):
        """Submit based on amount."""
        for record in self:
            if record.amount_total <= self.MANAGER_LIMIT:
                # Auto-approve small amounts
                record.state = 'approved'
            elif record.amount_total <= self.DIRECTOR_LIMIT:
                # Needs manager approval
                record.state = 'pending'
                record._request_manager_approval()
            else:
                # Needs director approval
                record.state = 'pending'
                record._request_director_approval()

    def action_manager_approve(self):
        """Manager approval."""
        for record in self:
            if record.amount_total <= self.DIRECTOR_LIMIT:
                record.state = 'approved'
            else:
                record.state = 'manager_approved'
                record._request_director_approval()

    def action_director_approve(self):
        """Director approval."""
        self.write({'state': 'approved'})
```

---

## Parallel States (Kanban Stages)

### Stage-Based Workflow
```python
class Task(models.Model):
    _name = 'my.task'

    name = fields.Char(required=True)

    stage_id = fields.Many2one(
        'my.task.stage',
        string='Stage',
        group_expand='_read_group_stage_ids',
        tracking=True,
        default=lambda self: self._get_default_stage(),
    )

    # Computed state from stage
    state = fields.Selection(related='stage_id.state', store=True)

    def _get_default_stage(self):
        """Get first stage."""
        return self.env['my.task.stage'].search([], limit=1)

    @api.model
    def _read_group_stage_ids(self, stages, domain, order):
        """Always show all stages in kanban."""
        return self.env['my.task.stage'].search([])


class TaskStage(models.Model):
    _name = 'my.task.stage'
    _order = 'sequence, id'

    name = fields.Char(required=True)
    sequence = fields.Integer(default=10)
    fold = fields.Boolean(string='Folded in Kanban')

    state = fields.Selection([
        ('draft', 'Draft'),
        ('in_progress', 'In Progress'),
        ('done', 'Done'),
    ], default='draft')

    is_closed = fields.Boolean(string='Closing Stage')
```

### Kanban View
```xml
<kanban default_group_by="stage_id" class="o_kanban_small_column">
    <field name="stage_id"/>
    <field name="color"/>
    <progressbar field="state"
                 colors='{"draft": "secondary", "in_progress": "warning", "done": "success"}'/>
    <templates>
        <t t-name="kanban-box">
            <div t-attf-class="oe_kanban_card oe_kanban_global_click">
                <div class="oe_kanban_content">
                    <field name="name"/>
                </div>
            </div>
        </t>
    </templates>
</kanban>
```

---

## State Change Tracking

### With Mail Thread
```python
class TrackedDocument(models.Model):
    _name = 'tracked.document'
    _inherit = ['mail.thread', 'mail.activity.mixin']

    name = fields.Char()
    state = fields.Selection([
        ('draft', 'Draft'),
        ('confirmed', 'Confirmed'),
        ('done', 'Done'),
    ], default='draft', tracking=True)  # tracking=True logs changes

    # Custom tracking message
    def action_confirm(self):
        for record in self:
            record.state = 'confirmed'
            record.message_post(
                body="Document confirmed.",
                subtype_xmlid='mail.mt_note',
            )
```

### Activity Scheduling
```python
def action_submit_for_approval(self):
    """Submit and create approval activity."""
    self.ensure_one()
    self.state = 'pending_approval'

    # Create activity for approver
    approver = self._get_approver()
    self.activity_schedule(
        'mail.mail_activity_data_todo',
        user_id=approver.id,
        summary='Approval Required',
        note=f'Please review and approve: {self.name}',
    )
```

---

## Scheduled State Transitions

### Auto-Close After Deadline
```python
class AutoCloseDocument(models.Model):
    _name = 'auto.close.document'

    state = fields.Selection([
        ('active', 'Active'),
        ('expired', 'Expired'),
    ], default='active')

    expiry_date = fields.Date()

    @api.model
    def _cron_check_expiry(self):
        """Scheduled action to expire documents."""
        expired = self.search([
            ('state', '=', 'active'),
            ('expiry_date', '<', fields.Date.today()),
        ])
        expired.write({'state': 'expired'})

        # Optional: notify owners
        for doc in expired:
            doc._notify_expiry()
```

### Cron Job Definition
```xml
<record id="ir_cron_check_document_expiry" model="ir.cron">
    <field name="name">Check Document Expiry</field>
    <field name="model_id" ref="model_auto_close_document"/>
    <field name="state">code</field>
    <field name="code">model._cron_check_expiry()</field>
    <field name="interval_number">1</field>
    <field name="interval_type">days</field>
    <field name="numbercall">-1</field>
</record>
```

---

## Best Practices

1. **Use Selection for state** - Clear, finite set of states
2. **Default to 'draft'** - Start in editable state
3. **Track state changes** - Use tracking=True with mail.thread
4. **Validate transitions** - Check current state before changing
5. **Use statusbar widget** - Visual representation of progress
6. **Restrict field editing** - Lock fields in certain states
7. **Log transitions** - Keep audit trail of state changes
8. **Handle cancellation** - Always provide cancel path
9. **Reset to draft** - Allow re-processing of cancelled items
10. **Test all paths** - Verify every transition works correctly

---


## Source: wizard-patterns.md

# Wizard and Transient Model Patterns

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  WIZARD PATTERNS                                                             ║
║  Complete reference for transient models and wizard implementation           ║
║  Use for user interactions, batch operations, and confirmation dialogs       ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## Overview

Wizards (TransientModel) are temporary records that:
- Auto-delete after a period (vacuum)
- Don't persist permanently in database
- Perfect for user dialogs and batch operations
- Support multi-record operations

---

## Basic Wizard Structure

### Model Definition (v18)
```python
from odoo import api, fields, models
from odoo.exceptions import UserError


class MyWizard(models.TransientModel):
    _name = 'my.module.wizard'
    _description = 'My Wizard'

    # Context-dependent fields
    model_id = fields.Many2one(
        comodel_name='my.model',
        string='Record',
        default=lambda self: self.env.context.get('active_id'),
    )
    model_ids = fields.Many2many(
        comodel_name='my.model',
        string='Records',
        default=lambda self: self.env.context.get('active_ids'),
    )

    # Wizard-specific fields
    date = fields.Date(
        string='Date',
        default=fields.Date.today,
        required=True,
    )
    note = fields.Text(string='Notes')

    def action_confirm(self) -> dict:
        """Execute wizard action."""
        self.ensure_one()

        if not self.model_id:
            raise UserError("No record selected.")

        # Perform operation
        self.model_id.write({
            'date': self.date,
            'notes': self.note,
        })

        return {'type': 'ir.actions.act_window_close'}
```

### View Definition
```xml
<?xml version="1.0" encoding="utf-8"?>
<odoo>
    <record id="my_wizard_view_form" model="ir.ui.view">
        <field name="name">my.module.wizard.form</field>
        <field name="model">my.module.wizard</field>
        <field name="arch" type="xml">
            <form string="My Wizard">
                <group>
                    <field name="model_id" readonly="1"
                           invisible="not model_id"/>
                    <field name="date"/>
                    <field name="note"/>
                </group>
                <footer>
                    <button name="action_confirm"
                            string="Confirm"
                            type="object"
                            class="btn-primary"/>
                    <button string="Cancel"
                            class="btn-secondary"
                            special="cancel"/>
                </footer>
            </form>
        </field>
    </record>

    <record id="my_wizard_action" model="ir.actions.act_window">
        <field name="name">My Wizard</field>
        <field name="res_model">my.module.wizard</field>
        <field name="view_mode">form</field>
        <field name="target">new</field>
        <field name="binding_model_id" ref="my_module.model_my_model"/>
        <field name="binding_view_types">form,list</field>
    </record>
</odoo>
```

### Security (ir.model.access.csv)
```csv
id,name,model_id:id,group_id:id,perm_read,perm_write,perm_create,perm_unlink
access_my_wizard_user,my.module.wizard.user,model_my_module_wizard,base.group_user,1,1,1,1
```

---

## Common Wizard Patterns

### 1. Confirmation Dialog

```python
class ConfirmationWizard(models.TransientModel):
    _name = 'my.confirm.wizard'
    _description = 'Confirmation Dialog'

    message = fields.Text(
        string='Message',
        readonly=True,
        default=lambda self: self._default_message(),
    )

    @api.model
    def _default_message(self) -> str:
        active_ids = self.env.context.get('active_ids', [])
        count = len(active_ids)
        return f"Are you sure you want to process {count} record(s)?"

    def action_confirm(self) -> dict:
        """Process confirmed action."""
        active_ids = self.env.context.get('active_ids', [])
        records = self.env['my.model'].browse(active_ids)

        for record in records:
            record.action_process()

        return {'type': 'ir.actions.act_window_close'}
```

```xml
<form string="Confirm">
    <group>
        <field name="message" nolabel="1"/>
    </group>
    <footer>
        <button name="action_confirm" string="Yes, Proceed"
                type="object" class="btn-primary"/>
        <button string="Cancel" class="btn-secondary" special="cancel"/>
    </footer>
</form>
```

### 2. Batch Update Wizard

```python
class BatchUpdateWizard(models.TransientModel):
    _name = 'my.batch.update.wizard'
    _description = 'Batch Update'

    record_ids = fields.Many2many(
        comodel_name='my.model',
        string='Records to Update',
        default=lambda self: self._default_records(),
    )
    new_state = fields.Selection(
        selection=[
            ('draft', 'Draft'),
            ('confirmed', 'Confirmed'),
            ('done', 'Done'),
        ],
        string='New Status',
        required=True,
    )
    update_date = fields.Boolean(
        string='Update Date',
        default=False,
    )
    date = fields.Date(
        string='New Date',
    )

    @api.model
    def _default_records(self) -> 'my.model':
        active_ids = self.env.context.get('active_ids', [])
        return self.env['my.model'].browse(active_ids)

    def action_update(self) -> dict:
        """Apply batch update."""
        vals = {'state': self.new_state}

        if self.update_date and self.date:
            vals['date'] = self.date

        self.record_ids.write(vals)

        return {
            'type': 'ir.actions.client',
            'tag': 'display_notification',
            'params': {
                'title': 'Success',
                'message': f'Updated {len(self.record_ids)} records.',
                'type': 'success',
                'sticky': False,
            }
        }
```

### 3. Report Wizard (Date Range)

```python
class ReportWizard(models.TransientModel):
    _name = 'my.report.wizard'
    _description = 'Report Parameters'

    date_from = fields.Date(
        string='From Date',
        required=True,
        default=lambda self: fields.Date.today().replace(day=1),
    )
    date_to = fields.Date(
        string='To Date',
        required=True,
        default=fields.Date.today,
    )
    partner_ids = fields.Many2many(
        comodel_name='res.partner',
        string='Partners',
        help='Leave empty for all partners',
    )
    output_format = fields.Selection(
        selection=[
            ('pdf', 'PDF'),
            ('xlsx', 'Excel'),
        ],
        string='Format',
        default='pdf',
        required=True,
    )

    @api.constrains('date_from', 'date_to')
    def _check_dates(self):
        for wizard in self:
            if wizard.date_from > wizard.date_to:
                raise UserError("Start date must be before end date.")

    def action_print(self) -> dict:
        """Generate report."""
        domain = [
            ('date', '>=', self.date_from),
            ('date', '<=', self.date_to),
        ]
        if self.partner_ids:
            domain.append(('partner_id', 'in', self.partner_ids.ids))

        records = self.env['my.model'].search(domain)

        if self.output_format == 'pdf':
            return self.env.ref('my_module.report_my_model').report_action(records)
        else:
            return self._export_xlsx(records)

    def _export_xlsx(self, records) -> dict:
        """Export to Excel."""
        # Implementation for Excel export
        pass
```

### 4. Import Wizard

```python
class ImportWizard(models.TransientModel):
    _name = 'my.import.wizard'
    _description = 'Import Data'

    file = fields.Binary(
        string='File',
        required=True,
        attachment=False,
    )
    filename = fields.Char(string='Filename')
    skip_errors = fields.Boolean(
        string='Skip Errors',
        default=False,
        help='Continue import even if some rows fail',
    )

    def action_import(self) -> dict:
        """Import data from file."""
        import base64
        import csv
        from io import StringIO

        if not self.file:
            raise UserError("Please select a file.")

        # Decode file
        file_content = base64.b64decode(self.file).decode('utf-8')
        reader = csv.DictReader(StringIO(file_content))

        created_count = 0
        error_count = 0
        errors = []

        for row_num, row in enumerate(reader, start=2):
            try:
                self._create_record(row)
                created_count += 1
            except Exception as e:
                error_count += 1
                errors.append(f"Row {row_num}: {str(e)}")
                if not self.skip_errors:
                    raise UserError(f"Error on row {row_num}: {str(e)}")

        message = f"Created {created_count} records."
        if error_count:
            message += f" {error_count} errors."

        return {
            'type': 'ir.actions.client',
            'tag': 'display_notification',
            'params': {
                'title': 'Import Complete',
                'message': message,
                'type': 'success' if not error_count else 'warning',
            }
        }

    def _create_record(self, row: dict) -> 'my.model':
        """Create record from row data."""
        return self.env['my.model'].create({
            'name': row['name'],
            'code': row.get('code'),
        })
```

### 5. Selection Wizard (Multi-step)

```python
class SelectionWizard(models.TransientModel):
    _name = 'my.selection.wizard'
    _description = 'Selection Wizard'

    step = fields.Selection(
        selection=[
            ('select', 'Selection'),
            ('configure', 'Configuration'),
            ('confirm', 'Confirmation'),
        ],
        string='Step',
        default='select',
    )

    # Step 1: Selection
    template_id = fields.Many2one(
        comodel_name='my.template',
        string='Template',
    )

    # Step 2: Configuration
    name = fields.Char(string='Name')
    date = fields.Date(string='Date')

    # Step 3: Confirmation
    summary = fields.Text(
        string='Summary',
        compute='_compute_summary',
    )

    @api.depends('template_id', 'name', 'date')
    def _compute_summary(self):
        for wizard in self:
            wizard.summary = f"""
Template: {wizard.template_id.name or 'None'}
Name: {wizard.name or 'Not set'}
Date: {wizard.date or 'Not set'}
            """

    def action_next(self) -> dict:
        """Go to next step."""
        self.ensure_one()

        if self.step == 'select':
            if not self.template_id:
                raise UserError("Please select a template.")
            self.step = 'configure'
        elif self.step == 'configure':
            if not self.name:
                raise UserError("Please enter a name.")
            self.step = 'confirm'

        return self._reopen()

    def action_previous(self) -> dict:
        """Go to previous step."""
        self.ensure_one()

        if self.step == 'configure':
            self.step = 'select'
        elif self.step == 'confirm':
            self.step = 'configure'

        return self._reopen()

    def action_create(self) -> dict:
        """Create record from wizard."""
        self.ensure_one()

        record = self.env['my.model'].create({
            'template_id': self.template_id.id,
            'name': self.name,
            'date': self.date,
        })

        return {
            'type': 'ir.actions.act_window',
            'res_model': 'my.model',
            'res_id': record.id,
            'view_mode': 'form',
            'target': 'current',
        }

    def _reopen(self) -> dict:
        """Reopen wizard at current state."""
        return {
            'type': 'ir.actions.act_window',
            'res_model': self._name,
            'res_id': self.id,
            'view_mode': 'form',
            'target': 'new',
        }
```

---

## Wizard Action Return Types

### Close Wizard
```python
return {'type': 'ir.actions.act_window_close'}
```

### Notification
```python
return {
    'type': 'ir.actions.client',
    'tag': 'display_notification',
    'params': {
        'title': 'Success',
        'message': 'Operation completed.',
        'type': 'success',  # success, warning, danger, info
        'sticky': False,
    }
}
```

### Open Record
```python
return {
    'type': 'ir.actions.act_window',
    'res_model': 'my.model',
    'res_id': record_id,
    'view_mode': 'form',
    'target': 'current',  # current, new, inline
}
```

### Open List
```python
return {
    'type': 'ir.actions.act_window',
    'name': 'Created Records',
    'res_model': 'my.model',
    'view_mode': 'tree,form',
    'domain': [('id', 'in', record_ids)],
    'target': 'current',
}
```

### Download Report
```python
return self.env.ref('my_module.report_action').report_action(records)
```

### Reload Page
```python
return {
    'type': 'ir.actions.client',
    'tag': 'reload',
}
```

---

## Binding to Models

### From Action Definition
```xml
<record id="my_wizard_action" model="ir.actions.act_window">
    <field name="name">My Wizard</field>
    <field name="res_model">my.wizard</field>
    <field name="view_mode">form</field>
    <field name="target">new</field>
    <!-- Bind to specific model's Action menu -->
    <field name="binding_model_id" ref="model_my_model"/>
    <field name="binding_view_types">form,list</field>
</record>
```

### From Python (Server Action)
```python
def action_open_wizard(self) -> dict:
    """Open wizard from button."""
    return {
        'type': 'ir.actions.act_window',
        'name': 'My Wizard',
        'res_model': 'my.wizard',
        'view_mode': 'form',
        'target': 'new',
        'context': {
            'default_model_id': self.id,
            'default_model_ids': self.ids,
            'active_id': self.id,
            'active_ids': self.ids,
            'active_model': self._name,
        },
    }
```

---

## Version-Specific Notes

### v17+ View Syntax
```xml
<!-- Use inline expressions -->
<field name="date" invisible="not update_date"/>
<button name="action_next" invisible="step == 'confirm'"/>
```

### v18+ Type Hints
```python
def action_confirm(self) -> dict:
    """Execute with type hints."""
    self.ensure_one()
    return {'type': 'ir.actions.act_window_close'}

@api.model
def _default_records(self) -> 'my.model':
    return self.env['my.model'].browse(
        self.env.context.get('active_ids', [])
    )
```

---

## Best Practices

1. **Always define security** - Wizards need ir.model.access.csv entries
2. **Use context defaults** - Pass active_id/active_ids through context
3. **Validate input** - Use @api.constrains and raise UserError
4. **Handle empty selection** - Check if records exist before processing
5. **Provide feedback** - Return notification or open result view
6. **Clean UI** - Use footer for buttons, proper grouping
7. **Multi-record support** - Use Many2many for batch operations

---

