# Feature: Odoo Charts

## Chart types observed in validated samples

| Type | Count | Description |
|------|------:|-------------|
| `odoo_line` | 7 | Odoo-native line chart backed by pivot/list data source |
| `line` | 6 | Spreadsheet-range line chart (local calculation) |
| `odoo_bar` | 5 | Odoo-native bar chart |
| `odoo_pie` | 4 | Odoo-native pie chart |
| `combo` | 2 | Combo line+bar (POS Restaurant sample) |

## Odoo chart figure structure

```json
{
  "id": "figure_5",
  "width": 480,
  "height": 280,
  "tag": "chart",
  "col": 4,
  "row": 2,
  "offset": { "x": 0, "y": 0 },
  "data": {
    "type": "odoo_line",
    "title": "Revenue Over Time",
    "legendPosition": "top",
    "stacked": false,
    "cumulative": false,
    "odooDataSets": [
      {
        "metaData": {
          "groupBy": ["date:month"],
          "measure": "price_subtotal",
          "order": "ASC",
          "resModel": "sale.report"
        },
        "domain": [],
        "context": {},
        "label": "Revenue"
      }
    ],
    "searchParams": {
      "groupBy": [],
      "orderBy": [],
      "domain": []
    }
  }
}
```

## Odoo vs spreadsheet-range charts

| Aspect | Odoo chart (`odoo_*`) | Spreadsheet-range chart (`line`, `bar`, `pie`) |
|--------|-----------------------|-----------------------------------------------|
| Data source | Odoo model via `odooDataSets` | Cell range in Data sheet |
| Global filter | Automatically wired via `searchParams` | Must be manually wired via formula |
| Permissions | Enforced by Odoo access rights | Inherits from formula evaluation context |
| When to use | Always preferred for Odoo model data | Only for calculated/non-model metrics |

## Key rules

1. Every chart figure MUST have `data.type` set (validator: CRITICAL if missing).
2. Odoo charts wire to global filters via `searchParams` fields — do not duplicate filter conditions in `domain`.
3. For `cumulative` line charts, set `"cumulative": true` and use `order: "ASC"`.
4. `combo` type requires `dataSetsHaveTitle: true` and per-dataset `type` overrides.
5. Use `stacked: true` for revenue-breakdown bar charts.
