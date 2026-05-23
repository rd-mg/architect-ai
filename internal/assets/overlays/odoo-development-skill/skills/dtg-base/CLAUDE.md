# DTG Base Dev Guide

AI guidance for DTG Base utilities in Odoo 18.

## What is DTG Base?

Custom abstract model (`dtg_base.DTGBase`) providing common utility methods for Odoo dev at DTG. Inherited by other models.

## Location

**Module:** `addons_customs/erp/dtg_base/`
**Main Model:** `dtg_base/models/dtg_base.py`

## When to Use DTG Base

| Task | Method |
|------|--------|
| First date of month/quarter/year | `find_first_date_of_period(date, 'month')` |
| Last date of month/quarter/year | `find_last_date_of_period(date, 'year')` |
| Local datetime to UTC | `convert_local_to_utc(local_dt, 'Asia/Ho_Chi_Minh')` |
| UTC to local datetime | `convert_utc_to_local(utc_dt, 'Asia/Ho_Chi_Minh')` |
| Barcode exists check | `barcode_exists('1234567890123')` |
| Generate EAN13 | `get_ean13('product_code')` |
| Batch process large recordsets | `splittor(limit=100)` |
| Remove Vietnamese accents | `strip_accents('Tiếng Việt')` |
| Zip directory | `zip_dir(source_path, output_path)` |
| File size readable | `_get_file_size(file_path)` |

## Inheritance

```python
from odoo import models

class MyModel(models.Model):
    _name = 'my.model'
    _inherit = ['dtg_base.dtg_base']

    def process_records(self):
        for batch in self.splittor(limit=100):
            pass
```

## Key Utilities

### Date & Period
```python
first_date = self.find_first_date_of_period(fields.Date.today(), 'month')
last_date = self.find_last_date_of_period(fields.Date.today(), 'quarter')
for start, end in self.period_iter('2024-01-01', '2024-12-31', 'month'):
    print(f"Period: {start} to {end}")
```

### Timezone
```python
utc_dt = self.convert_local_to_utc('2024-01-15 10:00:00', 'Asia/Ho_Chi_Minh')
local_dt = self.convert_utc_to_local(utc_dt, 'Asia/Ho_Chi_Minh')
```

### Batch Processing
```python
records = self.env['my.model'].search([])
for batch in records.splittor(limit=100):
    for record in batch:
        pass
```

### Barcode
```python
if self.barcode_exists('1234567890123'):
    raise UserError("Barcode already exists!")
ean13 = self.get_ean13('PRODUCT123')
```

### Vietnamese Text
```python
search_text = self.strip_accents('Tiếng Việt')  # -> 'Ties Viet'
if self.strip_accents(record.name) == self.strip_accents(search_term):
    pass
```

## Period Types

| Type | Description |
|------|-------------|
| `'month'` | Month |
| `'quarter'` | Quarter |
| `'year'` | Year |
| `'week'` | Week |

## Common Timezones

| Timezone | Offset |
|----------|--------|
| `'Asia/Ho_Chi_Minh'` | UTC+7 |
| `'UTC'` | UTC+0 |
| `'Asia/Bangkok'` | UTC+7 |
| `'Asia/Singapore'` | UTC+8 |

---

**Full ref:** [odoo-18-dtg-base-guide.md](./odoo-18-dtg-base-guide.md)
