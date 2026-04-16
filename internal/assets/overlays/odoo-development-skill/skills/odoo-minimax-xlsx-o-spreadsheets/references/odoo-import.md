# Odoo 19 XLSX Import Compatibility

When generating `.xlsx` files for Odoo import (e.g., via Chatter attachment or "Import Spreadsheet"), strict rules apply.

## Mandatory Rules

1. **Fonts**: ONLY **Arial** is supported. Any other font will be discarded or forced to Arial by Odoo.
2. **Fills**: Only **Solid Fills** are allowed. Gradient fills are ignored/broken.
3. **Borders**: **Diagonal borders** are not supported.
4. **Alignments**: Only `left`, `center`, `right` (horizontal) and `top`, `center`, `bottom` (vertical). `Justify` or `Distributed` are forced to defaults.
5. **Strings**: Clean all cell text from **newline characters** (`\n`, `\r`). Odoo strips them on import.

## Chart Restrictions

- Supported: `pie`, `doughnut`, `bar`, `line`.
- **Conversion**: If a `pie` chart has multiple data series, Odoo will convert it to a `doughnut`.

## Conditional Formatting Restrictions

- **Forbidden**: `AboveAverage`, `Top10`, `DataBar`, `DuplicateValues`.
- **IconSets**: Max **3 icons**. Never generate IconSets with empty nodes (causes Odoo side-panel crash).

## Validation Command

```bash
# Run the auditor before delivery
python3 scripts/formula_check.py output.xlsx --report
```
Check the `Odoo 19 Compatibility Audit` section in the output.
