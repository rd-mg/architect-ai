# Odoo 19 XLSX Import Compatibility

When generating `.xlsx` files for Odoo import (e.g., via Chatter attachment or "Import Spreadsheet"), strict rules apply.

## Mandatory Rules

1. **Fonts**: ONLY **Arial** supported. Any other font discarded or forced to Arial by Odoo.
2. **Fills**: Only **Solid Fills** allowed. Gradient fills ignored/broken.
3. **Borders**: **Diagonal borders** not supported.
4. **Alignments**: Only `left`, `center`, `right` (horizontal) and `top`, `center`, `bottom` (vertical). `Justify` or `Distributed` forced to defaults.
5. **Strings**: Clean all cell text from **newline characters** (`\n`, `\r`). Odoo strips them on import.

## Chart Restrictions

- Supported: `pie`, `doughnut`, `bar`, `line`.
- **Conversion**: If `pie` chart has multiple data series, Odoo converts it to `doughnut`.

## Conditional Formatting Restrictions

- **Forbidden**: `AboveAverage`, `Top10`, `DataBar`, `DuplicateValues`.
- **IconSets**: Max **3 icons**. Never generate IconSets with empty nodes (causes Odoo side-panel crash).

## Validation Command

```bash
# Run auditor before delivery
python3 scripts/formula_check.py output.xlsx --report
```
Check `Odoo 19 Compatibility Audit` section in output.
