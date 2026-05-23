---
name: dtg-base
description: Complete reference for DTG Base module utilities and helpers. DTGBase is an abstract model providing common utility methods for date/time handling, barcode generation, timezone conversion, file operations, and more.
globs: "**/addons_customs/erp/**/*.py"
license: MIT
author: UncleCat
version: 1.0.0
---

# DTG Base Skill

Complete reference for DTG Base module utilities and helpers in Odoo 18.

## What is DTG Base?

Custom abstract model (`dtg_base.DTGBase`) providing common utility methods. Inherit for access.

## Quick Reference

| Utility | Description |
|---------|-------------|
| [Date & Period](#date--period-utilities) | First/last date of period, iteration |
| [Timezone](#timezone-conversion) | Local ↔ UTC |
| [Barcode](#barcode-utilities) | Check exists, generate EAN13 |
| [Batch Processing](#batch-processing) | Split large recordsets |
| [after_commit](#after_commit-decorator) | Execute after txn commit |
| [Vietnamese Text](#string--text-utilities) | Strip accents |
| [File Utilities](#file-utilities) | Zip dirs, file size |
| [Number Utilities](#number-utilities) | Round to decimal places |

---

## Main Guide

**File:** `odoo-18-dtg-base-guide.md`

### When to use

DTG Odoo codebase, date/period calcs, timezone conversions, barcode validation, batch processing, Vietnamese text, file zipping.

---

## DTGBase Abstract Model

### Inherit from DTGBase

**Location:** `addons_customs/erp/dtg_base/models/dtg_base.py`

```python
from odoo import models

class MyModel(models.Model):
    _name = 'my.model'
    _inherit = ['dtg_base.dtg_base']

    def my_method(self):
        first_date = self.find_first_date_of_period('2024-01-15', 'month')
        utc_date = self.convert_local_to_utc('2024-01-15 10:00:00')
```

---

## File Structure

```
agent-skills/skills/dtg-base/
├── SKILL.md                       # Master index
├── odoo-18-dtg-base-guide.md      # Complete ref
└── README.md                      # Overview
```

---

## Utilities Overview

### Date & Period
- `find_first_date_of_period(date, period_type)` — First date of period
- `find_last_date_of_period(date, period_type)` — Last date of period
- `period_iter(start_date, end_date, period_type)` — Iterate periods

### Timezone
- `convert_local_to_utc(local_dt, tz=None)` — Local to UTC
- `convert_utc_to_local(utc_dt, tz=None)` — UTC to local

### Barcode
- `barcode_exists(barcode, exclude_id=0)` — Check barcode exists
- `get_ean13(barcode)` — Generate/check EAN13

### Batch
- `splittor(limit=None)` — Split recordset into batches

### String & Text
- `strip_accents(text)` — Remove Vietnamese accents
- `_no_accent_vietnamese(text)` — Convert Vietnamese text

### File
- `zip_dir(source_dir, output_file)` — Zip directory
- `zip_dirs(dirs, output_file)` — Zip multiple dirs
- `_get_file_size(file_path)` — Human-readable file size

### Number
- `round_decimal(value, decimal_places)` — Round to decimal places

---

**Detailed docs:** [odoo-18-dtg-base-guide.md](./odoo-18-dtg-base-guide.md)
