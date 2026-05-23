# Odoo o-spreadsheet JSON (Native Format)

Use this format for Odoo Documents, Dashboards, and any native Odoo 19 spreadsheet integration.

## Key Principles

1. **JSON Structure**: Final output must be single `.json` (or `.osps`) file following o-spreadsheet schema.
2. **Formula Sign**: All formulas in `content` field **MUST** start with `=`.
   - Correct: `"content": "=SUM(A1:A10)"`
   - Incorrect: `"content": "SUM(A1:A10)"`
3. **Styles & Formats**: Use indices to refer to global `styles` and `formats` dictionaries.
4. **Pivots & Lists (Odoo 19 / v20)**: 
   - **Pivot Type**: Must be uppercase `"ODOO"`.
   - **Formulas**: Legacy `=ODOO.PIVOT()` deprecated. Use `=PIVOT.VALUE("1", ...)`. Lists remain `=ODOO.LIST()`.
   - **Explicit IDs (CRITICAL)**: Both `pivots` and `lists` objects **must** contain `"id"` and `"formulaId"` explicitly inside definition objects. Missing these causes data to silently fail to bind (blank cells).

## Script Usage

```bash
# Generate base JSON structure with Pydantic validation
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
  },
  "pivots": {
    "1": {
      "id": "1",
      "formulaId": "1",
      "type": "ODOO",
      "model": "sale.report",
      "measures": [{"id": "amount_total", "fieldName": "amount_total"}]
    }
  }
}
```
