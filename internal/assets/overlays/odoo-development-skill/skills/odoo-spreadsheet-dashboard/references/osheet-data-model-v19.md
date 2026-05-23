# Reference: Odoo 19 .osheet Data Model

## Top-level keys (all 13 samples)

```json
{
  "version": "19.1.2",
  "revisionId": "<uuid>",
  "uniqueFigureIds": ["figure_1", "figure_2"],
  "isNotSquishable": true,
  "sheets": [...],
  "styles": [...],
  "formats": [...],
  "borders": [...],
  "settings": { "locale": { "name": "English (US)", "code": "en_US" } },
  "customTableStyles": {},
  "pivots": { "<uuid>": { ... } },
  "pivotNextId": 11,
  "lists": { "1": { ... } },
  "listNextId": 3,
  "odooLinkReferences": { "<id>": { ... } },
  "globalFilters": [...]
}
```

## Sheet object

```json
{
  "id": "sheet_1",
  "name": "Dashboard",
  "figures": [...],
  "cells": { "A1": { "content": "=...", "style": 0, "format": 1 } },
  "merges": ["A1:C2"],
  "cols": { "0": { "size": 240, "isHidden": false } },
  "rows": { "0": { "size": 30, "isHidden": false } },
  "conditionalFormats": [],
  "filterTables": [],
  "isVisible": true
}
```

## Figure object

```json
{
  "id": "figure_1",
  "tag": "chart",
  "width": 240,
  "height": 120,
  "col": 0,
  "row": 0,
  "offset": { "x": 0, "y": 0 },
  "data": { "type": "scorecard", ... }
}
```

## Pivot object

```json
{
  "<uuid>": {
    "id": "<uuid>",
    "formulaId": "1",
    "model": "sale.report",
    "domain": [],
    "context": {},
    "measures": [{ "id": "price_subtotal", "fieldName": "price_subtotal", "aggregator": "sum" }],
    "columns": [],
    "rows": [{ "fieldName": "date", "granularity": "month", "order": "asc" }],
    "sortedColumn": null,
    "name": "Sales"
  }
}
```

## List object

```json
{
  "1": {
    "id": "1",
    "model": "sale.order",
    "domain": [],
    "context": {},
    "orderBy": [{ "name": "date_order", "asc": false }],
    "columns": [{ "name": "name", "type": "char" }],
    "name": "Recent Orders"
  }
}
```

## Global filter object

```json
{
  "id": "filter_1",
  "type": "relation",
  "label": "Company",
  "modelName": "res.company",
  "fields": {
    "<pivot_uuid>": { "field": "company_id", "type": "many2one" }
  },
  "defaultValue": []
}
```

## odooLinkReference object

```json
{
  "<id>": {
    "url": "odoo://view/{\"model\":\"sale.order\",\"viewType\":\"list\",\"action\":{\"modelName\":\"sale.order\",\"domain\":[],\"viewType\":\"list\"}}",
    "label": "View Orders"
  }
}
```

## Constraints

- `version` MUST be `"19.1.2"` for Odoo 19 import.
- `pivotNextId` MUST be >= count of pivots + 1.
- `listNextId` MUST be >= count of lists + 1.
- All figure IDs MUST appear in `uniqueFigureIds`.
- Style/format/border references are 0-indexed arrays.
