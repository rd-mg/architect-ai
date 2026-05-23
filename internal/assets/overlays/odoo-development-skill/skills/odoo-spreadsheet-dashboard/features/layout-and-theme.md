# Feature: Layout and Theme

## Default two-sheet layout

All 13 validated samples use `Dashboard` + `Data` two-sheet pattern (one uses three sheets).

```
Sheet 1: "Dashboard"  — figures only (scorecards, charts)
Sheet 2: "Data"       — backing formulas (PIVOT.VALUE, ODOO.LIST, ODOO.BALANCE, calculations)
```

Rationale:
- Separates presentation from data plumbing.
- Allows hiding Data sheet in Odoo's dashboard view.
- Makes formula auditing tractable.
- Mirrors what Odoo's native dashboard builder generates.

## Sheet structure in .osheet JSON

```json
{
  "sheets": [
    {
      "id": "sheet_dashboard",
      "name": "Dashboard",
      "figures": [ /* all figure objects go here */ ],
      "cells": {},
      "merges": [],
      "cols": {},
      "rows": {},
      "conditionalFormats": [],
      "filterTables": [],
      "isVisible": true
    },
    {
      "id": "sheet_data",
      "name": "Data",
      "figures": [],
      "cells": {
        "A1": { "content": "=PIVOT.VALUE(1,\"revenue\")", "style": 0 }
      },
      "merges": [],
      "cols": {},
      "rows": {},
      "conditionalFormats": [],
      "filterTables": [],
      "isVisible": false
    }
  ]
}
```

## Global style dictionaries

Styles, formats, and borders defined globally and referenced by index.

```json
{
  "styles": [
    {},
    { "bold": true, "fontSize": 14, "textColor": "#1F4E79" }
  ],
  "formats": [
    "",
    "#,##0",
    "#,##0.00",
    "[$€-fr-FR]#,##0.00"
  ],
  "borders": [
    null,
    { "top": { "style": "thin", "color": "#000000" } }
  ]
}
```

Style reference by index: `"style": 1` = bold, 14pt, dark-blue.

## Colour conventions (from samples)

| Role | Suggested hex |
|------|--------------|
| Dashboard background | `#F0F4F8` or `#FFFFFF` |
| Scorecard background | `#FFFFFF` |
| Positive delta | `#00A04A` (green) |
| Negative delta | `#E30000` (red) |
| Accent / header | `#1F4E79` (dark blue) or brand colour |
| Chart grid lines | `#D9E1EC` |

## Figure positioning

- Columns and rows are 0-indexed in figure `col` and `row`.
- Width/height in pixels; typical scorecard: 240×120, chart: 480×280.
- `offset.x` and `offset.y` adjust sub-cell placement (usually 0).
- Do not place figures on `Data` sheet.

## `isNotSquishable`

12 of 13 samples have `"isNotSquishable": true` at top level. Set this flag on dashboards
to prevent Odoo from auto-collapsing columns when dashboard viewed in narrow panel.
