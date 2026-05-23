
---
name: coding-style-cudio-append
project_scope: cudio
---

## Cudio Coding Standards Appendix

Additional standards for Cudio Inc. projects. EXTEND base coding-style rules.

### Hardcoded Values

NEVER hardcode configurable values. Use:
- `res.config.settings` for user-facing global settings
- `ir.config_parameter` for technical/system settings
- Model-level configuration (e.g., `product.template` defaults)
- Module-level Python constants for values that NEVER change between deployments

Anti-pattern:
```python
# BAD — hardcoded threshold
if order.amount_total > 10000:
    require_approval()
```

Correct:
```python
# GOOD — configurable via system parameter
threshold = float(self.env['ir.config_parameter'].sudo().get_param(
    'acme_sale.approval_threshold', default='10000'
))
if order.amount_total > threshold:
    require_approval()
```

### XPath Class Selectors

ALWAYS use `hasclass()` for class-based selectors. NEVER `contains(@class, ...)`.

Anti-pattern:
```xml
<xpath expr="//div[contains(@class, 'o_form_view')]" position="inside">
```

Correct:
```xml
<xpath expr="//div[hasclass('o_form_view')]" position="inside">
```

### Super() Calls

Use Python 3 `super()` syntax. NOT `super(MyClass, self)`.

Anti-pattern:
```python
def create(self, vals_list):
    return super(MyModel, self).create(vals_list)
```

Correct:
```python
def create(self, vals_list):
    return super().create(vals_list)
```

### Python Encoding Declaration

Do NOT include `# -*- coding: utf-8 -*-`. Obsolete in Python 3.

### Field Ordering in Model Definitions

Within model class, order fields:
1. Constant fields (Selection options as class-level constants)
2. Inherited fields (parent field modifications)
3. New Char/Text/Html fields
4. New Integer/Float/Monetary fields
5. New Boolean fields
6. New Date/Datetime fields
7. New Selection/Enum fields
8. New Many2one fields
9. New One2many fields
10. New Many2many fields
11. Computed fields
12. Related fields
13. Property fields

### Method Ordering in Model Classes

1. Constants and class attributes
2. Default methods (`_default_*`)
3. Compute methods (`_compute_*`)
4. Search methods (`_search_*`)
5. Constraint methods (`_check_*`, `@api.constrains`)
6. Onchange methods (`@api.onchange`)
7. CRUD overrides (`create`, `write`, `unlink`, `copy`)
8. Action methods (`action_*`)
9. Business methods (private `_` prefix first, then public)

### Docstrings

Every public method MUST have docstring:

```python
def action_approve(self):
    """Approve the current request.

    Transitions state from 'pending' to 'approved' and triggers
    the downstream workflow via the approval.workflow engine.

    :returns: action dictionary to reload the view
    :raises UserError: if current user lacks approval rights
    """
    ...
```

### String Formatting

Use f-strings for simple interpolation, `.format()` for complex:

```python
# Simple: f-string
msg = f"Order {order.name} requires approval"

# Complex with formatting: .format()
report = "Total: {amount:,.2f} {currency}".format(
    amount=order.amount_total,
    currency=order.currency_id.name
)

# Never use % formatting in new code
```

### Module-Level Constants

UPPER_SNAKE_CASE at top of file, after imports:

```python
from odoo import models, fields, api

APPROVAL_STATES = [
    ('draft', 'Draft'),
    ('pending', 'Pending Approval'),
    ('approved', 'Approved'),
    ('rejected', 'Rejected'),
]

DEFAULT_THRESHOLD = 10000.0

class ApprovalRequest(models.Model):
    _name = 'acme.approval.request'
    ...
```

### Version-Specific Syntax Enforcement

Odoo 17+:
- Direct `invisible=`, `readonly=`, `required=` on fields, NEVER `attrs=`
- Avoid `name_get` override — use `_compute_display_name`

Odoo 18+:
- `<list>` not `<tree>` in XML views
- `<chatter/>` not manual chatter structure

Odoo 19+:
- `SQL()` builder mandatory for raw SQL
- Type hints required on all public methods
- OWL 3.x patterns (no OWL 2.x deprecated syntax)
