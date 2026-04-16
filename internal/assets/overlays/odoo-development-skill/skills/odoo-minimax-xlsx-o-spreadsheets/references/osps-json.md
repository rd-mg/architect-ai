# Odoo o-spreadsheet JSON (Native Format)

Use this format for Odoo Documents, Dashboards, and any native Odoo 19 spreadsheet integration.

## Key Principles

1. **JSON Structure**: The final output must be a single `.json` (or `.osps`) file following the o-spreadsheet schema.
2. **Formula Sign**: All formulas in the `content` field **MUST** start with `=`.
   - Correct: `"content": "=SUM(A1:A10)"`
   - Incorrect: `"content": "SUM(A1:A10)"`
3. **Styles & Formats**: Use indices to refer to the global `styles` and `formats` dictionaries.
4. **Pivots**: Define pivots in the `pivots` object to link spreadsheet cells to Odoo model data using `=PIVOT("pivot_id", "measure", ...)` functions.

## Script Usage

```bash
# Generate a base JSON structure with Pydantic validation
python3 scripts/json_builder.py > report.json
```

## JSON Schema Snippet

```json
{
  "sheets": [
    {
      "id": "sheet1",
      "cells": {
        "A1": { "content": "Total Revenue", "style": 1 },
        "B1": { "content": "=PIVOT.VALUE(\"1\", \"amount_total\")", "format": 2 }
      }
    }
  ],
  "styles": {
    "1": { "bold": true, "fontSize": 12 }
  },
  "formats": {
    "2": "#,##0.00"
  }
}
```
