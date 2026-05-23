# DTG Base Skill

Complete reference for DTG Base module utilities and helpers in Odoo 18.

## Overview

Custom abstract model providing common utility methods for Odoo dev. Comprehensive docs for all DTGBase utilities.

## Included

- **Date & Period** — First/last dates, iterate periods
- **Timezone** — Local ↔ UTC
- **Barcode** — Validate and generate EAN13
- **Batch** — Split large recordsets
- **after_commit** — Execute after txn commit
- **Vietnamese Text** — Strip accents for search
- **File** — Zip dirs, get file sizes
- **Number** — Round to decimal places

## Files

| File | Description |
|------|-------------|
| `SKILL.md` | Master index + quick reference |
| `CLAUDE.md` | AI agent guidance |
| `odoo-18-dtg-base-guide.md` | Complete utilities reference |

## Quick Start

```python
class MyModel(models.Model):
    _name = 'my.model'
    _inherit = ['dtg_base.dtg_base']

    def my_method(self):
        first_date = self.find_first_date_of_period('2024-01-15', 'month')
        utc_date = self.convert_local_to_utc('2024-01-15 10:00:00')
```

## Links

- [Full Docs](./odoo-18-dtg-base-guide.md)
- [SKILL.md](./SKILL.md) — Quick reference
