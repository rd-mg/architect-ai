---
name: odoo-18-dtg-base
description: Complete reference for DTG Base module utilities and helpers. DTGBase is an abstract model providing common utility methods for date/time handling, barcode generation, timezone conversion, file operations, and more.
globs: "**/addons_customs/erp/**/*.py"
topics:
  - DTGBase abstract model inheritance
  - Date/Period utilities (find_first_date_of_period, find_last_date_of_period, period_iter)
  - Timezone conversion (convert_local_to_utc, convert_utc_to_local)
  - Barcode utilities (barcode_exists, get_ean13)
  - Batch processing (splittor)
  - after_commit decorator
  - Vietnamese text utilities (strip_accents, _no_accent_vietnamese)
  - File utilities (zip_dir, zip_dirs, _get_file_size)
when_to_use:
  - Working with DTG Odoo codebase
  - Date/period calculations
  - Timezone conversions
  - Barcode validation
  - Batch processing large recordsets
  - Vietnamese text processing
---

# Odoo 18 DTG Base Guide

Complete reference for DTG Base module utilities and helpers.

## TOC

1. [DTGBase Abstract Model](#dtgbase-abstract-model)
2. [Date & Period Utilities](#date--period-utilities)
3. [Timezone Conversion](#timezone-conversion)
4. [Barcode Utilities](#barcode-utilities)
5. [Batch Processing](#batch-processing)
6. [after_commit Decorator](#after_commit-decorator)
7. [String & Text Utilities](#string--text-utilities)
8. [File Utilities](#file-utilities)
9. [Number Utilities](#number-utilities)

---

## DTGBase Abstract Model

### Inherit from DTGBase

**Location:** `addons_customs/erp/dtg_base/models/dtg_base.py`

```python
from odoo import models, fields

class MyModel(models.Model):
    _name = 'my.model'
    _inherit = ['dtg.base']  # Inherit DTGBase for all utilities
    name = fields.Char()
```

**When to inherit:** Date/period calcs, timezone conversion, barcode validation/generation, batch processing with memory mgmt, Vietnamese text, file zipping.

---

## Date & Period Utilities

### Period Names

Supported: `'hourly'`, `'daily'`, `'weekly'`, `'monthly'`, `'quarterly'`, `'biannually'`, `'annually'`
Aliases: `'hour'`, `'day'`, `'week'`, `'month'`, `'quarter'`, `'biannual'`, `'year'`, `'annual'`

### find_first_date_of_period()

```python
date = fields.Date.to_date('2024-02-15')
first_day = self.find_first_date_of_period('monthly', date)       # datetime(2024,2,1,0,0,0)
first_week = self.find_first_date_of_period('weekly', date)       # datetime(2024,2,12,0,0,0) — Mon
first_quarter = self.find_first_date_of_period('quarterly', date) # datetime(2024,1,1,0,0,0)
first_offset = self.find_first_date_of_period('monthly', date, start_day_offset=5) # datetime(2024,2,6,0,0,0)
```

### find_last_date_of_period()

```python
date = fields.Date.to_date('2024-02-15')
last_day = self.find_last_date_of_period('monthly', date)                    # datetime(2024,2,29,23,59,59,999999) leap
last_quarter = self.find_last_date_of_period('quarterly', date)              # datetime(2024,3,31,23,59,59,999999)
last_from_start = self.find_last_date_of_period('monthly', start_date, date_is_start_date=True)  # datetime(2024,2,29,23,59,59,999999)
last_2mo = self.find_last_date_of_period('monthly', date, cycle_value=2)     # datetime(2024,3,31,23,59,59,999999)
```

### period_iter()

```python
dt_start, dt_end = fields.Date.to_date('2024-01-15'), fields.Date.to_date('2024-06-20')
dates = self.period_iter('monthly', dt_start, dt_end)
# [date(2024,1,15), date(2024,1,31), date(2024,2,29), date(2024,3,31), date(2024,4,30), date(2024,5,31), date(2024,6,20)]

quarterly = self.period_iter('quarterly', dt_start, dt_end, start_day_offset=5)
```

### Date Difference

```python
days = self.get_days_between_dates(date_from, date_to)
hours = self.get_hours_between_dates(datetime_from, datetime_to)
weeks = self.get_weeks_between_dates(date_from, date_to)
months = self.get_months_between_dates(date_from, date_to)   # Jan 15 to Feb 14 = 0.9677
years = self.get_number_of_years_between_dates(date_from, date_to)
days_in_month = self.get_days_of_month_from_date(date)
day_of_year = self.get_day_of_year_from_date(date)            # 1-366
days_in_year = self.get_days_in_year(date)                    # 365/366
```

### Other Date

```python
year, month, day = self.split_date(date)
next_monday = self.next_weekday(date, weekday=0)              # 0=Mon, 6=Sun
same_wkday = self.next_weekday(date)                          # same weekday next week
# Break time range at midnight: [2024-02-02 20:00, 2024-02-03 00:00, 2024-02-03 04:00]
intervals = self.break_timerange_for_midnight(start_dt, end_dt)
```

### Period Ratio

```python
ratio = self.get_ratio_between_periods('monthly', 1, 'daily', 1, given_date=date(2024,2,1))  # 29/7
ratio = self.get_ratio_between_periods('quarterly', 1, 'monthly', 1)                         # 3.0
```

---

## Timezone Conversion

### get_company_tz()

```python
tz = self.get_company_tz()                             # 'Asia/Ho_Chi_Minh'/'UTC'/company tz
tz = self.get_company_tz(company=company_record)
```

### convert_local_to_utc()

```python
local_dt = datetime(2024,2,15,14,30,0)
utc_dt = self.convert_local_to_utc(local_dt, force_local_tz_name='Asia/Ho_Chi_Minh')  # datetime(2024,2,15,7,30,0)
utc_dt = self.convert_local_to_utc(local_dt)
utc_naive = self.convert_local_to_utc(local_dt, naive=True)                           # without tzinfo
utc_from_date = self.convert_local_to_utc(date(2024,2,15))
```

### convert_utc_to_local()

```python
utc_dt = datetime(2024,2,15,7,30,0)
local_dt = self.convert_utc_to_local(utc_dt, force_local_tz_name='Asia/Ho_Chi_Minh')  # datetime(2024,2,15,14,30,0)
local_dt = self.convert_utc_to_local(utc_dt, is_dst=False)
```

### Time Conversion

```python
float_hours = self.time_to_float_hour(datetime)        # datetime(2024,1,1,14,30,0) -> 14.5
time_obj = self.float_hours_to_time(14.5)               # 14.5 -> time(14,30,0)
time_str = self.hours_time_string(14.5)                 # "14:30"
dt = self.date_to_datetime(date_value)
```

---

## Barcode Utilities

### barcode_exists()

```python
exists = self.barcode_exists('8901234567890')
exists = self.barcode_exists('8901234567890', model_name='product.product')
exists = self.barcode_exists('8901234567890', barcode_field='default_code')
exists = self.barcode_exists('8901234567890', inactive_rec=True)
```

### get_ean13()

```python
barcode = self.get_ean13('123456789012')  # '1234567890128' — last digit checksum
barcode = self.get_ean13('123')            # '000000000123X' — padded + checksum
```

---

## Batch Processing

### splittor()

```python
# Default PREFETCH_MAX (1000)
for batch in self.splittor(large_recordset):
    batch.compute_expensive_field()

# Custom size
for batch in self.splittor(large_recordset, max_rec_in_batch=500):
    batch.write({'field': value})

# Maintain order
for batch in self.splittor(recordset, max_rec_in_batch=100, maintain_order=True):
    batch.process()

# No flush — keep cache
for batch in self.splittor(recordset, flush=False):
    batch.read_only_operation()
```

**Key features:** Divides collection into equal batches. Invalidates recordset after each batch (default) to free memory. `flush=False` keeps cache. `maintain_order=True` preserves order.

---

## after_commit Decorator

```python
from odoo.addons.dtg_base.models.dtg_base import after_commit

class MyModel(models.Model):
    _name = 'my.model'
    _inherit = ['dtg.base']

    @after_commit
    def send_notification_after_commit(self):
        for rec in self:
            rec.message_post(body=_("Record created"), message_type='notification')

    def action_process(self):
        self.send_notification_after_commit()
        return {'type': 'ir.actions.act_window_close'}
```

**Important:** Runs AFTER commit in new cursor. Use for notifications, external API calls, emails. Exception logged but doesn't rollback.

---

## String & Text Utilities

### strip_accents() & _no_accent_vietnamese()

```python
text = "Tiếng Việt có dấu"
no_accent = self.strip_accents(text)                     # "Tieng Viet khong dau"

vietnamese = "Xin chào, Đất Việt nước đẹp"
converted = self._no_accent_vietnamese(vietnamese)       # "Xin chao, Dat Viet nuoc dep"
```

---

## File Utilities

### zip_dir()

```python
path = '/path/to/dir'
zipped_bytes = self.zip_dir(path, incl_dir=False)
self.attachment_data = zipped_bytes
zipped_with_dir = self.zip_dir(path, incl_dir=True)
```

### zip_dirs()

```python
paths = ['/path/to/dir1', '/path/to/dir2']
zipped_bytes = self.zip_dirs(paths)
attachment = self.env['ir.attachment'].create({
    'name': 'archives.zip',
    'res_id': self.id,
    'res_model': self._name,
    'datas': zipped_bytes,
})
```

### _get_file_size()

```python
file_size = self._get_file_size('/path/to/file.pdf')
dir_size = self._get_file_size('/path/to/directory')  # recursive, excludes symlinks
```

---

## Number Utilities

### sum_digits()

```python
result = self.sum_digits(178)                      # 16 (1+7+8)
result = self.sum_digits(178, number_of_digit_return=1)  # 7
result = self.sum_digits(9999, number_of_digit_return=2) # 36
```

### find_nearest_lucky_number()

```python
lucky = self.find_nearest_lucky_number(178)             # 171 (1+7+1=9)
lucky = self.find_nearest_lucky_number(178999, rounding=2)  # 178900
lucky = self.find_nearest_lucky_number(100, round_up=True)  # 108 (1+0+8=9)
```

### calculate_weights()

```python
weights = self.calculate_weights(2, 6)                               # [0.25, 0.75]
weights = self.calculate_weights(2, 6, precision_digits=2)
assert sum(weights) == 1.0
```

### fibonacci()

```python
fib = self.fibonacci(5)                            # [0,1,1,2,3]
fib = self.fibonacci(5, deduplicate_1=True)        # [0,1,2,3]
```

---

## Other Utilities

### validate_year()

```python
year = self.validate_year('2024')  # 2024
year = self.validate_year(2024)    # 2024
# Raises ValidationError for invalid
```

### identical_images()

```python
is_same = self.identical_images(img1_field, img2_field)
# No SVG support (PIL limitation)
```

### Unit Conversion

```python
km = self.mile2km(10)      # 16.09344
miles = self.km2mile(16)   # 9.9419
```

### Week Utilities

```python
weekdays = self.get_weekdays_for_period(date_from, date_to)
# {0: date, 1: date, ...} 0=Mon, 6=Sun
```

---

## Common Patterns

### Pattern 1: Date Range by Period

```python
def _get_period_dates(self, date_from, date_to):
    return self.period_iter('monthly', date_from, date_to)

def action_report_by_period(self):
    date_from = fields.Date.to_date(self.env.context.get('date_from'))
    date_to = fields.Date.to_date(self.env.context.get('date_to'))
    period_dates = self._get_period_dates(date_from, date_to)
    for i in range(len(period_dates) - 1):
        self._process_period(period_dates[i], period_dates[i + 1])
```

### Pattern 2: Safe Timezone

```python
def action_schedule_meeting(self):
    tz = self.get_company_tz()
    utc_dt = self.convert_local_to_utc(self.meeting_date, force_local_tz_name=tz)
    self.meeting_date_utc = utc_dt
    local_dt = self.convert_utc_to_local(self.meeting_date_utc, force_local_tz_name=tz)
    self.meeting_date_display = local_dt
```

### Pattern 3: Batch Processing

```python
def action_recompute_all(self):
    records = self.search([])
    for batch in self.splittor(records, max_rec_in_batch=500):
        for rec in batch:
            rec._compute_expensive_field()
```

### Pattern 4: After-Commit Notification

```python
@after_commit
def _send_external_notification(self):
    for rec in self:
        requests.post('https://api.example.com/notify',
            json={'record_id': rec.id, 'state': rec.state})

def action_confirm(self):
    self.state = 'confirmed'
    self._send_external_notification()
```

### Pattern 5: Barcode Validation

```python
def _check_barcode_unique(self, barcode):
    if self.barcode_exists(barcode):
        raise UserError(_("Barcode %s already exists") % barcode)

def create(self, vals):
    if vals.get('barcode'):
        self._check_barcode_unique(vals['barcode'])
    return super().create(vals)
```

---

## Anti-Patterns

| Anti-Pattern | Why Bad | Fix |
|---|---|---|
| Manual date calc for periods | Error-prone, tz issues | Use `find_first_date_of_period()`, `find_last_date_of_period()` |
| Process all records at once | Memory issues | Use `splittor()` batch |
| Send notifications before commit | Sent even if txn rolls back | Use `@after_commit` |
| Manual tz conversion | DST issues | Use `convert_local_to_utc()`, `convert_utc_to_local()` |
| Check barcode with search() | Misses inactive records | Use `barcode_exists()` |

---

## Method Reference

### Date/Period

| Method | Description |
|--------|-------------|
| `find_first_date_of_period(period, date, offset)` | First date of period |
| `find_last_date_of_period(period, date, is_start, cycle)` | Last date of period |
| `period_iter(period, dt_start, dt_end, offset, cycle)` | All period dates in range |
| `get_days_between_dates(dt_from, dt_to)` | Days between |
| `get_months_between_dates(dt_from, dt_to)` | Months (float) |
| `get_number_of_years_between_dates(dt_from, dt_to)` | Years (float) |
| `get_hours_between_dates(dt_from, dt_to)` | Hours |
| `get_days_of_month_from_date(dt)` | Days in month |
| `get_day_of_year_from_date(dt)` | Day of year (1-366) |
| `get_days_in_year(dt)` | Days in year (365/366) |
| `split_date(date)` | Year, month, day |
| `next_weekday(date, weekday)` | Next week date |
| `break_timerange_for_midnight(start, end)` | Split at midnight |
| `get_ratio_between_periods(p1, d1, p2, d2, date)` | Ratio between periods |

### Timezone

| Method | Description |
|--------|-------------|
| `get_company_tz(company)` | Company timezone |
| `convert_local_to_utc(dt, tz, is_dst, naive)` | Local to UTC |
| `convert_utc_to_local(utc_dt, tz, is_dst, naive)` | UTC to local |
| `time_to_float_hour(dt)` | Datetime to float hours |
| `float_hours_to_time(hours, tz)` | Float to time |
| `hours_time_string(hours)` | Hours to "HH:MM" |
| `date_to_datetime(date)` | Date to datetime |

### Barcode

| Method | Description |
|--------|-------------|
| `barcode_exists(barcode, model, field, active)` | Check barcode |
| `get_ean13(base_number)` | Generate EAN-13 checksum |

### Batch

| Method | Description |
|--------|-------------|
| `splittor(collection, max, order, flush)` | Split into batches |

### String

| Method | Description |
|--------|-------------|
| `strip_accents(s)` | Remove all accents |
| `_no_accent_vietnamese(s)` | Vietnamese accent removal |

### File

| Method | Description |
|--------|-------------|
| `zip_dir(path, incl_dir)` | Zip directory |
| `zip_dirs(paths)` | Zip multiple dirs |
| `_get_file_size(path)` | File/dir size |

### Number

| Method | Description |
|--------|-------------|
| `sum_digits(n, digits)` | Sum digits |
| `find_nearest_lucky_number(n, round, up)` | Lucky number |
| `calculate_weights(*weights, ...)` | Percentages |
| `fibonacci(n, dedup)` | Fibonacci |

### Other

| Method | Description |
|--------|-------------|
| `validate_year(year)` | Validate (1-9999) |
| `identical_images(img1, img2)` | Compare images |
| `mile2km(miles)` | Miles to km |
| `km2mile(km)` | Km to miles |
| `get_weekdays_for_period(from, to)` | Weekdays dict |

---

## Module Info

**Module:** `dtg_base`
**Version:** 1.0.0
**Author:** AnhBT
**Location:** `addons_customs/erp/dtg_base/`
**License:** OPL-1
**Dependencies:** `base`

**Files:**
- `models/dtg_base.py` — DTGBase abstract model with all utilities
