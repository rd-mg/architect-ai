> **IMPORTANT (WORKSPACE RULE):** This reference explicitly targets older Odoo versions. Remember that for **Odoo 19.0**, you MUST use `<list>` (not `<tree>`), `<chatter/>`, `_compute_display_name`, and avoid `attrs`. For **Odoo 18.0**, use direct attributes. See `.github/copilot-instructions.md` for strict version variations, and consult Context7 documentation (https://context7.com/...) as your PREFERRED source of truth. Use local workspace assets only as a SECONDARY fallback.


# Odoo Models Reference

## Table of Contents
- [Model Types](#model-types)
- [Field Types](#field-types)
- [Computed Fields](#computed-fields)
- [Constraints](#constraints)
- [Inheritance Patterns](#inheritance-patterns)
- [CRUD Overrides](#crud-overrides)
- [Domain Filters](#domain-filters)
- [Common Mixins](#common-mixins)

## Model Types

### Regular Model (`models.Model`)

Persistent, stored in a database table.

```python
from odoo import models, fields, api

class LibraryBook(models.Model):
    _name = 'library.book'
    _description = 'Library Book'
    _order = 'name'
    _rec_name = 'name'

    name = fields.Char(string='Title', required=True)
    isbn = fields.Char(string='ISBN')
    active = fields.Boolean(default=True)
    date_published = fields.Date(string='Published Date')
    author_ids = fields.Many2many('res.partner', string='Authors')
    category_id = fields.Many2one('library.book.category', string='Category')
    state = fields.Selection([
        ('draft', 'Draft'),
        ('available', 'Available'),
        ('borrowed', 'Borrowed'),
        ('lost', 'Lost'),
    ], string='State', default='draft', required=True)
```

### Transient Model (`models.TransientModel`)

Temporary records, auto-cleaned. Used for wizards (see wizards.md).

### Abstract Model (`models.AbstractModel`)

Not stored in DB, used as mixin base classes.

```python
class MailTracking(models.AbstractModel):
    _name = 'mail.tracking.mixin'
    _description = 'Mail Tracking Mixin'

    tracking_field = fields.Char()
```

## Field Types

### Basic Fields

| Field | Usage | Key Attributes |
|-------|-------|----------------|
| `Char` | Short text | `size`, `trim`, `translate` |
| `Text` | Long text | `translate` |
| `Html` | Rich text | `sanitize`, `translate` |
| `Integer` | Whole numbers | — |
| `Float` | Decimal numbers | `digits=(precision, scale)` |
| `Monetary` | Currency amounts | `currency_field='currency_id'` |
| `Boolean` | True/False | — |
| `Date` | Date only | — |
| `Datetime` | Date + time | — |
| `Selection` | Dropdown | `selection=[('key', 'Label'), ...]` |
| `Binary` | File/image storage | `attachment=True` |
| `Image` | Image with resize | `max_width`, `max_height` |

### Relational Fields

```python
# Many2one — FK to another model
partner_id = fields.Many2one('res.partner', string='Customer', ondelete='restrict')

# One2many — Reverse of Many2one (virtual)
line_ids = fields.One2many('sale.order.line', 'order_id', string='Order Lines')

# Many2many — Junction table
tag_ids = fields.Many2many(
    'project.tags',
    'project_task_tag_rel',    # relation table name (optional)
    'task_id',                  # column for this model (optional)
    'tag_id',                   # column for comodel (optional)
    string='Tags',
)
```

### Common Field Attributes

```python
name = fields.Char(
    string='Label',           # UI label
    required=True,            # NOT NULL
    readonly=False,           # editable
    index=True,               # DB index
    default='New',            # default value or lambda
    help='Tooltip text',      # UI help text
    copy=True,                # copied on duplicate
    tracking=True,            # chatter tracking (requires mail.thread)
    groups='base.group_user', # access restriction
    company_dependent=True,   # per-company value (Odoo 17)
)
```

### `ondelete` options for Many2one

| Value | Behavior |
|-------|----------|
| `'set null'` | Set field to NULL (default) |
| `'restrict'` | Prevent deletion if referenced |
| `'cascade'` | Delete this record too |

## Computed Fields

```python
total = fields.Float(compute='_compute_total', store=True)

@api.depends('line_ids.price', 'line_ids.quantity')
def _compute_total(self):
    for record in self:
        record.total = sum(line.price * line.quantity for line in record.line_ids)
```

### Inverse (editable computed field)

```python
total = fields.Float(compute='_compute_total', inverse='_inverse_total', store=True)

def _inverse_total(self):
    for record in self:
        # logic to set dependent fields from total
        pass
```

### Search method (for non-stored computed fields)

```python
display_name = fields.Char(compute='_compute_display_name', search='_search_display_name')

def _search_display_name(self, operator, value):
    return [('name', operator, value)]
```

## Constraints

### Python Constraints

```python
from odoo.exceptions import ValidationError

@api.constrains('date_start', 'date_end')
def _check_dates(self):
    for record in self:
        if record.date_start and record.date_end and record.date_start > record.date_end:
            raise ValidationError("End date must be after start date.")
```

### SQL Constraints

```python
_sql_constraints = [
    ('isbn_unique', 'UNIQUE(isbn)', 'ISBN must be unique!'),
    ('price_positive', 'CHECK(price >= 0)', 'Price must be positive!'),
]
```

## Inheritance Patterns

### Class Inheritance (extend existing model)

Add fields/methods to an existing model without creating a new table.

```python
class ResPartner(models.Model):
    _inherit = 'res.partner'

    loyalty_points = fields.Integer(string='Loyalty Points', default=0)
    membership_date = fields.Date(string='Membership Date')
```

### Prototype Inheritance (copy features)

Create a new model copying fields from an existing model.

```python
class LibraryMember(models.Model):
    _name = 'library.member'
    _inherits = {'res.partner': 'partner_id'}

    partner_id = fields.Many2one('res.partner', required=True, ondelete='cascade')
    member_number = fields.Char(string='Member Number')
```

### Delegation Inheritance (`_inherits`)

Use `_inherits` when the new model IS-A variant of the parent (composition with delegation). The child record automatically delegates reads/writes to the parent record's fields.

## CRUD Overrides

```python
@api.model_create_multi
def create(self, vals_list):
    for vals in vals_list:
        if vals.get('name', 'New') == 'New':
            vals['name'] = self.env['ir.sequence'].next_by_code('library.book') or 'New'
    return super().create(vals_list)

def write(self, vals):
    # pre-write logic
    result = super().write(vals)
    # post-write logic
    return result

def unlink(self):
    for record in self:
        if record.state != 'draft':
            raise UserError("Cannot delete a non-draft record.")
    return super().unlink()
```

### `@api.model` vs `@api.model_create_multi`

- Odoo 17: Use `@api.model_create_multi` for `create()` (receives list of dicts).
- Odoo 16: `@api.model_create_multi` also supported; `create()` with single dict is deprecated.

## Domain Filters

Domains are lists of tuples `(field, operator, value)`:

```python
# Simple domain
[('state', '=', 'available')]

# AND (implicit)
[('state', '=', 'available'), ('author_ids', '!=', False)]

# OR
['|', ('state', '=', 'available'), ('state', '=', 'draft')]

# NOT
['!', ('active', '=', False)]

# Common operators: =, !=, >, <, >=, <=, like, ilike, in, not in, child_of, parent_of
```

## Common Mixins

### mail.thread (Chatter)

```python
class LibraryBook(models.Model):
    _name = 'library.book'
    _inherit = ['mail.thread', 'mail.activity.mixin']

    name = fields.Char(tracking=True)
    state = fields.Selection([...], tracking=True)
```

### mail.activity.mixin

Adds activity scheduling (To-do, Call, Meeting, etc.) to the form view.

### image.mixin

Provides `image_1920`, `image_1024`, `image_512`, `image_256`, `image_128` auto-resized image fields.

### portal.mixin

Enables portal access for external users (customers, vendors).
